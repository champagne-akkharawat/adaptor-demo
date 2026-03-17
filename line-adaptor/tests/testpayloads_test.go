package handler_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"line-adaptor/internal/handler"
	"line-adaptor/internal/line"
	"line-adaptor/internal/line/content"
	"line-adaptor/internal/logger"
)

// TestWebhook_RealPayloads replays every file under ../test_payloads/ through
// the webhook handler.  Because the files are real LINE payloads we don't have
// the original channel secret, so we re-sign each body with the test secret
// before sending.  The reply API is mocked so no real HTTP calls escape.
func TestWebhook_RealPayloads(t *testing.T) {
	// Mock the LINE reply API so the handler doesn't call out to LINE.
	mockReply := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockReply.Close()

	original := line.ReplyAPIURL
	line.ReplyAPIURL = mockReply.URL
	t.Cleanup(func() { line.ReplyAPIURL = original })

	// Collect payload files.
	entries, err := os.ReadDir("../test_payloads")
	if err != nil {
		t.Fatalf("read test_payloads dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no files found in test_payloads/")
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		tc := entry.Name()
		t.Run(tc, func(t *testing.T) {
			body, err := os.ReadFile(filepath.Join("../test_payloads", tc))
			if err != nil {
				t.Fatalf("read payload: %v", err)
			}

			logDir := t.TempDir()
			h := handler.New("test-secret", "test-token", logger.New(logDir), content.New("test-token"))
			srv := httptest.NewServer(http.HandlerFunc(h.Webhook))
			defer srv.Close()

			req, err := http.NewRequest(http.MethodPost, srv.URL+"/webhook", strings.NewReader(string(body)))
			if err != nil {
				t.Fatalf("new request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Line-Signature", makeSignature("test-secret", body))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("do request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected 200, got %d", resp.StatusCode)
			}
		})
	}
}
