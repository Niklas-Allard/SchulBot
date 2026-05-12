package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Provider is the AI generation interface.
// model overrides the configured default when non-empty.
type Provider interface {
	Generate(ctx context.Context, prompt, model string) (string, error)
	DefaultModel() string
}

// NewProvider returns the correct implementation based on the provider name.
//
//	"gemini"     – Google Gemini API
//	"cloudflare" – Cloudflare Workers AI
func NewProvider(provider, apiURL, apiKey, model string) (Provider, error) {
	base := &baseClient{
		client: &http.Client{Timeout: 90 * time.Second},
	}
	switch strings.ToLower(provider) {
	case "gemini":
		url := apiURL
		if url == "" {
			url = "https://generativelanguage.googleapis.com/v1beta"
		}
		return &geminiProvider{baseClient: base, apiURL: url, apiKey: apiKey, defaultModel: model}, nil
	case "cloudflare":
		if apiURL == "" {
			return nil, fmt.Errorf("ai: AI_API_URL is required for Cloudflare provider (e.g. https://api.cloudflare.com/client/v4/accounts/<ACCOUNT_ID>/ai/run)")
		}
		return &cloudflareProvider{baseClient: base, apiURL: apiURL, apiKey: apiKey, defaultModel: model}, nil
	default:
		return nil, fmt.Errorf("ai: unknown provider %q – supported: gemini, cloudflare", provider)
	}
}

// ── shared HTTP helper ────────────────────────────────────────────────────────

type baseClient struct {
	client *http.Client
}

func (b *baseClient) post(ctx context.Context, url string, headers map[string]string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, clip(string(data), 300))
	}
	return data, nil
}

// ── Google Gemini ─────────────────────────────────────────────────────────────

type geminiProvider struct {
	*baseClient
	apiURL       string
	apiKey       string
	defaultModel string
}

func (p *geminiProvider) DefaultModel() string { return p.defaultModel }

func (p *geminiProvider) Generate(ctx context.Context, prompt, model string) (string, error) {
	if model == "" {
		model = p.defaultModel
	}
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.apiURL, model, p.apiKey)

	payload, _ := json.Marshal(map[string]any{
		"contents": []map[string]any{
			{"role": "user", "parts": []map[string]string{{"text": prompt}}},
		},
	})

	data, err := p.post(ctx, url, nil, payload)
	if err != nil {
		return "", fmt.Errorf("gemini: %w", err)
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("gemini: parse response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("gemini: API error: %s", result.Error.Message)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: empty response")
	}
	return result.Candidates[0].Content.Parts[0].Text, nil
}

// ── Cloudflare Workers AI ─────────────────────────────────────────────────────

type cloudflareProvider struct {
	*baseClient
	apiURL       string
	apiKey       string
	defaultModel string
}

func (p *cloudflareProvider) DefaultModel() string { return p.defaultModel }

func (p *cloudflareProvider) Generate(ctx context.Context, prompt, model string) (string, error) {
	if model == "" {
		model = p.defaultModel
	}
	url := strings.TrimRight(p.apiURL, "/") + "/" + strings.TrimLeft(model, "/")

	payload, _ := json.Marshal(map[string]any{
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})

	data, err := p.post(ctx, url,
		map[string]string{"Authorization": "Bearer " + p.apiKey},
		payload,
	)
	if err != nil {
		return "", fmt.Errorf("cloudflare: %w", err)
	}

	var result struct {
		Result  struct{ Response string } `json:"result"`
		Success bool                      `json:"success"`
		Errors  []struct{ Message string } `json:"errors"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("cloudflare: parse response: %w", err)
	}
	if !result.Success || len(result.Errors) > 0 {
		if len(result.Errors) > 0 {
			return "", fmt.Errorf("cloudflare: API error: %s", result.Errors[0].Message)
		}
		return "", fmt.Errorf("cloudflare: request unsuccessful")
	}
	if result.Result.Response == "" {
		return "", fmt.Errorf("cloudflare: empty response")
	}
	return result.Result.Response, nil
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}