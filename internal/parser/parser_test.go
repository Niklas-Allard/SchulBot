package parser

import "testing"

func TestParseSubject(t *testing.T) {
	cases := []struct {
		subject string
		tag     string
		payload string
		ok      bool
	}{
		{"#ki Was ist Photosynthese?", "#ki", "Was ist Photosynthese?", true},
		{"/ki Erkläre Quantenmechanik", "/ki", "Erkläre Quantenmechanik", true},
		{"#KI Großschreibung", "#ki", "Großschreibung", true},
		{"#ki", "#ki", "", true},
		{"Hallo Welt", "", "", false},
		{"", "", "", false},
		{"#tasks add Mathe-Hausaufgaben", "#tasks", "add Mathe-Hausaufgaben", true},
		{"/calendar morgen 9 Uhr", "/calendar", "morgen 9 Uhr", true},
	}

	for _, tc := range cases {
		cmd, ok := Parse(tc.subject, "")
		if ok != tc.ok {
			t.Errorf("subject=%q: ok got %v want %v", tc.subject, ok, tc.ok)
			continue
		}
		if !ok {
			continue
		}
		if cmd.Tag != tc.tag {
			t.Errorf("subject=%q: tag got %q want %q", tc.subject, cmd.Tag, tc.tag)
		}
		if cmd.Payload != tc.payload {
			t.Errorf("subject=%q: payload got %q want %q", tc.subject, cmd.Payload, tc.payload)
		}
	}
}

func TestParseBodyFallback(t *testing.T) {
	cmd, ok := Parse("Hallo", "\n\n#ki Wie groß ist die Sonne?\nRest wird ignoriert")
	if !ok {
		t.Fatal("expected ok")
	}
	if cmd.Tag != "#ki" || cmd.Payload != "Wie groß ist die Sonne?" {
		t.Errorf("got tag=%q payload=%q", cmd.Tag, cmd.Payload)
	}
}

func TestParseBodyOnlyFirstLine(t *testing.T) {
	_, ok := Parse("Kein Tag", "Normal text\n#ki hier nicht mehr")
	if ok {
		t.Error("should not match tag on second body line")
	}
}

func TestParseNoMatch(t *testing.T) {
	_, ok := Parse("Frage", "Wie geht es dir?")
	if ok {
		t.Error("expected no match")
	}
}

func TestParseEmptyPayload(t *testing.T) {
	cmd, ok := Parse("#ki", "")
	if !ok || cmd.Payload != "" {
		t.Errorf("expected empty payload, got %q", cmd.Payload)
	}
}

func TestParseSubjectTakesPriority(t *testing.T) {
	cmd, ok := Parse("#ki Frage aus Betreff", "#tasks Frage aus Body")
	if !ok || cmd.Tag != "#ki" {
		t.Error("subject should take priority over body")
	}
}