package messages

import (
	"fmt"

	line "line-adaptor/internal/line"
)

// Text holds the parsed fields of a LINE text message.
type Text struct {
	Id              string
	QuoteToken      string
	Text            string
	Emojis          []line.Emoji
	Mention         *line.Mention
	QuotedMessageId string
}

// MessageType implements Parsed.
func (t *Text) MessageType() string { return "text" }

// ParseText parses a raw line.Message of type "text".
// Required fields: Id, Text. Returns an error if either is empty.
func ParseText(msg *line.Message) (*Text, error) {
	if msg.Id == "" {
		return nil, fmt.Errorf("messages/text: Id is required")
	}
	if msg.Text == "" {
		return nil, fmt.Errorf("messages/text: Text is required")
	}
	return &Text{
		Id:              msg.Id,
		QuoteToken:      msg.QuoteToken,
		Text:            msg.Text,
		Emojis:          msg.Emojis,
		Mention:         msg.Mention,
		QuotedMessageId: msg.QuotedMessageId,
	}, nil
}
