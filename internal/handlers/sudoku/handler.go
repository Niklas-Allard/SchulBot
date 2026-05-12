package sudoku

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"schulbot/internal/model"
	"schulbot/internal/store"
)

const (
	idChars   = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no 0/O, 1/I confusion
	puzzleTTL = 48 * time.Hour
)

// Handler implements the /sudoku command.
type Handler struct {
	api    *apiClient
	store  *store.Store
}

func New(apiKey string, s *store.Store) *Handler {
	return &Handler{api: newAPIClient(apiKey), store: s}
}

func (h *Handler) CanHandle(tag string) bool {
	return tag == "#sudoku" || tag == "/sudoku"
}

func (h *Handler) Handle(ctx context.Context, req model.CommandRequest) (model.CommandResponse, error) {
	payload := strings.TrimSpace(req.Payload)
	lower := strings.ToLower(payload)

	// /sudoku lösung ID  or  /sudoku solution ID
	if strings.HasPrefix(lower, "lösung ") || strings.HasPrefix(lower, "solution ") {
		parts := strings.Fields(payload)
		if len(parts) < 2 {
			return h.helpResponse(req.Subject), nil
		}
		id := strings.ToUpper(parts[len(parts)-1])
		return h.handleSolution(ctx, req.Subject, id)
	}

	// /sudoku [easy|medium|hard]
	difficulty := "easy"
	switch lower {
	case "medium", "mittel":
		difficulty = "medium"
	case "hard", "schwer":
		difficulty = "hard"
	case "easy", "einfach", "":
		difficulty = "easy"
	case "hilfe", "help":
		return h.helpResponse(req.Subject), nil
	}

	return h.handleNew(ctx, req.Subject, difficulty)
}

func (h *Handler) handleNew(ctx context.Context, subject, difficulty string) (model.CommandResponse, error) {
	p, err := h.api.fetch(ctx, difficulty)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("sudoku: fetch puzzle: %w", err)
	}

	id, err := generateID()
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("sudoku: generate id: %w", err)
	}

	now := time.Now().UTC()
	entry := store.SudokuEntry{
		ID:         id,
		Puzzle:     p.Puzzle,
		Solution:   p.Solution,
		Difficulty: p.Difficulty,
		CreatedAt:  now,
		ExpiresAt:  now.Add(puzzleTTL),
	}
	if err := h.store.SaveSudoku(entry); err != nil {
		return model.CommandResponse{}, fmt.Errorf("sudoku: save: %w", err)
	}

	// Render puzzle image (no solution shown)
	imgData, err := RenderPuzzle(p.Puzzle, "")
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("sudoku: render: %w", err)
	}

	diffLabel := map[string]string{"easy": "Einfach", "medium": "Mittel", "hard": "Schwer"}[p.Difficulty]
	body := fmt.Sprintf(
		"Dein Sudoku-Rätsel\n\n"+
			"ID: %s\n"+
			"Schwierigkeit: %s\n\n"+
			"Das Bild ist im Anhang.\n\n"+
			"Lösung abrufen (2 Tage gültig):\n  /sudoku lösung %s",
		id, diffLabel, id,
	)

	return model.CommandResponse{
		ReplySubject: reply(subject),
		ReplyBody:    body,
		Attachments: []model.Attachment{{
			Filename:    fmt.Sprintf("sudoku_%s.png", id),
			ContentType: "image/png",
			Data:        imgData,
		}},
	}, nil
}

func (h *Handler) handleSolution(ctx context.Context, subject, id string) (model.CommandResponse, error) {
	entry, err := h.store.GetSudoku(id)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("sudoku: lookup: %w", err)
	}
	if entry == nil {
		return model.CommandResponse{
			ReplySubject: reply(subject),
			ReplyBody:    fmt.Sprintf("Kein Puzzle mit ID %s gefunden (abgelaufen oder ungültig).", id),
		}, nil
	}

	imgData, err := RenderPuzzle(entry.Puzzle, entry.Solution)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("sudoku: render solution: %w", err)
	}

	body := fmt.Sprintf(
		"Lösung für Puzzle %s\n\nSchwierigkeit: %s\nErstellt: %s",
		id,
		entry.Difficulty,
		entry.CreatedAt.In(time.FixedZone("Europe/Berlin", 2*3600)).Format("02.01.2006 15:04"),
	)

	return model.CommandResponse{
		ReplySubject: reply(subject),
		ReplyBody:    body,
		Attachments: []model.Attachment{{
			Filename:    fmt.Sprintf("sudoku_%s_loesung.png", id),
			ContentType: "image/png",
			Data:        imgData,
		}},
	}, nil
}

func (h *Handler) helpResponse(subject string) model.CommandResponse {
	return model.CommandResponse{
		ReplySubject: reply(subject),
		ReplyBody: "Sudoku-Befehle:\n\n" +
			"  /sudoku              → neues Puzzle (Einfach)\n" +
			"  /sudoku easy         → Einfach\n" +
			"  /sudoku medium       → Mittel\n" +
			"  /sudoku hard         → Schwer\n" +
			"  /sudoku lösung ID    → Lösung für Puzzle-ID abrufen\n\n" +
			"Puzzles sind 2 Tage lang abrufbar.",
	}
}

func generateID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	result := make([]byte, 6)
	for i, c := range b {
		result[i] = idChars[int(c)%len(idChars)]
	}
	return string(result), nil
}

func reply(subject string) string {
	if strings.HasPrefix(subject, "Re: ") {
		return subject
	}
	return "Re: " + subject
}
