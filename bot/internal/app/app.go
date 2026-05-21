package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"schulbot/internal/ai"
	"schulbot/internal/commands"
	"schulbot/internal/config"
	imail "schulbot/internal/mail"
	"schulbot/internal/model"
	"schulbot/internal/parser"
	"schulbot/internal/store"
	"schulbot/internal/users"
)

// DispatcherFactory creates a handler dispatcher for a given AI provider and store.
type DispatcherFactory func(aiProvider ai.Provider, db *store.Store) *commands.Dispatcher

// UserLookup resolves a sender email to their SMTP + AI config.
type UserLookup interface {
	FindByEmail(ctx context.Context, email string) (*users.UserConfig, error)
}

type App struct {
	cfg        *config.Config
	imap       *imail.IMAPClient
	smtp       *imail.SMTPClient  // nil in multi-user mode (resolved per message)
	dispatcher *commands.Dispatcher // nil in multi-user mode (built per message)
	store      *store.Store
	log        *slog.Logger
	keyScope   string
	httpClient *http.Client
	// Multi-user fields (nil in single-user mode)
	userLookup        UserLookup
	dispatcherFactory DispatcherFactory
}

func New(
	cfg *config.Config,
	imap *imail.IMAPClient,
	smtp *imail.SMTPClient,
	dispatcher *commands.Dispatcher,
	store *store.Store,
	log *slog.Logger,
) *App {
	return &App{
		cfg:        cfg,
		imap:       imap,
		smtp:       smtp,
		dispatcher: dispatcher,
		store:      store,
		log:        log,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// NewMultiUser creates an App for the shared-inbox multi-user mode.
// SMTP and AI are resolved per message via userLookup + dispatcherFactory.
func NewMultiUser(
	cfg *config.Config,
	imap *imail.IMAPClient,
	store *store.Store,
	lookup UserLookup,
	factory DispatcherFactory,
	log *slog.Logger,
) *App {
	return &App{
		cfg:               cfg,
		imap:              imap,
		store:             store,
		log:               log,
		httpClient:        &http.Client{Timeout: 10 * time.Second},
		userLookup:        lookup,
		dispatcherFactory: factory,
	}
}

// NewScoped creates an App whose BoltDB keys are prefixed with the user email,
// preventing collisions between users in multi-user mode.
func NewScoped(
	cfg *config.Config,
	imap *imail.IMAPClient,
	smtp *imail.SMTPClient,
	dispatcher *commands.Dispatcher,
	store *store.Store,
	log *slog.Logger,
	userEmail string,
) *App {
	return &App{
		cfg:        cfg,
		imap:       imap,
		smtp:       smtp,
		dispatcher: dispatcher,
		store:      store,
		log:        log,
		keyScope:   "user:" + userEmail + ":",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Run blocks, polling for new messages on every tick, until ctx is cancelled.
func (a *App) Run(ctx context.Context) error {
	a.log.Info("schulbot started",
		"poll_interval", a.cfg.PollInterval,
		"imap_host", a.cfg.IMAP.Host,
		"mailbox", a.cfg.IMAP.Mailbox,
	)

	ticker := time.NewTicker(a.cfg.PollInterval)
	defer ticker.Stop()

	a.poll(ctx) // immediate first run

	for {
		select {
		case <-ctx.Done():
			a.log.Info("schulbot stopping")
			return nil
		case <-ticker.C:
			a.poll(ctx)
		}
	}
}

func (a *App) poll(ctx context.Context) {
	messages, err := a.imap.FetchUnseen()
	if err != nil {
		a.log.Error("imap fetch failed", "err", err)
		return
	}
	if len(messages) == 0 {
		return
	}
	a.log.Info("fetched unseen messages", "count", len(messages))

	var seenOK []uint32
	for _, msg := range messages {
		if a.processMessage(ctx, msg) {
			seenOK = append(seenOK, msg.SeqNum)
		}
	}

	// Mark processed messages as SEEN in IMAP so we don't fetch them again.
	// Non-processed messages (parse errors, skipped) are also marked to
	// avoid re-fetching on every tick.
	if len(seenOK) > 0 {
		if err := a.imap.MarkSeen(seenOK); err != nil {
			a.log.Warn("could not mark messages as seen", "err", err)
		}
	}
}

// processMessage handles one email. Returns true if the message was fully
// handled and should be marked seen in IMAP.
func (a *App) processMessage(ctx context.Context, msg imail.Message) bool {
	log := a.log.With("msg_id", msg.ID, "from", msg.From, "subject", msg.Subject)

	// ── Loop / self-reply guard ───────────────────────────────────────────────
	// Check for our own X-SchulBot: reply header rather than the From address,
	// because the bot and the user may share the same mailbox and address.
	if msg.IsSchulBotReply {
		log.Info("skipping own bot reply (loop guard)")
		return true
	}
	if isAutoReply(msg.Subject) {
		log.Info("skipping auto-reply")
		return true
	}

	// ── Duplicate guard ───────────────────────────────────────────────────────
	storeKey := a.keyScope + msg.ID
	processed, err := a.store.IsProcessed(storeKey)
	if err != nil {
		log.Error("store check failed", "err", err)
		return false
	}
	if processed {
		log.Info("already processed, skipping")
		return true
	}

	// ── Command parsing ───────────────────────────────────────────────────────
	cmd, ok := parser.Parse(msg.Subject, msg.Body)
	if !ok {
		log.Info("no command tag found, skipping")
		_ = a.store.MarkProcessed(storeKey) // don't fetch again
		return true
	}

	log = log.With("tag", cmd.Tag, "payload_len", len(cmd.Payload))
	log.Info("command detected")

	// ── Multi-user: resolve sender's SMTP + AI from web dashboard ────────────
	smtpClient := a.smtp
	dispatcher := a.dispatcher
	if a.userLookup != nil {
		uc, err := a.userLookup.FindByEmail(ctx, msg.From)
		if err != nil {
			log.Error("user lookup failed", "err", err)
			return false
		}
		if uc == nil {
			log.Info("sender not registered in dashboard, skipping")
			_ = a.store.MarkProcessed(storeKey)
			return true
		}
		aiProvider, err := ai.NewProvider(uc.AIProvider, uc.AIAPIURL, uc.AIAPIKey, uc.AIModel)
		if err != nil {
			log.Error("AI provider init failed", "err", err)
			return false
		}
		smtpClient = imail.NewSMTPClient(
			uc.SMTPHost, uc.SMTPPort,
			uc.SMTPUsername, uc.SMTPPassword,
			uc.SMTPFromName, uc.SMTPFromAddress,
			uc.SMTPSecurity,
		)
		dispatcher = a.dispatcherFactory(aiProvider, a.store)
	}

	// ── Dispatch ──────────────────────────────────────────────────────────────
	replyAddr := msg.From
	if msg.ReplyTo != "" {
		replyAddr = msg.ReplyTo
	}

	req := model.CommandRequest{
		MessageID:  msg.ID,
		From:       msg.From,
		Subject:    msg.Subject,
		Tag:        cmd.Tag,
		Payload:    cmd.Payload,
		RawText:    msg.Body,
		ReceivedAt: msg.Date,
	}

	resp, dispErr := dispatcher.Dispatch(ctx, req)
	if dispErr != nil {
		log.Error("dispatch failed", "err", dispErr)
		a.sendErrorReply(smtpClient, replyAddr, msg.Subject, dispErr)
		_ = a.store.MarkProcessed(storeKey)
		a.writeHistory(req.From, cmd.Tag, cmd.Payload, dispErr.Error(), "error")
		return true
	}

	// ── Send reply ────────────────────────────────────────────────────────────
	var sendErr error
	if len(resp.Attachments) > 0 {
		sendErr = smtpClient.SendWithAttachments(replyAddr, resp.ReplySubject, resp.ReplyBody, resp.Attachments)
	} else {
		sendErr = smtpClient.Send(replyAddr, resp.ReplySubject, resp.ReplyBody)
	}
	if err := sendErr; err != nil {
		if imail.IsPermanentSMTP(err) {
			log.Error("permanent smtp error – not retrying", "err", err)
			_ = a.store.MarkProcessed(msg.ID)
			return true
		}
		log.Error("smtp send failed – will retry on next poll", "err", err)
		return false
	}

	if err := a.store.MarkProcessed(storeKey); err != nil {
		log.Error("store mark failed", "err", err)
	}

	a.writeHistory(req.From, cmd.Tag, cmd.Payload, resp.ReplyBody, "ok")
	log.Info("replied successfully")
	return true
}

func (a *App) sendErrorReply(smtp *imail.SMTPClient, to, subject string, origErr error) {
	if smtp == nil {
		return
	}
	body := fmt.Sprintf(
		"Leider ist ein Fehler aufgetreten:\n\n%v\n\nBitte versuche es später erneut oder wende dich an den Administrator.",
		origErr,
	)
	if err := smtp.Send(to, "Re: "+subject+" [Fehler]", body); err != nil {
		a.log.Error("could not send error reply", "to", to, "err", err)
	}
}

// writeHistory posts a processed command to the Laravel web app for display in the dashboard.
// It is best-effort: failures are logged but never block message processing.
func (a *App) writeHistory(senderEmail, tag, payload, response, status string) {
	if a.cfg.LaravelAPIURL == "" || a.cfg.BotAPISecret == "" {
		return
	}

	body, _ := json.Marshal(map[string]string{
		"sender_email": senderEmail,
		"tag":          tag,
		"payload":      payload,
		"response":     response,
		"status":       status,
	})

	req, err := http.NewRequest(http.MethodPost, a.cfg.LaravelAPIURL+"/api/internal/history", bytes.NewReader(body))
	if err != nil {
		a.log.Warn("history: failed to build request", "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bot-Secret", a.cfg.BotAPISecret)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		a.log.Warn("history: post failed", "err", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		a.log.Warn("history: unexpected status", "code", resp.StatusCode)
	}
}

// isAutoReply detects common auto-reply patterns to avoid reply loops.
func isAutoReply(subject string) bool {
	lower := strings.ToLower(subject)
	autoPrefixes := []string{
		"auto:", "automatic reply", "out of office", "abwesenheit",
		"abwesenheitsbenachrichtigung", "autoreply", "auto-reply",
		"vacation", "away:", "out of office:",
	}
	for _, prefix := range autoPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}
