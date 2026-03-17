package content_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"line-adaptor/internal/line/content"
)

func TestFetch_200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	}))
	defer srv.Close()

	original := content.BaseURL
	content.BaseURL = srv.URL
	defer func() { content.BaseURL = original }()

	c := content.New("test-token")
	rc, err := c.Fetch(context.Background(), "msg123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer rc.Close()

	body, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	if string(body) != "hello" {
		t.Errorf("expected body %q, got %q", "hello", string(body))
	}
}

func TestFetch_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	original := content.BaseURL
	content.BaseURL = srv.URL
	defer func() { content.BaseURL = original }()

	c := content.New("test-token")
	_, err := c.Fetch(context.Background(), "msg123")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestFetch_500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	original := content.BaseURL
	content.BaseURL = srv.URL
	defer func() { content.BaseURL = original }()

	c := content.New("test-token")
	_, err := c.Fetch(context.Background(), "msg123")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestFetchPreview_200(t *testing.T) {
	const msgID = "msg456"
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("preview-data"))
	}))
	defer srv.Close()

	original := content.BaseURL
	content.BaseURL = srv.URL
	defer func() { content.BaseURL = original }()

	c := content.New("test-token")
	rc, err := c.FetchPreview(context.Background(), msgID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer rc.Close()

	body, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	if string(body) != "preview-data" {
		t.Errorf("expected body %q, got %q", "preview-data", string(body))
	}

	wantPath := "/" + msgID + "/content/preview"
	if gotPath != wantPath {
		t.Errorf("expected path %q, got %q", wantPath, gotPath)
	}
}

func TestFetchPreview_PathContainsPreview(t *testing.T) {
	const msgID = "msg789"
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	original := content.BaseURL
	content.BaseURL = srv.URL
	defer func() { content.BaseURL = original }()

	c := content.New("test-token")
	rc, err := c.FetchPreview(context.Background(), msgID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	rc.Close()

	if !strings.Contains(gotPath, "/content/preview") {
		t.Errorf("expected path to contain /content/preview, got %q", gotPath)
	}
}

func TestFetchPreview_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	original := content.BaseURL
	content.BaseURL = srv.URL
	defer func() { content.BaseURL = original }()

	c := content.New("test-token")
	_, err := c.FetchPreview(context.Background(), "msg000")
	if err == nil {
		t.Fatal("expected error for non-2xx, got nil")
	}
}
