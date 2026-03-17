package messages

import (
	"fmt"

	line "line-adaptor/internal/line"
)

// Parsed is the common interface for all typed message results.
type Parsed interface {
	MessageType() string
}

// Route dispatches msg to the correct typed parser.
// Returns an error if msg is nil, msg.Type is unknown, or the parser fails.
func Route(msg *line.Message) (Parsed, error) {
	if msg == nil {
		return nil, fmt.Errorf("messages: message is nil")
	}

	switch msg.Type {
	case "text":
		return ParseText(msg)
	case "image":
		return ParseImage(msg)
	case "video":
		return ParseVideo(msg)
	case "audio":
		return ParseAudio(msg)
	case "file":
		return ParseFile(msg)
	case "location":
		return ParseLocation(msg)
	case "sticker":
		return ParseSticker(msg)
	default:
		return nil, fmt.Errorf("messages: unknown message type %q", msg.Type)
	}
}
