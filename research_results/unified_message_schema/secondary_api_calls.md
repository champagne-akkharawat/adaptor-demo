# Secondary API Calls — Receiving Full Message Content

> **Date:** 2026-03-17
> **Usecase:** CS hub integration — additional calls required after a webhook event to retrieve complete content

These are API calls that must (or should) be made after receiving a webhook, because the webhook payload alone does not contain the full message content.

---

## Table of Contents

- [LINE](#line)
- [Facebook](#facebook)
- [Instagram](#instagram)
- [Cross-Platform Summary](#cross-platform-summary)
  - [Key Rules](#key-rules)

---

## LINE

All LINE-hosted media arrives in the webhook with a `message.id` only. The actual bytes must be fetched separately. External content providers (`contentProvider.type: "external"`) include a URL directly in the webhook and do not require a secondary call.

| Trigger | Endpoint | Returns | Notes |
|---|---|---|---|
| `message.type: "image"` and `contentProvider.type: "line"` | `GET /v2/bot/message/{messageId}/content` | Full-size image binary | Download promptly — content expires |
| Same | `GET /v2/bot/message/{messageId}/content/preview` | Thumbnail image binary | Optional but useful for inbox preview |
| `message.type: "video"` and `contentProvider.type: "line"` | `GET /v2/bot/message/{messageId}/content/transcoding` | `{ status: "processing"\|"succeeded"\|"failed" }` | **Must check before downloading** — video may still be transcoding |
| Same (after status is `succeeded`) | `GET /v2/bot/message/{messageId}/content` | Video file binary | Download promptly — content expires |
| Same | `GET /v2/bot/message/{messageId}/content/preview` | Video thumbnail binary | Useful for inbox preview |
| `message.type: "audio"` and `contentProvider.type: "line"` | `GET /v2/bot/message/{messageId}/content` | Audio file binary | Download promptly — content expires |
| `message.type: "file"` | `GET /v2/bot/message/{messageId}/content` | File binary (PDF, DOCX, XLSX, etc.) | Webhook provides `fileName` + `fileSize` only; always LINE-hosted |
| `message.type: "sticker"` | Construct CDN URL (no API call) | See URL patterns below | Sticker image or animation | All metadata (`packageId`, `stickerId`, `stickerResourceType`, `keywords`) is in the webhook. Build the URL directly from the IDs. |

**LINE Sticker CDN URL patterns** (no auth required):

| `stickerResourceType` | Image URL |
|---|---|
| `STATIC` | `https://stickershop.line-scdn.net/stickershop/v1/sticker/{stickerId}/android/sticker.png` |
| `ANIMATION` | `https://stickershop.line-scdn.net/stickershop/v1/sticker/{stickerId}/android/sticker_animation.png` |
| `SOUND` | `https://stickershop.line-scdn.net/stickershop/v1/sticker/{stickerId}/android/sticker.png` (static frame) |
| `ANIMATION_SOUND` | `https://stickershop.line-scdn.net/stickershop/v1/sticker/{stickerId}/android/sticker_animation.png` |
| `POPUP` / `POPUP_SOUND` | `https://stickershop.line-scdn.net/stickershop/v1/sticker/{stickerId}/android/sticker_popup.png` |

Package thumbnail (for sticker pack preview):
`https://stickershop.line-scdn.net/stickershop/v1/product/{packageId}/android/sticker_tab.png`
| `message.quotedMessageId` present | **Not retrievable** | — | The quoted message ID is provided but LINE has no API to fetch a past message by ID. Store the content when originally received. |

**Auth:** All `/v2/bot/message/*` calls require `Authorization: Bearer {channel_access_token}`.

---

## Facebook

Facebook delivers media as pre-signed CDN URLs directly in the webhook payload. No Graph API call is needed to get the binary — but URLs expire so they must be downloaded immediately. Quoted message content is the one case requiring a Graph API call.

| Trigger | Action | Endpoint | Returns | Notes |
|---|---|---|---|---|
| `attachments[].type: "image/video/audio/file"` | Download binary | `GET {payload.url}` (CDN) | Media binary | No auth needed — pre-signed URL. **Expires — download immediately.** |
| `attachments[].type: "image"` with `payload.sticker_id` present | Download binary | `GET {payload.url}` (CDN) | Sticker image | Same CDN download; `sticker_id` is a persistent identifier for display logic |
| `message.reply_to.mid` present | Fetch quoted message | `GET /{message-id}?fields=from,message,created_time,attachments` | Original message text/attachment | Requires Page Access Token + `pages_messaging` permission; message must not have been deleted |
| `message_reactions` event received (`reaction.mid`) | Look up reacted-to message | Local store lookup by `mid`; fallback: `GET /{message-id}?fields=from,message,created_time` | Original message to attach reaction to | **Prefer local store** — avoid Graph API call if message was already received and stored. Graph API fallback only if message is missing locally. |

---

## Instagram

Instagram follows the same CDN URL pattern as Facebook for media. Quoted replies and certain referral events require Graph API calls.

| Trigger | Action | Endpoint | Returns | Notes |
|---|---|---|---|---|
| `attachments[].type: "image/video/audio"` | Download binary | `GET {payload.url}` (CDN) | Media binary | **Expires — download immediately.** No file type supported on IG. |
| `attachments[].type: "story_mention"` | Download story media | `GET {payload.url}` (CDN) | Story image/video | URL expires; **do not store media** per Meta policy — store URL reference only |
| Same (optional extra context) | Fetch story metadata | `GET /{message-id}?fields=story` | Story metadata | Optional; use for displaying story context in CS inbox |
| `message.reply_to.story` present | Download story media | `GET {reply_to.story.url}` (CDN) | Story being replied to | URL expires; private accounts may omit `url` entirely |
| `message.reply_to.mid` present | Fetch quoted message | `GET /{message-id}?fields=from,message,created_time,attachments` | Original message content | Same as Facebook — requires appropriate Graph API permission |
| `attachments[].type: "ephemeral"` | **Not retrievable** | — | — | View-once media; no URL is provided by design |
| `message.referral.product.id` present | Fetch product details | `GET /{product-id}?fields=name,description,image_url,price,currency` | Product catalog entry | Requires catalog access permission |
| `reaction` event received (`reaction.mid`) | Look up reacted-to message | Local store lookup by `mid`; fallback: `GET /{message-id}?fields=from,message,created_time` | Original message to attach reaction to | **Prefer local store** — same pattern as Facebook. `reaction.reaction` + `reaction.emoji` + `reaction.action` (react/unreact) are all in the webhook. |

---

## Cross-Platform Summary

| Platform | Media binary | Quoted message content | Sticker image |
|---|---|---|---|
| **LINE** | `GET /v2/bot/message/{id}/content` (no URL in webhook) | Not retrievable via API — must be stored on first receipt | Rendered from LINE CDN using `packageId`/`stickerId`; no separate fetch call |
| **Facebook** | Download from `payload.url` in webhook (expires) | `GET /{message-id}?fields=...` via Graph API | Same as image — URL in webhook |
| **Instagram** | Download from `payload.url` in webhook (expires) | `GET /{message-id}?fields=...` via Graph API | Not applicable — no sticker support |

### Key Rules

1. **Download media immediately** — FB/IG CDN URLs are pre-signed and expire. Treat the webhook as a trigger to enqueue a download job, not a durable media store.
2. **Check LINE video transcoding status first** — calling `/content` before transcoding completes returns an error. Poll `/content/transcoding` until `status: "succeeded"`.
3. **LINE quoted messages are not fetchable** — the only way to show quoted context is to have stored the original message when it was first received. Do not rely on a secondary API call.
4. **IG story media must not be stored** — per Meta policy, only the URL reference may be stored, not the media binary itself.
5. **Stickers on LINE are metadata-only** — `packageId` + `stickerId` identify the sticker; the image is served from LINE's public CDN and does not require an authenticated fetch.
6. **Reactions do not require an API call for their own content** — the webhook carries the full reaction data (emoji, action, reacted-to `mid`). The only secondary step is finding which stored message to attach it to; prefer a local DB lookup over a Graph API call.
