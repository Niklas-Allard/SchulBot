package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"schulbot/internal/ai"
	"schulbot/internal/app"
	"schulbot/internal/commands"
	"schulbot/internal/config"
	"schulbot/internal/handlers/calendar"
	"schulbot/internal/handlers/hilfe"
	"schulbot/internal/handlers/ki"
	"schulbot/internal/handlers/news"
	"schulbot/internal/handlers/sudoku"
	"schulbot/internal/handlers/tasks"
	"schulbot/internal/handlers/translate"
	modelcmd "schulbot/internal/handlers/model"
	"schulbot/internal/mail"
	"schulbot/internal/store"
	"schulbot/internal/users"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration error", "err", err)
		os.Exit(1)
	}

	log := buildLogger(cfg.LogLevel)

	db, err := store.New(cfg.DBPath)
	if err != nil {
		log.Error("store init failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Multi-user mode: one shared IMAP inbox, SMTP + AI resolved per sender
	if cfg.LaravelAPIURL != "" {
		if cfg.BotAPISecret == "" {
			log.Error("LARAVEL_API_URL is set but BOT_API_SECRET is missing")
			os.Exit(1)
		}
		manager := users.NewManager(cfg.LaravelAPIURL, cfg.BotAPISecret)

		factory := func(aiProvider ai.Provider, d *store.Store) *commands.Dispatcher {
			return commands.NewDispatcher(
				ki.New(aiProvider, d, cfg.MaxPayloadChars, cfg.MaxResponseChars),
				modelcmd.New(d, aiProvider.DefaultModel()),
				sudoku.New(os.Getenv("SUDOKU_API_KEY"), d),
				news.New(),
				hilfe.New(),
				translate.New(os.Getenv("LIBRETRANSLATE_URL"), os.Getenv("LIBRETRANSLATE_API_KEY")),
				tasks.New(
					os.Getenv("GOOGLE_CLIENT_ID"),
					os.Getenv("GOOGLE_CLIENT_SECRET"),
					os.Getenv("GOOGLE_REFRESH_TOKEN"),
					os.Getenv("GOOGLE_TASKLIST_ID"),
				),
				calendar.New(
					os.Getenv("GOOGLE_CLIENT_ID"),
					os.Getenv("GOOGLE_CLIENT_SECRET"),
					os.Getenv("GOOGLE_REFRESH_TOKEN"),
					os.Getenv("GOOGLE_CALENDAR_ID"),
				),
			)
		}

		if err := app.RunMultiUser(ctx, cfg, manager, factory, db, log); err != nil {
			log.Error("multi-user runner error", "err", err)
			os.Exit(1)
		}
		return
	}

	// Single-user mode (legacy): use IMAP/SMTP/AI env vars directly
	aiProvider, err := ai.NewProvider(cfg.AI.Provider, cfg.AI.APIURL, cfg.AI.APIKey, cfg.AI.Model)
	if err != nil {
		log.Error("AI provider init failed", "err", err)
		os.Exit(1)
	}

	dispatcher := commands.NewDispatcher(
		ki.New(aiProvider, db, cfg.MaxPayloadChars, cfg.MaxResponseChars),
		modelcmd.New(db, aiProvider.DefaultModel()),
		sudoku.New(os.Getenv("SUDOKU_API_KEY"), db),
		news.New(),
		hilfe.New(),
		translate.New(os.Getenv("LIBRETRANSLATE_URL"), os.Getenv("LIBRETRANSLATE_API_KEY")),
		tasks.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			os.Getenv("GOOGLE_REFRESH_TOKEN"),
			os.Getenv("GOOGLE_TASKLIST_ID"),
		),
		calendar.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			os.Getenv("GOOGLE_REFRESH_TOKEN"),
			os.Getenv("GOOGLE_CALENDAR_ID"),
		),
	)

	imapClient := mail.NewIMAPClient(
		cfg.IMAP.Host, cfg.IMAP.Port,
		cfg.IMAP.Username, cfg.IMAP.Password,
		cfg.IMAP.Mailbox, cfg.IMAP.Security,
	)
	smtpClient := mail.NewSMTPClient(
		cfg.SMTP.Host, cfg.SMTP.Port,
		cfg.SMTP.Username, cfg.SMTP.Password,
		cfg.SMTP.FromName, cfg.SMTP.FromAddress,
		cfg.SMTP.Security,
	)

	application := app.New(cfg, imapClient, smtpClient, dispatcher, db, log)
	if err := application.Run(ctx); err != nil {
		log.Error("application error", "err", err)
		os.Exit(1)
	}
}

func buildLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}