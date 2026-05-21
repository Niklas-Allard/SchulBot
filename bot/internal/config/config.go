package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv   string
	LogLevel string

	IMAP IMAPConfig
	SMTP SMTPConfig
	AI   AIConfig

	PollInterval     time.Duration
	MaxPayloadChars  int
	MaxResponseChars int

	DBPath string

	// Multi-user mode: if set, the bot fetches user configs from the Laravel web app
	// instead of using the IMAP/SMTP/AI env vars above.
	LaravelAPIURL string
	BotAPISecret  string
}

type IMAPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Mailbox  string
	Security string // SSL or STARTTLS
}

type SMTPConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromName    string
	FromAddress string
	Security    string // SSL or STARTTLS
}

type AIConfig struct {
	Provider string
	APIURL   string
	APIKey   string
	Model    string
}

func Load() (*Config, error) {
	var errs []error

	requireEnv := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			errs = append(errs, fmt.Errorf("required env var %s is not set", key))
		}
		return v
	}

	parseInt := func(key, fallback string) int {
		s := getEnv(key, fallback)
		n, err := strconv.Atoi(s)
		if err != nil {
			errs = append(errs, fmt.Errorf("env var %s: invalid int %q", key, s))
			return 0
		}
		return n
	}

	parseDuration := func(key, fallback string) time.Duration {
		s := getEnv(key, fallback)
		d, err := time.ParseDuration(s)
		if err != nil {
			errs = append(errs, fmt.Errorf("env var %s: invalid duration %q", key, s))
			return 0
		}
		return d
	}

	multiUser := os.Getenv("LARAVEL_API_URL") != ""

	// In multi-user mode SMTP and AI come from the web dashboard per user,
	// so only IMAP (shared inbox) is required at startup.
	smtpRequired := requireEnv
	aiRequired := requireEnv
	if multiUser {
		smtpRequired = func(key string) string { return os.Getenv(key) }
		aiRequired = func(key string) string { return os.Getenv(key) }
	}

	cfg := &Config{
		AppEnv:   getEnv("APP_ENV", "production"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		IMAP: IMAPConfig{
			Host:     requireEnv("IMAP_HOST"),
			Port:     parseInt("IMAP_PORT", "993"),
			Username: requireEnv("IMAP_USERNAME"),
			Password: requireEnv("IMAP_PASSWORD"),
			Mailbox:  getEnv("IMAP_MAILBOX", "INBOX"),
			Security: getEnv("IMAP_SECURITY", "SSL"),
		},
		SMTP: SMTPConfig{
			Host:        smtpRequired("SMTP_HOST"),
			Port:        parseInt("SMTP_PORT", "465"),
			Username:    smtpRequired("SMTP_USERNAME"),
			Password:    smtpRequired("SMTP_PASSWORD"),
			FromName:    getEnv("SMTP_FROM_NAME", "SchulBot"),
			FromAddress: smtpRequired("SMTP_FROM_ADDRESS"),
			Security:    getEnv("SMTP_SECURITY", "SSL"),
		},
		AI: AIConfig{
			Provider: getEnv("AI_PROVIDER", "gemini"),
			APIURL:   getEnv("AI_API_URL", ""),
			APIKey:   aiRequired("AI_API_KEY"),
			Model:    getEnv("AI_MODEL", ""),
		},
		PollInterval:     parseDuration("POLL_INTERVAL", "30s"),
		MaxPayloadChars:  parseInt("MAX_PAYLOAD_CHARS", "4000"),
		MaxResponseChars: parseInt("MAX_RESPONSE_CHARS", "8000"),
		DBPath:           getEnv("DB_PATH", "data/processed.db"),
		LaravelAPIURL:    os.Getenv("LARAVEL_API_URL"),
		BotAPISecret:     os.Getenv("BOT_API_SECRET"),
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
