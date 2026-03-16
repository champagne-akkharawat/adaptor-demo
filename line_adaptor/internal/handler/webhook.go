package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"aura-wellness/line-adaptor/internal/line"
)

const eventLogDir = "event_logs"

// WebhookHandler handles inbound LINE webhook events.
type WebhookHandler struct {
	channelSecret string
	lineClient    *line.Client
}

// New creates a WebhookHandler.
func New(channelSecret string, lineClient *line.Client) *WebhookHandler {
	return &WebhookHandler{
		channelSecret: channelSecret,
		lineClient:    lineClient,
	}
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Read raw body first — needed for signature verification before parsing.
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}

	// Verify X-Line-Signature.
	sig := r.Header.Get("X-Line-Signature")
	if !line.Verify(h.channelSecret, sig, rawBody) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var req line.WebhookRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	for _, event := range req.Events {
		if err := writeEventLog(event); err != nil {
			log.Printf("write event log: %v", err)
		}
		if event.Type != "message" {
			continue
		}
		if event.Message.Type != "text" {
			if err := h.handleNonTextMessage(event); err != nil {
				log.Printf("handle non-text message: %v", err)
			}
			continue
		}
		if err := h.handleTextMessage(event); err != nil {
			log.Printf("handle text message: %v", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) handleTextMessage(event line.Event) error {
	log.Printf("message from %s: %q", event.Source.UserID, event.Message.Text)

	// TODO: replace this echo with real business logic.
	reply := line.NewTextMessage("You said: " + event.Message.Text)
	return h.lineClient.Reply(event.ReplyToken, []line.TextMessage{reply})
}

func (h *WebhookHandler) handleNonTextMessage(event line.Event) error {
	log.Printf("non-text message from %s: type=%q", event.Source.UserID, event.Message.Type)

	reply := line.NewTextMessage("I have received your " + event.Message.Type + ".")
	return h.lineClient.Reply(event.ReplyToken, []line.TextMessage{reply})
}

func writeEventLog(event line.Event) error {
	filename := fmt.Sprintf("%s/%s.json", eventLogDir, time.Now().Format("20060102_150405"))
	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return os.WriteFile(filename, data, 0o644)
}
