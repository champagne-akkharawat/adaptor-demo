package line

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ReplyAPIURL is the LINE reply API endpoint. Override in tests to point at a mock server.
var ReplyAPIURL = "https://api.line.me/v2/bot/message/reply"

type replyRequest struct {
	ReplyToken string         `json:"replyToken"`
	Messages   []replyMessage `json:"messages"`
}

type replyMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func Reply(accessToken, replyToken string) error {
	payload := replyRequest{
		ReplyToken: replyToken,
		Messages: []replyMessage{
			{Type: "text", Text: "received"},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, ReplyAPIURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("reply API returned status %d: %s", resp.StatusCode, respBody)
	}

	return nil
}
