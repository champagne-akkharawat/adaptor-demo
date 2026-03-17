# LINE Adaptor — Message Processing Plan

## Current State

| File | Concern | Status |
|---|---|---|
| `line/signature.go` | HMAC-SHA256 webhook verification | Done |
| `line/events.go` | Raw webhook payload structs | Done — flat `Message` struct missing many fields |
| `line/reply.go` | Reply API (hardcoded "received") | Done — stub only |
| `handler/webhook.go` | HTTP entry point | Done — no message-type awareness, replies to every event |

## Separate Concerns to Build

---

### Prerequisite — Expand `line/events.go`

**Purpose:** The existing `Message` struct only covers a subset of LINE's message
fields. The `messages/` parsers need all fields to be deserialized before they
can do their work. This is a cleanup/expansion of the existing file, not a new
package.

**Missing fields to add to `Message`:**

| Field | Used by |
|---|---|
| `QuoteToken` | text, image, video, sticker |
| `ContentProvider` | image, video, audio |
| `ImageSet` | image |
| `Duration` | video, audio |
| `Emojis` | text |
| `Mention` | text |
| `QuotedMessageId` | text, sticker |
| `StickerResourceType` | sticker |
| `Keywords` | sticker |

**New types to add alongside `Message`:**

```go
type ContentProvider struct {
    Type               string `json:"type"`                         // "line" or "external"
    OriginalContentUrl string `json:"originalContentUrl,omitempty"`
    PreviewImageUrl    string `json:"previewImageUrl,omitempty"`
}

type ImageSet struct {
    Id    string `json:"id"`
    Index int    `json:"index"`
    Total int    `json:"total"`
}

type Emoji struct {
    Index     int    `json:"index"`
    Length    int    `json:"length"`
    ProductId string `json:"productId"`
    EmojiId   string `json:"emojiId"`
}

type Mention struct {
    Mentionees []Mentionee `json:"mentionees"`
}

type Mentionee struct {
    Index  int    `json:"index"`
    Length int    `json:"length"`
    Type   string `json:"type"`             // "user" or "all"
    UserId string `json:"userId,omitempty"`
    IsSelf bool   `json:"isSelf,omitempty"`
}
```

**No changes to `Event`, `Source`, `WebhookPayload`, or `Postback`** — those
are complete for the current scope.

**Tests:** `line/events_test.go`
- JSON unmarshal round-trip for all 7 message types using fixture payloads
- Verify each new field (`QuoteToken`, `ContentProvider`, `ImageSet`, `Duration`,
  `Emojis`, `Mention`, `QuotedMessageId`, `StickerResourceType`, `Keywords`)
  deserializes correctly from realistic JSON
- Verify `omitempty` fields are absent when zero/nil (marshal direction)

---

### Concern 1 — Message Parsing (`line/messages/`)

**Purpose:** Convert the raw, flat `*line.Message` struct into a strongly-typed parsed representation. No I/O, no HTTP. Pure data transformation.

**Package:** `messages`

**File layout:**

```
line/messages/
├── router.go       -- Parsed interface + Route(*line.Message) dispatcher
├── text.go         -- Text struct + Parse
├── image.go        -- Image struct + Parse
├── video.go        -- Video struct + Parse
├── audio.go        -- Audio struct + Parse
├── file.go         -- File struct + Parse
├── location.go     -- Location struct + Parse
└── sticker.go      -- Sticker struct + Parse
```

**Key types:**

```go
// router.go
type Parsed interface {
    MessageType() string
}

func Route(msg *line.Message) (Parsed, error)
// Returns *Text | *Image | *Video | *Audio | *File | *Location | *Sticker
// Returns error if msg.Type is unknown or required fields are missing
```

**Each parser:**
- Validates required fields, returns `error` on failure
- Carries only fields relevant to its type (no cross-type pollution)
- Media types (`Image`, `Video`, `Audio`, `File`) expose `NeedsContentFetch bool`
  and `MessageId string` so callers know whether to proceed to Concern 2
- No HTTP calls

**Dependencies:** `line` package only (for `*line.Message`)

**Tests:** one `_test.go` file per parser + `router_test.go`, all in `line/messages/`
- Each parser: table-driven tests covering valid input, each required-field-missing
  error case, and optional fields absent (nil/zero)
- `router_test.go`: dispatches to correct type for each of the 7 `msg.Type` values;
  returns error for unknown type; returns error when `msg` is nil

---

### Concern 2 — Content Fetching (`line/content/`)

**Purpose:** Download media from LINE servers for message types where
`contentProvider.type == "line"`. Separate from parsing so the handler can
choose when (or whether) to fetch, e.g. only if downstream needs the bytes.

**Package:** `content`

**File layout:**

```
line/content/
├── client.go       -- Client struct, constructor
├── fetch.go        -- Fetch, FetchPreview
└── transcoding.go  -- CheckTranscoding (video readiness)
```

