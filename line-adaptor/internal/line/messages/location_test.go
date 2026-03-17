package messages_test

import (
	"testing"

	line "line-adaptor/internal/line"
	"line-adaptor/internal/line/messages"
)

func TestParseLocation(t *testing.T) {
	tests := []struct {
		name    string
		msg     *line.Message
		wantErr bool
		check   func(t *testing.T, got *messages.Location)
	}{
		{
			name: "valid with all fields",
			msg: &line.Message{
				Type:      "location",
				Id:        "loc-001",
				Title:     "Aura Wellness",
				Address:   "123 Main St, Bangkok",
				Latitude:  13.7563,
				Longitude: 100.5018,
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Location) {
				t.Helper()
				if got.Id != "loc-001" {
					t.Errorf("Id: got %q, want %q", got.Id, "loc-001")
				}
				if got.Title != "Aura Wellness" {
					t.Errorf("Title: got %q, want %q", got.Title, "Aura Wellness")
				}
				if got.Address != "123 Main St, Bangkok" {
					t.Errorf("Address: got %q, want %q", got.Address, "123 Main St, Bangkok")
				}
				if got.Latitude != 13.7563 {
					t.Errorf("Latitude: got %f, want %f", got.Latitude, 13.7563)
				}
				if got.Longitude != 100.5018 {
					t.Errorf("Longitude: got %f, want %f", got.Longitude, 100.5018)
				}
				if got.MessageType() != "location" {
					t.Errorf("MessageType: got %q, want %q", got.MessageType(), "location")
				}
			},
		},
		{
			name: "valid with only required fields",
			msg: &line.Message{
				Type:      "location",
				Id:        "loc-002",
				Latitude:  35.6762,
				Longitude: 139.6503,
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.Location) {
				t.Helper()
				if got.Title != "" {
					t.Errorf("Title should be empty, got %q", got.Title)
				}
				if got.Address != "" {
					t.Errorf("Address should be empty, got %q", got.Address)
				}
				if got.Latitude != 35.6762 {
					t.Errorf("Latitude: got %f, want %f", got.Latitude, 35.6762)
				}
				if got.Longitude != 139.6503 {
					t.Errorf("Longitude: got %f, want %f", got.Longitude, 139.6503)
				}
			},
		},
		{
			name:    "missing Id returns error",
			msg:     &line.Message{Type: "location", Latitude: 13.7563, Longitude: 100.5018},
			wantErr: true,
		},
		{
			name:    "missing Latitude (zero) returns error",
			msg:     &line.Message{Type: "location", Id: "loc-003", Longitude: 100.5018},
			wantErr: true,
		},
		{
			name:    "missing Longitude (zero) returns error",
			msg:     &line.Message{Type: "location", Id: "loc-004", Latitude: 13.7563},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := messages.ParseLocation(tc.msg)
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
