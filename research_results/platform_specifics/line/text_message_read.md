# LINE Messaging API — Receiving a Text Message (Webhooks)

## Overview

LINE delivers incoming messages by making an **HTTP POST request from the LINE Platform to your server** at a URL you configure. This is called a webhook. Your server must:

1. Accept HTTPS POST requests at the registered webhook URL.
2. Verify the `x-line-signature` header to confirm the request is authentic.
3. Parse the JSON body to extract event data.
4. Respond with HTTP `200 OK` quickly (within a few seconds) — LINE retries on failure.

---

## Webhook Setup

### Step 1 — Register Your Webhook URL

1. Open the [LINE Developers Console](https://developers.line.biz/console/).
2. Select your channel (Messaging API channel).
3. Go to the **Messaging API** tab.
4. Under **Webhook settings**, enter your HTTPS endpoint URL.
5. Toggle **Use webhook** to enabled.
6. Use the **Verify** button to confirm LINE can reach your server.

Requirements for the webhook URL:
- Must be **HTTPS** (TLS 1.2+).
- Must use port **443** or **80** (443 strongly recommended).
- Self-signed certificates are not accepted; use a trusted CA certificate.
- Must respond to LINE's verification POST with HTTP `200`.

### Step 2 — Optional: Enable Webhook Redelivery

In the same console section, you can enable **Webhook redelivery**. If your server fails to respond with `200 OK`, LINE will retry delivery automatically.

---

## Webhook Signature Verification

Before processing any webhook payload, **always verify the signature** to confirm the request came from LINE and was not tampered with.

### How It Works

1. LINE computes an HMAC-SHA256 hash of the raw request body using your **channel secret** as the key.
2. The resulting digest is Base64-encoded and placed in the `x-line-signature` request header.
3. Your server independently computes the same hash and compares it to the header value.

### What You Need

- **Channel secret** — found in the LINE Developers Console under **Basic settings**.
- **Raw request body** — the exact bytes received, before any parsing or transformation.

### Verification Algorithm

```
signature = Base64( HMAC-SHA256( channelSecret, requestBodyBytes ) )
```

Compare `signature` to the value of the `x-line-signature` header. If they match, the request is authentic.

### Critical Rules

- Use the **raw, unmodified** request body. Do NOT parse JSON first and re-serialize.
- Use **UTF-8** encoding throughout.
- Do NOT interpret escape characters — treat them literally.
- Reject requests where the signatures do not match.

### Example Verification (openssl CLI)

```bash
echo -n '{"destination":"U8e742f61d673b39c7fff3cecb7536ef0","events":[]}' \
  | openssl dgst -sha256 -hmac 'YOUR_CHANNEL_SECRET' -binary \
  | openssl base64
```

The output should match the `x-line-signature` header value.

### Example Verification (Node.js)

```javascript
const crypto = require('crypto');

function verifySignature(channelSecret, body, signature) {
  const hmac = crypto.createHmac('SHA256', channelSecret);
  hmac.update(body); // body must be the raw Buffer or string, not parsed JSON
  const computed = hmac.digest('base64');
  return computed === signature;
}

// In your route handler:
const rawBody = req.rawBody; // must be captured before JSON parsing
const signature = req.headers['x-line-signature'];
if (!verifySignature(process.env.LINE_CHANNEL_SECRET, rawBody, signature)) {
  return res.status(401).send('Invalid signature');
}
```

---

## Webhook Request Structure

LINE sends a POST request with `Content-Type: application/json`. The body always has this top-level shape:

```json
{
  "destination": "Uxxxxxxxxxxxxxxxxx",
  "events": [ ... ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `destination` | string | The user ID of your LINE Official Account (the bot receiving events) |
| `events` | array | One or more event objects |

---

## Text Message Event Object

A text message from a user produces a `message` event with `message.type = "text"`.

### Full Example Payload

```json
{
  "destination": "U8e742f61d673b39c7fff3cecb7536ef0",
  "events": [
    {
      "type": "message",
      "mode": "active",
      "timestamp": 1462629479859,
      "webhookEventId": "01FZ744A37Q019KEYBOARDBADN",
      "deliveryContext": {
        "isRedelivery": false
      },
      "replyToken": "nHuyWiB7yP5Zw52FIkcQT",
      "source": {
        "type": "user",
        "userId": "U206d25c2ea6bd87c17655609a1c37cb8"
      },
      "message": {
        "id": "100001",
        "type": "text",
        "text": "Hello, this is a text message",
        "quoteToken": "nHuyWiBzyw52FIkcQT"
      }
    }
  ]
}
```

---

## Field Reference

### Event Object Fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Always `"message"` for incoming messages |
| `mode` | string | `"active"` (normal) or `"standby"` (bot in standby in a group) |
| `timestamp` | number | Unix time in **milliseconds** when the event occurred |
| `webhookEventId` | string | Unique ID for this event delivery; use to detect duplicates on redelivery |
| `deliveryContext.isRedelivery` | boolean | `true` if this is a retry of a previously failed delivery |
| `replyToken` | string | One-time token used to send a reply via `POST /v2/bot/message/reply`; valid for a short window (~60 s) |

### Source Object — Sender Identification

The `source` object identifies where the message came from:

**Direct message (1:1 chat):**
```json
{
  "type": "user",
  "userId": "U206d25c2ea6bd87c17655609a1c37cb8"
}
```

**Group chat:**
```json
{
  "type": "group",
  "groupId": "C4af4980629...",
  "userId": "U206d25c2ea6bd87c17655609a1c37cb8"
}
```

**Multi-person chat (room):**
```json
{
  "type": "room",
  "roomId": "Ra8dbf32b45...",
  "userId": "U206d25c2ea6bd87c17655609a1c37cb8"
}
```

| Field | Description |
|-------|-------------|
| `source.type` | `"user"`, `"group"`, or `"room"` |
| `source.userId` | Stable, channel-scoped ID for the sending user — use this to identify the sender |
| `source.groupId` | Present when `type = "group"` |
| `source.roomId` | Present when `type = "room"` |

### Message Object Fields

| Field | Type | Description |
|-------|------|-------------|
| `message.id` | string | Unique message ID assigned by LINE |
| `message.type` | string | `"text"` for plain text messages |
| `message.text` | string | The actual text content sent by the user |
| `message.quoteToken` | string | Token you can pass when quoting this message in a reply |
| `message.mention` | object | Present if users are @-mentioned; contains `mentionees[]` with user details |

---

## Extracting Key Fields — Summary

To handle an incoming text message, extract these fields from the payload:

| Data | JSON Path |
|------|-----------|
| Sender user ID | `events[0].source.userId` |
| Message text | `events[0].message.text` |
| Timestamp (ms) | `events[0].timestamp` |
| Reply token | `events[0].replyToken` |
| Message ID | `events[0].message.id` |
| Is redelivery | `events[0].deliveryContext.isRedelivery` |
| Chat context | `events[0].source.type` + `groupId` / `roomId` if applicable |

---

## Responding to the Webhook

Your HTTP server must return `200 OK` promptly. LINE will treat any non-2xx response (or a timeout) as a failure and may retry delivery.

- Process events **asynchronously** if needed — return `200` immediately, then handle the event in a background job.
- Check `deliveryContext.isRedelivery` to avoid processing the same event twice; use `webhookEventId` as the deduplication key.

---

## Other Event Types (for Awareness)

The same webhook endpoint receives all event types. Filter by `events[].type`:

| `type` | Trigger |
|--------|---------|
| `message` | User sends a message (text, image, audio, video, etc.) |
| `follow` | User adds the bot as a friend |
| `unfollow` | User blocks the bot |
| `postback` | User taps a postback action button |
| `join` | Bot is added to a group or room |
| `leave` | Bot is removed from a group or room |

For text message handling, filter to events where `type === "message"` and `message.type === "text"`.
