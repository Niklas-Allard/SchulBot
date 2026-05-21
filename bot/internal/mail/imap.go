package mail

import (
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	gomessage "github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset" // register charset decoders
)

// Message is a parsed inbound email.
type Message struct {
	// ID is derived from the Message-ID header (or a fallback fingerprint).
	// Used as deduplication key in the store.
	ID              string
	SeqNum          uint32
	From            string
	ReplyTo         string // Reply-To header if set, otherwise empty
	Subject         string
	Body            string
	Date            time.Time
	IsSchulBotReply bool // true when outgoing X-SchulBot: reply header is present
}

// IMAPClient fetches unseen messages from an IMAP mailbox.
type IMAPClient struct {
	host     string
	port     int
	username string
	password string
	mailbox  string
	security string
}

func NewIMAPClient(host string, port int, username, password, mailbox, security string) *IMAPClient {
	return &IMAPClient{
		host:     host,
		port:     port,
		username: username,
		password: password,
		mailbox:  mailbox,
		security: security,
	}
}

// FetchUnseen connects, fetches all unseen messages, and disconnects.
// Messages are NOT marked seen here; the caller marks them via MarkSeen
// after successful processing.
func (c *IMAPClient) FetchUnseen() ([]Message, error) {
	cl, err := c.connect()
	if err != nil {
		return nil, fmt.Errorf("imap: connect: %w", err)
	}
	defer cl.Logout()

	if _, err := cl.Select(c.mailbox, false); err != nil {
		return nil, fmt.Errorf("imap: select %s: %w", c.mailbox, err)
	}

	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}

	seqNums, err := cl.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("imap: search: %w", err)
	}
	if len(seqNums) == 0 {
		return nil, nil
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums...)

	section := &imap.BodySectionName{Peek: true} // BODY.PEEK[] – does not auto-mark as seen
	items := []imap.FetchItem{
		imap.FetchEnvelope,
		imap.FetchUid,
		section.FetchItem(),
	}

	ch := make(chan *imap.Message, len(seqNums))
	fetchErr := make(chan error, 1)
	go func() {
		fetchErr <- cl.Fetch(seqSet, items, ch)
	}()

	var result []Message
	for raw := range ch {
		msg, err := parseIMAPMessage(raw, section)
		if err != nil {
			slog.Warn("imap: could not parse message", "seq", raw.SeqNum, "err", err)
			continue
		}
		result = append(result, msg)
	}
	if err := <-fetchErr; err != nil {
		return nil, fmt.Errorf("imap: fetch: %w", err)
	}
	return result, nil
}

// MarkSeen sets the \Seen flag on the given sequence numbers.
func (c *IMAPClient) MarkSeen(seqNums []uint32) error {
	if len(seqNums) == 0 {
		return nil
	}
	cl, err := c.connect()
	if err != nil {
		return fmt.Errorf("imap: connect for mark-seen: %w", err)
	}
	defer cl.Logout()

	if _, err := cl.Select(c.mailbox, false); err != nil {
		return fmt.Errorf("imap: select for mark-seen: %w", err)
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums...)

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag}
	return cl.Store(seqSet, item, flags, nil)
}

// ── Connection ────────────────────────────────────────────────────────────────

func (c *IMAPClient) connect() (*client.Client, error) {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	tlsCfg := &tls.Config{ServerName: c.host, MinVersion: tls.VersionTLS12}

	var (
		cl  *client.Client
		err error
	)
	switch strings.ToUpper(c.security) {
	case "SSL", "TLS", "SSL/TLS":
		cl, err = client.DialTLS(addr, tlsCfg)
	case "STARTTLS":
		cl, err = client.Dial(addr)
		if err != nil {
			return nil, err
		}
		if err = cl.StartTLS(tlsCfg); err != nil {
			cl.Logout()
			return nil, err
		}
	default:
		cl, err = client.Dial(addr)
	}
	if err != nil {
		return nil, err
	}

	if err := cl.Login(c.username, c.password); err != nil {
		cl.Logout()
		return nil, fmt.Errorf("login: %w", err)
	}
	return cl, nil
}

// ── Message parsing ───────────────────────────────────────────────────────────

func parseIMAPMessage(raw *imap.Message, section *imap.BodySectionName) (Message, error) {
	env := raw.Envelope
	if env == nil {
		return Message{}, fmt.Errorf("nil envelope")
	}

	msgID := env.MessageId
	if msgID == "" {
		// Fallback fingerprint when Message-ID is absent
		from := ""
		if len(env.From) > 0 {
			from = env.From[0].Address()
		}
		msgID = fmt.Sprintf("%s|%s|%d", from, env.Subject, env.Date.Unix())
	}

	from := ""
	if len(env.From) > 0 {
		from = env.From[0].Address()
	}

	replyTo := ""
	if len(env.ReplyTo) > 0 {
		replyTo = env.ReplyTo[0].Address()
	}

	var body string
	var isSchulBotReply bool
	if literal := raw.GetBody(section); literal != nil {
		body, isSchulBotReply = parseRawMessage(literal)
	}

	return Message{
		ID:              msgID,
		SeqNum:          raw.SeqNum,
		From:            from,
		ReplyTo:         replyTo,
		Subject:         env.Subject,
		Body:            body,
		Date:            env.Date,
		IsSchulBotReply: isSchulBotReply,
	}, nil
}

// parseRawMessage parses the full RFC 2822 message and returns the plain-text
// body and whether the X-SchulBot reply header is present.
func parseRawMessage(r io.Reader) (body string, isSchulBotReply bool) {
	entity, err := gomessage.Read(r)
	if err != nil {
		return "", false
	}
	if entity.Header.Get("X-SchulBot") == "reply" {
		return "", true
	}
	return walkForPlainText(entity), false
}

func walkForPlainText(e *gomessage.Entity) string {
	if mr := e.MultipartReader(); mr != nil {
		for {
			part, err := mr.NextPart()
			if err != nil {
				break
			}
			if text := walkForPlainText(part); text != "" {
				return text
			}
		}
		return ""
	}

	t, _, _ := e.Header.ContentType()
	if !strings.EqualFold(t, "text/plain") {
		return ""
	}

	data, err := io.ReadAll(e.Body)
	if err != nil {
		return ""
	}
	return string(data)
}
