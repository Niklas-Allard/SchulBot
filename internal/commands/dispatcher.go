package commands

import (
	"context"
	"fmt"

	"schulbot/internal/model"
)

// Handler is the interface every command must implement.
type Handler interface {
	// CanHandle reports whether this handler is responsible for the given tag.
	CanHandle(tag string) bool
	// Handle executes the command and returns a reply.
	Handle(ctx context.Context, req model.CommandRequest) (model.CommandResponse, error)
}

// Dispatcher routes incoming requests to the first matching handler.
type Dispatcher struct {
	handlers []Handler
}

func NewDispatcher(handlers ...Handler) *Dispatcher {
	return &Dispatcher{handlers: handlers}
}

// Dispatch finds the right handler and calls it. Returns an error if no
// handler claims the tag.
func (d *Dispatcher) Dispatch(ctx context.Context, req model.CommandRequest) (model.CommandResponse, error) {
	for _, h := range d.handlers {
		if h.CanHandle(req.Tag) {
			return h.Handle(ctx, req)
		}
	}
	return model.CommandResponse{}, fmt.Errorf("no handler registered for tag %q", req.Tag)
}