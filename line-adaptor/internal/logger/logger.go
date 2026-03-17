package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	logDir string
}

func New(logDir string) *Logger {
	return &Logger{logDir: logDir}
}

func (l *Logger) LogWebhookEvent(raw []byte, parsed []byte) error {
	now := time.Now()
	filename := fmt.Sprintf("%s_%d.json", now.Format("20060102T150405"), now.Nanosecond())

	rawDir := filepath.Join(l.logDir, "webhook-events", "raw")
	parsedDir := filepath.Join(l.logDir, "webhook-events", "parsed")

	if err := os.MkdirAll(rawDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(parsedDir, 0755); err != nil {
		return err
	}

	rawFormatted, err := formatJSON(raw)
	if err != nil {
		rawFormatted = raw
	}
	if err := os.WriteFile(filepath.Join(rawDir, filename), rawFormatted, 0644); err != nil {
		return err
	}

	parsedFormatted, err := formatJSON(parsed)
	if err != nil {
		parsedFormatted = parsed
	}
	if err := os.WriteFile(filepath.Join(parsedDir, filename), parsedFormatted, 0644); err != nil {
		return err
	}

	return nil
}

func formatJSON(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
