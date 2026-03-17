package messages_test

import (
	"testing"

	line "line-adaptor/internal/line"
	"line-adaptor/internal/line/messages"
)

func TestParseSticker(t *testing.T) {
	keywords := []string{"happy", "smile", "cute"}

	tests := []struct {
		name    string
		msg     *line.Message
		wantErr bool
		check   func(t *testing.T, got *messages.Sticker)
	}{
		{
			name: "valid with all fields",
			msg: &line.Message{
				Type:                "sticker",
				Id:                  "stk-001",
				QuoteToken:          "qt-stk",
				PackageId:           "pkg-123",
				StickerId:           "stk-456",
				StickerResourceType: "MESSAGE",
				Keywords:            keywords,
				Text:                "Happy!",
				QuotedMessageId:     "quoted-789",
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Sticker) {
				t.Helper()
				if got.Id != "stk-001" {
					t.Errorf("Id: got %q, want %q", got.Id, "stk-001")
				}
				if got.QuoteToken != "qt-stk" {
					t.Errorf("QuoteToken: got %q, want %q", got.QuoteToken, "qt-stk")
				}
				if got.PackageId != "pkg-123" {
					t.Errorf("PackageId: got %q, want %q", got.PackageId, "pkg-123")
				}
				if got.StickerId != "stk-456" {
					t.Errorf("StickerId: got %q, want %q", got.StickerId, "stk-456")
				}
				if got.StickerResourceType != "MESSAGE" {
					t.Errorf("StickerResourceType: got %q, want %q", got.StickerResourceType, "MESSAGE")
				}
				if len(got.Keywords) != 3 {
					t.Errorf("Keywords length: got %d, want 3", len(got.Keywords))
				}
				if got.Text != "Happy!" {
					t.Errorf("Text: got %q, want %q", got.Text, "Happy!")
				}
				if got.QuotedMessageId != "quoted-789" {
					t.Errorf("QuotedMessageId: got %q, want %q", got.QuotedMessageId, "quoted-789")
				}
				if got.MessageType() != "sticker" {
					t.Errorf("MessageType: got %q, want %q", got.MessageType(), "sticker")
				}
			},
		},
		{
			name: "valid with only required fields",
			msg: &line.Message{
				Type:                "sticker",
				Id:                  "stk-002",
				PackageId:           "pkg-001",
				StickerId:           "stk-001",
				StickerResourceType: "STATIC",
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Sticker) {
				t.Helper()
				if got.QuoteToken != "" {
					t.Errorf("QuoteToken should be empty, got %q", got.QuoteToken)
				}
				if got.Keywords != nil {
					t.Errorf("Keywords should be nil, got %v", got.Keywords)
				}
				if got.Text != "" {
					t.Errorf("Text should be empty, got %q", got.Text)
				}
				if got.QuotedMessageId != "" {
					t.Errorf("QuotedMessageId should be empty, got %q", got.QuotedMessageId)
				}
			},
		},
		{
			name:    "missing Id returns error",
			msg:     &line.Message{Type: "sticker", PackageId: "p1", StickerId: "s1", StickerResourceType: "STATIC"},
			wantErr: true,
		},
		{
			name:    "missing PackageId returns error",
			msg:     &line.Message{Type: "sticker", Id: "stk-003", StickerId: "s1", StickerResourceType: "STATIC"},
			wantErr: true,
		},
		{
			name:    "missing StickerId returns error",
			msg:     &line.Message{Type: "sticker", Id: "stk-004", PackageId: "p1", StickerResourceType: "STATIC"},
			wantErr: true,
		},
		{
			name:    "missing StickerResourceType returns error",
			msg:     &line.Message{Type: "sticker", Id: "stk-005", PackageId: "p1", StickerId: "s1"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := messages.ParseSticker(tc.msg)
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
