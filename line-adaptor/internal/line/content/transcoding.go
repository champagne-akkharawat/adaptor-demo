package content

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// TranscodingStatus represents the readiness of a video on LINE servers.
type TranscodingStatus string

const (
	TranscodingProcessing TranscodingStatus = "processing"
	TranscodingSucceeded  TranscodingStatus = "succeeded"
	TranscodingFailed     TranscodingStatus = "failed"
)

// CheckTranscoding checks whether a video message has finished processing on LINE servers.
// Call before Fetch for video messages to avoid downloading an incomplete file.
func (c *Client) CheckTranscoding(ctx context.Context, messageId string) (TranscodingStatus, error) {
	url := fmt.Sprintf("%s/%s/content/transcoding", BaseURL, messageId)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("content: unexpected status %d", resp.StatusCode)
	}

	var payload struct {
		Status TranscodingStatus `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("content: failed to decode transcoding response: %w", err)
	}

	if payload.Status == "" {
		return "", fmt.Errorf("content: transcoding status field is missing or empty")
	}

	return payload.Status, nil
}
