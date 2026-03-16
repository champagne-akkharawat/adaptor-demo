# LINE Messaging API Webhook Integration

Comprehensive reference for integrating LINE Messaging API webhooks into a bot server. All information sourced from the official LINE Developers documentation at https://developers.line.biz.

---

## Table of Contents

1. [Setup (Manual Steps)](#setup-manual-steps)
2. [Authentication / Signature Verification](#authentication--signature-verification)
3. [Webhook Event Types and Payloads](#webhook-event-types-and-payloads)
   - [Common Envelope Fields](#common-envelope-fields)
   - [Source Object](#source-object)
   - [message — text](#message-event--text)
   - [message — image](#message-event--image)
   - [message — video](#message-event--video)
   - [message — audio](#message-event--audio)
   - [message — file](#message-event--file)
   - [message — location](#message-event--location)
   - [message — sticker](#message-event--sticker)
   - [follow](#follow-event)
   - [unfollow](#unfollow-event)
   - [join](#join-event)
   - [leave](#leave-event)
   - [memberJoined](#memberjoined-event)
   - [memberLeft](#memberleft-event)
   - [postback](#postback-event)
   - [beacon](#beacon-event)
   - [accountLink](#accountlink-event)
   - [things](#things-event)
   - [unsend](#unsend-event)
   - [videoPlayComplete](#videoplaycomplete-event)

---

## Setup (Manual Steps)

> Relevant documentation pages:
> - Getting started: https://developers.line.biz/en/docs/messaging-api/getting-started/
> - Building a bot: https://developers.line.biz/en/docs/messaging-api/building-bot/
> - Verify webhook URL: https://developers.line.biz/en/docs/messaging-api/verify-webhook-url/

### Step 1 — Create a LINE Official Account

**Note:** As of September 4, 2024, it is no longer possible to create Messaging API channels directly from the LINE Developers Console. Channels must be created through the LINE Official Account Manager first.

#### 1-1. Register for a Business ID

Go to https://account.line.biz/signup and register using either your personal LINE account or an email address.

#### 1-2. Complete the Entry Form

Fill out the LINE Official Account entry form at https://entry.line.biz/form/entry/unverified with the required business information. Your LINE Official Account is created upon submission.

#### 1-3. Verify Account Creation

Open the LINE Official Account Manager at https://manager.line.biz/ and confirm your new account appears in the list.

---

### Step 2 — Enable the Messaging API

#### 2-1. Activate Messaging API

Inside the LINE Official Account Manager, navigate to your account and enable the Messaging API. If your login has never been used on the LINE Developers Console, a developer registration screen will appear — enter your name and email to create your developer profile.

You will be prompted to **select a Provider**. This selection is permanent and cannot be changed or reassigned after creation, so choose carefully. A Provider represents the company or individual that owns the account.

#### 2-2. Access the LINE Developers Console

Go to https://console.line.biz/ and log in with the same credentials used in the LINE Official Account Manager.

#### 2-3. Confirm Channel Creation

In the console, select your Provider from the left sidebar. A Messaging API channel should have been automatically created. Click it to open its settings.

---

### Step 3 — Configure the Webhook URL

#### 3-1. Navigate to the Messaging API Tab

Inside the channel settings, click the **Messaging API** tab.

#### 3-2. Set the Webhook URL

Click **Edit** under the **Webhook URL** field. Enter the HTTPS URL of your bot server's webhook endpoint (e.g., `https://your-server.example.com/webhook`). Click **Update**.

**TLS / HTTPS requirements:**

| Requirement | Detail |
|---|---|
| Protocol | HTTPS only — plain HTTP is rejected |
| Certificate authority | Must be issued by a CA widely trusted by general web browsers |
| Self-signed certificates | Not permitted |
| Certificate chain | Intermediate certificates must be properly installed; incomplete chains cause verification failure |
| Port | Standard HTTPS port (443) recommended |

#### 3-3. Verify the Webhook URL

Click the **Verify** button. The LINE Platform sends a test HTTP POST request to your endpoint with an empty events array:

```json
{
  "destination": "xxxxxxxxxx",
  "events": []
}
```

Your server must respond with **HTTP 200** for the console to display **"Success"**. If verification fails, check:
- Your server is publicly reachable over the internet
- TLS certificate is valid and the chain is complete
- The endpoint returns HTTP 200 for POST requests
- Review the webhook error logs and statistics in the console

#### 3-4. Enable "Use webhook"

After a successful verification, toggle the **"Use webhook"** switch to **On**. Without this, the LINE Platform will not send webhook events even if a URL is configured.

---

### Step 4 — Additional Configuration

#### 4-1. Disable Auto-responses (Recommended)

In the LINE Official Account Manager under **Messaging API Settings**, set both **"Greeting messages"** and **"Auto-reply messages"** to **Disabled**. Leaving these enabled causes LINE's built-in auto-replies to fire in addition to your bot's webhook responses, resulting in duplicate messages.

#### 4-2. Issue a Channel Access Token

In the LINE Developers Console under the **Messaging API** tab, issue a **Channel Access Token**. The recommended type is **Channel access token v2.1** (user-specified expiration). Other options include:
- Stateless channel access token
- Short-lived channel access token (expires in 30 days)
- Long-lived channel access token (does not expire, less secure)

The channel access token is required to call the LINE Messaging API (to send messages, manage rich menus, etc.) — it is separate from webhook receipt.

#### 4-3. Add Your Official Account as a Friend (For Testing)

On the **Messaging API** tab in the console, scan the QR code with the LINE mobile app to add your LINE Official Account as a friend. This allows you to send test messages and trigger webhook events during development.

#### 4-4. Optional: IP Restriction

For long-lived channel access tokens, visit the **Security** tab in the channel settings to restrict API access to specific IP addresses or CIDR ranges.

---

## Authentication / Signature Verification

> Relevant documentation pages:
> - Verify webhook signature: https://developers.line.biz/en/docs/messaging-api/verify-webhook-signature/
> - Receiving messages: https://developers.line.biz/en/docs/messaging-api/receiving-messages/

Every webhook POST request from the LINE Platform includes an `x-line-signature` header. You **must** verify this signature before processing any event. Skipping verification allows any malicious actor who discovers your endpoint URL to inject fake events.

### The `x-line-signature` Header

The header contains a **Base64-encoded HMAC-SHA256 digest** computed by the LINE Platform using:
- **Key**: Your channel's Channel Secret (found in the LINE Developers Console under **Basic Settings**)
- **Message**: The raw, unmodified HTTP request body (bytes, UTF-8 encoded)

### Verification Algorithm

| Step | Action |
|---|---|
| 1 | Receive the HTTP POST request |
| 2 | Extract the `x-line-signature` header value |
| 3 | Read the **raw request body** — do not parse, deserialize, or reformat it |
| 4 | Retrieve your **Channel Secret** from secure storage |
| 5 | Compute `HMAC-SHA256(key=channel_secret, message=raw_body)` using UTF-8 encoding for both |
| 6 | Base64-encode the resulting digest |
| 7 | Compare your computed signature to the `x-line-signature` header value |
| 8 | If they match: proceed to process events. If they differ: reject the request (return HTTP 401 or 400) |

### Code Examples

#### Python (manual implementation)

```python
import base64
import hashlib
import hmac

def verify_line_signature(channel_secret: str, body: bytes, signature: str) -> bool:
    """
    Verify the x-line-signature header value.

    :param channel_secret: Your LINE channel secret (from Basic Settings tab)
    :param body: The raw request body as bytes (do NOT decode/re-encode)
    :param signature: The value of the x-line-signature header
    :return: True if valid, False otherwise
    """
    digest = hmac.new(
        channel_secret.encode("utf-8"),
        body,
        hashlib.sha256
    ).digest()
    computed = base64.b64encode(digest).decode("utf-8")
    return hmac.compare_digest(computed, signature)

# Flask usage example
from flask import Flask, request, abort

app = Flask(__name__)
CHANNEL_SECRET = "YOUR_CHANNEL_SECRET_HERE"

@app.route("/webhook", methods=["POST"])
def webhook():
    signature = request.headers.get("X-Line-Signature", "")
    body = request.get_data()  # raw bytes, before any parsing

    if not verify_line_signature(CHANNEL_SECRET, body, signature):
        abort(401)

    import json
    payload = json.loads(body)
    for event in payload.get("events", []):
        handle_event(event)

    return "OK", 200
```

#### Python (using official SDK)

```python
from linebot.v3.webhook import WebhookHandler
from linebot.v3.exceptions import InvalidSignatureError

handler = WebhookHandler(channel_secret="YOUR_CHANNEL_SECRET")

@app.route("/webhook", methods=["POST"])
def webhook():
    signature = request.headers.get("X-Line-Signature", "")
    body = request.get_data(as_text=True)

    try:
        handler.handle(body, signature)
    except InvalidSignatureError:
        abort(401)

    return "OK", 200
```

#### Node.js (manual implementation)

```javascript
const crypto = require("crypto");

function verifyLineSignature(channelSecret, rawBody, signature) {
  // rawBody must be the original Buffer or string — do not JSON.parse first
  const computed = crypto
    .createHmac("SHA256", channelSecret)
    .update(rawBody)
    .digest("base64");
  // Use timingSafeEqual to prevent timing attacks
  return crypto.timingSafeEqual(
    Buffer.from(computed),
    Buffer.from(signature)
  );
}

// Express usage example
const express = require("express");
const app = express();
const CHANNEL_SECRET = "YOUR_CHANNEL_SECRET_HERE";

// IMPORTANT: use express.raw() to preserve the original body buffer
app.post("/webhook", express.raw({ type: "application/json" }), (req, res) => {
  const signature = req.headers["x-line-signature"] || "";
  const rawBody = req.body; // Buffer

  if (!verifyLineSignature(CHANNEL_SECRET, rawBody, signature)) {
    return res.status(401).send("Invalid signature");
  }

  const payload = JSON.parse(rawBody.toString("utf-8"));
  payload.events.forEach(handleEvent);
  res.status(200).send("OK");
});
```

#### Node.js (using official SDK)

```javascript
import express from "express";
import { middleware, SignatureValidationFailed, JSONParseError } from "@line/bot-sdk";

const app = express();
const config = { channelSecret: "YOUR_CHANNEL_SECRET" };

app.post("/webhook", middleware(config), (req, res) => {
  const events = req.body.events;
  const destination = req.body.destination;
  events.forEach(handleEvent);
  res.status(200).send("OK");
});

// Error handler for signature failures
app.use((err, req, res, next) => {
  if (err instanceof SignatureValidationFailed) {
    return res.status(401).send(err.signature);
  }
  if (err instanceof JSONParseError) {
    return res.status(400).send(err.raw);
  }
  next(err);
});
```

#### OpenSSL command-line verification (for debugging)

```sh
echo -n '{"destination":"Uxxxxx","events":[]}' \
  | openssl dgst -sha256 -hmac 'YOUR_CHANNEL_SECRET' -binary \
  | openssl base64
```

### Common Verification Failure Causes

| Cause | Explanation |
|---|---|
| JSON parsed before verification | Deserializing and re-serializing changes whitespace and key order, altering the body bytes |
| JSON reformatted | Pretty-printing or minifying changes the byte sequence |
| Wrong channel secret | Each channel has its own secret; using a secret from another channel always fails |
| Reissued channel secret | If the secret was rotated in the console, old signatures computed with the previous secret will not match |
| Non-UTF-8 encoding | Both the channel secret and body must be treated as UTF-8 |
| Escape characters interpreted | Characters like `\n` or `\\` must remain as literal escape sequences; do not interpret them before hashing |
| Body modified by proxy | Some reverse proxies strip or modify the raw body; ensure it reaches your handler unchanged |
| Wrong algorithm | Only HMAC-SHA256 is used; SHA-1, MD5, or SHA-512 will not match |

### When Verification Fails

Do not process the webhook. Return an HTTP error status (401 or 400) and log the failure. Do not reveal the channel secret in error responses.

---

## Webhook Event Types and Payloads

> Relevant documentation page: https://developers.line.biz/en/reference/messaging-api/#webhook-event-objects

All webhook requests are HTTP POST requests sent to your webhook URL with `Content-Type: application/json`. Each request body is a JSON object containing a `destination` field and an `events` array. A single HTTP request may contain multiple events.

### Common Envelope Fields

Every webhook request body has this top-level structure:

```json
{
  "destination": "U1234567890abcdef1234567890abcdef",
  "events": [
    { "...": "..." }
  ]
}
```

| Field | Type | Description |
|---|---|---|
| `destination` | String | User ID of the bot that should receive these events. This is the bot's user ID, not the sender's. |
| `events` | Array | Array of webhook event objects. May be empty in verification requests. |

Every individual event object inside `events` shares these common fields:

```json
{
  "type": "message",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "...": "..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": {
    "isRedelivery": false
  },
  "replyToken": "757913772c4646b784d4b7ce46d12671"
}
```

| Field | Type | Description |
|---|---|---|
| `type` | String | Event type identifier (e.g., `"message"`, `"follow"`, `"postback"`). |
| `mode` | String | Channel state. `"active"` — the channel is responding normally. `"standby"` — another module channel is handling responses (multiple bots scenario). |
| `timestamp` | Number | UNIX time in **milliseconds** when the event occurred on the LINE Platform. |
| `source` | Object | Information about the user and context that triggered the event. See [Source Object](#source-object). |
| `webhookEventId` | String | Unique ID for this webhook delivery, in ULID format. Use this for idempotency checks. |
| `deliveryContext.isRedelivery` | Boolean | `false` for first delivery. `true` if this is a redelivery attempt (LINE retries failed webhooks). |
| `replyToken` | String | Token used to reply to this event using the Reply Message API. Present only on events that support replies. Expires after a short time (approximately 1 minute). |

---

### Source Object

The `source` object identifies who triggered the event and where.

#### User source (1-on-1 chat)

```json
{
  "type": "user",
  "userId": "U4af4980629..."
}
```

#### Group chat source

```json
{
  "type": "group",
  "groupId": "Ca56f94637c...",
  "userId": "U4af4980629..."
}
```

#### Multi-person chat (room) source

```json
{
  "type": "room",
  "roomId": "Ra8dbf4673c...",
  "userId": "U4af4980629..."
}
```

| Field | Type | Description |
|---|---|---|
| `type` | String | Context type: `"user"`, `"group"`, or `"room"`. |
| `userId` | String | User ID of the user who triggered the event. May be absent in some leave events. |
| `groupId` | String | Group chat ID. Present only when `type` is `"group"`. |
| `roomId` | String | Multi-person chat ID. Present only when `type` is `"room"`. |

---

### Message Event — Text

Fires when a user sends a text message.

```json
{
  "type": "message",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": {
    "type": "user",
    "userId": "U4af4980629..."
  },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "757913772c4646b784d4b7ce46d12671",
  "message": {
    "type": "text",
    "id": "444573844083572737",
    "quoteToken": "q3Plxr4AgKd...",
    "markAsReadToken": "30yhdy232...",
    "text": "@All @example Good morning!! (love)",
    "emojis": [
      {
        "index": 29,
        "length": 6,
        "productId": "5ac1bfd5040ab15980c9b435",
        "emojiId": "001"
      }
    ],
    "mention": {
      "mentionees": [
        {
          "index": 0,
          "length": 4,
          "type": "all"
        },
        {
          "index": 5,
          "length": 8,
          "userId": "U49585cd0d5...",
          "type": "user",
          "isSelf": false
        }
      ]
    },
    "quotedMessageId": "444573844083572737"
  }
}
```

#### `message` fields (text)

| Field | Type | Description |
|---|---|---|
| `type` | String | Always `"text"`. |
| `id` | String | Unique message ID. |
| `quoteToken` | String | Token used to quote this message in a reply. |
| `markAsReadToken` | String | Token to mark the message as read. |
| `text` | String | The text content of the message. LINE emoji characters are included as Unicode. |
| `emojis` | Array | Present when the message contains LINE emoji. Each entry describes one emoji. |
| `emojis[].index` | Number | Zero-based character index in `text` where the emoji placeholder starts. |
| `emojis[].length` | Number | Length of the emoji placeholder string in `text`. |
| `emojis[].productId` | String | ID of the emoji product set. |
| `emojis[].emojiId` | String | ID of the specific emoji within the product set. |
| `mention` | Object | Present when the message contains `@` mentions. |
| `mention.mentionees[]` | Array | Array of mentioned users or the `@All` mention. |
| `mention.mentionees[].index` | Number | Zero-based character index in `text` where the mention starts. |
| `mention.mentionees[].length` | Number | Length of the mention text. |
| `mention.mentionees[].type` | String | `"user"` for a specific user mention, `"all"` for `@All`. |
| `mention.mentionees[].userId` | String | User ID of the mentioned user. Present only when `type` is `"user"`. |
| `mention.mentionees[].isSelf` | Boolean | `true` if the mention refers to the bot itself. Present only when `type` is `"user"`. |
| `quotedMessageId` | String | Message ID of the message being quoted/replied to. Present only if the user replied to a previous message. |

---

### Message Event — Image

Fires when a user sends an image.

```json
{
  "type": "message",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "757913772c4646b784d4b7ce46d12671",
  "message": {
    "type": "image",
    "id": "354718705033693859",
    "quoteToken": "q3Plxr4AgKd...",
    "markAsReadToken": "30yhdy232...",
    "contentProvider": {
      "type": "line"
    },
    "imageSet": {
      "id": "E005D41A7288F41B65593ED38FF6E9834B046AB36A37921A56BC236F13A91855",
      "index": 1,
      "total": 2
    }
  }
}
```

When content is hosted externally:

```json
{
  "message": {
    "type": "image",
    "id": "354718705033693859",
    "quoteToken": "q3Plxr4AgKd...",
    "contentProvider": {
      "type": "external",
      "originalContentUrl": "https://example.com/image.jpg",
      "previewImageUrl": "https://example.com/image-preview.jpg"
    }
  }
}
```

#### `message` fields (image)

| Field | Type | Description |
|---|---|---|
| `type` | String | Always `"image"`. |
| `id` | String | Unique message ID. Use with the [Get Content](https://developers.line.biz/en/reference/messaging-api/#get-content) API to download the image when `contentProvider.type` is `"line"`. |
| `quoteToken` | String | Token to quote this message. |
| `markAsReadToken` | String | Token to mark as read. |
| `contentProvider.type` | String | `"line"` — content stored on LINE servers (use Get Content API). `"external"` — content hosted by a third party (use the URLs directly). |
| `contentProvider.originalContentUrl` | String | URL of the full-size image. Present only when `type` is `"external"`. |
| `contentProvider.previewImageUrl` | String | URL of the thumbnail/preview image. Present only when `type` is `"external"`. |
| `imageSet.id` | String | Shared ID for a group of images sent simultaneously. |
| `imageSet.index` | Number | 1-based index of this image within the set. |
| `imageSet.total` | Number | Total number of images in the set. |

---

### Message Event — Video

Fires when a user sends a video file.

```json
{
  "type": "message",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "757913772c4646b784d4b7ce46d12671",
  "message": {
    "type": "video",
    "id": "325708",
    "quoteToken": "q3Plxr4AgKd...",
    "markAsReadToken": "30yhdy232...",
    "duration": 60000,
    "contentProvider": {
      "type": "line"
    }
  }
}
```

#### `message` fields (video)

| Field | Type | Description |
|---|---|---|
| `type` | String | Always `"video"`. |
| `id` | String | Unique message ID. Used with the Get Content API when `contentProvider.type` is `"line"`. |
| `quoteToken` | String | Token to quote this message. |
| `markAsReadToken` | String | Token to mark as read. |
| `duration` | Number | Length of the video in **milliseconds**. |
| `contentProvider.type` | String | `"line"` or `"external"`. |
| `contentProvider.originalContentUrl` | String | URL of the video file. Present only when `type` is `"external"`. |
| `contentProvider.previewImageUrl` | String | URL of the thumbnail image. Present only when `type` is `"external"`. |

---

### Message Event — Audio

Fires when a user sends a voice or audio message.

```json
{
  "type": "message",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "757913772c4646b784d4b7ce46d12671",
  "message": {
    "type": "audio",
    "id": "325708",
    "markAsReadToken": "30yhdy232...",
    "duration": 60000,
    "contentProvider": {
      "type": "line"
    }
  }
}
```

#### `message` fields (audio)

| Field | Type | Description |
|---|---|---|
| `type` | String | Always `"audio"`. |
| `id` | String | Unique message ID. |
| `markAsReadToken` | String | Token to mark as read. |
| `duration` | Number | Length of the audio in **milliseconds**. Optional — may be absent. |
| `contentProvider.type` | String | `"line"` or `"external"`. |
| `contentProvider.originalContentUrl` | String | URL of the audio file. Present only when `type` is `"external"`. Audio does not have a preview URL. |

---

### Message Event — File

Fires when a user sends a file (e.g., PDF, Word document).

```json
{
  "type": "message",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "757913772c4646b784d4b7ce46d12671",
  "message": {
    "type": "file",
    "id": "325708",
    "markAsReadToken": "30yhdy232...",
    "fileName": "example-report.pdf",
    "fileSize": 2138
  }
}
```

#### `message` fields (file)

| Field | Type | Description |
|---|---|---|
| `type` | String | Always `"file"`. |
| `id` | String | Unique message ID. Use with the Get Content API to download the file. |
| `markAsReadToken` | String | Token to mark as read. |
| `fileName` | String | Original filename as provided by the sender. |
| `fileSize` | Number | File size in **bytes**. |

---

### Message Event — Location

Fires when a user sends a location pin.

```json
{
  "type": "message",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "757913772c4646b784d4b7ce46d12671",
  "message": {
    "type": "location",
    "id": "325708",
    "markAsReadToken": "30yhdy232...",
    "title": "my location",
    "address": "1-3 Kioicho, Chiyoda-ku, Tokyo, 102-8282 Japan",
    "latitude": 35.67966,
    "longitude": 139.73669
  }
}
```

#### `message` fields (location)

| Field | Type | Description |
|---|---|---|
| `type` | String | Always `"location"`. |
| `id` | String | Unique message ID. |
| `markAsReadToken` | String | Token to mark as read. |
| `title` | String | Location label set by the user. Optional — may be absent. |
| `address` | String | Human-readable address string. Optional — may be absent. |
| `latitude` | Number | Latitude in decimal degrees. |
| `longitude` | Number | Longitude in decimal degrees. |

---

### Message Event — Sticker

Fires when a user sends a sticker.

```json
{
  "type": "message",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "757913772c4646b784d4b7ce46d12671",
  "message": {
    "type": "sticker",
    "id": "1501597916",
    "quoteToken": "q3Plxr4AgKd...",
    "markAsReadToken": "30yhdy232...",
    "packageId": "11537",
    "stickerId": "52002738",
    "stickerResourceType": "ANIMATION",
    "keywords": ["cony", "sally", "Staring", "thinking"],
    "text": "Let's hang out this weekend!"
  }
}
```

#### `message` fields (sticker)

| Field | Type | Description |
|---|---|---|
| `type` | String | Always `"sticker"`. |
| `id` | String | Unique message ID. |
| `quoteToken` | String | Token to quote this sticker message. |
| `markAsReadToken` | String | Token to mark as read. |
| `packageId` | String | Sticker package (set) ID. |
| `stickerId` | String | Individual sticker ID within the package. |
| `stickerResourceType` | String | Resource type: `STATIC`, `ANIMATION`, `SOUND`, `ANIMATION_SOUND`, `POPUP`, `POPUP_SOUND`, `CUSTOM`, `MESSAGE`, `NAME_TEXT`, or `PER_STICKER_TEXT`. |
| `keywords` | Array of String | Up to 15 descriptive keywords for the sticker. |
| `text` | String | Text entered by the user for message stickers (`MESSAGE` or `PER_STICKER_TEXT` types). Absent for non-text stickers. |

---

### Follow Event

Fires when a user adds the LINE Official Account as a friend, or when a user who had previously blocked the account unblocks it.

```json
{
  "type": "follow",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "85cbe770fa8b4f45bbe077b1d4be4a36",
  "follow": {
    "isUnblocked": false
  }
}
```

#### Event-specific fields

| Field | Type | Description |
|---|---|---|
| `replyToken` | String | Token available for sending a welcome/greeting reply. |
| `follow.isUnblocked` | Boolean | `false` — the user is adding the account as a friend for the first time. `true` — the user previously blocked the account and is now unblocking it. |

---

### Unfollow Event

Fires when a user blocks the LINE Official Account. No reply token is provided because the user will not receive replies after blocking.

```json
{
  "type": "unfollow",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false }
}
```

There are no event-specific fields beyond the common fields. No `replyToken` is present.

---

### Join Event

Fires when the LINE Official Account is invited to and joins a **group chat** or **multi-person chat**. For group chats, it fires when the bot is invited. For rooms (multi-person chats), it fires when the first event occurs after the bot is added.

```json
{
  "type": "join",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": {
    "type": "group",
    "groupId": "Ca56f94637c..."
  },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "nHuyWiB7yP5Zw52FIkcQobQuGDXCTA"
}
```

#### Event-specific fields

| Field | Type | Description |
|---|---|---|
| `replyToken` | String | Token to send a greeting message to the group or room. |
| `source.type` | String | `"group"` or `"room"`. The `userId` of the inviter may not be present. |

---

### Leave Event

Fires when the LINE Official Account is removed from a group chat by a user, or when the bot itself calls the Leave API to exit a group or room.

```json
{
  "type": "leave",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": {
    "type": "group",
    "groupId": "Ca56f94637c..."
  },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false }
}
```

There are no event-specific fields. No `replyToken` is present (the bot is no longer in the chat).

---

### memberJoined Event

Fires when one or more users **join a group chat or multi-person chat** that the LINE Official Account is already a member of.

```json
{
  "type": "memberJoined",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": {
    "type": "group",
    "groupId": "Ca56f94637c..."
  },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "0f3779fba3b349968c5d07db31eabf65",
  "joined": {
    "members": [
      {
        "type": "user",
        "userId": "U4af4980629..."
      },
      {
        "type": "user",
        "userId": "U91eeaf62d9..."
      }
    ]
  }
}
```

#### Event-specific fields

| Field | Type | Description |
|---|---|---|
| `replyToken` | String | Token to reply welcoming the new members. |
| `joined.members` | Array | Array of user objects representing the users who joined. |
| `joined.members[].type` | String | Always `"user"`. |
| `joined.members[].userId` | String | User ID of the joining member. |

---

### memberLeft Event

Fires when one or more users **leave or are removed from a group chat or multi-person chat** that the bot is a member of.

```json
{
  "type": "memberLeft",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": {
    "type": "group",
    "groupId": "Ca56f94637c..."
  },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "left": {
    "members": [
      {
        "type": "user",
        "userId": "U4af4980629..."
      },
      {
        "type": "user",
        "userId": "U91eeaf62d9..."
      }
    ]
  }
}
```

No `replyToken` is provided for this event.

#### Event-specific fields

| Field | Type | Description |
|---|---|---|
| `left.members` | Array | Array of user objects representing the users who left. |
| `left.members[].type` | String | Always `"user"`. |
| `left.members[].userId` | String | User ID of the departing member. |

---

### Postback Event

Fires when a user triggers a **postback action** — for example, by tapping a button in a template message, a quick reply, or a rich menu item that has a postback action configured.

```json
{
  "type": "postback",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "b60d432864f44d079f6d8efe86cf404b",
  "postback": {
    "data": "storeId=12345&action=buy"
  }
}
```

#### With datetime picker result

```json
{
  "type": "postback",
  "replyToken": "b60d432864f44d079f6d8efe86cf404b",
  "postback": {
    "data": "reservationDate",
    "params": {
      "date": "2024-12-25",
      "time": "14:30",
      "datetime": "2024-12-25T14:30"
    }
  }
}
```

#### With rich menu switch action result

```json
{
  "type": "postback",
  "replyToken": "b60d432864f44d079f6d8efe86cf404b",
  "postback": {
    "data": "richmenu-changed-to-b",
    "params": {
      "newRichMenuAliasId": "richmenu-alias-b",
      "status": "SUCCESS"
    }
  }
}
```

#### Event-specific fields

| Field | Type | Description |
|---|---|---|
| `replyToken` | String | Token to reply to this postback action. |
| `postback.data` | String | The data string you defined in the postback action configuration. Up to 300 characters. |
| `postback.params` | Object | Present only for datetime picker or rich menu switch actions. |
| `postback.params.date` | String | Selected date in `YYYY-MM-DD` format. Present for date picker actions. |
| `postback.params.time` | String | Selected time in `HH:mm` format. Present for time picker actions. |
| `postback.params.datetime` | String | Selected datetime in `YYYY-MM-DDThh:mm` format. Present for datetime picker actions. |
| `postback.params.newRichMenuAliasId` | String | Alias ID of the rich menu switched to. Present for rich menu switch actions. |
| `postback.params.status` | String | Result of the rich menu switch: `"SUCCESS"`, `"RICHMENU_ALIAS_ID_NOTFOUND"`, `"RICHMENU_NOTFOUND"`, or `"FAILED"`. |

---

### Beacon Event

Fires when a user's LINE app detects a **LINE Beacon** device. Beacons are physical Bluetooth Low Energy devices that LINE-compatible devices can detect.

```json
{
  "type": "beacon",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "nHuyWiB7yP5Zw52FIkcQobQuGDXCTA",
  "beacon": {
    "hwid": "d41d8cd98f00b204",
    "type": "enter",
    "dm": "48656c6c6f"
  }
}
```

#### Event-specific fields

| Field | Type | Description |
|---|---|---|
| `replyToken` | String | Token to reply with a location-relevant message. |
| `beacon.hwid` | String | Hardware ID of the beacon device that was detected. |
| `beacon.type` | String | Beacon interaction type: `"enter"` — user entered the beacon range; `"banner"` — user tapped the beacon banner; `"stay"` — periodic event while user remains in range. |
| `beacon.dm` | String | Optional device message in hexadecimal. A string set by the beacon owner. May be absent if the beacon does not support device messages. |

---

### accountLink Event

Fires when a user completes or fails the **account linking** flow — the process of connecting their LINE account to an account in your own service.

```json
{
  "type": "accountLink",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "b60d432864f44d079f6d8efe86cf404b",
  "link": {
    "result": "ok",
    "nonce": "xxxxxxxxxxxxxxx"
  }
}
```

**Note:** If account linking fails, the `source` and `replyToken` fields may be omitted.

#### Event-specific fields

| Field | Type | Description |
|---|---|---|
| `replyToken` | String | Token to send a confirmation message. May be absent on failure. |
| `link.result` | String | `"ok"` — linking succeeded. `"failed"` — linking failed (e.g., user canceled, impersonation attempt detected). |
| `link.nonce` | String | The nonce (number used once) that your service generated at the start of the linking flow, used to match this event to the linking session and verify the user's identity. |

---

### things Event

The `things` event covers **LINE Things** (IoT device integration). It fires for three sub-scenarios: a device being linked, a device being unlinked, and a scenario result (automated device interaction result).

#### deviceLink — Device linked to the user's LINE account

```json
{
  "type": "things",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "nHuyWiB7yP5Zw52FIkcQobQuGDXCTA",
  "things": {
    "deviceId": "t2c449c9d1...",
    "type": "link"
  }
}
```

#### deviceUnlink — Device unlinked from the user's LINE account

```json
{
  "type": "things",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "nHuyWiB7yP5Zw52FIkcQobQuGDXCTA",
  "things": {
    "deviceId": "t2c449c9d1...",
    "type": "unlink"
  }
}
```

#### scenarioResult — Result from an automated scenario execution

```json
{
  "type": "things",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "things": {
    "deviceId": "t2c449c9d1...",
    "type": "scenarioResult",
    "result": {
      "scenarioId": "scenario-id-here",
      "revision": 2,
      "startTime": 1547817845537,
      "endTime": 1547817845557,
      "resultCode": "success",
      "bleNotificationPayload": "AQ==",
      "actionResults": [
        {
          "type": "binary",
          "data": "/w=="
        }
      ]
    }
  }
}
```

No `replyToken` is present on `scenarioResult` events.

#### Event-specific fields

| Field | Type | Description |
|---|---|---|
| `replyToken` | String | Token to send a notification. Present on `link` and `unlink` events; absent on `scenarioResult`. |
| `things.deviceId` | String | Unique ID of the LINE Things device. |
| `things.type` | String | Sub-type: `"link"`, `"unlink"`, or `"scenarioResult"`. |
| `things.result.scenarioId` | String | ID of the scenario that was executed. |
| `things.result.revision` | Number | Revision number of the scenario. |
| `things.result.startTime` | Number | UNIX timestamp in milliseconds when the scenario started. |
| `things.result.endTime` | Number | UNIX timestamp in milliseconds when the scenario ended. |
| `things.result.resultCode` | String | Execution result: `"success"`, `"gatt_error"`, `"runtime_error"`, etc. |
| `things.result.bleNotificationPayload` | String | Base64-encoded BLE notification payload if received during scenario. |
| `things.result.actionResults` | Array | Array of per-action results from the scenario. |
| `things.result.actionResults[].type` | String | Result type: `"binary"` or `"void"`. |
| `things.result.actionResults[].data` | String | Base64-encoded binary data for `"binary"` type results. |

---

### unsend Event

Fires when a user **unsends (deletes) a previously sent message**. The original message content is no longer accessible after this event.

```json
{
  "type": "unsend",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "unsend": {
    "messageId": "325708"
  }
}
```

No `replyToken` is present. There is no way to retrieve the unsent message content after this event.

#### Event-specific fields

| Field | Type | Description |
|---|---|---|
| `unsend.messageId` | String | The message ID of the message that was unsent. Use this to identify which previously received message was removed. |

---

### videoPlayComplete Event

Fires when a user **finishes watching a video message** that was sent by the bot and tagged with a `trackingId`. This event fires only for videos the bot itself sent, not for user-sent videos.

```json
{
  "type": "videoPlayComplete",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": { "type": "user", "userId": "U4af4980629..." },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "nHuyWiB7yP5Zw52FIkcQobQuGDXCTA",
  "videoPlayComplete": {
    "trackingId": "track-id-here"
  }
}
```

#### Event-specific fields

| Field | Type | Description |
|---|---|---|
| `replyToken` | String | Token to reply after the user finishes watching (e.g., a survey or follow-up message). |
| `videoPlayComplete.trackingId` | String | The `trackingId` you assigned when sending the original video message. Use this to correlate which video was watched. |

---

## Additional Notes

### Webhook Redelivery

The LINE Platform retries webhook delivery when your server does not respond with HTTP 2xx within the timeout window. The `deliveryContext.isRedelivery` flag is `true` on retry attempts. Implement idempotency using `webhookEventId` to avoid processing the same event twice.

### Mode: standby

When `mode` is `"standby"`, the bot is operating in a multi-bot (module channel) setup and another channel is currently the active responder. You should not send replies from standby mode but may still process events for logging or analytics.

### Rate Limits and Timeouts

Your webhook endpoint should respond with HTTP 2xx within **2 seconds**. The LINE Platform treats a lack of response or a non-2xx status code as a failed delivery and will retry. Process events asynchronously if needed — respond 200 immediately and handle work in a background thread or queue.

### Port and IP Range

The LINE Platform sends webhooks from a specific set of IP addresses. For production environments, ensure your firewall does not block LINE's outbound IP ranges. Check the current list in the [LINE Developers documentation](https://developers.line.biz/en/docs/messaging-api/receiving-messages/).

---

*Sources:*
- https://developers.line.biz/en/docs/messaging-api/getting-started/
- https://developers.line.biz/en/docs/messaging-api/building-bot/
- https://developers.line.biz/en/docs/messaging-api/verify-webhook-url/
- https://developers.line.biz/en/docs/messaging-api/verify-webhook-signature/
- https://developers.line.biz/en/docs/messaging-api/receiving-messages/
- https://developers.line.biz/en/reference/messaging-api/#webhook-event-objects
- https://line.github.io/line-bot-sdk-nodejs/guide/webhook.html
- https://line-bot-sdk-python.readthedocs.io/en/stable/_modules/linebot/models/events.html
