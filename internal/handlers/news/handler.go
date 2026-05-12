package news

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"schulbot/internal/model"
)

const maxItems = 8 // max news items per response

// Handler implements the /news command using the public Tagesschau API.
// Rate limit: 60 requests/hour — the bot's usage is well within this.
type Handler struct {
	api *apiClient
}

func New() *Handler {
	return &Handler{api: newAPIClient()}
}

func (h *Handler) CanHandle(tag string) bool {
	return tag == "#news" || tag == "/news"
}

func (h *Handler) Handle(ctx context.Context, req model.CommandRequest) (model.CommandResponse, error) {
	payload := strings.TrimSpace(req.Payload)
	lower := strings.ToLower(payload)

	// /news suche <query>
	if strings.HasPrefix(lower, "suche ") || strings.HasPrefix(lower, "search ") {
		query := strings.TrimSpace(payload[strings.Index(payload, " ")+1:])
		return h.handleSearch(ctx, req.Subject, query)
	}

	// /news region <name|nr>
	if strings.HasPrefix(lower, "region ") {
		regionArg := strings.TrimSpace(lower[7:])
		return h.handleRegion(ctx, req.Subject, regionArg)
	}

	// /news <ressort> or /news (top news)
	if payload == "" || lower == "top" || lower == "aktuell" {
		return h.handleHomepage(ctx, req.Subject)
	}

	// Check if it's a known ressort
	if apiRessort, ok := Ressorts[lower]; ok {
		return h.handleRessort(ctx, req.Subject, apiRessort)
	}

	return h.helpResponse(req.Subject), nil
}

func (h *Handler) handleHomepage(ctx context.Context, subject string) (model.CommandResponse, error) {
	items, err := h.api.homepage(ctx)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("news: homepage: %w", err)
	}
	body := formatNewsItems("Tagesschau – Top-Nachrichten", items)
	return model.CommandResponse{
		ReplySubject: reply(subject),
		ReplyBody:    body,
	}, nil
}

func (h *Handler) handleRessort(ctx context.Context, subject, ressort string) (model.CommandResponse, error) {
	items, err := h.api.news(ctx, ressort, "")
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("news: ressort %s: %w", ressort, err)
	}
	label := strings.ToUpper(ressort[:1]) + ressort[1:]
	body := formatNewsItems("Tagesschau – "+label, items)
	return model.CommandResponse{
		ReplySubject: reply(subject),
		ReplyBody:    body,
	}, nil
}

func (h *Handler) handleRegion(ctx context.Context, subject, regionArg string) (model.CommandResponse, error) {
	regionID := resolveRegion(regionArg)
	if regionID == "" {
		return model.CommandResponse{
			ReplySubject: reply(subject),
			ReplyBody:    "Unbekannte Region. Beispiele:\n  /news region nrw\n  /news region berlin\n  /news region 2 (Bayern)\n\nBundesländer-IDs: 1=BW, 2=BY, 3=BE, 4=BB, 5=HB, 6=HH, 7=HE,\n8=MV, 9=NI, 10=NW, 11=RP, 12=SL, 13=SN, 14=ST, 15=SH, 16=TH",
		}, nil
	}

	items, err := h.api.news(ctx, "", regionID)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("news: region %s: %w", regionID, err)
	}
	label := regionName(regionID)
	body := formatNewsItems("Tagesschau – "+label, items)
	return model.CommandResponse{
		ReplySubject: reply(subject),
		ReplyBody:    body,
	}, nil
}

