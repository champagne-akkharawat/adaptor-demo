package messages_test

import (
	"testing"

	line "line-adaptor/internal/line"
	"line-adaptor/internal/line/messages"
)

func TestParseFile(t *testing.T) {
	tests := []struct {
		name    string
		msg     *line.Message
		wantErr bool
		check   func(t *testing.T, got *messages.File)
	}{
		{
			name: "valid with all fields",
			msg: &line.Message{
				Type:     "file",
				Id:       "file-001",
				FileName: "report.pdf",
				FileSize: 204800,
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.File) {
				t.Helper()
				if got.Id != "file-001" {
					t.Errorf("Id: got %q, want %q", got.Id, "file-001")
				}
				if got.FileName != "report.pdf" {
					t.Errorf("FileName: got %q, want %q", got.FileName, "report.pdf")
				}
				if got.FileSize != 204800 {
					t.Errorf("FileSize: got %d, want %d", got.FileSize, 204800)
				}
				if !got.NeedsContentFetch {
					t.Error("NeedsContentFetch should always be true for file messages")
				}
				if got.MessageType() != "file" {
					t.Errorf("MessageType: got %q, want %q", got.MessageType(), "file")
				}
			},
		},
		{
			name: "valid with only required fields (same as all fields for file)",
			msg: &line.Message{
				Type:     "file",
				Id:       "file-002",
				FileName: "data.csv",
				FileSize: 512,
			},
			wantErr: false,
			check: func(t *testing.T, got *messages.File) {
				t.Helper()
				if got.Id != "file-002" {
					t.Errorf("Id: got %q, want %q", got.Id, "file-002")
				}
				if !got.NeedsContentFetch {
					t.Error("NeedsContentFetch should always be true")
				}
			},
		},
		{
			name:    "missing Id returns error",
			msg:     &line.Message{Type: "file", FileName: "x.txt", FileSize: 100},
			wantErr: true,
		},
		{
			name:    "missing FileName returns error",
			msg:     &line.Message{Type: "file", Id: "file-003", FileSize: 100},
			wantErr: true,
		},
		{
			name:    "missing FileSize (zero) returns error",
			msg:     &line.Message{Type: "file", Id: "file-004", FileName: "x.txt"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := messages.ParseFile(tc.msg)
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
