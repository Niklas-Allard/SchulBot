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
			Host:        requireEnv("SMTP_HOST"),
			Port:        parseInt("SMTP_PORT", "465"),
			Username:    requireEnv("SMTP_USERNAME"),
			Password:    requireEnv("SMTP_PASSWORD"),
			FromName:    getEnv("SMTP_FROM_NAME", "SchulBot"),
			FromAddress: requireEnv("SMTP_FROM_ADDRESS"),
			Security:    getEnv("SMTP_SECURITY", "SSL"),
		},
		AI: AIConfig{
			Provider: getEnv("AI_PROVIDER", "openai"),
			APIURL:   getEnv("AI_API_URL", "https://api.openai.com/v1"),
			APIKey:   requireEnv("AI_API_KEY"),
			Model:    getEnv("AI_MODEL", "gpt-4o-mini"),
		},
		PollInterval:     parseDuration("POLL_INTERVAL", "30s"),
		MaxPayloadChars:  parseInt("MAX_PAYLOAD_CHARS", "4000"),
		MaxResponseChars: parseInt("MAX_RESPONSE_CHARS", "8000"),
		DBPath:           getEnv("DB_PATH", "data/processed.db"),
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
