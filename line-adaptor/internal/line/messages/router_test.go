package messages_test

import (
	"testing"

	line "line-adaptor/internal/line"
	"line-adaptor/internal/line/messages"
)

func TestRoute(t *testing.T) {
	t.Run("nil msg returns error", func(t *testing.T) {
		_, err := messages.Route(nil)
		if err == nil {
			t.Fatal("expected error for nil msg, got nil")
		}
	})

	t.Run("unknown type returns error", func(t *testing.T) {
		msg := &line.Message{Type: "unknown", Id: "1"}
		_, err := messages.Route(msg)
		if err == nil {
			t.Fatal("expected error for unknown type, got nil")
		}
	})

	t.Run("text type returns *Text", func(t *testing.T) {
		msg := &line.Message{Type: "text", Id: "1", Text: "hello"}
		parsed, err := messages.Route(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := parsed.(*messages.Text); !ok {
			t.Fatalf("expected *messages.Text, got %T", parsed)
		}
	})

	t.Run("image type returns *Image", func(t *testing.T) {
		msg := &line.Message{
			Type: "image",
			Id:   "1",
			ContentProvider: &line.ContentProvider{Type: "line"},
		}
		parsed, err := messages.Route(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := parsed.(*messages.Image); !ok {
			t.Fatalf("expected *messages.Image, got %T", parsed)
		}
	})

	t.Run("video type returns *Video", func(t *testing.T) {
		msg := &line.Message{
			Type: "video",
			Id:   "1",
			ContentProvider: &line.ContentProvider{Type: "line"},
		}
		parsed, err := messages.Route(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := parsed.(*messages.Video); !ok {
			t.Fatalf("expected *messages.Video, got %T", parsed)
		}
	})

	t.Run("audio type returns *Audio", func(t *testing.T) {
		msg := &line.Message{
			Type: "audio",
			Id:   "1",
			ContentProvider: &line.ContentProvider{Type: "line"},
		}
		parsed, err := messages.Route(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := parsed.(*messages.Audio); !ok {
			t.Fatalf("expected *messages.Audio, got %T", parsed)
		}
	})

	t.Run("file type returns *File", func(t *testing.T) {
		msg := &line.Message{
			Type:     "file",
			Id:       "1",
			FileName: "report.pdf",
			FileSize: 1024,
		}
		parsed, err := messages.Route(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := parsed.(*messages.File); !ok {
			t.Fatalf("expected *messages.File, got %T", parsed)
		}
	})

	t.Run("location type returns *Location", func(t *testing.T) {
		msg := &line.Message{
			Type:      "location",
			Id:        "1",
			Latitude:  13.7563,
			Longitude: 100.5018,
		}
		parsed, err := messages.Route(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := parsed.(*messages.Location); !ok {
			t.Fatalf("expected *messages.Location, got %T", parsed)
		}
	})

	t.Run("sticker type returns *Sticker", func(t *testing.T) {
		msg := &line.Message{
			Type:                "sticker",
			Id:                  "1",
			PackageId:           "pkg1",
			StickerId:           "stk1",
			StickerResourceType: "STATIC",
		}
		parsed, err := messages.Route(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := parsed.(*messages.Sticker); !ok {
			t.Fatalf("expected *messages.Sticker, got %T", parsed)
		}
	})
}
