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
	ReplyToken      string          `json:"replyToken,omitempty"`
	Message         *Message        `json:"message,omitempty"`
	Postback        *Postback       `json:"postback,omitempty"`
}

type DeliveryContext struct {
	IsRedelivery bool `json:"isRedelivery"`
}

type Source struct {
	Type    string `json:"type"`
	UserId  string `json:"userId,omitempty"`
	GroupId string `json:"groupId,omitempty"`
	RoomId  string `json:"roomId,omitempty"`
}

type Message struct {
	Type      string  `json:"type"`
	Id        string  `json:"id"`
	Text      string  `json:"text,omitempty"`
	Title     string  `json:"title,omitempty"`
	Address   string  `json:"address,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	PackageId string  `json:"packageId,omitempty"`
	StickerId string  `json:"stickerId,omitempty"`
	FileName  string  `json:"fileName,omitempty"`
	FileSize  int64   `json:"fileSize,omitempty"`
}

type Postback struct {
	Data   string            `json:"data"`
	Params map[string]string `json:"params,omitempty"`
}
