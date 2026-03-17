package messages

import (
	"fmt"

	line "line-adaptor/internal/line"
)

// Image holds the parsed fields of a LINE image message.
type Image struct {
	Id                string
	QuoteToken        string
	ContentProvider   line.ContentProvider
	ImageSet          *line.ImageSet // nil if not present
	NeedsContentFetch bool           // true when ContentProvider.Type == "line"
}

// MessageType implements Parsed.
func (i *Image) MessageType() string { return "image" }

// ParseImage parses a raw line.Message of type "image".
// Required fields: Id, ContentProvider (non-nil). Returns an error if either is missing.
func ParseImage(msg *line.Message) (*Image, error) {
	if msg.Id == "" {
		return nil, fmt.Errorf("messages/image: Id is required")
	}
	if msg.ContentProvider == nil {
		return nil, fmt.Errorf("messages/image: ContentProvider is required")
	}
	return &Image{
		Id:                msg.Id,
		QuoteToken:        msg.QuoteToken,
		ContentProvider:   *msg.ContentProvider,
		ImageSet:          msg.ImageSet,
		NeedsContentFetch: msg.ContentProvider.Type == "line",
	}, nil
}
