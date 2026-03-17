package messages_test

import (
	"testing"

	line "line-adaptor/internal/line"
	"line-adaptor/internal/line/messages"
)

func TestParseText(t *testing.T) {
	mention := &line.Mention{
		Mentionees: []line.Mentionee{
			{Index: 0, Length: 5, Type: "user", UserId: "U123"},
		},
	}
	emojis := []line.Emoji{
		{Index: 6, Length: 6, ProductId: "prod1", EmojiId: "emoji1"},
	}

	tests := []struct {
		name    string
		msg     *line.Message
		wantErr bool
		check   func(t *testing.T, got *messages.Text)
	}{
		{
			name: "valid with all fields",
			msg: &line.Message{
				Type:            "text",
				Id:              "msg-001",
				QuoteToken:      "qt-abc",
				Text:            "Hello world",
				Emojis:          emojis,
				Mention:         mention,
				QuotedMessageId: "quoted-001",
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Text) {
				t.Helper()
				if got.Id != "msg-001" {
					t.Errorf("Id: got %q, want %q", got.Id, "msg-001")
				}
				if got.QuoteToken != "qt-abc" {
					t.Errorf("QuoteToken: got %q, want %q", got.QuoteToken, "qt-abc")
				}
				if got.Text != "Hello world" {
					t.Errorf("Text: got %q, want %q", got.Text, "Hello world")
				}
				if len(got.Emojis) != 1 {
					t.Errorf("Emojis length: got %d, want 1", len(got.Emojis))
				}
				if got.Mention == nil || len(got.Mention.Mentionees) != 1 {
					t.Errorf("Mention not copied correctly")
				}
				if got.QuotedMessageId != "quoted-001" {
					t.Errorf("QuotedMessageId: got %q, want %q", got.QuotedMessageId, "quoted-001")
				}
				if got.MessageType() != "text" {
					t.Errorf("MessageType: got %q, want %q", got.MessageType(), "text")
				}
			},
		},
		{
			name: "valid with only required fields",
			msg: &line.Message{
				Type: "text",
				Id:   "msg-002",
				Text: "Hi",
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Text) {
				t.Helper()
				if got.Id != "msg-002" {
					t.Errorf("Id: got %q, want %q", got.Id, "msg-002")
				}
				if got.Text != "Hi" {
					t.Errorf("Text: got %q, want %q", got.Text, "Hi")
				}
				if got.QuoteToken != "" {
					t.Errorf("QuoteToken should be empty, got %q", got.QuoteToken)
				}
				if got.Emojis != nil {
					t.Errorf("Emojis should be nil, got %v", got.Emojis)
				}
				if got.Mention != nil {
					t.Errorf("Mention should be nil, got %v", got.Mention)
				}
				if got.QuotedMessageId != "" {
					t.Errorf("QuotedMessageId should be empty, got %q", got.QuotedMessageId)
				}
			},
		},
		{
			name:    "missing Id returns error",
			msg:     &line.Message{Type: "text", Text: "Hello"},
			wantErr: true,
		},
		{
			name:    "missing Text returns error",
			msg:     &line.Message{Type: "text", Id: "msg-003"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := messages.ParseText(tc.msg)
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
