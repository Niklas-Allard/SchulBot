package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"schulbot/internal/model"
)

// Handler connects to the Google Tasks API via OAuth2 refresh-token flow.
//
// Required env vars (passed into New):
//
//	GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GOOGLE_REFRESH_TOKEN
//	GOOGLE_TASKLIST_ID (optional, defaults to "@default")
type Handler struct {
	clientID     string
	clientSecret string
	refreshToken string
	tasklistID   string
	httpClient   *http.Client
}

func New(clientID, clientSecret, refreshToken, tasklistID string) *Handler {
	if tasklistID == "" {
		tasklistID = "@default"
	}
	return &Handler{
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		tasklistID:   tasklistID,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (h *Handler) CanHandle(tag string) bool {
	return tag == "#tasks" || tag == "/tasks"
}

func (h *Handler) Handle(ctx context.Context, req model.CommandRequest) (model.CommandResponse, error) {
	if h.clientID == "" || h.refreshToken == "" {
		return model.CommandResponse{
			ReplySubject: reply(req.Subject),
			ReplyBody:    "Google Tasks ist nicht konfiguriert. Bitte setze GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET und GOOGLE_REFRESH_TOKEN.",
		}, nil
	}

	token, err := h.refreshAccessToken(ctx)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("tasks: token refresh: %w", err)
	}

	cmd, arg := splitFirstWord(req.Payload)
	switch cmd {
	case "liste", "list":
		return h.listTasks(ctx, token, req.Subject)
	case "add", "neu", "hinzufügen":
		return h.addTask(ctx, token, arg, req.Subject)
	case "fertig", "done", "erledigt":
		return h.completeTask(ctx, token, arg, req.Subject)
	default:
		return model.CommandResponse{
			ReplySubject: reply(req.Subject),
			ReplyBody: "Unbekannter Unterbefehl. Verfügbare Befehle:\n\n" +
				"  #tasks liste               – offene Aufgaben anzeigen\n" +
				"  #tasks add <Titel>         – neue Aufgabe anlegen\n" +
				"  #tasks fertig <Titel>      – Aufgabe als erledigt markieren",
		}, nil
	}
}

// ── OAuth2 token refresh ──────────────────────────────────────────────────────

func (h *Handler) refreshAccessToken(ctx context.Context) (string, error) {
	data := url.Values{
		"client_id":     {h.clientID},
		"client_secret": {h.clientSecret},
		"refresh_token": {h.refreshToken},
		"grant_type":    {"refresh_token"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://oauth2.googleapis.com/token",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("OAuth error: %s", result.Error)
	}
	return result.AccessToken, nil
}

// ── List tasks ────────────────────────────────────────────────────────────────

type task struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"` // "needsAction" | "completed"
}

func (h *Handler) listTasks(ctx context.Context, token, subject string) (model.CommandResponse, error) {
	u := fmt.Sprintf(
		"https://tasks.googleapis.com/tasks/v1/lists/%s/tasks?showCompleted=false&maxResults=20",
		url.PathEscape(h.tasklistID),
	)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("tasks list: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return model.CommandResponse{}, fmt.Errorf("tasks list HTTP %d: %s", resp.StatusCode, clip(string(body), 200))
	}

	var result struct {
		Items []task `json:"items"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return model.CommandResponse{}, fmt.Errorf("tasks list parse: %w", err)
	}

	if len(result.Items) == 0 {
		return model.CommandResponse{
			ReplySubject: reply(subject),
			ReplyBody:    "Keine offenen Aufgaben.",
		}, nil
	}

	var sb strings.Builder
	for i, t := range result.Items {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, t.Title)
	}
	return model.CommandResponse{
		ReplySubject: reply(subject),
		ReplyBody:    "Offene Aufgaben:\n\n" + strings.TrimRight(sb.String(), "\n"),
	}, nil
}

// ── Add task ──────────────────────────────────────────────────────────────────

func (h *Handler) addTask(ctx context.Context, token, title, subject string) (model.CommandResponse, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return model.CommandResponse{
			ReplySubject: reply(subject),
			ReplyBody:    "Bitte gib einen Titel an: #tasks add <Titel>",
		}, nil
	}

	body, _ := json.Marshal(map[string]string{"title": title})
	u := fmt.Sprintf("https://tasks.googleapis.com/tasks/v1/lists/%s/tasks", url.PathEscape(h.tasklistID))
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("tasks add: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return model.CommandResponse{}, fmt.Errorf("tasks add HTTP %d: %s", resp.StatusCode, clip(string(respBody), 200))
	}

	return model.CommandResponse{
		ReplySubject: reply(subject),
		ReplyBody:    fmt.Sprintf("Aufgabe angelegt: %s", title),
	}, nil
}

// ── Complete task ─────────────────────────────────────────────────────────────

func (h *Handler) completeTask(ctx context.Context, token, titleHint, subject string) (model.CommandResponse, error) {
	titleHint = strings.TrimSpace(titleHint)
	if titleHint == "" {
		return model.CommandResponse{
			ReplySubject: reply(subject),
			ReplyBody:    "Bitte gib den Titel (oder Teil davon) an: #tasks fertig <Titel>",
		}, nil
	}

	// Fetch task list to find ID by title match
	u := fmt.Sprintf(
		"https://tasks.googleapis.com/tasks/v1/lists/%s/tasks?showCompleted=false&maxResults=100",
		url.PathEscape(h.tasklistID),
	)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("tasks complete fetch: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Items []task `json:"items"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return model.CommandResponse{}, fmt.Errorf("tasks complete parse: %w", err)
	}

	var matched *task
	hint := strings.ToLower(titleHint)
	for i := range result.Items {
		if strings.Contains(strings.ToLower(result.Items[i].Title), hint) {
			matched = &result.Items[i]
			break
		}
	}
	if matched == nil {
		return model.CommandResponse{
			ReplySubject: reply(subject),
			ReplyBody:    fmt.Sprintf("Keine offene Aufgabe gefunden, die %q enthält.", titleHint),
		}, nil
	}

	// Mark as completed via PATCH
	patchBody, _ := json.Marshal(map[string]string{"status": "completed"})
	patchURL := fmt.Sprintf("https://tasks.googleapis.com/tasks/v1/lists/%s/tasks/%s",
		url.PathEscape(h.tasklistID), url.PathEscape(matched.ID))
	patchReq, _ := http.NewRequestWithContext(ctx, http.MethodPatch, patchURL, strings.NewReader(string(patchBody)))
	patchReq.Header.Set("Authorization", "Bearer "+token)
	patchReq.Header.Set("Content-Type", "application/json")

	patchResp, err := h.httpClient.Do(patchReq)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("tasks complete patch: %w", err)
	}
	defer patchResp.Body.Close()

	patchRespBody, _ := io.ReadAll(patchResp.Body)
	if patchResp.StatusCode >= 400 {
		return model.CommandResponse{}, fmt.Errorf("tasks complete HTTP %d: %s", patchResp.StatusCode, clip(string(patchRespBody), 200))
	}

	return model.CommandResponse{
		ReplySubject: reply(subject),
		ReplyBody:    fmt.Sprintf("Aufgabe als erledigt markiert: %s", matched.Title),
	}, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func splitFirstWord(s string) (word, rest string) {
	s = strings.TrimSpace(s)
	parts := strings.SplitN(s, " ", 2)
	word = strings.ToLower(parts[0])
	if len(parts) > 1 {
		rest = strings.TrimSpace(parts[1])
	}
	return
}

func reply(subject string) string {
	if strings.HasPrefix(subject, "Re: ") {
		return subject
	}
	return "Re: " + subject
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
