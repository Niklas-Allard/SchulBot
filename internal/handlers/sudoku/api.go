package sudoku

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type apiPuzzle struct {
	Difficulty string `json:"difficulty"`
	Puzzle     string `json:"puzzle"`
	Solution   string `json:"solution"`
}

type apiClient struct {
	apiKey string
	client *http.Client
}

func newAPIClient(apiKey string) *apiClient {
	return &apiClient{
		apiKey: apiKey,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *apiClient) fetch(ctx context.Context, difficulty string) (*apiPuzzle, error) {
	body, _ := json.Marshal(map[string]any{
		"difficulty": difficulty,
		"solution":   true,
		"array":      false,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://youdosudoku.com/api/", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("sudoku API HTTP %d: %s", resp.StatusCode, clip(string(data), 200))
	}

	var p apiPuzzle
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse puzzle: %w", err)
	}
	if len(p.Puzzle) != 81 || len(p.Solution) != 81 {
		return nil, fmt.Errorf("unexpected puzzle length: puzzle=%d solution=%d", len(p.Puzzle), len(p.Solution))
	}
	return &p, nil
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
