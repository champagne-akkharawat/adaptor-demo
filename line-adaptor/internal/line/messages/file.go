package messages

import (
	"fmt"

	line "line-adaptor/internal/line"
)

// File holds the parsed fields of a LINE file message.
type File struct {
	Id                string
	FileName          string
	FileSize          int64
	NeedsContentFetch bool // always true for file (always LINE-hosted)
}

// MessageType implements Parsed.
func (f *File) MessageType() string { return "file" }

// ParseFile parses a raw line.Message of type "file".
// Required fields: Id, FileName, FileSize (non-zero). NeedsContentFetch is always true.
func ParseFile(msg *line.Message) (*File, error) {
	if msg.Id == "" {
		return nil, fmt.Errorf("messages/file: Id is required")
	}
	if msg.FileName == "" {
		return nil, fmt.Errorf("messages/file: FileName is required")
	}
	if msg.FileSize == 0 {
		return nil, fmt.Errorf("messages/file: FileSize is required")
	}
	return &File{
		Id:                msg.Id,
		FileName:          msg.FileName,
		FileSize:          msg.FileSize,
		NeedsContentFetch: true,
	}, nil
}
