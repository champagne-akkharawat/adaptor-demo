# LINE Adaptor — Implementation Plan

## Overview

A Go HTTP server that acts as a bidirectional adaptor for the LINE Messaging API:

- **Inbound**: Exposes `POST /webhook` to receive events from LINE, verifies the signature, and processes messages.
- **Outbound**: Provides a LINE API client to send reply/push messages back to users.

---

## Directory Structure

```
line_adaptor/
├── go.mod                          # module: aura-wellness/line-adaptor
├── go.sum
├── cmd/
│   └── server/
│       └── main.go                 # entry point: wires config, client, handler, HTTP server
├── internal/
│   ├── config/
│   │   └── config.go               # reads env vars (CHANNEL_ACCESS_TOKEN, CHANNEL_SECRET, PORT)
│   ├── line/
│   │   ├── types.go                # LINE webhook & message payload structs
│   │   ├── client.go               # outgoing LINE API calls (Reply, Push)
│   │   └── signature.go            # X-Line-Signature HMAC-SHA256 verification
│   └── handler/
│       └── webhook.go              # inbound webhook HTTP handler
└── PLAN.md
```

---

## Components

### 1. Config (`internal/config/config.go`)

Reads configuration from environment variables at startup. No external dependency.

| Env Var                | Required | Description                          |
|------------------------|----------|--------------------------------------|
| `CHANNEL_ACCESS_TOKEN` | Yes      | LINE channel access token for Bearer auth |
| `CHANNEL_SECRET`       | Yes      | LINE channel secret for signature verification |
| `PORT`                 | No       | HTTP listen port (default: `8080`)   |

```go
type Config struct {
    ChannelAccessToken string
    ChannelSecret      string
    Port               string
}
```

---

### 2. Types (`internal/line/types.go`)

Structs matching the LINE webhook payload and message API shapes.

**Inbound (webhook received from LINE):**
```go
type WebhookRequest struct {
    Destination string  `json:"destination"`
    Events      []Event `json:"events"`
}

type Event struct {
    Type            string  `json:"type"`
    Mode            string  `json:"mode"`
    Timestamp       int64   `json:"timestamp"`
    Source          Source  `json:"source"`
    WebhookEventId  string  `json:"webhookEventId"`
    ReplyToken      string  `json:"replyToken"`
    Message         Message `json:"message"`
}

type Source struct {
    Type   string `json:"type"`
    UserID string `json:"userId"`
}

type Message struct {
    ID         string `json:"id"`
    Type       string `json:"type"`
    Text       string `json:"text"`
    QuoteToken string `json:"quoteToken"`
}
```

**Outbound (sent to LINE API):**
```go
type ReplyRequest struct {
    ReplyToken string        `json:"replyToken"`
    Messages   []TextMessage `json:"messages"`
}

type PushRequest struct {
    To       string        `json:"to"`
    Messages []TextMessage `json:"messages"`
}

type TextMessage struct {
    Type string `json:"type"` // always "text"
    Text string `json:"text"`
}
```

---

### 3. Signature Verification (`internal/line/signature.go`)

LINE sends a `X-Line-Signature` header containing Base64(HMAC-SHA256(rawBody, channelSecret)).

```go
func Verify(channelSecret, signature string, body []byte) bool {
    mac := hmac.New(sha256.New, []byte(channelSecret))
    mac.Write(body)
    expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

---

### 4. LINE Client (`internal/line/client.go`)

Handles all outgoing calls to the LINE Messaging API.

```go
type Client struct {
    channelAccessToken string
    baseURL            string       // https://api.line.me
    httpClient         *http.Client
}

func NewClient(token string) *Client

// Reply to a webhook event using its replyToken (single-use, within 30s)
func (c *Client) Reply(replyToken string, messages []TextMessage) error

// Push a proactive message to a user by userId (requires push-capable plan)
func (c *Client) Push(to string, messages []TextMessage) error
```

**Endpoints used:**

| Method | Direction | Endpoint |
|--------|-----------|----------|
| Reply  | Outbound  | `POST https://api.line.me/v2/bot/message/reply` |
| Push   | Outbound  | `POST https://api.line.me/v2/bot/message/push`  |

Both require header: `Authorization: Bearer <CHANNEL_ACCESS_TOKEN>`

---

### 5. Webhook Handler (`internal/handler/webhook.go`)

```go
type WebhookHandler struct {
    channelSecret string
    lineClient    *line.Client
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)
```

**Request handling flow:**

```
1. Read raw request body (preserve bytes for signature check)
2. Verify X-Line-Signature → 401 if invalid
3. Unmarshal JSON into WebhookRequest
4. Iterate over events:
   - Skip non-"message" type events
   - Skip non-"text" message types
   - Call lineClient.Reply(event.ReplyToken, echoMessages)
5. Respond 200 OK
```

---

### 6. Entry Point (`cmd/server/main.go`)

```go
func main() {
    cfg := config.Load()
    lineClient := line.NewClient(cfg.ChannelAccessToken)
    webhookHandler := handler.New(cfg.ChannelSecret, lineClient)

    mux := http.NewServeMux()
    mux.HandleFunc("GET /health", healthCheck)
    mux.Handle("POST /webhook", webhookHandler)

    log.Printf("LINE adaptor listening on :%s", cfg.Port)
    http.ListenAndServe(":"+cfg.Port, mux)
}
```

---

## HTTP Endpoints

| Method | Path       | Description                          |
|--------|------------|--------------------------------------|
| POST   | `/webhook` | Receives inbound LINE events         |
| GET    | `/health`  | Health check (returns 200 OK)        |

---

## Implementation Steps

- [ ] **Step 1** — `go mod init aura-wellness/line-adaptor` (stdlib only, no external deps)
- [ ] **Step 2** — `internal/config/config.go` — env var loading
- [ ] **Step 3** — `internal/line/types.go` — webhook & message structs
- [ ] **Step 4** — `internal/line/signature.go` — HMAC-SHA256 verification
- [ ] **Step 5** — `internal/line/client.go` — Reply and Push API calls
- [ ] **Step 6** — `internal/handler/webhook.go` — inbound webhook handler
- [ ] **Step 7** — `cmd/server/main.go` — wire everything, start HTTP server
- [ ] **Step 8** — Local test: `go run ./cmd/server` + `curl` to hit `/health` and `/webhook`

---

## Environment Variables (`.env` example)

```env
CHANNEL_ACCESS_TOKEN=<your-line-channel-access-token>
CHANNEL_SECRET=<your-line-channel-secret>
PORT=8080
```

---

## LINE Developer Console Setup (external steps)

1. Create a **Messaging API channel** at https://developers.line.biz/
2. Set webhook URL to `https://<your-host>/webhook` and enable webhooks
3. Copy **Channel Access Token** (long-lived) and **Channel Secret** to env vars
4. Disable auto-reply in the LINE Official Account Manager

---

## Notes

- `replyToken` from LINE is **single-use** and expires after **30 seconds** — Reply must be called promptly in the webhook handler.
- Push messages require a paid LINE plan; Reply is available on all plans.
- All LINE API responses with non-2xx status should be logged and surfaced as errors.
- stdlib only (`net/http`, `crypto/hmac`, `crypto/sha256`, `encoding/base64`, `encoding/json`) — zero external dependencies.
