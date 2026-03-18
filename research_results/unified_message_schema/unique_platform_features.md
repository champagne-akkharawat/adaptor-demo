# Unique Platform Features

Features that are genuinely unique to a specific platform — not different implementations of the same concept.

---

## LINE

### Flex Messages
Custom JSON-defined UI component system for rich interactive cards. Supports complex layouts (boxes, hero images, buttons, carousels) with its own internal versioning (`"version": "1"` or `"2"`). Outbound-only — users cannot send flex messages to the bot. Has no cross-platform equivalent.

### imageSet — Multi-image Grouping
When a user sends multiple images together, LINE attaches an `imageSet` object to each image event with a shared `id`, an `index`, and a `total` count. This allows reliable grouping of images from the same send batch without time-window heuristics.

```json
"imageSet": {
  "id": "abc123",
  "index": 1,
  "total": 3
}
```

### No Media URL in Webhook (messageId-only Content Fetching)
LINE never includes a media URL in the webhook payload for image/video/audio/file messages. The webhook only provides a `messageId`. Content must be fetched separately via `GET /v2/bot/message/{messageId}/content`. This makes lazy download the only option by design, and `messageId` is permanent.

### replyToken
LINE issues a short-lived `replyToken` (~30 second validity) per incoming event. Replies sent via this token are free-of-charge quota-wise. Responding outside the window requires a push message (quota-counted). No equivalent exists on Facebook or Instagram.

---

## Facebook

### attachment_id — Reusable Permanent Asset ID
When a user sends a media attachment, Facebook includes a permanent `attachment_id` alongside the expiring CDN URL (~1 hour TTL). This ID can be used indefinitely to re-fetch a fresh signed URL via the Graph API, and can also be reused to send the same asset outbound without re-uploading.

### One-Time Notification (OTN)
Allows a page to send a single follow-up message to a user outside the standard 24-hour messaging window, after the user explicitly opts in via a notification request button. No direct equivalent on Instagram or LINE.

### HUMAN_AGENT Messaging Type (7-day window)
`messaging_type: HUMAN_AGENT` unlocks a 7-day reply window for human-agent interactions, beyond the standard 24-hour `RESPONSE` window. Shared with Instagram but initiated differently.

---

## Instagram

### Story Mentions
Instagram delivers a webhook event when a user mentions the account in their Story (`story_mention` attachment type). This is a passive inbound signal — the user did not DM the account directly. Per Meta policy, the story media URL should not be stored. No equivalent on Facebook Messenger or LINE.

### Ephemeral / View-Once Messages
Instagram delivers a `type: "ephemeral"` event for view-once media — no content is included in the payload by design. No equivalent on Facebook Messenger or LINE.

### Private Reply
Allows the bot to reply to a user's comment or story mention as a DM. Bridges public engagement and private messaging. No equivalent on Facebook Messenger or LINE.

---

## Meta (Facebook & Instagram Shared)

### Explicit API Versioning with Deprecation Cycle
Meta releases new Graph API versions every ~6 months (e.g., v18.0, v19.0, v20.0) and deprecates old versions after ~2 years. Webhook subscriptions are tied to the version used at subscription time. No version identifier is included in the webhook payload itself — version must be tracked externally by the integrator.
