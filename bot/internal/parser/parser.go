package parser

import (
	"strings"
)

// knownTags lists every tag the system recognises, in priority order.
// Adding a new command is a one-line change here.
var knownTags = []string{
	"#ki", "/ki",
	"#tasks", "/tasks",
	"#calendar", "/calendar",
	"#model", "/model",
	"#sudoku", "/sudoku",
	"#news", "/news",
	"#hilfe", "/hilfe",
	"#help", "/help",
	"#translate", "/translate",
	"#übersetzen", "/übersetzen",
}

// ParsedCommand holds the normalised tag and the remaining payload text.
type ParsedCommand struct {
	Tag     string // always lowercase, e.g. "#ki"
	Payload string
}

// Parse tries to find a command tag in the email subject first, then in the
// first non-empty line of the plain-text body.
func Parse(subject, body string) (*ParsedCommand, bool) {
	if cmd, ok := extractCommand(subject); ok {
		return cmd, true
	}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if cmd, ok := extractCommand(line); ok {
			return cmd, true
		}
		break // only check the very first non-empty line
	}
	return nil, false
}

// extractCommand checks whether text starts with a known tag.
func extractCommand(text string) (*ParsedCommand, bool) {
	text = strings.TrimSpace(text)
	lower := strings.ToLower(text)

	for _, tag := range knownTags {
		if lower == tag {
			return &ParsedCommand{Tag: tag, Payload: ""}, true
		}
		if strings.HasPrefix(lower, tag+" ") || strings.HasPrefix(lower, tag+"\t") {
			payload := strings.TrimSpace(text[len(tag):])
			return &ParsedCommand{Tag: tag, Payload: payload}, true
		}
	}
	return nil, false
}