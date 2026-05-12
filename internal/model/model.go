package model

import "time"

// CommandRequest carries everything a handler needs to process a command.
type CommandRequest struct {
	MessageID  string
	From       string
	Subject    string
	Tag        string
	Payload    string
	RawText    string
	ReceivedAt time.Time
}

// Attachment holds a file to be sent as email attachment.
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// CommandResponse is what a handler returns for the outgoing reply.
type CommandResponse struct {
	ReplySubject string
	ReplyBody    string
	Attachments  []Attachment // nil for text-only replies
}
