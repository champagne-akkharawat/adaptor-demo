package line

import (
	"encoding/json"
	"testing"
)

func TestMessageUnmarshal(t *testing.T) {
	t.Run("text", func(t *testing.T) {
		raw := `{
			"type": "text",
			"id": "468789577898913",
			"quoteToken": "q3Plxr4AgKd...",
			"text": "Hello, @All @example\uFEFF",
			"emojis": [
				{
					"index": 0,
					"length": 6,
					"productId": "5ac21a8c040ab15980c9b43f",
					"emojiId": "001"
				}
			],
			"mention": {
				"mentionees": [
					{
						"index": 7,
						"length": 4,
						"type": "all"
					},
					{
						"index": 12,
						"length": 8,
						"type": "user",
						"userId": "U49585cd0d5...",
						"isSelf": true
					}
				]
			},
			"quotedMessageId": "468789577898900"
		}`

		var m Message
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if m.Type != "text" {
			t.Errorf("Type: got %q, want %q", m.Type, "text")
		}
		if m.Id != "468789577898913" {
			t.Errorf("Id: got %q, want %q", m.Id, "468789577898913")
		}
		if m.QuoteToken != "q3Plxr4AgKd..." {
			t.Errorf("QuoteToken: got %q, want %q", m.QuoteToken, "q3Plxr4AgKd...")
		}
		if m.Text != "Hello, @All @example\uFEFF" {
			t.Errorf("Text: got %q", m.Text)
		}
		if len(m.Emojis) != 1 {
			t.Fatalf("Emojis len: got %d, want 1", len(m.Emojis))
		}
		e := m.Emojis[0]
		if e.Index != 0 {
			t.Errorf("Emoji.Index: got %d, want 0", e.Index)
		}
		if e.Length != 6 {
			t.Errorf("Emoji.Length: got %d, want 6", e.Length)
		}
		if e.ProductId != "5ac21a8c040ab15980c9b43f" {
			t.Errorf("Emoji.ProductId: got %q", e.ProductId)
		}
		if e.EmojiId != "001" {
			t.Errorf("Emoji.EmojiId: got %q, want %q", e.EmojiId, "001")
		}
		if m.Mention == nil {
			t.Fatal("Mention: got nil, want non-nil")
		}
		if len(m.Mention.Mentionees) != 2 {
			t.Fatalf("Mentionees len: got %d, want 2", len(m.Mention.Mentionees))
		}
		all := m.Mention.Mentionees[0]
		if all.Index != 7 || all.Length != 4 || all.Type != "all" {
			t.Errorf("Mentionees[0]: got %+v", all)
		}
		user := m.Mention.Mentionees[1]
		if user.Index != 12 || user.Length != 8 || user.Type != "user" {
			t.Errorf("Mentionees[1] base fields: got %+v", user)
		}
		if user.UserId != "U49585cd0d5..." {
			t.Errorf("Mentionees[1].UserId: got %q", user.UserId)
		}
		if !user.IsSelf {
			t.Errorf("Mentionees[1].IsSelf: got false, want true")
		}
		if m.QuotedMessageId != "468789577898900" {
			t.Errorf("QuotedMessageId: got %q", m.QuotedMessageId)
		}
	})

	t.Run("image_line", func(t *testing.T) {
		raw := `{
			"type": "image",
			"id": "354718705273720",
			"quoteToken": "q3Plxr4AgKd...",
			"contentProvider": {
				"type": "line"
			},
			"imageSet": {
				"id": "E005D41A7288F41B65593ED38FF6E9834B046AB99289A775950A9BF7D78E2DD",
				"index": 1,
				"total": 3
			}
		}`

		var m Message
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if m.Type != "image" {
			t.Errorf("Type: got %q, want %q", m.Type, "image")
		}
		if m.Id != "354718705273720" {
			t.Errorf("Id: got %q", m.Id)
		}
		if m.QuoteToken != "q3Plxr4AgKd..." {
			t.Errorf("QuoteToken: got %q", m.QuoteToken)
		}
		if m.ContentProvider == nil {
			t.Fatal("ContentProvider: got nil")
		}
		if m.ContentProvider.Type != "line" {
			t.Errorf("ContentProvider.Type: got %q, want %q", m.ContentProvider.Type, "line")
		}
		if m.ImageSet == nil {
			t.Fatal("ImageSet: got nil")
		}
		if m.ImageSet.Id != "E005D41A7288F41B65593ED38FF6E9834B046AB99289A775950A9BF7D78E2DD" {
			t.Errorf("ImageSet.Id: got %q", m.ImageSet.Id)
		}
		if m.ImageSet.Index != 1 {
			t.Errorf("ImageSet.Index: got %d, want 1", m.ImageSet.Index)
		}
		if m.ImageSet.Total != 3 {
			t.Errorf("ImageSet.Total: got %d, want 3", m.ImageSet.Total)
		}
	})

	t.Run("image_external", func(t *testing.T) {
		raw := `{
			"type": "image",
			"id": "354718705273721",
			"quoteToken": "q3Plxr4AgKd...",
			"contentProvider": {
				"type": "external",
				"originalContentUrl": "https://example.com/original.jpg",
				"previewImageUrl": "https://example.com/preview.jpg"
			}
		}`

		var m Message
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if m.ContentProvider == nil {
			t.Fatal("ContentProvider: got nil")
		}
		if m.ContentProvider.Type != "external" {
			t.Errorf("ContentProvider.Type: got %q, want %q", m.ContentProvider.Type, "external")
		}
		if m.ContentProvider.OriginalContentUrl != "https://example.com/original.jpg" {
			t.Errorf("ContentProvider.OriginalContentUrl: got %q", m.ContentProvider.OriginalContentUrl)
		}
		if m.ContentProvider.PreviewImageUrl != "https://example.com/preview.jpg" {
			t.Errorf("ContentProvider.PreviewImageUrl: got %q", m.ContentProvider.PreviewImageUrl)
		}
	})

	t.Run("video", func(t *testing.T) {
		raw := `{
			"type": "video",
			"id": "354718705273722",
			"quoteToken": "q3Plxr4AgKd...",
			"duration": 60000,
			"contentProvider": {
				"type": "line"
			}
		}`

		var m Message
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if m.Type != "video" {
			t.Errorf("Type: got %q, want %q", m.Type, "video")
		}
		if m.Id != "354718705273722" {
			t.Errorf("Id: got %q", m.Id)
		}
		if m.QuoteToken != "q3Plxr4AgKd..." {
			t.Errorf("QuoteToken: got %q", m.QuoteToken)
		}
		if m.Duration != 60000 {
			t.Errorf("Duration: got %d, want 60000", m.Duration)
		}
		if m.ContentProvider == nil {
			t.Fatal("ContentProvider: got nil")
		}
		if m.ContentProvider.Type != "line" {
			t.Errorf("ContentProvider.Type: got %q, want %q", m.ContentProvider.Type, "line")
		}
	})

	t.Run("audio", func(t *testing.T) {
		raw := `{
			"type": "audio",
			"id": "354718705273723",
			"duration": 30000,
			"contentProvider": {
				"type": "line"
			}
		}`

		var m Message
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if m.Type != "audio" {
			t.Errorf("Type: got %q, want %q", m.Type, "audio")
		}
		if m.Id != "354718705273723" {
			t.Errorf("Id: got %q", m.Id)
		}
		if m.QuoteToken != "" {
			t.Errorf("QuoteToken: expected empty for audio, got %q", m.QuoteToken)
		}
		if m.Duration != 30000 {
			t.Errorf("Duration: got %d, want 30000", m.Duration)
		}
		if m.ContentProvider == nil {
			t.Fatal("ContentProvider: got nil")
		}
		if m.ContentProvider.Type != "line" {
			t.Errorf("ContentProvider.Type: got %q, want %q", m.ContentProvider.Type, "line")
		}
	})

	t.Run("file", func(t *testing.T) {
		raw := `{
			"type": "file",
			"id": "325708",
			"fileName": "file.txt",
			"fileSize": 2138
		}`

		var m Message
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if m.Type != "file" {
			t.Errorf("Type: got %q, want %q", m.Type, "file")
		}
		if m.Id != "325708" {
			t.Errorf("Id: got %q", m.Id)
		}
		if m.FileName != "file.txt" {
			t.Errorf("FileName: got %q, want %q", m.FileName, "file.txt")
		}
		if m.FileSize != 2138 {
			t.Errorf("FileSize: got %d, want 2138", m.FileSize)
		}
		if m.ContentProvider != nil {
			t.Errorf("ContentProvider: expected nil for file type, got %+v", m.ContentProvider)
		}
		if m.QuoteToken != "" {
			t.Errorf("QuoteToken: expected empty for file type, got %q", m.QuoteToken)
		}
	})

	t.Run("location", func(t *testing.T) {
		raw := `{
			"type": "location",
			"id": "325708",
			"title": "my location",
			"address": "1-6-1 Yaesu, Chuo-ku, Tokyo",
			"latitude": 35.65910807942215,
			"longitude": 139.70372892916718
		}`

		var m Message
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if m.Type != "location" {
			t.Errorf("Type: got %q, want %q", m.Type, "location")
		}
		if m.Id != "325708" {
			t.Errorf("Id: got %q", m.Id)
		}
		if m.Title != "my location" {
			t.Errorf("Title: got %q, want %q", m.Title, "my location")
		}
		if m.Address != "1-6-1 Yaesu, Chuo-ku, Tokyo" {
			t.Errorf("Address: got %q", m.Address)
		}
		if m.Latitude != 35.65910807942215 {
			t.Errorf("Latitude: got %v, want 35.65910807942215", m.Latitude)
		}
		if m.Longitude != 139.70372892916718 {
			t.Errorf("Longitude: got %v, want 139.70372892916718", m.Longitude)
		}
		if m.QuoteToken != "" {
			t.Errorf("QuoteToken: expected empty for location, got %q", m.QuoteToken)
		}
	})

	t.Run("sticker", func(t *testing.T) {
		raw := `{
			"type": "sticker",
			"id": "1501597916",
			"quoteToken": "q3Plxr4AgKd...",
			"packageId": "446",
			"stickerId": "1988",
			"stickerResourceType": "ANIMATION",
			"keywords": ["cony", "sticker", "cute"],
			"quotedMessageId": "468789577898900"
		}`

		var m Message
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if m.Type != "sticker" {
			t.Errorf("Type: got %q, want %q", m.Type, "sticker")
		}
		if m.Id != "1501597916" {
			t.Errorf("Id: got %q", m.Id)
		}
		if m.QuoteToken != "q3Plxr4AgKd..." {
			t.Errorf("QuoteToken: got %q", m.QuoteToken)
		}
		if m.PackageId != "446" {
			t.Errorf("PackageId: got %q, want %q", m.PackageId, "446")
		}
		if m.StickerId != "1988" {
			t.Errorf("StickerId: got %q, want %q", m.StickerId, "1988")
		}
		if m.StickerResourceType != "ANIMATION" {
			t.Errorf("StickerResourceType: got %q, want %q", m.StickerResourceType, "ANIMATION")
		}
		if len(m.Keywords) != 3 {
			t.Fatalf("Keywords len: got %d, want 3", len(m.Keywords))
		}
		wantKeywords := []string{"cony", "sticker", "cute"}
		for i, kw := range wantKeywords {
			if m.Keywords[i] != kw {
				t.Errorf("Keywords[%d]: got %q, want %q", i, m.Keywords[i], kw)
			}
		}
		if m.QuotedMessageId != "468789577898900" {
			t.Errorf("QuotedMessageId: got %q", m.QuotedMessageId)
		}
	})

	t.Run("omitempty_fields_absent", func(t *testing.T) {
		// Audio has no quoteToken; text has no contentProvider.
		// Verify that absent optional fields deserialise to zero values.
		audioRaw := `{
			"type": "audio",
			"id": "999",
			"duration": 5000,
			"contentProvider": {"type": "line"}
		}`
		var audio Message
		if err := json.Unmarshal([]byte(audioRaw), &audio); err != nil {
			t.Fatalf("unmarshal audio: %v", err)
		}
		if audio.QuoteToken != "" {
			t.Errorf("audio QuoteToken should be empty, got %q", audio.QuoteToken)
		}
		if audio.ImageSet != nil {
			t.Errorf("audio ImageSet should be nil, got %+v", audio.ImageSet)
		}
		if audio.Mention != nil {
			t.Errorf("audio Mention should be nil, got %+v", audio.Mention)
		}

		textRaw := `{
			"type": "text",
			"id": "111",
			"quoteToken": "tok",
			"text": "hi"
		}`
		var text Message
		if err := json.Unmarshal([]byte(textRaw), &text); err != nil {
			t.Fatalf("unmarshal text: %v", err)
		}
		if text.ContentProvider != nil {
			t.Errorf("text ContentProvider should be nil, got %+v", text.ContentProvider)
		}
		if text.Duration != 0 {
			t.Errorf("text Duration should be 0, got %d", text.Duration)
		}
		if len(text.Emojis) != 0 {
			t.Errorf("text Emojis should be empty, got %v", text.Emojis)
		}
	})
}
