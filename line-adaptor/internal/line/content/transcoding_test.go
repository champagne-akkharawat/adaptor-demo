package content_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"line-adaptor/internal/line/content"
)

func TestCheckTranscoding_Processing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"processing"}`))
	}))
	defer srv.Close()

	original := content.BaseURL
	content.BaseURL = srv.URL
	defer func() { content.BaseURL = original }()

	c := content.New("test-token")
	status, err := c.CheckTranscoding(context.Background(), "vid001")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if status != content.TranscodingProcessing {
		t.Errorf("expected %q, got %q", content.TranscodingProcessing, status)
	}
}

func TestCheckTranscoding_Succeeded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"succeeded"}`))
	}))
	defer srv.Close()

	original := content.BaseURL
	content.BaseURL = srv.URL
	defer func() { content.BaseURL = original }()

	c := content.New("test-token")
	status, err := c.CheckTranscoding(context.Background(), "vid002")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if status != content.TranscodingSucceeded {
		t.Errorf("expected %q, got %q", content.TranscodingSucceeded, status)
	}
}

func TestCheckTranscoding_Failed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"failed"}`))
	}))
	defer srv.Close()

	original := content.BaseURL
	content.BaseURL = srv.URL
	defer func() { content.BaseURL = original }()

	c := content.New("test-token")
	status, err := c.CheckTranscoding(context.Background(), "vid003")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if status != content.TranscodingFailed {
		t.Errorf("expected %q, got %q", content.TranscodingFailed, status)
	}
}

func TestCheckTranscoding_500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	original := content.BaseURL
	content.BaseURL = srv.URL
	defer func() { content.BaseURL = original }()

	c := content.New("test-token")
	_, err := c.CheckTranscoding(context.Background(), "vid004")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestCheckTranscoding_EmptyStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":""}`))
	}))
	defer srv.Close()

	original := content.BaseURL
	content.BaseURL = srv.URL
	defer func() { content.BaseURL = original }()

	c := content.New("test-token")
	_, err := c.CheckTranscoding(context.Background(), "vid005")
	if err == nil {
		t.Fatal("expected error for empty status, got nil")
	}
}

func TestCheckTranscoding_MissingStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	original := content.BaseURL
	content.BaseURL = srv.URL
	defer func() { content.BaseURL = original }()

	c := content.New("test-token")
	_, err := c.CheckTranscoding(context.Background(), "vid006")
	if err == nil {
		t.Fatal("expected error for missing status field, got nil")
	}
}

func TestCheckTranscoding_Path(t *testing.T) {
	const msgID = "vid007"
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"succeeded"}`))
	}))
	defer srv.Close()

	original := content.BaseURL
	content.BaseURL = srv.URL
	defer func() { content.BaseURL = original }()

	c := content.New("test-token")
	_, err := c.CheckTranscoding(context.Background(), msgID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	wantPath := "/" + msgID + "/content/transcoding"
	if gotPath != wantPath {
		t.Errorf("expected path %q, got %q", wantPath, gotPath)
	}
}
