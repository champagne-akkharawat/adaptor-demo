# Unified Message Schema — Overall Analysis

> **Date:** 2026-03-17
> **Sources analysed:**
> - LINE Messaging API webhook & send API (deep_research_2)
> - Instagram Messaging API webhook & send API (deep_research)
> - Facebook Messenger webhook & send API (deep_research)

---

## Table of Contents

- [What the Platforms Share (the Common CS Core)](#what-the-platforms-share-the-common-cs-core)
- [Key Divergences That Matter for the Schema](#key-divergences-that-matter-for-the-schema)
- [Schema Architecture Recommendation](#schema-architecture-recommendation)
  - [Unified Inbound Message Shape](#unified-inbound-message-shape)
  - [Unified Outbound Reply Shape](#unified-outbound-reply-shape)
- [Reply Window Constraints](#reply-window-constraints)
- [Signature Verification Differences](#signature-verification-differences)
- [Webhook Envelope Differences](#webhook-envelope-differences)

---

## What the Platforms Share (the Common CS Core)

All three platforms converge on the same fundamental CS loop:

1. **Inbound trigger** — user sends a message → webhook POST with a sender ID + message content
2. **Reply window** — LINE has no window (push any time with a token); Facebook/Instagram enforce a **24-hour window** after the user's last message, with limited tag-based escapes (`HUMAN_AGENT`)
3. **Sender identity** — LINE uses `userId`, Facebook uses `PSID` (Page-Scoped), Instagram uses `IGSID` (IG-Scoped). All are platform-scoped and must be stored per-platform per-account
4. **Core message types** — text + image + video + audio are universally supported inbound and outbound
5. **Typing indicators / read receipts** — all three support some form of read state and typing actions
6. **Postback / interactive actions** — all three support button-triggered postback events
7. **Quick replies** — all three support ephemeral chips above the input

---

## Key Divergences That Matter for the Schema

| Concern | LINE | Facebook | Instagram |
|---|---|---|---|
| Auth model | Channel Secret (HMAC-SHA256, Base64) | App Secret (HMAC-SHA256, hex) | App Secret (HMAC-SHA256, hex) — same as FB |
| Reply mechanism | `replyToken` (one-shot, ~1 min) | Stateless send to PSID within 24h | Stateless send to IGSID within 24h |
| File send (outbound) | Yes | Yes | **No** — text link workaround only |
| Rich card type | Flex Message / Template | Generic/Button/Media/Receipt Template | Generic Template only (limited buttons) |
| Carousel | Yes (up to 10) | Yes (legacy, currently 1 element) | Yes (up to 10) |
| Proactive push | Yes (`push` endpoint, consumes quota) | Yes (`UPDATE`/`MESSAGE_TAG`) | Severely restricted — user must message first |
| Group/room context | Yes | No | No |
| Sticker | Yes (inbound + outbound) | Yes (inbound only, treated as image) | No |
| Location share | Yes inbound + outbound | Inbound only | No |

---

## Schema Architecture Recommendation

The unified schema needs **two layers**:

1. **Normalized layer** — platform-agnostic fields the CS UI always renders (sender, timestamp, text, attachment type/url, message ID). This powers the inbox display.
2. **Platform envelope** — the raw or lightly-typed platform-specific payload preserved alongside. This is what allows a CS agent to reply with platform-native features (LINE Flex, FB Template, etc.) without the hub needing to understand every variant.

### Unified Inbound Message Shape

```
UnifiedMessage {
  id:           string          // internal hub ID
  platform:     "line"|"facebook"|"instagram"
  channel_id:   string          // LINE: channelId / FB: pageId / IG: igUserId
  sender_id:    string          // LINE: userId / FB: PSID / IG: IGSID
  timestamp:    int64           // ms epoch, normalized

  direction:    "inbound"|"outbound"

  // Normalized content (for display)
  type:         "text"|"image"|"video"|"audio"|"file"|"location"|
                "sticker"|"postback"|"reaction"|"unsend"|"read"|
                "delivery"|"typing"|"system"
  text:         string?         // extracted text across all platforms
  attachments:  Attachment[]    // [{kind, url, mime_hint}]

  // Reply context
  reply_token:  string?         // LINE only, short-lived (~1 min)
  reply_to_mid: string?         // thread reply / quote reference

  // Platform-native payload (preserved for rich replies)
  raw:          json            // original webhook event object
}
```

### Unified Outbound Reply Shape

Outbound replies carry a `platform_message` field typed per-platform so a CS agent can send a LINE Flex message, a Facebook carousel, or an Instagram Generic Template exactly as each platform specifies — without the hub needing to re-model every variant.

```
OutboundMessage {
  platform:         "line"|"facebook"|"instagram"
  channel_id:       string
  recipient_id:     string          // userId / PSID / IGSID

  // Normalized hint (for logging / routing)
  type:             "text"|"image"|"video"|"audio"|"file"|
                    "template"|"flex"|"quick_reply"|"sender_action"

  // Platform-native send payload (passed through to platform API)
  platform_message: json            // LINE message object / FB-IG Send API body
}
```

---

## Reply Window Constraints

| Platform | Reactive reply | Proactive push | Out-of-window escape |
|---|---|---|---|
| LINE | Reply token (~1 min) or push (any time) | Yes — `push` endpoint, consumes monthly quota | No window concept; push always works |
| Facebook | RESPONSE within 24h of last user message | UPDATE within 24h; MESSAGE_TAG after | `HUMAN_AGENT` (7 days, human-sent only); Recurring Notifications (opt-in) |
| Instagram | RESPONSE within 24h of last user message | Not allowed — user must message first | `HUMAN_AGENT` (7 days, human-sent only); NOTIFICATION_MESSAGE (opt-in only) |

The CS hub must track `last_user_message_at` per conversation per platform to enforce this correctly.

---

## Signature Verification Differences

| Platform | Header | Algorithm | Encoding |
|---|---|---|---|
| LINE | `x-line-signature` | HMAC-SHA256(channel_secret, raw_body) | Base64 |
| Facebook | `X-Hub-Signature-256` | HMAC-SHA256(app_secret, raw_body) | `sha256=` + hex |
| Instagram | `X-Hub-Signature-256` | HMAC-SHA256(app_secret, raw_body) | `sha256=` + hex |

All require the **raw body bytes before JSON parsing**. Facebook/Instagram use identical verification logic.

---

## Webhook Envelope Differences

| Platform | Top-level key | Batching |
|---|---|---|
| LINE | `{ destination, events: [] }` | Multiple events per POST |
| Facebook | `{ object: "page", entry[].messaging[] }` | Multiple entries and events per POST |
| Instagram | `{ object: "instagram", entry[].messaging[] }` | Multiple entries per POST; messaging[] typically 1 item |

The hub must iterate all arrays correctly — a single POST can carry events from multiple users on Facebook/LINE.
