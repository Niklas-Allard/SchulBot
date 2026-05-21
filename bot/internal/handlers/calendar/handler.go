package calendar

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

// Handler connects to Google Calendar API via OAuth2 refresh-token flow.
//
// Required env vars (loaded via Config, passed into New):
//
//	GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GOOGLE_REFRESH_TOKEN
//	GOOGLE_CALENDAR_ID (optional, defaults to "primary")
type Handler struct {
	clientID     string
	clientSecret string
	refreshToken string
	calendarID   string
	httpClient   *http.Client
}

func New(clientID, clientSecret, refreshToken, calendarID string) *Handler {
	if calendarID == "" {
		calendarID = "primary"
	}
	return &Handler{
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		calendarID:   calendarID,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (h *Handler) CanHandle(tag string) bool {
	return tag == "#calendar" || tag == "/calendar"
}

func (h *Handler) Handle(ctx context.Context, req model.CommandRequest) (model.CommandResponse, error) {
	if h.clientID == "" || h.refreshToken == "" {
		return model.CommandResponse{
			ReplySubject: reply(req.Subject),
			ReplyBody:    "Google Calendar ist nicht konfiguriert. Bitte setze GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET und GOOGLE_REFRESH_TOKEN.",
		}, nil
	}

	token, err := h.refreshAccessToken(ctx)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("calendar: token refresh: %w", err)
	}

	cmd, arg := parseSubcommand(req.Payload)
	switch cmd {
	case "heute":
		return h.listEvents(ctx, token, today(), today().Add(24*time.Hour), req.Subject)
	case "woche":
		return h.listEvents(ctx, token, today(), today().Add(7*24*time.Hour), req.Subject)
	case "neu":
		return h.createEvent(ctx, token, arg, req.Subject)
	default:
		return model.CommandResponse{
			ReplySubject: reply(req.Subject),
			ReplyBody: "Unbekannter Unterbefehl. Verfügbare Befehle:\n\n" +
				"  #calendar heute            – heutige Termine\n" +
				"  #calendar woche            – Termine dieser Woche\n" +
				"  #calendar neu <Titel> <Datum [HH:MM]>  – Termin anlegen\n\n" +
				"Datumsformat: 2026-05-15  oder  15.5.2026  oder  morgen  oder  übermorgen",
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
		return "", fmt.Errorf("parse token response: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("OAuth error: %s", result.Error)
	}
	return result.AccessToken, nil
}

// ── List events ───────────────────────────────────────────────────────────────

type calEvent struct {
	Summary string    `json:"summary"`
	Start   eventTime `json:"start"`
	End     eventTime `json:"end"`
}

type eventTime struct {
	DateTime string `json:"dateTime"`
	Date     string `json:"date"`
}

func (h *Handler) listEvents(ctx context.Context, token string, from, to time.Time, subject string) (model.CommandResponse, error) {
	u := fmt.Sprintf(
		"https://www.googleapis.com/calendar/v3/calendars/%s/events?timeMin=%s&timeMax=%s&singleEvents=true&orderBy=startTime&maxResults=15",
		url.PathEscape(h.calendarID),
		url.QueryEscape(from.UTC().Format(time.RFC3339)),
		url.QueryEscape(to.UTC().Format(time.RFC3339)),
	)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("calendar list: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return model.CommandResponse{}, fmt.Errorf("calendar list HTTP %d: %s", resp.StatusCode, clip(string(body), 200))
	}

	var result struct {
		Items []calEvent `json:"items"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return model.CommandResponse{}, fmt.Errorf("calendar list parse: %w", err)
	}

	if len(result.Items) == 0 {
		return model.CommandResponse{
			ReplySubject: reply(subject),
			ReplyBody:    "Keine Termine in diesem Zeitraum gefunden.",
		}, nil
	}

	var sb strings.Builder
	for _, ev := range result.Items {
		timeStr := formatEventTime(ev.Start)
		fmt.Fprintf(&sb, "• %s  –  %s\n", timeStr, ev.Summary)
	}
	return model.CommandResponse{
		ReplySubject: reply(subject),
		ReplyBody:    strings.TrimRight(sb.String(), "\n"),
	}, nil
}

// ── Create event ──────────────────────────────────────────────────────────────

func (h *Handler) createEvent(ctx context.Context, token, arg, subject string) (model.CommandResponse, error) {
	title, start, err := parseNewEventArgs(arg)
	if err != nil {
		return model.CommandResponse{
			ReplySubject: reply(subject),
			ReplyBody:    fmt.Sprintf("Konnte den Termin nicht anlegen: %v\n\nBeispiele:\n  #calendar neu Mathe-Test 2026-05-20 09:00\n  #calendar neu Elternabend morgen 19:00", err),
		}, nil
	}

	end := start.Add(time.Hour)
	body, _ := json.Marshal(map[string]any{
		"summary": title,
		"start":   map[string]string{"dateTime": start.Format(time.RFC3339), "timeZone": "Europe/Berlin"},
		"end":     map[string]string{"dateTime": end.Format(time.RFC3339), "timeZone": "Europe/Berlin"},
	})

	u := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/%s/events", url.PathEscape(h.calendarID))
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("calendar create: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return model.CommandResponse{}, fmt.Errorf("calendar create HTTP %d: %s", resp.StatusCode, clip(string(respBody), 200))
	}

	return model.CommandResponse{
		ReplySubject: reply(subject),
		ReplyBody:    fmt.Sprintf("Termin angelegt: %s am %s", title, start.Format("02.01.2006 15:04")),
	}, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func parseSubcommand(payload string) (cmd, rest string) {
	payload = strings.TrimSpace(payload)
	parts := strings.SplitN(payload, " ", 2)
	cmd = strings.ToLower(parts[0])
	if len(parts) > 1 {
		rest = strings.TrimSpace(parts[1])
	}
	return
}

// parseNewEventArgs extracts title and start time from strings like:
//
//	"Mathe-Test 2026-05-20 09:00"
//	"Elternabend morgen 19:00"
//	"Sport heute 14:00"
func parseNewEventArgs(arg string) (title string, start time.Time, err error) {
	if arg == "" {
		return "", time.Time{}, fmt.Errorf("Bitte Titel und Datum angeben")
	}

	loc, _ := time.LoadLocation("Europe/Berlin")
	now := time.Now().In(loc)
	base := today().In(loc)

	tokens := strings.Fields(arg)
	if len(tokens) < 2 {
		return "", time.Time{}, fmt.Errorf("mindestens Titel und Datum erforderlich")
	}

	// Try to find date and optional time at the end of the token list.
	// Strategy: walk backwards to find a date token.
	dateIdx := -1
	var parsedDate time.Time

	// Check last token for HH:MM time
	last := tokens[len(tokens)-1]
	var timeOffset time.Duration
	hasTime := false
	if len(last) == 5 && last[2] == ':' {
		var h, m int
		if _, scanErr := fmt.Sscanf(last, "%d:%d", &h, &m); scanErr == nil {
			timeOffset = time.Duration(h)*time.Hour + time.Duration(m)*time.Minute
			hasTime = true
			tokens = tokens[:len(tokens)-1]
		}
	}

	// Check the (now possibly shifted) last token for a date
	if len(tokens) > 0 {
		last = tokens[len(tokens)-1]
		switch strings.ToLower(last) {
		case "heute":
			parsedDate = base
			dateIdx = len(tokens) - 1
		case "morgen":
			parsedDate = base.Add(24 * time.Hour)
			dateIdx = len(tokens) - 1
		case "übermorgen", "uebermorgen":
			parsedDate = base.Add(48 * time.Hour)
			dateIdx = len(tokens) - 1
		default:
			// Try ISO: 2026-05-15
			if t, e := time.ParseInLocation("2006-01-02", last, loc); e == nil {
				parsedDate = t
				dateIdx = len(tokens) - 1
			} else if t, e := time.ParseInLocation("2.1.2006", last, loc); e == nil {
				parsedDate = t
				dateIdx = len(tokens) - 1
			} else if t, e := time.ParseInLocation("02.01.2006", last, loc); e == nil {
				parsedDate = t
				dateIdx = len(tokens) - 1
			}
		}
	}

	if dateIdx < 0 {
		return "", time.Time{}, fmt.Errorf("kein gültiges Datum gefunden – unterstützt: JJJJ-MM-TT, TT.MM.JJJJ, heute, morgen, übermorgen")
	}

	title = strings.Join(tokens[:dateIdx], " ")
	if title == "" {
		return "", time.Time{}, fmt.Errorf("Titel fehlt")
	}

	start = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, loc)
	if hasTime {
		start = start.Add(timeOffset)
	} else {
		// Default to 8:00 for school context
		start = start.Add(8 * time.Hour)
	}
	_ = now
	return title, start, nil
}

func today() time.Time {
	loc, _ := time.LoadLocation("Europe/Berlin")
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
}

func formatEventTime(et eventTime) string {
	if et.DateTime != "" {
		t, err := time.Parse(time.RFC3339, et.DateTime)
		if err == nil {
			loc, _ := time.LoadLocation("Europe/Berlin")
			return t.In(loc).Format("02.01. 15:04")
		}
	}
	if et.Date != "" {
		t, err := time.Parse("2006-01-02", et.Date)
		if err == nil {
			return t.Format("02.01.")
		}
	}
	return "?"
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