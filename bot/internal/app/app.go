package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"schulbot/internal/commands"
	"schulbot/internal/config"
	imail "schulbot/internal/mail"
	"schulbot/internal/model"
	"schulbot/internal/parser"
	"schulbot/internal/store"
)

type App struct {
	cfg        *config.Config
	imap       *imail.IMAPClient
	smtp       *imail.SMTPClient
	dispatcher *commands.Dispatcher
	store      *store.Store
	log        *slog.Logger
}

func New(
	cfg *config.Config,
	imap *imail.IMAPClient,
	smtp *imail.SMTPClient,
	dispatcher *commands.Dispatcher,
	store *store.Store,
	log *slog.Logger,
) *App {
	return &App{cfg: cfg, imap: imap, smtp: smtp, dispatcher: dispatcher, store: store, log: log}
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
	processed, err := a.store.IsProcessed(msg.ID)
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
		_ = a.store.MarkProcessed(msg.ID) // don't fetch again
		return true
	}

	log = log.With("tag", cmd.Tag, "payload_len", len(cmd.Payload))
	log.Info("command detected")

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

	resp, dispErr := a.dispatcher.Dispatch(ctx, req)
	if dispErr != nil {
		log.Error("dispatch failed", "err", dispErr)
		a.sendErrorReply(replyAddr, msg.Subject, dispErr)
		_ = a.store.MarkProcessed(msg.ID)
		return true
	}

	// ── Send reply ────────────────────────────────────────────────────────────
	var sendErr error
	if len(resp.Attachments) > 0 {
		sendErr = a.smtp.SendWithAttachments(replyAddr, resp.ReplySubject, resp.ReplyBody, resp.Attachments)
	} else {
		sendErr = a.smtp.Send(replyAddr, resp.ReplySubject, resp.ReplyBody)
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

	if err := a.store.MarkProcessed(msg.ID); err != nil {
		log.Error("store mark failed", "err", err)
	}

	log.Info("replied successfully")
	return true
}

func (a *App) sendErrorReply(to, subject string, origErr error) {
	body := fmt.Sprintf(
		"Leider ist ein Fehler aufgetreten:\n\n%v\n\nBitte versuche es später erneut oder wende dich an den Administrator.",
		origErr,
	)
	if err := a.smtp.Send(to, "Re: "+subject+" [Fehler]", body); err != nil {
		a.log.Error("could not send error reply", "to", to, "err", err)
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
