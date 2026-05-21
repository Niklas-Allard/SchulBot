package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"schulbot/internal/model"
)

// langAliases maps German names and short codes to ISO 639-1 codes used by LibreTranslate.
var langAliases = map[string]string{
	"de": "de", "deutsch": "de", "german": "de",
	"en": "en", "englisch": "en", "english": "en",
	"fr": "fr", "französisch": "fr", "franz": "fr", "french": "fr",
	"es": "es", "spanisch": "es", "spanish": "es",
	"it": "it", "italienisch": "it", "italian": "it",
	"pt": "pt", "portugiesisch": "pt", "portuguese": "pt",
	"nl": "nl", "niederländisch": "nl", "dutch": "nl",
	"pl": "pl", "polnisch": "pl", "polish": "pl",
	"ru": "ru", "russisch": "ru", "russian": "ru",
	"ja": "ja", "japanisch": "ja", "japanese": "ja",
	"zh": "zh", "chinesisch": "zh", "chinese": "zh",
	"ar": "ar", "arabisch": "ar", "arabic": "ar",
	"ko": "ko", "koreanisch": "ko", "korean": "ko",
	"tr": "tr", "türkisch": "tr", "turkish": "tr",
	"uk": "uk", "ukrainisch": "uk", "ukrainian": "uk",
	"sv": "sv", "schwedisch": "sv", "swedish": "sv",
	"da": "da", "dänisch": "da", "danish": "da",
	"fi": "fi", "finnisch": "fi", "finnish": "fi",
	"cs": "cs", "tschechisch": "cs", "czech": "cs",
	"hi": "hi", "hindi": "hi",
	"id": "id", "indonesisch": "id", "indonesian": "id",
}

var langNames = map[string]string{
	"de": "Deutsch", "en": "Englisch", "fr": "Französisch",
	"es": "Spanisch", "it": "Italienisch", "pt": "Portugiesisch",
	"nl": "Niederländisch", "pl": "Polnisch", "ru": "Russisch",
	"ja": "Japanisch", "zh": "Chinesisch", "ar": "Arabisch",
	"ko": "Koreanisch", "tr": "Türkisch", "uk": "Ukrainisch",
	"sv": "Schwedisch", "da": "Dänisch", "fi": "Finnisch",
	"cs": "Tschechisch", "hi": "Hindi", "id": "Indonesisch",
}

// Handler translates text using a LibreTranslate instance.
type Handler struct {
	apiURL string
	apiKey string
	client *http.Client
}

func New(apiURL, apiKey string) *Handler {
	if apiURL == "" {
		apiURL = "https://libretranslate.com"
	}
	return &Handler{
		apiURL: strings.TrimRight(apiURL, "/"),
		apiKey: apiKey,
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (h *Handler) CanHandle(tag string) bool {
	return tag == "#translate" || tag == "/translate" ||
		tag == "#übersetzen" || tag == "/übersetzen"
}

func (h *Handler) Handle(ctx context.Context, req model.CommandRequest) (model.CommandResponse, error) {
	payload := strings.TrimSpace(req.Payload)
	lower := strings.ToLower(payload)

	if payload == "" || lower == "hilfe" || lower == "help" {
		return h.helpResponse(req.Subject), nil
	}

	// Parse optional target language as first word
	targetCode := "de" // default: translate to German
	text := payload

	parts := strings.SplitN(payload, " ", 2)
	if len(parts) == 2 {
		if code, ok := langAliases[strings.ToLower(parts[0])]; ok {
			targetCode = code
			text = strings.TrimSpace(parts[1])
		}
	}

	if text == "" {
		return model.CommandResponse{
			ReplySubject: reply(req.Subject),
			ReplyBody:    "Bitte Text angeben.\n\nBeispiel: /translate en Das ist ein Test",
		}, nil
	}

	translated, detectedCode, err := h.translate(ctx, text, targetCode)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("translate: %w", err)
	}

	srcLabel := detectedCode
	if n, ok := langNames[detectedCode]; ok {
		srcLabel = n
	}
	tgtLabel := targetCode
	if n, ok := langNames[targetCode]; ok {
		tgtLabel = n
	}

	body := fmt.Sprintf("%s\n\n[%s -> %s]", translated, srcLabel, tgtLabel)
	return model.CommandResponse{
		ReplySubject: reply(req.Subject),
		ReplyBody:    body,
	}, nil
}

// ── LibreTranslate API ────────────────────────────────────────────────────────

func (h *Handler) translate(ctx context.Context, text, target string) (translated, detectedSrc string, err error) {
	// Detect source language first so we can show it in the response
	detectedSrc, err = h.detect(ctx, text)
	if err != nil {
		detectedSrc = "auto"
	}

	payload := map[string]any{
		"q":      text,
		"source": detectedSrc,
		"target": target,
		"format": "text",
	}
	if h.apiKey != "" {
		payload["api_key"] = h.apiKey
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.apiURL+"/translate", bytes.NewReader(body))
	if err != nil {
		return "", detectedSrc, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return "", detectedSrc, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		json.Unmarshal(data, &errResp)
		if errResp.Error != "" {
			return "", detectedSrc, fmt.Errorf("LibreTranslate: %s", errResp.Error)
		}
		return "", detectedSrc, fmt.Errorf("LibreTranslate HTTP %d", resp.StatusCode)
	}

	var result struct {
		TranslatedText string `json:"translatedText"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", detectedSrc, fmt.Errorf("parse response: %w", err)
	}
	return result.TranslatedText, detectedSrc, nil
}

func (h *Handler) detect(ctx context.Context, text string) (string, error) {
	payload := map[string]any{"q": text}
	if h.apiKey != "" {
		payload["api_key"] = h.apiKey
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.apiURL+"/detect", bytes.NewReader(body))
	if err != nil {
		return "auto", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return "auto", err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "auto", fmt.Errorf("detect HTTP %d", resp.StatusCode)
	}

	// Response: [{"confidence": 0.95, "language": "en"}, ...]
	var detections []struct {
		Language   string  `json:"language"`
		Confidence float64 `json:"confidence"`
	}
	if err := json.Unmarshal(data, &detections); err != nil || len(detections) == 0 {
		return "auto", nil
	}
	return detections[0].Language, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (h *Handler) helpResponse(subject string) model.CommandResponse {
	return model.CommandResponse{
		ReplySubject: reply(subject),
		ReplyBody: "Uebersetzer (LibreTranslate)\n\n" +
			"Verwendung:\n" +
			"  /translate <Text>             -> nach Deutsch (Standard)\n" +
			"  /translate <Sprache> <Text>   -> in Zielsprache\n\n" +
			"Beispiele:\n" +
			"  /translate I love coding\n" +
			"  /translate en Das ist cool\n" +
			"  /translate fr Guten Morgen\n" +
			"  /translate ja Hello World\n\n" +
			"Sprachen:\n" +
			"  de  Deutsch        en  Englisch\n" +
			"  fr  Franzoesisch   es  Spanisch\n" +
			"  it  Italienisch    pt  Portugiesisch\n" +
			"  nl  Niederlaend.   pl  Polnisch\n" +
			"  ru  Russisch       uk  Ukrainisch\n" +
			"  ja  Japanisch      zh  Chinesisch\n" +
			"  ar  Arabisch       ko  Koreanisch\n" +
			"  tr  Tuerkisch      hi  Hindi\n" +
			"  sv  Schwedisch     da  Daenisch\n" +
			"  fi  Finnisch       cs  Tschechisch",
	}
}

func reply(subject string) string {
	if strings.HasPrefix(subject, "Re: ") {
		return subject
	}
	return "Re: " + subject
}
