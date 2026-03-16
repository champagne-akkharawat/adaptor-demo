package line

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	replyEndpoint = "https://api.line.me/v2/bot/message/reply"
	pushEndpoint  = "https://api.line.me/v2/bot/message/push"
)

// Client sends messages to the LINE Messaging API.
type Client struct {
	channelAccessToken string
	httpClient         *http.Client
}

// NewClient creates a Client using the given channel access token.
func NewClient(channelAccessToken string) *Client {
	return &Client{
		channelAccessToken: channelAccessToken,
		httpClient:         &http.Client{},
	}
}

// Reply sends messages back using the replyToken from a webhook event.
// The replyToken is single-use and expires 30 seconds after the event.
func (c *Client) Reply(replyToken string, messages []TextMessage) error {
	payload := ReplyRequest{
		ReplyToken: replyToken,
		Messages:   messages,
	}
	return c.post(replyEndpoint, payload)
}

// Push sends proactive messages to a user by their userId.
// Requires a LINE plan that supports push messages.
func (c *Client) Push(to string, messages []TextMessage) error {
	payload := PushRequest{
		To:       to,
		Messages: messages,
	}
	return c.post(pushEndpoint, payload)
}

func (c *Client) post(url string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("line client: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("line client: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.channelAccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("line client: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("line client: unexpected status %d: %s", resp.StatusCode, respBody)
	}
	return nil
}
