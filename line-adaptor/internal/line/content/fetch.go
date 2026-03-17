package content

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Fetch downloads the full content of a LINE-hosted message.
// The caller is responsible for closing the returned ReadCloser.
// Use for: image, video, audio, file — when contentProvider.type == "line".
func (c *Client) Fetch(ctx context.Context, messageId string) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/%s/content", BaseURL, messageId)
	return c.doGet(ctx, url)
}

// FetchPreview downloads the preview/thumbnail of a LINE-hosted message.
// The caller is responsible for closing the returned ReadCloser.
// Use for: image, video only.
func (c *Client) FetchPreview(ctx context.Context, messageId string) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/%s/content/preview", BaseURL, messageId)
	return c.doGet(ctx, url)
}

func (c *Client) doGet(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("content: unexpected status %d", resp.StatusCode)
	}

	return resp.Body, nil
}