**Key types:**

```go
// client.go
type Client struct { /* accessToken, http.Client */ }
func New(accessToken string) *Client

// fetch.go
func (c *Client) Fetch(ctx context.Context, messageId string) (io.ReadCloser, error)
// GET /v2/bot/message/{messageId}/content
// Used by: image (line-hosted), video, audio, file

func (c *Client) FetchPreview(ctx context.Context, messageId string) (io.ReadCloser, error)
// GET /v2/bot/message/{messageId}/content/preview
// Used by: image, video only

// transcoding.go
func (c *Client) CheckTranscoding(ctx context.Context, messageId string) (string, error)
// GET /v2/bot/message/{messageId}/content/transcoding
// Returns status string: "processing" | "succeeded" | "failed"
// Used by: video only, before Fetch
```

**Notes:**
- Callers are responsible for closing the returned `io.ReadCloser`
- Content on LINE servers expires (undisclosed TTL) — fetch promptly after webhook receipt
- `FetchPreview` is not applicable to audio or file types

**Dependencies:** standard library only (`net/http`, `context`, `io`)

**Tests:** `line/content/fetch_test.go`, `line/content/transcoding_test.go`
- Use `httptest.NewServer` as a mock LINE content API
- `Fetch`: 200 with body returned as `io.ReadCloser`; non-2xx returns error
- `FetchPreview`: same coverage; verify correct URL path (`/content/preview`)
- `CheckTranscoding`: parses `"processing"`, `"succeeded"`, `"failed"` status
  strings from mock response; non-2xx returns error

---

### Concern 3 — Handler Wiring (`handler/webhook.go`)

**Purpose:** Connect the two new concerns into the existing HTTP handler. This
file already owns the request lifecycle; it gains message-type awareness.

**Changes to `webhook.go`:**
1. After unmarshalling, route each `message` event through `messages.Route()`
2. For media messages where `NeedsContentFetch == true`, call `content.Client`
   if the handler needs the bytes (or pass the `MessageId` downstream — TBD)
3. Replace the blanket `event.ReplyToken != ""` reply with per-type logic once
   reply building is richer (out of scope here — see below)

**New dependency injection:**
```go
type Handler struct {
    channelSecret  string
    accessToken    string
    log            *logger.Logger
    contentClient  *content.Client   // added
}
```

**Tests:** extend `tests/webhook_api_test.go`
- Existing tests must continue to pass unchanged
- Add cases for each of the 7 message types: verify `messages.Route()` is called
  and the correct typed result flows through (assert via a test spy or log output,
  not internal state)
- Add a case where `messages.Route()` returns an error (unknown message type):
  handler must still return 200 (don't reject valid webhooks over unknown types)

---

## What Is Out of Scope Here

- **Reply building** — `reply.go` currently sends a hardcoded `"received"`.
  Building per-type reply messages (e.g. echoing text, confirming file receipt)
  is a separate concern that builds on top of the parsed types once they exist.
- **Downstream forwarding** — routing parsed events to an external system (e.g.
  a webhook fan-out or message queue) is not part of this plan.
- **Non-message event types** — `follow`, `unfollow`, `join`, `leave`,
  `postback`, etc. are not addressed here. The message type processor is the
  first slice; event-type routing is next.

---

## Implementation Order

0. `line/events.go` — clean up and expand `Message` struct + `line/events_test.go`
   - Remove fields that belong to specific message types (location, sticker, file
     fields currently embedded flat) and replace with properly typed optional fields
   - Add missing fields (`QuoteToken`, `ContentProvider`, `ImageSet`, `Duration`,
     `Emojis`, `Mention`, `QuotedMessageId`, `StickerResourceType`, `Keywords`)
   - Add supporting types: `ContentProvider`, `ImageSet`, `Emoji`, `Mention`, `Mentionee`
   - Write `events_test.go`: JSON unmarshal round-trips for all 7 message types
1. `line/messages/` — parsers + router + per-file `_test.go` files
2. `line/content/` — content client + `fetch_test.go` + `transcoding_test.go`
3. `handler/webhook.go` — wire in both, update `Handler` constructor,
   extend `tests/webhook_api_test.go`

## Subagent Execution Plan

```
[Wave 1]  events-cleanup
          (step 0: events.go + events_test.go)
                    ↓
[Wave 2]  messages-parser          ║  content-client
          (step 1: messages/ +     ║  (step 2: content/ +
           per-type _test.go)      ║   fetch_test, transcoding_test)
                    ↓                       ↓
[Wave 3]            handler-wiring
                    (step 3: webhook.go + webhook_api_test.go)
```

Each agent writes its tests alongside its implementation code. Wave 2 agents run
in parallel — they have no shared files and no mutual dependency. Wave 3 waits
for both Wave 2 agents to complete before starting.
