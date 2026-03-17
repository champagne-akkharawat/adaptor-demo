package messages_test

import (
	"testing"

	line "line-adaptor/internal/line"
	"line-adaptor/internal/line/messages"
)

func TestParseAudio(t *testing.T) {
	tests := []struct {
		name    string
		msg     *line.Message
		wantErr bool
		check   func(t *testing.T, got *messages.Audio)
	}{
		{
			name: "valid with all fields - line provider",
			msg: &line.Message{
				Type:     "audio",
				Id:       "aud-001",
				Duration: 3000,
				ContentProvider: &line.ContentProvider{
					Type: "line",
				},
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Audio) {
				t.Helper()
				if got.Id != "aud-001" {
					t.Errorf("Id: got %q, want %q", got.Id, "aud-001")
				}
				if got.Duration != 3000 {
					t.Errorf("Duration: got %d, want %d", got.Duration, 3000)
				}
				if got.ContentProvider.Type != "line" {
					t.Errorf("ContentProvider.Type: got %q, want %q", got.ContentProvider.Type, "line")
				}
				if !got.NeedsContentFetch {
					t.Error("NeedsContentFetch should be true for line provider")
				}
				if got.MessageType() != "audio" {
					t.Errorf("MessageType: got %q, want %q", got.MessageType(), "audio")
				}
			},
		},
		{
			name: "valid with all fields - external provider",
			msg: &line.Message{
				Type:     "audio",
				Id:       "aud-002",
				Duration: 7500,
				ContentProvider: &line.ContentProvider{
					Type:               "external",
					OriginalContentUrl: "https://example.com/audio.m4a",
				},
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Audio) {
				t.Helper()
				if got.NeedsContentFetch {
					t.Error("NeedsContentFetch should be false for external provider")
				}
				if got.ContentProvider.OriginalContentUrl != "https://example.com/audio.m4a" {
					t.Errorf("OriginalContentUrl mismatch")
				}
			},
		},
		{
			name: "valid with only required fields - duration 0",
			msg: &line.Message{
				Type: "audio",
				Id:   "aud-003",
				ContentProvider: &line.ContentProvider{Type: "line"},
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Audio) {
				t.Helper()
				if got.Duration != 0 {
					t.Errorf("Duration should be 0, got %d", got.Duration)
				}
			},
		},
		{
			name:    "missing Id returns error",
			msg:     &line.Message{Type: "audio", ContentProvider: &line.ContentProvider{Type: "line"}},
			wantErr: true,
		},
		{
			name:    "missing ContentProvider returns error",
			msg:     &line.Message{Type: "audio", Id: "aud-004"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := messages.ParseAudio(tc.msg)
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
