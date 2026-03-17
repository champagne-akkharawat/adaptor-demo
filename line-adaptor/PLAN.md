# LINE Adaptor Service — Implementation Plan

## Overview

Prototype LINE webhook adapter written in Go 1.24. Minimal functionality focused on:
- Receiving and verifying LINE webhook events
- Logging each request to JSON files (raw and parsed)
- Replying to every message with "received"

---

## Project Structure

```
line-adaptor/
├── main.go
├── go.mod
└── internal/
    ├── config/
    │   └── config.go           # Env var loading
    ├── line/
    │   ├── events.go           # LINE webhook payload structs
    │   ├── signature.go        # HMAC-SHA256 verification
    │   └── reply.go            # Reply API client
    ├── handler/
    │   └── webhook.go          # HTTP handler (wires line/ + logger/)
    └── logger/
        └── logger.go           # JSON file logger (raw + parsed subdirs)
└── tests/
    └── webhook_api_test.go     # Automated API tests
```

---

## HTTP Endpoints

| Method | Path       | Description              |
|--------|------------|--------------------------|
| `POST` | `/webhook` | LINE webhook receiver    |
| `GET`  | `/health`  | Health check (200 OK)    |

---

## Configuration (env vars)

| Variable                    | Default    | Description                  |
|-----------------------------|------------|------------------------------|
| `LINE_CHANNEL_SECRET`       | (required) | For signature verification   |
| `LINE_CHANNEL_ACCESS_TOKEN` | (required) | For reply API calls          |
| `PORT`                      | `8080`     | HTTP listen port             |
| `LOG_DIR`                   | `./logs`   | Root log directory           |

---

## Request Flow

```
POST /webhook
    │
    ├── io.ReadAll (capture raw body)
    ├── Signature verification (HMAC-SHA256 vs X-Line-Signature header)
    │       └── fail → 401, stop
    │
    ├── logger: write raw bytes  → ./logs/webhook-events/raw/<timestamp>.json
    ├── json.Unmarshal → WebhookPayload struct
    ├── logger: write parsed JSON → ./logs/webhook-events/parsed/<timestamp>.json
    │
    └── for each event:
            └── if replyToken != "" → POST reply "received" to LINE Reply API

    → 200 OK
```

---

## Logging

Both logs share the same timestamp-based filename per request, making them easy to diff.

**Filename format:** `20060102T150405_<nanoseconds>.json`

| Directory                          | Source                          | Notes                              |
|------------------------------------|---------------------------------|------------------------------------|
| `./logs/webhook-events/raw/`       | `[]byte` from `io.ReadAll`      | Exact bytes LINE sent              |
| `./logs/webhook-events/parsed/`    | `json.Marshal(WebhookPayload)`  | Unknown/unmodeled fields dropped   |

Log write failures are logged to stderr but do not return an error to LINE (avoids retry storms).

---

## `internal/line/events.go` — Struct Reference

```go
// Envelope
type WebhookPayload struct {
    Destination string  `json:"destination"`
    Events      []Event `json:"events"`
}

// Common fields across all event types
type Event struct {
    Type            string          `json:"type"`
    Mode            string          `json:"mode"`
    Timestamp       int64           `json:"timestamp"`
    WebhookEventId  string          `json:"webhookEventId"`
    DeliveryContext DeliveryContext `json:"deliveryContext"`
    Source          Source          `json:"source"`
    ReplyToken      string          `json:"replyToken,omitempty"`
    Message         *Message        `json:"message,omitempty"`
    Postback        *Postback       `json:"postback,omitempty"`
}

type DeliveryContext struct {
    IsRedelivery bool `json:"isRedelivery"`
}

type Source struct {
    Type    string `json:"type"`
    UserId  string `json:"userId,omitempty"`
    GroupId string `json:"groupId,omitempty"`
    RoomId  string `json:"roomId,omitempty"`
}

// Union of all message subtypes discriminated by Type field
type Message struct {
    Type      string  `json:"type"`
    Id        string  `json:"id"`
    Text      string  `json:"text,omitempty"`
    Title     string  `json:"title,omitempty"`
    Address   string  `json:"address,omitempty"`
    Latitude  float64 `json:"latitude,omitempty"`
    Longitude float64 `json:"longitude,omitempty"`
    PackageId string  `json:"packageId,omitempty"`
    StickerId string  `json:"stickerId,omitempty"`
    FileName  string  `json:"fileName,omitempty"`
    FileSize  int64   `json:"fileSize,omitempty"`
}

type Postback struct {
    Data   string            `json:"data"`
    Params map[string]string `json:"params,omitempty"`
}
```

Event types covered by `Event.Type`: `message`, `follow`, `unfollow`, `join`, `leave`, `postback`, `memberJoined`, `memberLeft`, `unsend`, `videoPlayComplete`, `beacon`, `accountLink`, `things`.

Reply structs (outbound) live in `reply.go`, not here.

---

## `internal/line/signature.go`

- Read raw body **before** JSON parsing
- `HMAC-SHA256(key=channelSecret, message=rawBody)` → base64
- Compare to `X-Line-Signature` header using `hmac.Equal` (timing-safe)
- Reject with 401 on mismatch

---

## `internal/line/reply.go`

- Endpoint: `POST https://api.line.me/v2/bot/message/reply`
- Auth: `Authorization: Bearer {channel_access_token}`
- Body: `{"replyToken": "...", "messages": [{"type": "text", "text": "received"}]}`
- Only called when `event.ReplyToken != ""`
- Reply API errors logged to stderr, do not fail the webhook response

---

## `tests/webhook_api_test.go` — Test Cases

Uses `net/http/httptest` — no real LINE credentials needed. Reply API calls intercepted by a mock `httptest.Server`.

| Test | Scenario | Expected |
|------|----------|----------|
| Happy path | Valid signature + text message event | 200, both log files created, reply called |
| Invalid signature | Bad `X-Line-Signature` | 401, no log files written |
| Empty events array | Valid signature, `events: []` | 200, logs written, no reply attempted |
| No replyToken | e.g. unfollow event | 200, logs written, no reply attempted |

---

## Agent Execution Plan

### Wave 1 — Parallel (no dependencies)

| Agent | Files | Notes |
|-------|-------|-------|
| **A** | `go.mod`, `internal/config/config.go`, `internal/line/events.go` | Foundational, zero internal deps |
| **B** | `internal/line/signature.go` | Pure stdlib crypto |
| **C** | `internal/logger/logger.go` | Writes to raw/ and parsed/ subdirs |

### Wave 2 — Parallel (after Wave 1)

| Agent | Files | Depends on |
|-------|-------|------------|
| **D** | `internal/line/reply.go` | `events.go` |
| **E** | `internal/handler/webhook.go` | `events.go`, `signature.go`, `logger/` |

### Wave 3 — Parallel (after Wave 2)

| Agent | Files | Depends on |
|-------|-------|------------|
| **F** | `main.go` | `config/`, `handler/` |
| **G** | `tests/webhook_api_test.go` | `handler/`, `line/`, `logger/` |
