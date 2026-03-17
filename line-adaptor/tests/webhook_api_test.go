package handler_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"line-adaptor/internal/handler"
	"line-adaptor/internal/line"
	"line-adaptor/internal/line/content"
	"line-adaptor/internal/logger"
)

// makeSignature computes the X-Line-Signature for a request body using HMAC-SHA256 + base64.
func makeSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// logFilesExist returns true when at least one file is present in both the
// raw/ and parsed/ webhook-events subdirectories under logDir.
func logFilesExist(t *testing.T, logDir string) bool {
	t.Helper()
	rawDir := filepath.Join(logDir, "webhook-events", "raw")
	parsedDir := filepath.Join(logDir, "webhook-events", "parsed")

	rawEntries, err := os.ReadDir(rawDir)
	if err != nil || len(rawEntries) == 0 {
		return false
	}
	parsedEntries, err := os.ReadDir(parsedDir)
	if err != nil || len(parsedEntries) == 0 {
		return false
	}
	return true
}

// TestWebhook_HappyPath sends a valid signed request containing one text message
// event with a replyToken. The handler must respond 200, write log files, and
// call the reply API exactly once.
func TestWebhook_HappyPath(t *testing.T) {
	// Start a mock LINE reply server.
	var replyCalls atomic.Int32
	mockReply := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		replyCalls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockReply.Close()

	// Redirect the reply package to the mock server and restore after the test.
	original := line.ReplyAPIURL
	line.ReplyAPIURL = mockReply.URL
	t.Cleanup(func() { line.ReplyAPIURL = original })

	logDir := t.TempDir()
	h := handler.New("test-secret", "test-token", logger.New(logDir), content.New("test-token"))
	srv := httptest.NewServer(http.HandlerFunc(h.Webhook))
	defer srv.Close()

	payload := map[string]interface{}{
		"destination": "Udeadbeefdeadbeefdeadbeefdeadbeef",
		"events": []map[string]interface{}{
			{
				"type":           "message",
				"mode":           "active",
				"timestamp":      1625665242211,
				"webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
				"deliveryContext": map[string]bool{"isRedelivery": false},
				"source":         map[string]string{"type": "user", "userId": "Udeadbeef"},
				"replyToken":     "nHuyWiB7yP5Zw52FIkcQobQuGDXCTA",
				"message": map[string]string{
					"type": "text",
					"id":   "100001",
					"text": "Hello",
				},
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	sig := makeSignature("test-secret", body)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/webhook", strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Line-Signature", sig)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !logFilesExist(t, logDir) {
		t.Error("expected log files in raw/ and parsed/ but found none")
	}
	if got := replyCalls.Load(); got != 1 {
		t.Errorf("expected reply API to be called once, called %d time(s)", got)
	}
}

// TestWebhook_InvalidSignature sends a request with a tampered signature.
// The handler must respond 401 and must not create any log files.
func TestWebhook_InvalidSignature(t *testing.T) {
	logDir := t.TempDir()
	h := handler.New("test-secret", "test-token", logger.New(logDir), content.New("test-token"))
	srv := httptest.NewServer(http.HandlerFunc(h.Webhook))
	defer srv.Close()

	payload := map[string]interface{}{
		"destination": "Udeadbeefdeadbeefdeadbeefdeadbeef",
		"events":      []interface{}{},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/webhook", strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Line-Signature", "invalidsignature==")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	if logFilesExist(t, logDir) {
		t.Error("expected no log files after rejected request, but files were found")
	}
}

// TestWebhook_EmptyEvents sends a valid signed request with an empty events
// array. The handler must respond 200, write log files, and never call the
// reply API (there is no replyToken to act on).
func TestWebhook_EmptyEvents(t *testing.T) {
	logDir := t.TempDir()
	h := handler.New("test-secret", "test-token", logger.New(logDir), content.New("test-token"))
	srv := httptest.NewServer(http.HandlerFunc(h.Webhook))
	defer srv.Close()

	payload := map[string]interface{}{
		"destination": "Udeadbeefdeadbeefdeadbeefdeadbeef",
		"events":      []interface{}{},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	sig := makeSignature("test-secret", body)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/webhook", strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Line-Signature", sig)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !logFilesExist(t, logDir) {
		t.Error("expected log files in raw/ and parsed/ but found none")
	}
}

// TestWebhook_NoReplyToken sends a valid signed unfollow event that carries no
// replyToken. The handler must respond 200, write log files, and must not call
// the reply API.
func TestWebhook_NoReplyToken(t *testing.T) {
	var replyCalls atomic.Int32
	mockReply := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		replyCalls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockReply.Close()

	original := line.ReplyAPIURL
	line.ReplyAPIURL = mockReply.URL
	t.Cleanup(func() { line.ReplyAPIURL = original })

	logDir := t.TempDir()
	h := handler.New("test-secret", "test-token", logger.New(logDir), content.New("test-token"))
	srv := httptest.NewServer(http.HandlerFunc(h.Webhook))
	defer srv.Close()

	payload := map[string]interface{}{
		"destination": "Udeadbeefdeadbeefdeadbeefdeadbeef",
		"events": []map[string]interface{}{
			{
				"type":           "unfollow",
				"mode":           "active",
				"timestamp":      1625665242211,
				"webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZS",
				"deliveryContext": map[string]bool{"isRedelivery": false},
				"source":         map[string]string{"type": "user", "userId": "Udeadbeef"},
				// replyToken intentionally absent
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	sig := makeSignature("test-secret", body)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/webhook", strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Line-Signature", sig)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !logFilesExist(t, logDir) {
		t.Error("expected log files in raw/ and parsed/ but found none")
	}
	if got := replyCalls.Load(); got != 0 {
		t.Errorf("expected reply API to never be called, called %d time(s)", got)
	}
}

// TestWebhook_MessageTypes checks that the handler returns 200 for every
// supported message type.
func TestWebhook_MessageTypes(t *testing.T) {
	cases := []struct {
		name    string
		message map[string]interface{}
	}{
		{
			name: "text",
			message: map[string]interface{}{
				"type": "text",
				"id":   "200001",
				"text": "Hello",
			},
		},
		{
			name: "image",
			message: map[string]interface{}{
				"type":            "image",
				"id":              "200002",
				"contentProvider": map[string]string{"type": "line"},
			},
		},
		{
			name: "video",
			message: map[string]interface{}{
				"type":            "video",
				"id":              "200003",
				"contentProvider": map[string]string{"type": "line"},
			},
		},
		{
			name: "audio",
			message: map[string]interface{}{
				"type":            "audio",
				"id":              "200004",
				"contentProvider": map[string]string{"type": "line"},
			},
		},
		{
			name: "file",
			message: map[string]interface{}{
				"type":     "file",
				"id":       "200005",
				"fileName": "report.pdf",
				"fileSize": 1024,
			},
		},
		{
			name: "location",
			message: map[string]interface{}{
				"type":      "location",
				"id":        "200006",
				"latitude":  35.65910807942215,
				"longitude": 139.70372892916718,
			},
		},
		{
			name: "sticker",
			message: map[string]interface{}{
				"type":                "sticker",
				"id":                  "200007",
				"packageId":           "446",
				"stickerId":           "1988",
				"stickerResourceType": "STATIC",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logDir := t.TempDir()
			h := handler.New("test-secret", "test-token", logger.New(logDir), content.New("test-token"))
			srv := httptest.NewServer(http.HandlerFunc(h.Webhook))
			defer srv.Close()

			payload := map[string]interface{}{
				"destination": "Udeadbeefdeadbeefdeadbeefdeadbeef",
				"events": []map[string]interface{}{
					{
						"type":           "message",
						"mode":           "active",
						"timestamp":      1625665242211,
						"webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
						"deliveryContext": map[string]bool{"isRedelivery": false},
						"source":         map[string]string{"type": "user", "userId": "Udeadbeef"},
						"message":        tc.message,
					},
				},
			}
			body, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("marshal payload: %v", err)
			}

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

// TestWebhook_UnknownMessageType sends a valid signed request with an
// unrecognised message type. The handler must still return 200.
func TestWebhook_UnknownMessageType(t *testing.T) {
	logDir := t.TempDir()
	h := handler.New("test-secret", "test-token", logger.New(logDir), content.New("test-token"))
	srv := httptest.NewServer(http.HandlerFunc(h.Webhook))
	defer srv.Close()

	payload := map[string]interface{}{
		"destination": "Udeadbeefdeadbeefdeadbeefdeadbeef",
		"events": []map[string]interface{}{
			{
				"type":           "message",
				"mode":           "active",
				"timestamp":      1625665242211,
				"webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
				"deliveryContext": map[string]bool{"isRedelivery": false},
				"source":         map[string]string{"type": "user", "userId": "Udeadbeef"},
				"message": map[string]string{
					"type": "unknown_type",
					"id":   "300001",
				},
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

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
		t.Errorf("expected 200 for unknown message type, got %d", resp.StatusCode)
	}
}
