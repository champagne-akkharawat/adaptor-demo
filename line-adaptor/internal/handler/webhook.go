package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"line-adaptor/internal/line"
	"line-adaptor/internal/line/content"
	"line-adaptor/internal/line/messages"
	"line-adaptor/internal/logger"
)

type Handler struct {
	channelSecret string
	accessToken   string
	log           *logger.Logger
	content       *content.Client
}

func New(channelSecret, accessToken string, log *logger.Logger, content *content.Client) *Handler {
	return &Handler{
		channelSecret: channelSecret,
		accessToken:   accessToken,
		log:           log,
		content:       content,
	}
}

func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}

	signature := r.Header.Get("X-Line-Signature")
	if !line.Verify(h.channelSecret, body, signature) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var payload line.WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "failed to parse body", http.StatusInternalServerError)
		return
	}

	parsedJSON, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to marshal parsed body", http.StatusInternalServerError)
		return
	}

	if err := h.log.LogWebhookEvent(body, parsedJSON); err != nil {
		log.Printf("failed to log webhook event: %v", err)
	}

	for _, event := range payload.Events {
		if event.Type == "message" && event.Message != nil {
			parsed, err := messages.Route(event.Message)
			if err != nil {
				log.Printf("unrecognised message type: %v", err)
			} else {
				log.Printf("message type: %s", parsed.MessageType())
			}
		}

		if event.ReplyToken != "" {
			if err := line.Reply(h.accessToken, event.ReplyToken); err != nil {
				log.Printf("failed to send reply: %v", err)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}
