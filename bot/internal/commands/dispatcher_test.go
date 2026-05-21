package commands

import (
	"context"
	"testing"

	"schulbot/internal/model"
)

type stubHandler struct {
	tag   string
	reply string
}

func (h *stubHandler) CanHandle(tag string) bool { return tag == h.tag }
func (h *stubHandler) Handle(_ context.Context, req model.CommandRequest) (model.CommandResponse, error) {
	return model.CommandResponse{ReplySubject: "Re: " + req.Subject, ReplyBody: h.reply}, nil
}

func TestDispatchKnownTag(t *testing.T) {
	d := NewDispatcher(&stubHandler{tag: "#ki", reply: "Antwort"})
	resp, err := d.Dispatch(context.Background(), model.CommandRequest{Tag: "#ki", Subject: "Test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ReplyBody != "Antwort" {
		t.Errorf("unexpected body: %q", resp.ReplyBody)
	}
}

func TestDispatchUnknownTag(t *testing.T) {
	d := NewDispatcher(&stubHandler{tag: "#ki", reply: "x"})
	_, err := d.Dispatch(context.Background(), model.CommandRequest{Tag: "#unknown"})
	if err == nil {
		t.Error("expected error for unknown tag")
	}
}

func TestDispatchFirstMatchWins(t *testing.T) {
	d := NewDispatcher(
		&stubHandler{tag: "#ki", reply: "first"},
		&stubHandler{tag: "#ki", reply: "second"},
	)
	resp, err := d.Dispatch(context.Background(), model.CommandRequest{Tag: "#ki"})
	if err != nil || resp.ReplyBody != "first" {
		t.Errorf("expected first handler, got body=%q err=%v", resp.ReplyBody, err)
	}
}