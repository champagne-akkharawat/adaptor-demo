# LINE Messaging API — Sending a Text Message

## Overview

LINE offers two primary patterns for sending messages to users:

- **Reply** — respond to a user-initiated event (message, postback, etc.) using a one-time `replyToken` delivered via webhook. Free of charge and strongly preferred for reactive bots.
- **Push** — proactively send a message to any user, group, or room at any time using the recipient's ID. Counts against monthly messaging quotas.

Both patterns share the same message object format; only the endpoint and top-level request body differ.

---

## Endpoints

| Pattern | Method | URL |
|---------|--------|-----|
| Reply   | POST   | `https://api.line.me/v2/bot/message/reply` |
| Push    | POST   | `https://api.line.me/v2/bot/message/push` |
| Multicast (multiple users) | POST | `https://api.line.me/v2/bot/message/multicast` |
| Broadcast (all friends)    | POST | `https://api.line.me/v2/bot/message/broadcast` |
| Narrowcast (segmented)     | POST | `https://api.line.me/v2/bot/message/narrowcast` |

For basic text messaging, only **reply** and **push** are needed.

---

## Required Headers

```
Authorization: Bearer {CHANNEL_ACCESS_TOKEN}
Content-Type: application/json
```

- The channel access token is issued per-channel in the LINE Developers Console under **Messaging API > Channel access token**.
- For idempotent push/multicast/broadcast/narrowcast requests, you may also include:
  ```
  X-Line-Retry-Key: {UUID}
  ```
  This ensures the same message is not sent twice if the network times out and you retry.

---

## Request Body — Reply Message

Use this when responding to a webhook event. The `replyToken` is provided in the incoming webhook payload and is valid for a short time (typically under 60 seconds).

```json
{
  "replyToken": "nHuyWiB7yP5Zw52FIkcQobQuGDXCTA",
  "messages": [
    {
      "type": "text",
      "text": "Hello! How can I help you?"
    }
  ],
  "notificationDisabled": false
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `replyToken` | string | Yes | One-time token from the webhook event |
| `messages` | array | Yes | 1–5 message objects |
| `messages[].type` | string | Yes | `"text"` for plain text |
| `messages[].text` | string | Yes | The text content to send |
| `notificationDisabled` | boolean | No | If `true`, suppresses push notification on recipient's device. Default: `false` |

---

## Request Body — Push Message

Use this to initiate a message without a preceding user action.

```json
{
  "to": "U4af4980629...",
  "messages": [
    {
      "type": "text",
      "text": "Hello, world!"
    }
  ],
  "notificationDisabled": false
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `to` | string | Yes | User ID, group ID, or room ID of the recipient |
| `messages` | array | Yes | 1–5 message objects |
| `messages[].type` | string | Yes | `"text"` for plain text |
| `messages[].text` | string | Yes | The text content to send |
| `notificationDisabled` | boolean | No | Suppresses device notification. Default: `false` |
| `customAggregationUnits` | array of strings | No | For aggregating statistics; max 1 unit name |

---

## Example — curl (Reply)

```bash
curl -X POST https://api.line.me/v2/bot/message/reply \
  -H "Authorization: Bearer YOUR_CHANNEL_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "replyToken": "nHuyWiB7yP5Zw52FIkcQobQuGDXCTA",
    "messages": [
      {
        "type": "text",
        "text": "Got your message!"
      }
    ]
  }'
```

## Example — curl (Push)

```bash
curl -X POST https://api.line.me/v2/bot/message/push \
  -H "Authorization: Bearer YOUR_CHANNEL_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "U206d25c2ea6bd87c17655609a1c37cb8",
    "messages": [
      {
        "type": "text",
        "text": "Hello from the bot!"
      }
    ]
  }'
```

---

## Successful Response (200 OK)

Both endpoints return the same response shape:

```json
{
  "sentMessages": [
    {
      "id": "461704574915325955",
      "quoteToken": "I042asfA8..."
    }
  ]
}
```

- `id` — the message ID assigned by LINE
- `quoteToken` — can be used to quote this message in a follow-up

---

## Error Responses

| Status | Meaning |
|--------|---------|
| 400 | Bad Request — malformed body or invalid field values |
| 403 | Forbidden — channel access token lacks permission |
| 409 | Conflict — duplicate request (retry key already used) |
| 429 | Too Many Requests — rate limit exceeded |

---

## Key Notes

### Reply vs. Push — Which to Use

| Consideration | Reply | Push |
|---------------|-------|------|
| Triggered by user action | Yes (required) | Not required |
| Token lifetime | Short (use immediately) | N/A — uses persistent user ID |
| Quota cost | Free | Counts against monthly limit |
| Use for proactive outreach | No | Yes |

### Rate Limits and Quotas

- Push messages count against the monthly free messaging quota for the channel plan.
- Check current quota and consumption via:
  - `GET https://api.line.me/v2/bot/message/quota` — total monthly limit
  - `GET https://api.line.me/v2/bot/message/quota/consumption` — messages sent so far
- Exceeding quota returns a `429` response.
- Reply messages do **not** count toward the quota.

### Messages Per Request

- A single API call can include up to **5 message objects** in the `messages` array.
- Each message object in the array is delivered as a separate chat bubble.

### Channel Types

- Only **Messaging API channels** support these endpoints. LINE Login channels and LIFF apps do not have access to the bot messaging endpoints.
- The channel access token must belong to the same channel as the target bot.

### Text Message Constraints

- `text` field: plain UTF-8 string. LINE does not enforce a strict character limit in the API spec, but the LINE app UI truncates very long messages; keep texts under 5,000 characters to be safe.
- Newlines (`\n`) are supported and render as line breaks in the chat.
- Basic LINE emoji can be embedded using the `emojis` field (an extension to the text message object not covered here).
