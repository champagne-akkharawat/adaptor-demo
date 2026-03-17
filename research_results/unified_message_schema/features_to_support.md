# Unified Message Schema — Features to Support

> **Date:** 2026-03-17
> **Usecase:** CS hub integration — reactive 1-on-1 customer service inbox with agent replies
>
> **Source documents:**
> - LINE webhook events: [platform_specifics/line/deep_research_2/webhook.md](../platform_specifics/line/deep_research_2/webhook.md)
> - LINE send API: [platform_specifics/line/deep_research_2/message_api.md](../platform_specifics/line/deep_research_2/message_api.md)
> - Instagram webhook events: [platform_specifics/instagram/deep_research/webhook.md](../platform_specifics/instagram/deep_research/webhook.md)
> - Instagram send API: [platform_specifics/instagram/deep_research/message_api.md](../platform_specifics/instagram/deep_research/message_api.md)
> - Facebook webhook events: [platform_specifics/facebook/deep_research/webhook_reference.md](../platform_specifics/facebook/deep_research/webhook_reference.md)
> - Facebook send API: [platform_specifics/facebook/deep_research/message_api.md](../platform_specifics/facebook/deep_research/message_api.md)

---

## Table of Contents

- [LINE](#line)
  - [Inbound (Webhook Events)](#inbound-webhook-events)
  - [Outbound (Send API)](#outbound-send-api)
- [Facebook](#facebook)
  - [Inbound (Webhook Events)](#inbound-webhook-events-1)
  - [Outbound (Send API)](#outbound-send-api-1)
- [Instagram](#instagram)
  - [Inbound (Webhook Events)](#inbound-webhook-events-2)
  - [Outbound (Send API)](#outbound-send-api-2)
- [Cross-Platform Requirements](#cross-platform-requirements)

---

## LINE

### Inbound (Webhook Events)

| Feature | Notes |
|---|---|
| `message` — text | Plain text from user |
| `message` — image | No URL in webhook; fetch binary via LINE Content API using `message.id` |
| `message` — video | No URL in webhook; fetch binary via LINE Content API; includes optional `duration` (ms) |
| `message` — audio | No URL in webhook; fetch binary via LINE Content API; includes `duration` (ms) |
| `message` — file | No URL in webhook; fetch binary via LINE Content API; includes `fileName`, `fileSize` |
| `message` — location | `title`, `address`, `latitude`, `longitude` |
| `message` — sticker | `packageId`, `stickerId`; content not fetchable — display as sticker indicator |
| `follow` | User adds the bot as a friend; open conversation in CS hub |
| `unfollow` | User blocks/removes the bot; close or archive conversation |
| `postback` | Button/quick reply action; carries `data` string and optional `params` |
| `unsend` | User deleted a message; redact local copy, show tombstone |

### Outbound (Send API)

| Feature | Notes |
|---|---|
| Reply — text | Via `replyToken` (one-shot, ~1 min validity) |
| Reply — image | Requires `originalContentUrl` + `previewImageUrl` (both HTTPS) |
| Reply — video | Requires `originalContentUrl` + `previewImageUrl` + `altText` |
| Reply — audio | Requires `originalContentUrl` + `duration` (ms) |
| Reply — file | Requires `originalContentUrl` + `fileName` |
| Reply — location | `title`, `address`, `latitude`, `longitude` |
| Reply — sticker | `packageId` + `stickerId` |
| Reply — Template (buttons) | Up to 4 action buttons; title + text + thumbnail optional |
| Reply — Template (confirm) | 2-button yes/no prompt |
| Reply — Template (carousel) | Up to 10 cards, each with image + title + text + buttons |
| Reply — Template (image carousel) | Up to 10 image-only cards with one action each |
| Reply — Flex Message | Free-form rich layout (bubble or carousel container) |
| Reply — Quick Reply chips | Ephemeral action chips above input (up to 13) |
| Push — any message type | Quota-consuming; used when reply token has expired or no reply token exists |
| Sender action — typing indicator | `chatting` action via LINE Chat Loading API |

---

## Facebook

### Inbound (Webhook Events)

| Feature | Notes |
|---|---|
| `messages` — text | Plain text from user |
| `messages` — image | `attachments[].payload.url` delivered inline |
| `messages` — video | `attachments[].payload.url` delivered inline |
| `messages` — audio | `attachments[].payload.url` delivered inline |
| `messages` — file | `attachments[].payload.url` delivered inline |
| `messages` — location | `attachments[].payload` with `coordinates.lat` + `coordinates.long` |
| `messages` — sticker | `attachments[].payload.sticker_id`; treat as image (URL provided) |
| `messaging_postbacks` | Button/quick reply action; carries `payload` string |
| `message_reads` | User read bot's message; update `read_at` on stored outbound — do not create new event |
| `message_deliveries` | Bot's message delivered; update `delivered_at` on stored outbound — do not create new event |
| `message_reactions` | User reacted to a message; attach as metadata to the reacted-to message |
| Unsend (`is_deleted: true`) | User deleted a message; redact local copy, show tombstone |

### Outbound (Send API)

| Feature | Notes |
|---|---|
| Send — text | `message.text`; `messaging_type: RESPONSE` within 24h |
| Send — image | `message.attachment` with `type: image` + `payload.url`; supports `is_reusable` |
| Send — video | `message.attachment` with `type: video` + `payload.url`; supports `is_reusable` |
| Send — audio | `message.attachment` with `type: audio` + `payload.url`; supports `is_reusable` |
| Send — file | `message.attachment` with `type: file` + `payload.url`; supports `is_reusable` |
| Send — Generic Template | Up to 10 cards; each with image, title, subtitle, url, up to 3 buttons |
| Send — Button Template | Text + up to 3 buttons; no image |
| Send — Media Template | Single image or video with up to 1 button |
| Send — Quick Reply chips | Ephemeral chips; up to 13; each with `content_type`, `title`, `payload` |
| Sender action — `typing_on` | Show typing indicator to user |
| Sender action — `typing_off` | Hide typing indicator |
| Sender action — `mark_seen` | Mark last message as seen |
| `messaging_type: HUMAN_AGENT` | Out-of-window escape; 7-day window; must be human-sent |

---

## Instagram

### Inbound (Webhook Events)

| Feature | Notes |
|---|---|
| `messages` — text | Plain text from user |
| `messages` — image | `attachments[].payload.url` delivered inline |
| `messages` — video | `attachments[].payload.url` delivered inline |
| `messages` — audio | `attachments[].payload.url` delivered inline |
| `messaging_postbacks` | Button/quick reply action; carries `payload` string |
| `messaging_seen` | User read bot's message; update `read_at` on stored outbound — do not create new event |
| `reaction` | User reacted to a message; attach as metadata to the reacted-to message |
| Unsend (`is_deleted: true`) | User deleted a message; redact local copy, show tombstone |
| Story mention | Inbound reference to user's story; URL expires; surface as informational note only — do not store media per Meta policy |
| Ephemeral / view-once | `type: "ephemeral"` — no content delivered; log as system event, no CS action |

### Outbound (Send API)

| Feature | Notes |
|---|---|
| Send — text | `message.text`; `messaging_type: RESPONSE` within 24h |
| Send — image | `message.attachment` with `type: image` + `payload.url` |
| Send — video | `message.attachment` with `type: video` + `payload.url` |
| Send — audio | `message.attachment` with `type: audio` + `payload.url` |
| Send — file | **Not supported** — use a text message with a link as workaround |
| Send — Generic Template | Up to 10 cards; `web_url` + `postback` button types only (no call/share) |
| Send — Quick Reply chips | Ephemeral chips; up to 13 |
| Sender action — `typing_on` | Show typing indicator |
| Sender action — `typing_off` | Hide typing indicator |
| Private Reply | Reply to a user comment or story mention as a DM |
| `messaging_type: HUMAN_AGENT` | Out-of-window escape; 7-day window; must be human-sent |

---

## Cross-Platform Requirements

| Requirement | Detail |
|---|---|
| Webhook signature verification | LINE: `x-line-signature` HMAC-SHA256 + Base64; FB/IG: `X-Hub-Signature-256` HMAC-SHA256 + `sha256=` hex. Raw body bytes required before JSON parsing. |
| Webhook batch iteration | One POST may carry multiple events (LINE `events[]`; FB `entry[].messaging[]`; IG `entry[].messaging[]`). Must iterate all arrays. |
| Sender identity storage | Store `userId` (LINE), `PSID` (FB), `IGSID` (IG) — all platform-scoped, stored per-platform per-account |
| Reply window enforcement | Track `last_user_message_at` per conversation. FB/IG: 24h window for `RESPONSE`; 7-day window with `HUMAN_AGENT`. LINE: no window. |
| Raw payload preservation | Store original webhook event as `raw` JSON alongside normalized fields to enable platform-native rich replies |
| Message status updates | Read receipts, delivery receipts — update status on stored outbound message; do not create new conversation events |
| Reaction metadata | Attach reaction to the reacted-to message as metadata; do not create a new conversation event |
| Unsend / delete handling | Redact message content locally and show a tombstone marker per platform policy |
