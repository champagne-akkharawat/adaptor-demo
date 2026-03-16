# Slack — Sending a Text Message

## Endpoint

```
POST https://slack.com/api/chat.postMessage
```

## Authentication

| Item | Value |
|---|---|
| Header | `Authorization: Bearer <token>` |
| Token types | Bot token (`xoxb-…`) or User token (`xoxp-…`) |
| Required scope | `chat:write` |
| Public channels (new apps) | Also requires `chat:write.public` |
| Custom username / icon | Also requires `chat:write.customize` |

The token must be passed either in the `Authorization` header (preferred) or as the `token` field in the request body.

## Request Body

Content-Type: `application/json; charset=utf-8`

### Required fields

| Field | Type | Description |
|---|---|---|
| `channel` | string | Channel ID, private group ID, or IM ID (e.g. `C123ABC456`) |
| `text` | string | Message body. Required when `blocks` is absent; acts as fallback notification text when `blocks` is present. Max 4,000 chars (hard truncation at 40,000). |

### Commonly used optional fields

| Field | Type | Description |
|---|---|---|
| `blocks` | array | Block Kit layout blocks (JSON-encoded). When present, `text` becomes the notification-only fallback. |
| `thread_ts` | string | Timestamp of the parent message to post as a threaded reply. |
| `reply_broadcast` | boolean | When `true`, also shows the threaded reply in the channel. Default: `false`. |
| `mrkdwn` | boolean | Enable Slack markdown parsing. Default: `true`. |
| `unfurl_links` | boolean | Enable link unfurling. |
| `unfurl_media` | boolean | Enable media unfurling. Default: `true`. |
| `username` | string | Override the bot display name (requires `chat:write.customize`). |
| `icon_emoji` | string | Override the bot icon with an emoji (requires `chat:write.customize`). |
| `icon_url` | string | Override the bot icon with an image URL (requires `chat:write.customize`). |

## Example Request

```bash
curl -X POST https://slack.com/api/chat.postMessage \
  -H "Authorization: Bearer xoxb-your-bot-token" \
  -H "Content-Type: application/json; charset=utf-8" \
  -d '{
    "channel": "C123ABC456",
    "text": "Hello from the integration!"
  }'
```

## Example Success Response

```json
{
  "ok": true,
  "channel": "C123ABC456",
  "ts": "1503435956.000247",
  "message": {
    "text": "Hello from the integration!",
    "username": "my-bot",
    "bot_id": "B123ABC456",
    "type": "message",
    "ts": "1503435956.000247"
  }
}
```

`ok: false` responses include an `error` string field explaining the failure.

## Rate Limits

- 1 message per second per channel (burst allowance provided).
- Workspace-level cap of several hundred messages per minute.
- Slack uses a tiered rate-limit system; `chat.postMessage` is in the Special tier.

## Common Errors

| Error | Cause |
|---|---|
| `channel_not_found` | Invalid or inaccessible channel value |
| `missing_scope` | Token lacks required `chat:write` scope |
| `not_in_channel` | Bot is not a member of the channel |
| `rate_limited` | Rate limit exceeded |
| `invalid_blocks` | Malformed Block Kit JSON |
| `too_many_attachments` | Exceeded 100 attachments per message |

## Key Notes

- Always use channel **IDs** (e.g. `C123ABC456`), not display names, to avoid `channel_not_found` errors when channels are renamed.
- Bot users cannot post to user-to-user DMs; they can only post to app-initiated or multi-person DMs.
- For simple text messages, omit `blocks` entirely and put content directly in `text`.
- The `as_user` parameter is a legacy option; new apps should use bot tokens and ignore it.
- The `ts` value in the response is used as `thread_ts` when posting a reply to this message.
