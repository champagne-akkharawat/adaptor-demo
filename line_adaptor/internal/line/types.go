package line

// WebhookRequest is the top-level payload POSTed by LINE to your webhook endpoint.
type WebhookRequest struct {
	Destination string  `json:"destination"`
	Events      []Event `json:"events"`
}

// Event represents a single LINE webhook event.
type Event struct {
	Type           string  `json:"type"`
	Mode           string  `json:"mode"`
	Timestamp      int64   `json:"timestamp"`
	Source         Source  `json:"source"`
	WebhookEventID string  `json:"webhookEventId"`
	ReplyToken     string  `json:"replyToken"`
	Message        Message `json:"message"`
}

// Source identifies who triggered the event.
type Source struct {
	Type   string `json:"type"`
	UserID string `json:"userId"`
}

// Message is the message payload within a message event.
type Message struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Text       string `json:"text"`
	QuoteToken string `json:"quoteToken"`
}

// ReplyRequest is the body sent to the LINE reply endpoint.
type ReplyRequest struct {
	ReplyToken string        `json:"replyToken"`
	Messages   []TextMessage `json:"messages"`
}

// PushRequest is the body sent to the LINE push endpoint.
type PushRequest struct {
	To       string        `json:"to"`
	Messages []TextMessage `json:"messages"`
}

// TextMessage is a plain-text LINE message object.
type TextMessage struct {
	Type string `json:"type"` // always "text"
	Text string `json:"text"`
}

// NewTextMessage returns a TextMessage with type pre-filled.
func NewTextMessage(text string) TextMessage {
	return TextMessage{Type: "text", Text: text}
}
