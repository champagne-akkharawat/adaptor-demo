package messages

import (
	"fmt"

	line "line-adaptor/internal/line"
)

// Video holds the parsed fields of a LINE video message.
type Video struct {
	Id                string
	QuoteToken        string
	Duration          int64 // milliseconds, 0 if absent
	ContentProvider   line.ContentProvider
	NeedsContentFetch bool
}

// MessageType implements Parsed.
func (v *Video) MessageType() string { return "video" }

// ParseVideo parses a raw line.Message of type "video".
// Required fields: Id, ContentProvider (non-nil). Duration is optional (0 if absent).
func ParseVideo(msg *line.Message) (*Video, error) {
	if msg.Id == "" {
		return nil, fmt.Errorf("messages/video: Id is required")
	}
	if msg.ContentProvider == nil {
		return nil, fmt.Errorf("messages/video: ContentProvider is required")
	}
	return &Video{
		Id:                msg.Id,
		QuoteToken:        msg.QuoteToken,
		Duration:          msg.Duration,
		ContentProvider:   *msg.ContentProvider,
		NeedsContentFetch: msg.ContentProvider.Type == "line",
	}, nil
}
