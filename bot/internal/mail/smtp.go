package mail

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"schulbot/internal/model"
)

// PermanentSMTPError wraps a 5xx SMTP error. The message should NOT be retried.
type PermanentSMTPError struct{ Err error }

func (e *PermanentSMTPError) Error() string { return e.Err.Error() }
func (e *PermanentSMTPError) Unwrap() error { return e.Err }

// IsPermanentSMTP reports whether err is a permanent (5xx) SMTP failure.
func IsPermanentSMTP(err error) bool {
	var p *PermanentSMTPError
	return errors.As(err, &p)
}

// SMTPClient sends reply emails via SMTP with SSL or STARTTLS.
type SMTPClient struct {
	host        string
	port        int
	username    string
	password    string
	fromName    string
	fromAddress string
	security    string
}

func NewSMTPClient(host string, port int, username, password, fromName, fromAddress, security string) *SMTPClient {
	return &SMTPClient{
		host:        host,
		port:        port,
		username:    username,
		password:    password,
		fromName:    fromName,
		fromAddress: fromAddress,
		security:    security,
	}
}

// Send delivers a plain-text reply email.
func (c *SMTPClient) Send(to, subject, body string) error {
	return c.deliver(to, subject, body, nil)
}

// SendWithAttachments delivers a multipart email with file attachments.
func (c *SMTPClient) SendWithAttachments(to, subject, body string, attachments []model.Attachment) error {
	return c.deliver(to, subject, body, attachments)
}

func (c *SMTPClient) deliver(to, subject, body string, attachments []model.Attachment) error {
	sc, err := c.dial()
	if err != nil {
		return fmt.Errorf("smtp: dial: %w", err)
	}
	defer sc.Close()

	if err := sc.Auth(smtp.PlainAuth("", c.username, c.password, c.host)); err != nil {
		return fmt.Errorf("smtp: auth: %w", err)
	}
	if err := sc.Mail(c.fromAddress); err != nil {
		return fmt.Errorf("smtp: MAIL FROM: %w", err)
	}
	if err := sc.Rcpt(to); err != nil {
		wrapped := fmt.Errorf("smtp: RCPT TO: %w", err)
		var tpErr *textproto.Error
		if errors.As(err, &tpErr) && tpErr.Code >= 500 {
			return &PermanentSMTPError{wrapped}
		}
		return wrapped
	}

	w, err := sc.Data()
	if err != nil {
		return fmt.Errorf("smtp: DATA: %w", err)
	}

	var msg string
	if len(attachments) == 0 {
		msg = c.buildPlain(to, subject, body)
	} else {
		msg = c.buildMultipart(to, subject, body, attachments)
	}

	if _, err := fmt.Fprint(w, msg); err != nil {
		return fmt.Errorf("smtp: write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp: close DATA: %w", err)
	}
	return sc.Quit()
}

// ── Connection ────────────────────────────────────────────────────────────────

func (c *SMTPClient) dial() (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	tlsCfg := &tls.Config{ServerName: c.host, MinVersion: tls.VersionTLS12}

	var (
		conn net.Conn
		err  error
	)
	switch strings.ToUpper(c.security) {
	case "SSL", "TLS", "SSL/TLS":
		conn, err = tls.Dial("tcp", addr, tlsCfg)
	default:
		conn, err = net.DialTimeout("tcp", addr, 15*time.Second)
	}
	if err != nil {
		return nil, err
	}

	sc, err := smtp.NewClient(conn, c.host)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if strings.ToUpper(c.security) == "STARTTLS" {
		if err := sc.StartTLS(tlsCfg); err != nil {
			sc.Close()
			return nil, fmt.Errorf("starttls: %w", err)
		}
	}
	return sc, nil
}

// ── Message building ──────────────────────────────────────────────────────────

func (c *SMTPClient) commonHeaders(to, subject string) string {
	from := mime.QEncoding.Encode("utf-8", c.fromName) + " <" + c.fromAddress + ">"
	subj := mime.QEncoding.Encode("utf-8", subject)
	var sb strings.Builder
	sb.WriteString("From: " + from + "\r\n")
	sb.WriteString("To: " + to + "\r\n")
	sb.WriteString("Subject: " + subj + "\r\n")
	sb.WriteString("Date: " + time.Now().UTC().Format(time.RFC1123Z) + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("X-Mailer: SchulBot\r\n")
	sb.WriteString("X-SchulBot: reply\r\n")
	return sb.String()
}

func (c *SMTPClient) buildPlain(to, subject, body string) string {
	var qp strings.Builder
	qpw := quotedprintable.NewWriter(&qp)
	_, _ = qpw.Write([]byte(body))
	_ = qpw.Close()

	var sb strings.Builder
	sb.WriteString(c.commonHeaders(to, subject))
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	sb.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(qp.String())
	return sb.String()
}

func (c *SMTPClient) buildMultipart(to, subject, body string, attachments []model.Attachment) string {
	var buf strings.Builder
	mw := multipart.NewWriter(&buf)

	// Headers
	buf2 := c.commonHeaders(to, subject)
	buf2 += "Content-Type: multipart/mixed; boundary=\"" + mw.Boundary() + "\"\r\n\r\n"

	// Text part
	th := textproto.MIMEHeader{}
	th.Set("Content-Type", "text/plain; charset=UTF-8")
	th.Set("Content-Transfer-Encoding", "quoted-printable")
	tp, _ := mw.CreatePart(th)
	qpw := quotedprintable.NewWriter(tp)
	_, _ = qpw.Write([]byte(body))
	_ = qpw.Close()

	// Attachment parts
	for _, a := range attachments {
		ah := textproto.MIMEHeader{}
		ah.Set("Content-Type", a.ContentType)
		ah.Set("Content-Transfer-Encoding", "base64")
		ah.Set("Content-Disposition", `attachment; filename="`+a.Filename+`"`)
		ap, _ := mw.CreatePart(ah)

		enc := base64.StdEncoding.EncodeToString(a.Data)
		for i := 0; i < len(enc); i += 76 {
			end := i + 76
			if end > len(enc) {
				end = len(enc)
			}
			fmt.Fprintf(ap, "%s\r\n", enc[i:end])
		}
	}
	_ = mw.Close()

	return buf2 + buf.String()
}
