# Inbound Prototype Plan

## Naming Convention

Files follow the pattern: `inbound_{message_type}_{provider}[_{variant}].json`

- `message_type`: `text`, `image`, `video`, `audio`, `file`, `sticker`, `quoted`, `location`
- `provider`: `line`, `facebook`, `instagram`
- `variant` (optional): `external_provider`, `attachment_id`, `static`, `popup`, `sound`, `animation_sound`, `story_mention`

## All 33 Prototype Files

| # | Filename | Provider | Message Type | Variant |
|---|----------|----------|--------------|---------|
| 1 | `inbound_text_line.json` | line | text | (existing) |
| 2 | `inbound_text_facebook.json` | facebook | text | |
| 3 | `inbound_text_instagram.json` | instagram | text | |
| 4 | `inbound_image_line.json` | line | image | content API (LINE-hosted) |
| 5 | `inbound_image_line_external_provider.json` | line | image | external provider URL |
| 6 | `inbound_image_facebook.json` | facebook | image | CDN URL only |
| 7 | `inbound_image_facebook_attachment_id.json` | facebook | image | with attachment_id |
| 8 | `inbound_image_instagram.json` | instagram | image | |
| 9 | `inbound_image_instagram_story_mention.json` | instagram | image | story_mention type |
| 10 | `inbound_video_line.json` | line | video | content API (LINE-hosted) |
| 11 | `inbound_video_line_external_provider.json` | line | video | external provider URL |
| 12 | `inbound_video_facebook.json` | facebook | video | CDN URL only |
| 13 | `inbound_video_facebook_attachment_id.json` | facebook | video | with attachment_id |
| 14 | `inbound_video_instagram.json` | instagram | video | |
| 15 | `inbound_audio_line.json` | line | audio | content API (LINE-hosted) |
| 16 | `inbound_audio_facebook.json` | facebook | audio | CDN URL only |
| 17 | `inbound_audio_facebook_attachment_id.json` | facebook | audio | with attachment_id |
| 18 | `inbound_audio_instagram.json` | instagram | audio | |
| 19 | `inbound_file_line.json` | line | file | content API (LINE-hosted) |
| 20 | `inbound_file_facebook.json` | facebook | file | CDN URL only |
| 21 | `inbound_file_facebook_attachment_id.json` | facebook | file | with attachment_id |
| 22 | `inbound_sticker_line.json` | line | sticker | ANIMATION |
| 23 | `inbound_sticker_line_static.json` | line | sticker | STATIC |
| 24 | `inbound_sticker_line_popup.json` | line | sticker | POPUP |
| 25 | `inbound_sticker_line_sound.json` | line | sticker | SOUND |
| 26 | `inbound_sticker_line_animation_sound.json` | line | sticker | ANIMATION_SOUND |
| 27 | `inbound_sticker_facebook.json` | facebook | sticker | |
| 28 | `inbound_quoted_line.json` | line | text | with reply_to |
| 29 | `inbound_quoted_facebook.json` | facebook | text | with reply_to |
| 30 | `inbound_quoted_instagram.json` | instagram | text | with reply_to |
| 31 | `inbound_location_line.json` | line | location | |
| 32 | `inbound_location_facebook.json` | facebook | location | |
| 33 | _(total: 32 new + 1 existing = 33)_ | | | |

## Excluded Event Types

The following LINE webhook event types are out of scope for inbound message prototypes (they are system/control events, not user messages):

- `follow` — user follows the channel
- `unfollow` — user blocks the channel
- `join` — bot joins a group/room
- `leave` — bot leaves a group/room
- `postback` — user taps a postback action button
- `beacon` — beacon device trigger
- `accountLink` — account linking result
- `memberJoined` / `memberLeft` — group membership changes
- `things` — LINE Things device events
- `unsend` — user unsends a message (no content)

Facebook/Instagram equivalents excluded:
- `delivery` — message delivery confirmation
- `read` — message read receipt
- `postback` — quick reply / button postback
- `referral` — referral link click
- `optin` — opt-in / checkbox plugin
- `reaction` — emoji reaction to a message

## Special Cases

### LINE Content API (contentProvider.type: "line")

For LINE image, video, audio, and file messages where LINE hosts the content, the binary cannot be retrieved from the webhook payload itself. The adaptor must call the LINE Content API to download the file:

```
GET https://api-data.line.me/v2/bot/message/{messageId}/content
```

In the prototype, `permanent_file_url` is set to:
```
"null (LINE Content API required: GET /v2/bot/message/{id}/content)"
```

### LINE External Provider (contentProvider.type: "external")

When a LINE user shares media originally hosted externally, the webhook includes `originalContentUrl` (and optionally `previewImageUrl`). The adaptor can use the URL directly.

In the prototype, `permanent_file_url` is set to the `originalContentUrl` value from the webhook.

### Facebook / Instagram CDN URL (no attachment_id)

Facebook and Instagram media URLs are temporary CDN links that expire. The adaptor must download the file immediately on receipt.

In the prototype, `permanent_file_url` is set to:
```
"null (download from CDN URL before it expires)"
```

### Facebook attachment_id

When Facebook sends a reusable media attachment, it includes an `attachment_id` that can be used to re-fetch the media later via the Graph API. The `attachment_id` is surfaced in `meta_data.content`.

In the prototype, `permanent_file_url` is set to:
```
"null (download from CDN URL or re-fetch via attachment_id)"
```

### Instagram Story Mention

When a user mentions the business account in their Instagram Story, the webhook delivers an attachment with `type: "story_mention"`. The URL is ephemeral (expires ~24 hours).

- `message_type` is set to `"image"` (the closest content-type equivalent)
- `file_attachments[0].type` is `"story_mention"`
- `permanent_file_url` is set to:
  ```
  "null (ephemeral story URL — download immediately on receipt, expires ~24h)"
  ```

### LINE Sticker CDN URLs

Sticker preview images are publicly accessible on the LINE sticker CDN. The URL pattern depends on `stickerResourceType`:

| stickerResourceType | URL pattern |
|---------------------|-------------|
| `STATIC` | `https://stickershop.line-scdn.net/stickershop/v1/sticker/{stickerId}/android/sticker.png` |
| `ANIMATION` | `https://stickershop.line-scdn.net/stickershop/v1/sticker/{stickerId}/android/sticker_animation.png` |
| `POPUP` | `https://stickershop.line-scdn.net/stickershop/v1/sticker/{stickerId}/android/sticker_popup.png` |
| `SOUND` | `https://stickershop.line-scdn.net/stickershop/v1/sticker/{stickerId}/android/sticker.png` |
| `ANIMATION_SOUND` | `https://stickershop.line-scdn.net/stickershop/v1/sticker/{stickerId}/android/sticker_animation.png` |

### Facebook Stickers

Facebook delivers stickers as image attachments with an additional `sticker_id` field. The sticker URL comes directly from `attachments[0].payload.url`.

### Quoted / Reply Messages

All providers surface a reference to the original message being replied to. The internal `reply_to.message_id` cannot be resolved at inbound-parse time (it requires a database lookup), so it is set to `"null (to be filled later)"`. The raw provider reference is preserved in `reply_to.meta_data.content`.
