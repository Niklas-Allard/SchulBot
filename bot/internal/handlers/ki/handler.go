package ki

import (
	"context"
	"fmt"
	"strings"

	"schulbot/internal/ai"
	"schulbot/internal/model"
	"schulbot/internal/store"
)

type Handler struct {
	ai          ai.Provider
	store       *store.Store
	maxPayload  int
	maxResponse int
}

func New(provider ai.Provider, s *store.Store, maxPayload, maxResponse int) *Handler {
	return &Handler{ai: provider, store: s, maxPayload: maxPayload, maxResponse: maxResponse}
}

func (h *Handler) CanHandle(tag string) bool {
	return tag == "#ki" || tag == "/ki"
}

func (h *Handler) Handle(ctx context.Context, req model.CommandRequest) (model.CommandResponse, error) {
	payload := strings.TrimSpace(req.Payload)
	if payload == "" {
		return model.CommandResponse{
			ReplySubject: replySubject(req.Subject),
			ReplyBody: "Bitte gib eine Frage oder Aufgabe nach dem Tag an.\n\n" +
				"Beispiel:\n  #ki Was ist Photosynthese?\n  /ki Erkläre den Kreislauf des Wassers.\n\n" +
				"Modell wechseln: /model",
		}, nil
	}

	if len([]rune(payload)) > h.maxPayload {
		payload = string([]rune(payload)[:h.maxPayload])
	}

	// Look up user's preferred model (empty = provider default)
	userModel := h.userModel(req.From)

	answer, err := h.ai.Generate(ctx, payload, userModel)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("ki: %w", err)
	}

	runes := []rune(answer)
	if len(runes) > h.maxResponse {
		answer = string(runes[:h.maxResponse]) + "\n\n[Antwort wurde aufgrund der Längenbegrenzung gekürzt.]"
	}

	footer := ""
	if userModel != "" && userModel != h.ai.DefaultModel() {
		footer = fmt.Sprintf("\n\n─\nModell: %s", userModel)
	}

	return model.CommandResponse{
		ReplySubject: replySubject(req.Subject),
		ReplyBody:    answer + footer,
	}, nil
}

func (h *Handler) userModel(email string) string {
	if h.store == nil {
		return ""
	}
	prefs, _ := h.store.GetUserPrefs(email)
	return prefs.AIModel
}

func replySubject(subject string) string {
	if strings.HasPrefix(subject, "Re: ") {
		return subject
	}
	return "Re: " + subject
}