package messages

import (
	"fmt"

	line "line-adaptor/internal/line"
)

// Audio holds the parsed fields of a LINE audio message.
type Audio struct {
	Id                string
	Duration          int64 // milliseconds, 0 if absent
	ContentProvider   line.ContentProvider
	NeedsContentFetch bool
}

// MessageType implements Parsed.
func (a *Audio) MessageType() string { return "audio" }

// ParseAudio parses a raw line.Message of type "audio".
// Required fields: Id, ContentProvider (non-nil). Duration is optional (0 if absent).
// Note: audio messages cannot be quoted, so there is no QuoteToken field.
func ParseAudio(msg *line.Message) (*Audio, error) {
	if msg.Id == "" {
		return nil, fmt.Errorf("messages/audio: Id is required")
	}
	if msg.ContentProvider == nil {
		return nil, fmt.Errorf("messages/audio: ContentProvider is required")
	}
	return &Audio{
		Id:                msg.Id,
		Duration:          msg.Duration,
		ContentProvider:   *msg.ContentProvider,
		NeedsContentFetch: msg.ContentProvider.Type == "line",
	}, nil
}
