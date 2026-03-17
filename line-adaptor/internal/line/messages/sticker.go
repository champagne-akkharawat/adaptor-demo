package messages

import (
	"fmt"

	line "line-adaptor/internal/line"
)

// Sticker holds the parsed fields of a LINE sticker message.
type Sticker struct {
	Id                  string
	QuoteToken          string
	PackageId           string
	StickerId           string
	StickerResourceType string
	Keywords            []string // nil if absent
	Text                string   // only for MESSAGE/PER_STICKER_TEXT resource types
	QuotedMessageId     string   // optional
}

// MessageType implements Parsed.
func (s *Sticker) MessageType() string { return "sticker" }

// ParseSticker parses a raw line.Message of type "sticker".
// Required fields: Id, PackageId, StickerId, StickerResourceType.
// Returns an error if any required field is empty.
func ParseSticker(msg *line.Message) (*Sticker, error) {
	if msg.Id == "" {
		return nil, fmt.Errorf("messages/sticker: Id is required")
	}
	if msg.PackageId == "" {
		return nil, fmt.Errorf("messages/sticker: PackageId is required")
	}
	if msg.StickerId == "" {
		return nil, fmt.Errorf("messages/sticker: StickerId is required")
	}
	if msg.StickerResourceType == "" {
		return nil, fmt.Errorf("messages/sticker: StickerResourceType is required")
	}
	return &Sticker{
		Id:                  msg.Id,
		QuoteToken:          msg.QuoteToken,
		PackageId:           msg.PackageId,
		StickerId:           msg.StickerId,
		StickerResourceType: msg.StickerResourceType,
		Keywords:            msg.Keywords,
		Text:                msg.Text,
		QuotedMessageId:     msg.QuotedMessageId,
	}, nil
}
