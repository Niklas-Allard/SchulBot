package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// UserConfig mirrors the JSON returned by /api/internal/bot-users.
// IMAP is shared (lives in bot .env); each user has their own SMTP sender + AI config.
type UserConfig struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"user_email"`

	SMTPHost        string `json:"smtp_host"`
	SMTPPort        int    `json:"smtp_port"`
	SMTPUsername    string `json:"smtp_username"`
	SMTPPassword    string `json:"smtp_password"`
	SMTPSecurity    string `json:"smtp_security"`
	SMTPFromName    string `json:"smtp_from_name"`
	SMTPFromAddress string `json:"smtp_from_address"`

	AIProvider string `json:"ai_provider"`
	AIAPIURL   string `json:"ai_api_url"`
	AIAPIKey   string `json:"ai_api_key"`
	AIModel    string `json:"ai_model"`
}

// Manager fetches active user configs from the Laravel web app.
type Manager struct {
	apiURL string
	secret string
	client *http.Client
}

func NewManager(apiURL, secret string) *Manager {
	return &Manager{
		apiURL: apiURL,
		secret: secret,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// FindByEmail looks up a single user's config by their sender email address.
// Returns nil, nil when the email is not registered or the bot is inactive.
func (m *Manager) FindByEmail(ctx context.Context, email string) (*UserConfig, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		m.apiURL+"/api/internal/user-config?email="+email, nil)
	if err != nil {
		return nil, fmt.Errorf("users: build request: %w", err)
	}
	req.Header.Set("X-Bot-Secret", m.secret)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("users: http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("users: api returned %d", resp.StatusCode)
	}

	var result struct {
		Found bool `json:"found"`
		UserConfig
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("users: decode response: %w", err)
	}
	if !result.Found {
		return nil, nil
	}
	uc := result.UserConfig
	return &uc, nil
}

// LoadUsers fetches all active user configs from the Laravel internal API.
func (m *Manager) LoadUsers(ctx context.Context) ([]UserConfig, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.apiURL+"/api/internal/bot-users", nil)
	if err != nil {
		return nil, fmt.Errorf("users: build request: %w", err)
	}
	req.Header.Set("X-Bot-Secret", m.secret)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("users: http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("users: api returned %d", resp.StatusCode)
	}

	var configs []UserConfig
	if err := json.NewDecoder(resp.Body).Decode(&configs); err != nil {
		return nil, fmt.Errorf("users: decode response: %w", err)
	}
	return configs, nil
}