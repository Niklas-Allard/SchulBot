package modelcmd

import (
	"context"
	"fmt"
	"strings"

	"schulbot/internal/model"
	"schulbot/internal/store"
)

// AvailableModels lists curated Cloudflare Workers AI text models.
// Users can pick any entry by index number or full model name.
var AvailableModels = []struct {
	ID          string
	Description string
}{
	{"@cf/meta/llama-3.1-8b-instruct", "Llama 3.1 8B – schnell, Standard"},
	{"@cf/meta/llama-3.3-70b-instruct-fp8-fast", "Llama 3.3 70B – schlau, etwas langsamer"},
	{"@cf/google/gemma-3-12b-it", "Google Gemma 3 12B"},
	{"@cf/mistral/mistral-7b-instruct-v0.2", "Mistral 7B"},
	{"@cf/deepseek-ai/deepseek-r1-distill-qwen-32b", "DeepSeek R1 32B – stark bei Logik"},
	{"@cf/qwen/qwen2.5-72b-instruct", "Qwen 2.5 72B"},
}

// Handler implements the /model command for per-user model selection.
type Handler struct {
	store        *store.Store
	defaultModel string
}

func New(s *store.Store, defaultModel string) *Handler {
	return &Handler{store: s, defaultModel: defaultModel}
}

func (h *Handler) CanHandle(tag string) bool {
	return tag == "#model" || tag == "/model"
}

func (h *Handler) Handle(_ context.Context, req model.CommandRequest) (model.CommandResponse, error) {
	payload := strings.TrimSpace(req.Payload)

	// No argument or "list" → show available models + current selection
	if payload == "" || strings.ToLower(payload) == "list" {
		return h.listModels(req)
	}

	// "reset" → remove user override
	if strings.ToLower(payload) == "reset" {
		return h.resetModel(req)
	}

	// Numeric index (1-based)
	var idx int
	if _, err := fmt.Sscanf(payload, "%d", &idx); err == nil {
		if idx < 1 || idx > len(AvailableModels) {
			return model.CommandResponse{
				ReplySubject: reply(req.Subject),
				ReplyBody:    fmt.Sprintf("Ungültige Nummer. Wähle zwischen 1 und %d.", len(AvailableModels)),
			}, nil
		}
		return h.setModel(req, AvailableModels[idx-1].ID)
	}

	// Full model name (must start with @cf/ or be in the list)
	candidate := payload
	if !strings.HasPrefix(candidate, "@cf/") && !strings.HasPrefix(candidate, "@") {
		return model.CommandResponse{
			ReplySubject: reply(req.Subject),
			ReplyBody: fmt.Sprintf(
				"Unbekanntes Format. Nutze eine Nummer (1–%d) oder den vollen Modellnamen.\n\n%s",
				len(AvailableModels), modelListText(h.currentModel(req.From)),
			),
		}, nil
	}
	return h.setModel(req, candidate)
}

func (h *Handler) listModels(req model.CommandRequest) (model.CommandResponse, error) {
	current := h.currentModel(req.From)
	return model.CommandResponse{
		ReplySubject: reply(req.Subject),
		ReplyBody:    "Verfügbare Modelle:\n\n" + modelListText(current),
	}, nil
}

func (h *Handler) setModel(req model.CommandRequest, modelID string) (model.CommandResponse, error) {
	prefs, err := h.store.GetUserPrefs(req.From)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("model: load prefs: %w", err)
	}
	prefs.AIModel = modelID
	if err := h.store.SetUserPrefs(req.From, prefs); err != nil {
		return model.CommandResponse{}, fmt.Errorf("model: save prefs: %w", err)
	}
	return model.CommandResponse{
		ReplySubject: reply(req.Subject),
		ReplyBody:    fmt.Sprintf("Modell gesetzt: %s\n\nDeine nächste #ki-Anfrage verwendet dieses Modell.", modelID),
	}, nil
}

func (h *Handler) resetModel(req model.CommandRequest) (model.CommandResponse, error) {
	prefs, err := h.store.GetUserPrefs(req.From)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("model: load prefs: %w", err)
	}
	prefs.AIModel = ""
	if err := h.store.SetUserPrefs(req.From, prefs); err != nil {
		return model.CommandResponse{}, fmt.Errorf("model: save prefs: %w", err)
	}
	return model.CommandResponse{
		ReplySubject: reply(req.Subject),
		ReplyBody:    fmt.Sprintf("Modell zurückgesetzt auf Standard: %s", h.defaultModel),
	}, nil
}

func (h *Handler) currentModel(email string) string {
	prefs, _ := h.store.GetUserPrefs(email)
	if prefs.AIModel != "" {
		return prefs.AIModel
	}
	return h.defaultModel
}

func modelListText(current string) string {
	var sb strings.Builder
	for i, m := range AvailableModels {
		marker := "  "
		if m.ID == current {
			marker = "→ "
		}
		fmt.Fprintf(&sb, "%s%d. %s\n   %s\n\n", marker, i+1, m.ID, m.Description)
	}
	sb.WriteString("Antworten:\n")
	sb.WriteString("  /model 2          → Modell Nr. 2 wählen\n")
	sb.WriteString("  /model @cf/…      → Modell direkt angeben\n")
	sb.WriteString("  /model reset      → Standard wiederherstellen")
	return sb.String()
}

func reply(subject string) string {
	if strings.HasPrefix(subject, "Re: ") {
		return subject
	}
	return "Re: " + subject
}
