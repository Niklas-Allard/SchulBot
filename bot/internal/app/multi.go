package app

import (
	"context"
	"log/slog"

	"schulbot/internal/config"
	imail "schulbot/internal/mail"
	"schulbot/internal/store"
	"schulbot/internal/users"
)

// RunMultiUser starts a single IMAP poller for the shared inbox.
// Per-message, the sender's SMTP + AI config is fetched from the web dashboard.
func RunMultiUser(
	ctx context.Context,
	cfg *config.Config,
	manager *users.Manager,
	factory DispatcherFactory,
	db *store.Store,
	log *slog.Logger,
) error {
	log.Info("schulbot multi-user mode starting", "imap_host", cfg.IMAP.Host)

	imapClient := imail.NewIMAPClient(
		cfg.IMAP.Host, cfg.IMAP.Port,
		cfg.IMAP.Username, cfg.IMAP.Password,
		cfg.IMAP.Mailbox, cfg.IMAP.Security,
	)

	application := NewMultiUser(cfg, imapClient, db, manager, factory, log)
	return application.Run(ctx)
}