func (h *Handler) handleSearch(ctx context.Context, subject, query string) (model.CommandResponse, error) {
	if query == "" {
		return model.CommandResponse{
			ReplySubject: reply(subject),
			ReplyBody:    "Bitte Suchbegriff angeben: /news suche <Begriff>",
		}, nil
	}

	items, err := h.api.search(ctx, query)
	if err != nil {
		return model.CommandResponse{}, fmt.Errorf("news: search %q: %w", query, err)
	}

	if len(items) == 0 {
		return model.CommandResponse{
			ReplySubject: reply(subject),
			ReplyBody:    fmt.Sprintf("Keine Ergebnisse fuer \"%s\".", query),
		}, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Tagesschau - Suche: \"%s\"\nStand: %s\n\n", query, now())

	limit := min(maxItems, len(items))
	for i := range limit {
		it := items[i]
		fmt.Fprintf(&sb, "%d. %s\n", i+1, it.Title)
		if it.TeaserText != "" {
			fmt.Fprintf(&sb, "   %s\n", truncate(it.TeaserText, 120))
		}
		if it.ShareURL != "" {
			fmt.Fprintf(&sb, "   %s\n", it.ShareURL)
		}
		sb.WriteString("\n")
	}
	return model.CommandResponse{
		ReplySubject: reply(subject),
		ReplyBody:    strings.TrimRight(sb.String(), "\n"),
	}, nil
}

func (h *Handler) helpResponse(subject string) model.CommandResponse {
	return model.CommandResponse{
		ReplySubject: reply(subject),
		ReplyBody: "Tagesschau-Befehle:\n\n" +
			"  /news                  → Top-Nachrichten\n" +
			"  /news inland           → Deutschland\n" +
			"  /news ausland          → International\n" +
			"  /news sport            → Sport\n" +
			"  /news wirtschaft       → Wirtschaft\n" +
			"  /news wissenschaft     → Wissenschaft & Tech\n" +
			"  /news gesundheit       → Gesundheit\n" +
			"  /news faktenfinder     → Faktenfinder\n" +
			"  /news region nrw       → Regionales (Bundesland-Name oder Nr. 1–16)\n" +
			"  /news suche <Begriff>  → Volltextsuche\n\n" +
			"Quelle: tagesschau.de · nur für privaten, nicht-kommerziellen Gebrauch.",
	}
}

// ── Formatting helpers ────────────────────────────────────────────────────────

func formatNewsItems(title string, items []newsItem) string {
	if len(items) == 0 {
		return "Aktuell keine Nachrichten verfügbar."
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\nStand: %s\n\n", title, now())

	limit := min(maxItems, len(items))
	for i := range limit {
		it := items[i]
		topline := it.Topline
		heading := it.Title
		teaser := it.TeaserText
		if teaser == "" {
			teaser = it.FirstSentence
		}

		if topline != "" {
			fmt.Fprintf(&sb, "%d. [%s] %s\n", i+1, topline, heading)
		} else {
			fmt.Fprintf(&sb, "%d. %s\n", i+1, heading)
		}
		if teaser != "" {
			fmt.Fprintf(&sb, "   %s\n", truncate(teaser, 130))
		}
		if it.ShareURL != "" {
			fmt.Fprintf(&sb, "   %s\n", it.ShareURL)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("─\nQuelle: tagesschau.de · privater Gebrauch")
	return strings.TrimRight(sb.String(), "\n")
}

func truncate(s string, n int) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, s)
	s = strings.Join(strings.Fields(s), " ")
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}

func now() string {
	loc, _ := time.LoadLocation("Europe/Berlin")
	return time.Now().In(loc).Format("02.01.2006 15:04")
}

func resolveRegion(arg string) string {
	// Direct numeric ID
	if len(arg) <= 2 && isNumeric(arg) {
		n := arg
		if n >= "1" && n <= "16" {
			return n
		}
	}
	// Named lookup
	if id, ok := Regions[arg]; ok {
		return id
	}
	return ""
}

func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

var regionNames = map[string]string{
	"1": "Baden-Württemberg", "2": "Bayern", "3": "Berlin",
	"4": "Brandenburg", "5": "Bremen", "6": "Hamburg",
	"7": "Hessen", "8": "Mecklenburg-Vorpommern", "9": "Niedersachsen",
	"10": "Nordrhein-Westfalen", "11": "Rheinland-Pfalz", "12": "Saarland",
	"13": "Sachsen", "14": "Sachsen-Anhalt", "15": "Schleswig-Holstein",
	"16": "Thüringen",
}

func regionName(id string) string {
	if n, ok := regionNames[id]; ok {
		return n
	}
	return "Region " + id
}

func reply(subject string) string {
	if strings.HasPrefix(subject, "Re: ") {
		return subject
	}
	return "Re: " + subject
}

