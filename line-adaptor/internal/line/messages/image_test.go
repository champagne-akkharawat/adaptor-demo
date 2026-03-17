package messages_test

import (
	"testing"

	line "line-adaptor/internal/line"
	"line-adaptor/internal/line/messages"
)

func TestParseImage(t *testing.T) {
	imageSet := &line.ImageSet{Id: "set-001", Index: 1, Total: 3}

	tests := []struct {
		name    string
		msg     *line.Message
		wantErr bool
		check   func(t *testing.T, got *messages.Image)
	}{
		{
			name: "valid with all fields - line provider",
			msg: &line.Message{
				Type:       "image",
				Id:         "img-001",
				QuoteToken: "qt-img",
				ContentProvider: &line.ContentProvider{
					Type:            "line",
					PreviewImageUrl: "",
				},
				ImageSet: imageSet,
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Image) {
				t.Helper()
				if got.Id != "img-001" {
					t.Errorf("Id: got %q, want %q", got.Id, "img-001")
				}
				if got.QuoteToken != "qt-img" {
					t.Errorf("QuoteToken: got %q, want %q", got.QuoteToken, "qt-img")
				}
				if got.ContentProvider.Type != "line" {
					t.Errorf("ContentProvider.Type: got %q, want %q", got.ContentProvider.Type, "line")
				}
				if !got.NeedsContentFetch {
					t.Error("NeedsContentFetch should be true for line provider")
				}
				if got.ImageSet == nil || got.ImageSet.Id != "set-001" {
					t.Errorf("ImageSet not copied correctly")
				}
				if got.MessageType() != "image" {
					t.Errorf("MessageType: got %q, want %q", got.MessageType(), "image")
				}
			},
		},
		{
			name: "valid with all fields - external provider",
			msg: &line.Message{
				Type:       "image",
				Id:         "img-002",
				QuoteToken: "qt-img2",
				ContentProvider: &line.ContentProvider{
					Type:               "external",
					OriginalContentUrl: "https://example.com/image.jpg",
					PreviewImageUrl:    "https://example.com/preview.jpg",
				},
				ImageSet: imageSet,
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Image) {
				t.Helper()
				if got.NeedsContentFetch {
					t.Error("NeedsContentFetch should be false for external provider")
				}
				if got.ContentProvider.OriginalContentUrl != "https://example.com/image.jpg" {
					t.Errorf("OriginalContentUrl mismatch")
				}
			},
		},
		{
			name: "valid with only required fields",
			msg: &line.Message{
				Type: "image",
				Id:   "img-003",
				ContentProvider: &line.ContentProvider{Type: "line"},
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Image) {
				t.Helper()
				if got.QuoteToken != "" {
					t.Errorf("QuoteToken should be empty, got %q", got.QuoteToken)
				}
				if got.ImageSet != nil {
					t.Errorf("ImageSet should be nil, got %v", got.ImageSet)
				}
			},
		},
		{
			name:    "missing Id returns error",
			msg:     &line.Message{Type: "image", ContentProvider: &line.ContentProvider{Type: "line"}},
			wantErr: true,
		},
		{
			name:    "missing ContentProvider returns error",
			msg:     &line.Message{Type: "image", Id: "img-004"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := messages.ParseImage(tc.msg)
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
