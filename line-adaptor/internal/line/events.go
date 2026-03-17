package line

type WebhookPayload struct {
	Destination string  `json:"destination"`
	Events      []Event `json:"events"`
}

type Event struct {
	Type            string          `json:"type"`
	Mode            string          `json:"mode"`
	Timestamp       int64           `json:"timestamp"`
	WebhookEventId  string          `json:"webhookEventId"`
	DeliveryContext DeliveryContext `json:"deliveryContext"`
	Source          Source          `json:"source"`
	ReplyToken      string          `json:"replyToken,omitempty"`  // Present only for events that support replies (e.g. message, postback, beacon)
	Message         *Message        `json:"message,omitempty"`     // Present only when Type is "message"
	Postback        *Postback       `json:"postback,omitempty"`    // Present only when Type is "postback"
}

type DeliveryContext struct {
	IsRedelivery bool `json:"isRedelivery"`
}

type Source struct {
	Type    string `json:"type"`
	UserId  string `json:"userId,omitempty"`  // Present when source type is "user", "group", or "room"
	GroupId string `json:"groupId,omitempty"` // Present only when source type is "group"
	RoomId  string `json:"roomId,omitempty"`  // Present only when source type is "room"
}

type Message struct {
	// Always present
	Type string `json:"type"`
	Id   string `json:"id"`

	// text, image, video, sticker
	QuoteToken string `json:"quoteToken,omitempty"`

	// text only
	Text    string  `json:"text,omitempty"`
	Emojis  []Emoji `json:"emojis,omitempty"`
	Mention *Mention `json:"mention,omitempty"`

	// text, sticker
	QuotedMessageId string `json:"quotedMessageId,omitempty"`

	// image, video, audio
	ContentProvider *ContentProvider `json:"contentProvider,omitempty"`

	// image only
	ImageSet *ImageSet `json:"imageSet,omitempty"`

	// video, audio (milliseconds)
	Duration int64 `json:"duration,omitempty"`

	// location only
	Title     string  `json:"title,omitempty"`
	Address   string  `json:"address,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`

	// sticker only
	PackageId           string   `json:"packageId,omitempty"`
	StickerId           string   `json:"stickerId,omitempty"`
	StickerResourceType string   `json:"stickerResourceType,omitempty"`
	Keywords            []string `json:"keywords,omitempty"`

	// file only
	FileName string `json:"fileName,omitempty"`
	FileSize int64  `json:"fileSize,omitempty"`
}

type ContentProvider struct {
	Type               string `json:"type"`                         // "line" or "external"
	OriginalContentUrl string `json:"originalContentUrl,omitempty"`
	PreviewImageUrl    string `json:"previewImageUrl,omitempty"`
}

type ImageSet struct {
	Id    string `json:"id"`
	Index int    `json:"index"`
	Total int    `json:"total"`
}

type Emoji struct {
	Index     int    `json:"index"`
	Length    int    `json:"length"`
	ProductId string `json:"productId"`
	EmojiId   string `json:"emojiId"`
}

type Mention struct {
	Mentionees []Mentionee `json:"mentionees"`
}

type Mentionee struct {
	Index  int    `json:"index"`
	Length int    `json:"length"`
	Type   string `json:"type"`             // "user" or "all"
	UserId string `json:"userId,omitempty"`
	IsSelf bool   `json:"isSelf,omitempty"`
}

type Postback struct {
	Data   string            `json:"data"`
	Params map[string]string `json:"params,omitempty"` // Present only when the postback originates from a date/time picker action
}
