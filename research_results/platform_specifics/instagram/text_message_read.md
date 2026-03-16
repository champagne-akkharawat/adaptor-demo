# Instagram Messaging API — Receiving Text Messages (Webhooks)

## Overview

Incoming messages from Instagram users are delivered to your server as HTTP POST requests from Meta's webhook infrastructure. You register a publicly accessible HTTPS endpoint; Meta calls it whenever a subscribed event occurs.

---

## Webhook Event Structure (Incoming Text Message)

```json
{
  "object": "instagram",
  "entry": [
    {
      "id": "{IG-PAGE-ID}",
      "time": 1569262486134,
      "messaging": [
        {
          "sender": {
            "id": "{IGSID}"
          },
          "recipient": {
            "id": "{IG-PAGE-ID}"
          },
          "timestamp": 1569262485349,
          "message": {
            "mid": "{MESSAGE-ID}",
            "text": "{MESSAGE-TEXT}"
          }
        }
      ]
    }
  ]
}
```

### Top-level fields

| Field | Type | Description |
|---|---|---|
| `object` | string | Always `"instagram"` for Instagram messaging events. |
| `entry` | array | One or more entry objects. Each entry represents one Instagram account. |

### `entry` object

| Field | Type | Description |
|---|---|---|
| `id` | string | The Instagram Professional Account ID (also the linked Facebook Page ID). |
| `time` | number | Unix timestamp (milliseconds) when Meta sent the notification. |
| `messaging` | array | Array of messaging event objects. Typically contains one item per notification. |

### `messaging` object (text message)

| Field | Type | Description |
|---|---|---|
| `sender.id` | string | **Instagram-Scoped ID (IGSID)** of the user who sent the message. Use this as `recipient.id` when replying. |
| `recipient.id` | string | The IG Professional Account ID that received the message (your account). |
| `timestamp` | number | Unix timestamp (milliseconds) of when the message was sent. |
| `message.mid` | string | Unique message identifier. |
| `message.text` | string | The plain text content of the message. Present only for text messages. |

### Optional fields on `message`

| Field | Type | Description |
|---|---|---|
| `quick_reply.payload` | string | Present if the user tapped a quick-reply button; contains the developer-defined payload string. |
| `attachments` | array | Present instead of (or alongside) `text` when the user sends media. Types include `image`, `audio`, `video`, `file`, `reel`, `ig_reel`, `fallback`. |
| `reply_to` | object | Present when the user replies to a specific message in the thread. |

---

## Extracting Sender ID and Message Text

```
sender_id  = payload["entry"][0]["messaging"][0]["sender"]["id"]
message_text = payload["entry"][0]["messaging"][0]["message"]["text"]
```

Defensive extraction pattern (pseudocode):

```
for each entry in payload.entry:
    for each event in entry.messaging:
        sender_id = event.sender.id
        if event.message exists and event.message.text exists:
            # plain text message
            handle_text(sender_id, event.message.text, event.message.mid)
        elif event.message exists and event.message.attachments exists:
            # media message — handle separately
```

Always check that `message.text` is present before treating the event as a text message, because the same `messages` webhook subscription delivers attachment events too.

---

## Signature Verification

Meta signs every POST notification using HMAC-SHA256. Verifying the signature ensures the request genuinely came from Meta and the payload was not tampered with.

### Header

```
X-Hub-Signature-256: sha256={HMAC_HEX_DIGEST}
```

### Verification steps

1. Read the raw request body bytes (before any JSON parsing).
2. Compute `HMAC-SHA256(key=APP_SECRET, message=raw_body)`.
3. Encode the result as a lowercase hex string.
4. Compare your computed digest to the value in `X-Hub-Signature-256` after stripping the `sha256=` prefix.
5. Use a **constant-time comparison** (e.g. `hmac.compare_digest` in Python) to prevent timing attacks.
6. Reject the request with HTTP 403 if the signatures do not match.

### Important note on encoding
Meta hashes the **Unicode-escaped version** of the payload as lowercase hexadecimal. If you hash the decoded bytes directly your digest may not match. Hash the raw body bytes as received over the wire.

### Example (Python)

```python
import hmac
import hashlib

def verify_signature(raw_body: bytes, app_secret: str, header_value: str) -> bool:
    expected = "sha256=" + hmac.new(
        app_secret.encode("utf-8"),
        raw_body,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(expected, header_value)
```

---

## Webhook Setup

### Prerequisites

- A publicly accessible HTTPS endpoint with a valid TLS/SSL certificate (self-signed certificates are not accepted).
- A Meta Developer App with a Facebook Page connected to an Instagram Professional account.
- The Instagram account must have **"Connected Tools"** messaging enabled.

### Required permissions

| Permission | Purpose |
|---|---|
| `instagram_basic` | Basic Instagram account access |
| `instagram_manage_messages` | Receive and send messages |
| `pages_manage_metadata` | Subscribe to page/Instagram webhook events |

**Advanced Access** is required to receive webhook events involving users who do not have a role on your app. Standard Access limits events to app developers/testers only.

### Verification challenge (one-time setup)

When you register your webhook URL in the App Dashboard, Meta sends a GET request to verify ownership:

```
GET {YOUR_WEBHOOK_URL}
  ?hub.mode=subscribe
  &hub.verify_token={YOUR_VERIFY_TOKEN}
  &hub.challenge={CHALLENGE_INTEGER}
```

Your endpoint must:
1. Confirm `hub.mode` equals `subscribe`.
2. Confirm `hub.verify_token` matches the value you configured in the dashboard.
3. Respond with HTTP 200 and the plain integer value of `hub.challenge` as the body.

### Subscribing to the `messages` field

In the App Dashboard:
1. Go to **App Settings > Webhooks** (or the **Instagram Settings** section for Instagram-specific subscriptions).
2. Add your callback URL and verify token, then click **Verify and Save**.
3. Subscribe to the `messages` field (and optionally `message_deliveries`, `message_reads`).

Alternatively, subscribe via API:

```bash
curl -X POST \
  "https://graph.facebook.com/v25.0/{PAGE-ID}/subscribed_apps" \
  -d "subscribed_fields=messages&access_token={PAGE_ACCESS_TOKEN}"
```

### Responding to webhook POST requests

Your endpoint must return **HTTP 200** within **5 seconds** of receiving a webhook event. If Meta does not receive a 200 response it will retry delivery. Process events asynchronously (enqueue and acknowledge immediately) if handling takes longer than 5 seconds.

---

## Key Notes

- **`entry` may contain multiple items** if Meta batches notifications; iterate over all entries and all `messaging` items within each entry.
- **Deduplication**: Use `message.mid` to detect and discard duplicate deliveries.
- **Echo events**: If your app sends a message, Meta may deliver an echo event back with `message.is_echo: true`. Filter these out to avoid processing your own outbound messages as inbound.
- **24-hour window tracking**: The `timestamp` field on the `messaging` object indicates when the customer's message was sent. Track this to enforce the 24-hour reply window.
