package messages_test

import (
	"testing"

	line "line-adaptor/internal/line"
	"line-adaptor/internal/line/messages"
)

func TestParseVideo(t *testing.T) {
	tests := []struct {
		name    string
		msg     *line.Message
		wantErr bool
		check   func(t *testing.T, got *messages.Video)
	}{
		{
			name: "valid with all fields - line provider",
			msg: &line.Message{
				Type:       "video",
				Id:         "vid-001",
				QuoteToken: "qt-vid",
				Duration:   5000,
				ContentProvider: &line.ContentProvider{
					Type: "line",
				},
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Video) {
				t.Helper()
				if got.Id != "vid-001" {
					t.Errorf("Id: got %q, want %q", got.Id, "vid-001")
				}
				if got.QuoteToken != "qt-vid" {
					t.Errorf("QuoteToken: got %q, want %q", got.QuoteToken, "qt-vid")
				}
				if got.Duration != 5000 {
					t.Errorf("Duration: got %d, want %d", got.Duration, 5000)
				}
				if got.ContentProvider.Type != "line" {
					t.Errorf("ContentProvider.Type: got %q, want %q", got.ContentProvider.Type, "line")
				}
				if !got.NeedsContentFetch {
					t.Error("NeedsContentFetch should be true for line provider")
				}
				if got.MessageType() != "video" {
					t.Errorf("MessageType: got %q, want %q", got.MessageType(), "video")
				}
			},
		},
		{
			name: "valid with all fields - external provider",
			msg: &line.Message{
				Type:     "video",
				Id:       "vid-002",
				Duration: 10000,
				ContentProvider: &line.ContentProvider{
					Type:               "external",
					OriginalContentUrl: "https://example.com/video.mp4",
					PreviewImageUrl:    "https://example.com/preview.jpg",
				},
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Video) {
				t.Helper()
				if got.NeedsContentFetch {
					t.Error("NeedsContentFetch should be false for external provider")
				}
				if got.ContentProvider.OriginalContentUrl != "https://example.com/video.mp4" {
					t.Errorf("OriginalContentUrl mismatch")
				}
			},
		},
		{
			name: "valid with only required fields - duration 0",
			msg: &line.Message{
				Type: "video",
				Id:   "vid-003",
				ContentProvider: &line.ContentProvider{Type: "line"},
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Video) {
				t.Helper()
				if got.Duration != 0 {
					t.Errorf("Duration should be 0, got %d", got.Duration)
				}
				if got.QuoteToken != "" {
					t.Errorf("QuoteToken should be empty, got %q", got.QuoteToken)
				}
			},
		},
		{
			name:    "missing Id returns error",
			msg:     &line.Message{Type: "video", ContentProvider: &line.ContentProvider{Type: "line"}},
			wantErr: true,
		},
		{
			name:    "missing ContentProvider returns error",
			msg:     &line.Message{Type: "video", Id: "vid-004"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := messages.ParseVideo(tc.msg)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, got)
			}
		})
	}
}
