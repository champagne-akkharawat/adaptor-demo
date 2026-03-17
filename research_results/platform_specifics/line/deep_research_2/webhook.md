# LINE Messaging API Webhook Integration

> **Sources:**
> - https://developers.line.biz/en/docs/messaging-api/getting-started/
> - https://developers.line.biz/en/docs/messaging-api/building-bot/
> - https://developers.line.biz/en/docs/messaging-api/verify-webhook-url/
> - https://developers.line.biz/en/docs/messaging-api/verify-webhook-signature/
> - https://developers.line.biz/en/docs/messaging-api/receiving-messages/
> - https://developers.line.biz/en/docs/messaging-api/use-membership-features/
> - https://developers.line.biz/en/reference/messaging-api/#webhooks
> - https://developers.line.biz/en/reference/messaging-api/#webhook-event-objects
> - https://developers.line.biz/en/reference/messaging-api/index.html.md (raw markdown dump — primary reference for field-level specs)

**Last verified:** March 2026

---

## Table of Contents

1. [Setup (Manual Steps)](#1-setup-manual-steps)
   - 1.1 [Create a LINE Official Account](#11-create-a-line-official-account)
   - 1.2 [Enable the Messaging API](#12-enable-the-messaging-api)
   - 1.3 [Configure the Webhook URL](#13-configure-the-webhook-url)
   - 1.4 [Disable Auto-reply and Greeting Messages](#14-disable-auto-reply-and-greeting-messages)
2. [Authentication — Signature Verification](#2-authentication--signature-verification)
   - 2.1 [X-Line-Signature Header](#21-x-line-signature-header)
   - 2.2 [HMAC-SHA256 Verification Algorithm](#22-hmac-sha256-verification-algorithm)
   - 2.3 [Code Examples](#23-code-examples)
   - 2.4 [Common Verification Failure Causes](#24-common-verification-failure-causes)
3. [Webhook Envelope](#3-webhook-envelope)
   - 3.1 [Common Envelope Fields](#31-common-envelope-fields)
   - 3.2 [Source Object](#32-source-object)
4. [Webhook Event Types and Payloads](#4-webhook-event-types-and-payloads)
   - 4.1  [message — text](#41-message--text)
   - 4.2  [message — image](#42-message--image)
   - 4.3  [message — video](#43-message--video)
   - 4.4  [message — audio](#44-message--audio)
   - 4.5  [message — file](#45-message--file)
   - 4.6  [message — location](#46-message--location)
   - 4.7  [message — sticker](#47-message--sticker)
   - 4.8  [follow](#48-follow)
   - 4.9  [unfollow](#49-unfollow)
   - 4.10 [join](#410-join)
   - 4.11 [leave](#411-leave)
   - 4.12 [memberJoined](#412-memberjoined)
   - 4.13 [memberLeft](#413-memberleft)
   - 4.14 [postback](#414-postback)
   - 4.15 [beacon](#415-beacon)
   - 4.16 [accountLink](#416-accountlink)
   - 4.17 [things](#417-things)
   - 4.18 [unsend](#418-unsend)
   - 4.19 [videoPlayComplete](#419-videoplaycomplete)
   - 4.20 [membership](#420-membership)

---

## 1. Setup (Manual Steps)

### 1.1 Create a LINE Official Account

> **Reference:** https://developers.line.biz/en/docs/messaging-api/getting-started/

As of September 4, 2024, Messaging API channels can no longer be created directly from the LINE Developers Console. The account must first be created through the LINE Official Account Manager.

**Step 1-1 — Register for a Business ID**

Go to https://account.line.biz/signup and register using either your personal LINE account or an email address. This creates the Business ID required to manage LINE Official Accounts.

**Step 1-2 — Complete the Entry Form**

Once your Business ID is created, fill out the LINE Official Account entry form at https://entry.line.biz/form/entry/unverified with the required business information. Your LINE Official Account is created upon submission.

**Step 1-3 — Verify Account Creation**

Open the LINE Official Account Manager at https://manager.line.biz/ and confirm your new account appears in the account list.

---

### 1.2 Enable the Messaging API

> **Reference:** https://developers.line.biz/en/docs/messaging-api/getting-started/

**Step 2-1 — Activate Messaging API**

Inside the LINE Official Account Manager, navigate to your account settings and enable the Messaging API. This automatically creates a Messaging API channel.

If your login has never been used on the LINE Developers Console, a developer registration screen appears — enter your name and email to create your developer profile.

You will be prompted to **select a Provider**. A Provider represents the company or individual that owns the account. **This selection is permanent and cannot be changed or reassigned after creation.** Choose carefully, especially if you plan to integrate with other LINE services (e.g., LINE Login) that are already under a specific provider.

**Step 2-2 — Access the LINE Developers Console**

Go to https://developers.line.biz/console/ and log in with the same credentials used in the LINE Official Account Manager.

**Step 2-3 — Confirm Channel Creation**

In the console, select your Provider. A Messaging API channel should have been automatically created. Click it to open its settings.

---

### 1.3 Configure the Webhook URL

> **Reference:** https://developers.line.biz/en/docs/messaging-api/building-bot/
> **Reference:** https://developers.line.biz/en/docs/messaging-api/verify-webhook-url/

**Step 3-1 — Issue a Channel Access Token**

Before configuring the webhook, issue a channel access token from the **Messaging API** tab of the channel settings. The recommended type is **Channel access token v2.1** (user-specified expiration). The token is needed to call the Messaging API (e.g., to send reply messages) but is distinct from webhook receipt, which requires no token.

**Step 3-2 — Set the Webhook URL**

In the channel settings, click the **Messaging API** tab, then click **Edit** under the **Webhook URL** field. Enter the full HTTPS URL of your bot server endpoint and click **Update**.

**TLS/HTTPS requirements:**

| Requirement | Detail |
|---|---|
| Protocol | HTTPS only — plain HTTP is rejected |
| Certificate authority | Must be issued by a CA widely trusted by general web browsers |
| Self-signed certificates | Not permitted |
| Certificate chain | Intermediate certificates must be correctly installed; incomplete chains cause verification failure |
| Port | Standard HTTPS port (443) is typical |

**Step 3-3 — Verify the Webhook URL**

Click the **Verify** button. The LINE Platform sends a test HTTP POST request with an empty events array:

```json
{
  "destination": "xxxxxxxxxx",
  "events": []
}
```

Your server must respond with **HTTP 200** for the console to show **"Success"**. If verification fails:
- Confirm your server is publicly reachable
- Confirm the TLS certificate is valid and the chain is complete
- Confirm the endpoint returns HTTP 200 for POST requests with an empty body
- Check webhook error statistics at: https://developers.line.biz/en/docs/messaging-api/check-webhook-error-statistics/

You can also verify programmatically using the [Test webhook endpoint](https://developers.line.biz/en/reference/messaging-api/#test-webhook-endpoint) API.

**Step 3-4 — Enable "Use webhook"**

After a successful verification, toggle the **"Use webhook"** switch to **On**. Without this, the LINE Platform will not send webhook events even if a URL is configured.

---

### 1.4 Disable Auto-reply and Greeting Messages

> **Reference:** https://developers.line.biz/en/docs/messaging-api/building-bot/

By default, the LINE Official Account has **Greeting messages** and **Auto-reply messages** set to **Enabled**. These cause LINE's built-in auto-replies to fire in addition to your bot's webhook-driven responses, resulting in duplicate or unexpected messages.

To disable:
1. Open the **LINE Official Account Manager** at https://manager.line.biz/
2. Navigate to your account
3. Under **Messaging API Settings**, set both **"Greeting messages"** and **"Auto-reply messages"** to **Disabled**

You can use both systems together — for example, use greeting messages only for the follow event, and the Messaging API for all other responses. However, it is difficult to distinguish which system sent a given auto-reply, so disabling both is recommended for bot-only deployments.

---

## 2. Authentication — Signature Verification

### 2.1 X-Line-Signature Header

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#signature-validation
> **Reference:** https://developers.line.biz/en/docs/messaging-api/verify-webhook-signature/

Every webhook HTTP POST request from the LINE Platform includes a header named `x-line-signature`.

**Important notes on the header name:**
- The header field name is case-insensitive per HTTP spec (RFC 7230). LINE's documentation notes the name may change between `X-Line-Signature` (capitalized) and `x-line-signature` (lowercase) without notice.
- Your bot server must match the header without case distinction.

The `x-line-signature` value is a **Base64-encoded HMAC-SHA256 digest** computed by the LINE Platform using:
- **Key:** Your channel's Channel Secret (found in the **Basic Settings** tab of the channel in the LINE Developers Console)
- **Message:** The raw, unmodified HTTP request body (UTF-8 bytes, exactly as received)

The LINE Platform does not disclose the IP addresses it sends webhooks from. Signature verification is the only supported security mechanism — do not rely on IP allowlisting.

**Getting the Channel Secret:**

Open the channel's **Basic settings** tab in the LINE Developers Console. The Channel Secret is listed there. You need Admin privileges to view it. If you suspect the secret has been compromised, click **Issue** to reissue it — this immediately invalidates the old secret.

---

### 2.2 HMAC-SHA256 Verification Algorithm

> **Reference:** https://developers.line.biz/en/docs/messaging-api/verify-webhook-signature/

| Step | Action |
|---|---|
| 1 | Receive the HTTP POST request |
| 2 | Store the raw request body as bytes — do not parse, deserialize, re-serialize, or reformat it |
| 3 | Extract the `x-line-signature` header value |
| 4 | Retrieve your **Channel Secret** from secure storage |
| 5 | Compute `HMAC-SHA256(key=channel_secret_bytes, message=raw_body_bytes)` using UTF-8 encoding for both the key and message |
| 6 | Base64-encode the resulting 32-byte digest |
| 7 | Compare your computed signature to the `x-line-signature` header value |
| 8 | If they match: proceed. If they differ or the header is absent: reject the request (HTTP 401 or 400), do not process events |

**Pseudocode:**
```
channel_secret = get_from_secure_storage()
raw_body       = read_request_body_as_raw_bytes()
received_sig   = request_header("x-line-signature")

digest         = HMAC_SHA256(key=channel_secret.encode("utf-8"),
                             message=raw_body)
computed_sig   = base64_encode(digest)

if constant_time_compare(computed_sig, received_sig):
    process_events()
else:
    reject(401)
```

---

### 2.3 Code Examples

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#signature-validation

**Go:**
```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "io/ioutil"
)

func verifySignature(channelSecret string, r *http.Request) bool {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        return false
    }
    decoded, err := base64.StdEncoding.DecodeString(r.Header.Get("x-line-signature"))
    if err != nil {
        return false
    }
    mac := hmac.New(sha256.New, []byte(channelSecret))
    mac.Write(body)
    // Use hmac.Equal for constant-time comparison
    return hmac.Equal(decoded, mac.Sum(nil))
}
```

**Python:**
```python
import base64
import hashlib
import hmac

def verify_line_signature(channel_secret: str, body: bytes, signature: str) -> bool:
    """
    Verify the x-line-signature header.
    body must be the raw request body bytes — do NOT decode/re-encode.
    """
    digest = hmac.new(
        channel_secret.encode("utf-8"),
        body,
        hashlib.sha256
    ).digest()
    computed = base64.b64encode(digest).decode("utf-8")
    return hmac.compare_digest(computed, signature)
```

**Node.js:**
```javascript
const crypto = require("crypto");

const channelSecret = "..."; // Channel secret string
const body = "...";          // Request body string
const signature = crypto
  .createHmac("SHA256", channelSecret)
  .update(body)
  .digest("base64");
// Compare x-line-signature request header and the signature
```

**Java:**
```java
String channelSecret = "..."; // Channel secret string
String httpRequestBody = "..."; // Request body string
SecretKeySpec key = new SecretKeySpec(channelSecret.getBytes(), "HmacSHA256");
Mac mac = Mac.getInstance("HmacSHA256");
mac.init(key);
byte[] source = httpRequestBody.getBytes("UTF-8");
String signature = Base64.encodeBase64String(mac.doFinal(source));
// Compare x-line-signature request header string and the signature
```

**Ruby:**
```ruby
CHANNEL_SECRET = '...' # Channel secret string
http_request_body = request.raw_post # Request body string
hash = OpenSSL::HMAC::digest(OpenSSL::Digest::SHA256.new, CHANNEL_SECRET, http_request_body)
signature = Base64.strict_encode64(hash)
# Compare x-line-signature request header string and the signature
```

**PHP:**
```php
$channelSecret = '...'; // Channel secret string
$httpRequestBody = '...'; // Request body string
$hash = hash_hmac('sha256', $httpRequestBody, $channelSecret, true);
$signature = base64_encode($hash);
// Compare x-line-signature request header string and the signature
```

**OpenSSL command-line (debugging):**
```sh
echo -n '{"destination":"U8e742f61d673b39c7fff3cecb7536ef0","events":[]}' \
  | openssl dgst -sha256 -hmac '8c570fa6dd201bb328f1c1eac23a96d8' -binary \
  | openssl base64
# Output: GhRKmvmHys4Pi8DxkF4+EayaH0OqtJtaZxgTD9fMDLs=
```

---

### 2.4 Common Verification Failure Causes

| Cause | Explanation |
|---|---|
| JSON parsed or re-serialized before verification | Deserializing and re-serializing changes whitespace and key ordering, altering the byte sequence |
| JSON formatted/pretty-printed | Adding indentation or newlines to the raw body changes the bytes |
| Wrong channel secret | Each channel has a unique secret; using the wrong channel's secret always fails |
| Channel secret was reissued | If a team member rotated the secret in the console, the old cached secret will no longer match |
| Non-UTF-8 encoding | Both the channel secret and request body must be treated as UTF-8 |
| Escape characters interpreted | Characters like `\n` or `\\` must remain as literal escape sequences before hashing |
| Request body modified by proxy or middleware | Some reverse proxies or web frameworks buffer and re-emit the body differently; ensure raw bytes reach your handler |
| Wrong algorithm | Only HMAC-SHA256 is correct; HMAC-SHA1, SHA-512, or MD5 will not match |

---

## 3. Webhook Envelope

### 3.1 Common Envelope Fields

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#request-body
> **Reference:** https://developers.line.biz/en/reference/messaging-api/#webhook-event-objects

Every webhook is an HTTP POST with `Content-Type: application/json`. The top-level body is:

```json
{
  "destination": "U1234567890abcdef1234567890abcdef",
  "events": [
    {
      "type": "message",
      "message": { "type": "text", "id": "14353798921116", "text": "Hello, world" },
      "timestamp": 1625665242211,
      "source": { "type": "user", "userId": "U80696558e1aa831..." },
      "replyToken": "757913772c4646b784d4b7ce46d12671",
      "mode": "active",
      "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
      "deliveryContext": { "isRedelivery": false }
    }
  ]
}
```

**Delivery behavior:**
- A single webhook POST may contain **multiple event objects** in the `events` array. There is not necessarily one user per request. A message event from person A and a follow event from person B can be batched in the same POST.
- The LINE Platform may send a POST with an **empty `events` array** to verify that the webhook URL is reachable. The server must return HTTP 200.
- The server must respond with **HTTP 200** within a reasonable time. LINE documentation notes that asynchronous processing is recommended to avoid delaying the response while handling events.
- If the server fails to return a 2xx response, LINE may redeliver the webhook (see `deliveryContext.isRedelivery`).

**Top-level envelope fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `destination` | String | Yes | User ID of the bot that should receive these events. Matches the bot's `userId`, not the event sender. Format: `U[0-9a-f]{32}`. |
| `events` | Array | Yes | Array of webhook event objects. May be empty for verification requests. |

**Common per-event fields** (present in every event object):

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | String | Yes | Event type identifier. Values: `message`, `unsend`, `follow`, `unfollow`, `join`, `leave`, `memberJoined`, `memberLeft`, `postback`, `videoPlayComplete`, `beacon`, `accountLink`, `membership`. |
| `mode` | String | Yes | Channel state. `active` — normal operation, bot may send replies. `standby` — another module channel is active; this bot should not send replies. |
| `timestamp` | Number | Yes | UNIX time in **milliseconds** when the event occurred. For redelivered webhooks, this is the original event time, not the redelivery time. |
| `source` | Object | No | Identifies the user and context. See §3.2. May be absent on some events (e.g., leave events where the user who removed the bot cannot be identified). |
| `webhookEventId` | String | Yes | Unique ID for this webhook event delivery, in ULID format. Use for idempotency — the same value is preserved on redelivery. |
| `deliveryContext` | Object | Yes | Contains `isRedelivery`. |
| `deliveryContext.isRedelivery` | Boolean | Yes | `false` on first delivery. `true` if LINE is retrying a previously failed delivery. |
| `replyToken` | String | No | Token used to reply via the Reply Message API. Present only on events that support replies. Valid for a short time only (approximately 1 minute). Not present on events that do not support replies (e.g., `unfollow`, `leave`, `memberLeft`, `unsend`). |

**Webhook redelivery:**
- Redelivery is **disabled by default**. Enable it in the LINE Developers Console under **Messaging API** tab → **Webhook redelivery**.
- When enabled, LINE retries failed webhooks (non-2xx responses) for a pre-defined number of attempts at undisclosed intervals.
- Redelivery does not guarantee delivery. If redeliveries surge and affect LINE Platform operations, LINE may force-disable redelivery.
- Use `webhookEventId` as the deduplication key. With redelivery enabled, the ordering of delivered events may differ from the order they occurred — use `timestamp` to reconstruct ordering if needed.

---

### 3.2 Source Object

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#webhook-event-objects

The `source` object identifies who triggered the event and in what context.

**User source (1-on-1 chat / DM):**

```json
{
  "type": "user",
  "userId": "U4af4980629..."
}
```

**Group chat source:**

```json
{
  "type": "group",
  "groupId": "Ca56f94637c...",
  "userId": "U4af4980629..."
}
```

**Multi-person chat (room) source:**

```json
{
  "type": "room",
  "roomId": "Ra8dbf4673c...",
  "userId": "U4af4980629..."
}
```

**Source object field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | String | Yes | Context type. Values: `user`, `group`, `room`. |
| `userId` | String | No | User ID of the user who triggered the event. May be absent in events where the triggering user cannot be identified (e.g., `memberLeft` events in some cases). |
| `groupId` | String | No | Group chat ID. Present only when `type` is `group`. |
| `roomId` | String | No | Multi-person chat (room) ID. Present only when `type` is `room`. |

**Notes on source variants:**
- `type: user` — events from 1-on-1 chats between a user and the bot. `userId` is always present.
- `type: group` — events from group chats (3+ participants). `userId` of the event originator is present on most events but may be absent on `join` events (when the inviter is unknown) and `leave` events.
- `type: room` — events from multi-person chats (LINE's "rooms" feature, distinct from groups). Same rules as group regarding `userId` presence.

---

## 4. Webhook Event Types and Payloads

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#webhook-event-objects

Each subsection below follows this pattern:
1. One-line description
2. Reference URL
3. Full JSON example
4. Field table (Field / Type / Required / Notes)
5. Constraints (where applicable)

---

### 4.1 message — text

Fires when a user sends a text message in a 1-on-1 chat, group chat, or multi-person chat.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#wh-text

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
    "text": "@All @example_bot Good morning!! (love)",
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
          "length": 12,
          "userId": "U49585cd0d5...",
          "type": "user",
          "isSelf": true
        }
      ]
    },
    "quotedMessageId": "444573844083572700"
  }
}
```

**`message` object fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | String | Yes | Always `text`. |
| `id` | String | Yes | Unique message ID. Text content is delivered only in this webhook; there is no API to retrieve it again after the fact. |
| `quoteToken` | String | Yes | Token used to quote this message in a subsequent reply. |
| `text` | String | Yes | Text content of the message. LINE emoji shortcodes appear as their Unicode characters. Max 5000 characters. |
| `emojis` | Array | No | Present when the message contains LINE-native emoji. Each entry describes one emoji occurrence in `text`. |
| `emojis[].index` | Number | Yes | Zero-based character index in `text` where the emoji placeholder starts. |
| `emojis[].length` | Number | Yes | Length (in characters) of the emoji placeholder in `text`. |
| `emojis[].productId` | String | Yes | ID of the emoji product set. |
| `emojis[].emojiId` | String | Yes | ID of the specific emoji within the product set. |
| `mention` | Object | No | Present when the message contains `@` mentions. |
| `mention.mentionees` | Array | Yes | Array of mention objects. |
| `mention.mentionees[].index` | Number | Yes | Zero-based character index in `text` where the mention starts. |
| `mention.mentionees[].length` | Number | Yes | Length of the mention text. |
| `mention.mentionees[].type` | String | Yes | `user` for a specific user mention; `all` for `@All` (mention all members). |
| `mention.mentionees[].userId` | String | No | User ID of the mentioned user. Present only when `type` is `user`. |
| `mention.mentionees[].isSelf` | Boolean | No | `true` if the mention targets the bot itself. Present only when `type` is `user`. |
| `quotedMessageId` | String | No | Message ID of the message being quoted/replied-to. Present only if the user used LINE's quote/reply feature. The quoted message content cannot be retrieved via API. |

**Constraints:**
- `text` has a maximum length of 5000 characters.
- Template messages and Flex Messages sent via `liff.sendMessages()` do not generate a webhook to the bot server.
- When a message is quoted, `quotedMessageId` is set, but the quoted message content cannot be fetched again — only the ID is provided.

---

### 4.2 message — image

Fires when a user sends an image.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#wh-image

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

When the image is hosted externally (e.g., via LIFF):

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

**`message` object fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | String | Yes | Always `image`. |
| `id` | String | Yes | Unique message ID. Use with `GET /v2/bot/message/{messageId}/content` when `contentProvider.type` is `line`. |
| `quoteToken` | String | Yes | Token to quote this message. |
| `contentProvider` | Object | Yes | Describes where the image content is hosted. |
| `contentProvider.type` | String | Yes | `line` — stored on LINE servers, download via Get Content API. `external` — hosted by a third party, use the URLs directly. |
| `contentProvider.originalContentUrl` | String | No | URL of the full-size image. Present only when `type` is `external`. |
| `contentProvider.previewImageUrl` | String | No | URL of the preview/thumbnail image. Present only when `type` is `external`. |
| `imageSet` | Object | No | Present when the user sent multiple images at once as a set. |
| `imageSet.id` | String | Yes | Shared ID for all images in the set. |
| `imageSet.index` | Number | Yes | 1-based index of this image within the set. |
| `imageSet.total` | Number | Yes | Total number of images in the set. May be `0` if the total cannot be determined at delivery time. |

**Constraints:**
- Content hosted on LINE servers (`contentProvider.type: line`) expires after a certain period. Retrieve content promptly after receiving the webhook. The exact expiry window is not publicly documented.
- Use `GET /v2/bot/message/{messageId}/content/preview` to retrieve a smaller preview image.

---

### 4.3 message — video

Fires when a user sends a video.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#wh-video

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
    "duration": 60000,
    "contentProvider": {
      "type": "line"
    }
  }
}
```

**`message` object fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | String | Yes | Always `video`. |
| `id` | String | Yes | Unique message ID. Use with Get Content API when `contentProvider.type` is `line`. |
| `quoteToken` | String | Yes | Token to quote this message. |
| `duration` | Number | No | Length of the video in **milliseconds**. May be absent. |
| `contentProvider` | Object | Yes | Where the video is hosted. |
| `contentProvider.type` | String | Yes | `line` or `external`. |
| `contentProvider.originalContentUrl` | String | No | Video file URL. Present only when `type` is `external`. |
| `contentProvider.previewImageUrl` | String | No | Thumbnail image URL. Present only when `type` is `external`. |

**Constraints:**
- Content on LINE servers expires after an undisclosed period; retrieve promptly.
- Use `GET /v2/bot/message/{messageId}/content/preview` for a preview image of the video.
- Before downloading, you can check preparation status with `GET /v2/bot/message/{messageId}/content/transcoding`.

---

### 4.4 message — audio

Fires when a user sends a voice message or audio file.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#wh-audio

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
    "duration": 60000,
    "contentProvider": {
      "type": "line"
    }
  }
}
```

**`message` object fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | String | Yes | Always `audio`. |
| `id` | String | Yes | Unique message ID. Use with Get Content API when `contentProvider.type` is `line`. |
| `duration` | Number | No | Length of the audio in **milliseconds**. Optional — may be absent. |
| `contentProvider` | Object | Yes | Where the audio is hosted. |
| `contentProvider.type` | String | Yes | `line` or `external`. |
| `contentProvider.originalContentUrl` | String | No | Audio file URL. Present only when `type` is `external`. |

**Constraints:**
- Audio does not have a preview image endpoint.
- Content on LINE servers expires after an undisclosed period.
- `quoteToken` is not present in audio message objects (audio messages cannot be quoted).

---

### 4.5 message — file

Fires when a user sends a file (e.g., PDF, Word document, Excel spreadsheet).

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#wh-file

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
    "fileName": "example-report.pdf",
    "fileSize": 2138
  }
}
```

**`message` object fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | String | Yes | Always `file`. |
| `id` | String | Yes | Unique message ID. Use with `GET /v2/bot/message/{messageId}/content` to download the file. |
| `fileName` | String | Yes | Original filename as provided by the sender. |
| `fileSize` | Number | Yes | File size in **bytes**. |

**Constraints:**
- File messages can only be sent in 1-on-1 chats (not in group chats or rooms).
- Content on LINE servers expires after an undisclosed period; download promptly.
- `quoteToken` is not present in file message objects.
- `contentProvider` is not present for file messages — files are always stored on LINE servers.

---

### 4.6 message — location

Fires when a user sends a location pin.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#wh-location

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
    "title": "my location",
    "address": "1-3 Kioicho, Chiyoda-ku, Tokyo, 102-8282 Japan",
    "latitude": 35.67966,
    "longitude": 139.73669
  }
}
```

**`message` object fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | String | Yes | Always `location`. |
| `id` | String | Yes | Unique message ID. |
| `title` | String | No | Location label set by the user. May be absent. |
| `address` | String | No | Human-readable address string. May be absent. |
| `latitude` | Number | Yes | Latitude in decimal degrees. |
| `longitude` | Number | Yes | Longitude in decimal degrees. |

**Constraints:**
- `quoteToken` is not present in location message objects.
- There is no Get Content API call for location — all data is delivered inline in the webhook.

---

### 4.7 message — sticker

Fires when a user sends a sticker.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#wh-sticker

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
    "packageId": "11537",
    "stickerId": "52002738",
    "stickerResourceType": "ANIMATION",
    "keywords": ["cony", "sally", "Staring", "thinking"],
    "text": "Let's hang out this weekend!",
    "quotedMessageId": "444573844083572700"
  }
}
```

**`message` object fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | String | Yes | Always `sticker`. |
| `id` | String | Yes | Unique message ID. |
| `quoteToken` | String | Yes | Token to quote this sticker message. |
| `packageId` | String | Yes | Sticker package (set) ID. |
| `stickerId` | String | Yes | Individual sticker ID within the package. |
| `stickerResourceType` | String | Yes | Resource type. Values: `STATIC`, `ANIMATION`, `SOUND`, `ANIMATION_SOUND`, `POPUP`, `POPUP_SOUND`, `CUSTOM`, `MESSAGE`, `NAME_TEXT`, `PER_STICKER_TEXT`. |
| `keywords` | Array of String | No | Up to 15 descriptive keywords associated with the sticker. |
| `text` | String | No | Text entered by the user for message stickers (`MESSAGE` or `PER_STICKER_TEXT` resource types). Absent for non-text stickers. |
| `quotedMessageId` | String | No | ID of the message being quoted. Present only if the user used the quote/reply feature with this sticker. |

---

### 4.8 follow

Fires when a user adds the LINE Official Account as a friend, or when a user who had previously blocked the account unblocks it.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#follow-event

```json
{
  "type": "follow",
  "mode": "active",
  "timestamp": 1625665242214,
  "source": {
    "type": "user",
    "userId": "Ufc729a925b3abef..."
  },
  "webhookEventId": "01FZ74ASS536FW97EX38NKCZQK",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "bb173f4d9cf64aed9d408ab4e36339ad",
  "follow": {
    "isUnblocked": false
  }
}
```

**Event-specific fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `replyToken` | String | Yes | Token for sending a welcome/greeting reply. |
| `follow` | Object | Yes | Contains follow-specific data. |
| `follow.isUnblocked` | Boolean | Yes | `false` — the user is adding the account as a friend for the first time (or re-adding after unfollowing). `true` — the user had previously blocked the account and is now unblocking it. |

**Constraints:**
- Only fires in 1-on-1 chats (`source.type: user`). Does not fire in group or room contexts.
- A reply using the `replyToken` can be used to send a welcome message.

---

### 4.9 unfollow

Fires when a user **blocks** the LINE Official Account. No `replyToken` is provided because the user will not receive replies after blocking.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#unfollow-event

```json
{
  "type": "unfollow",
  "mode": "active",
  "timestamp": 1625665242215,
  "source": {
    "type": "user",
    "userId": "Ubbd4f124aee5113..."
  },
  "webhookEventId": "01FZ74B5Y0F4TNKA5SCAVKPEDM",
  "deliveryContext": { "isRedelivery": false }
}
```

**Event-specific fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| *(none beyond common fields)* | — | — | No event-specific payload. No `replyToken`. |

**Constraints:**
- Only fires in 1-on-1 chat context (`source.type: user`). Does not fire in group or room contexts.
- The user has blocked the account; no messages can be sent to them until they unblock.

---

### 4.10 join

Fires when the LINE Official Account is **invited to and joins** a group chat or multi-person chat (room).

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#join-event

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

**Event-specific fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `replyToken` | String | Yes | Token for sending a greeting message to the group or room. |
| `source.type` | String | Yes | `group` or `room`. |
| `source.userId` | String | No | User ID of the inviter. May be absent if the inviter's identity cannot be determined. |

**Constraints:**
- Does not fire in 1-on-1 chat contexts.
- For groups, fires when the bot is explicitly invited. For rooms, fires when the first event in the room is generated after the bot is added.

---

### 4.11 leave

Fires when the LINE Official Account is **removed from a group chat or room** by a user, or when the bot itself calls the Leave Group or Leave Room API.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#leave-event

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

**Event-specific fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| *(none beyond common fields)* | — | — | No `replyToken` (the bot is no longer in the chat). |

**Constraints:**
- Does not fire in 1-on-1 chat contexts.
- No `userId` is included in the source — the identity of who removed the bot is not disclosed.

---

### 4.12 memberJoined

Fires when one or more users **join a group chat or multi-person chat** that the LINE Official Account is already a member of.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#member-joined-event

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

**Event-specific fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `replyToken` | String | Yes | Token for sending a welcome message to the new members. |
| `joined` | Object | Yes | Contains the list of users who joined. |
| `joined.members` | Array | Yes | Array of user objects. May contain multiple users if several joined simultaneously. |
| `joined.members[].type` | String | Yes | Always `user`. |
| `joined.members[].userId` | String | Yes | User ID of the joining member. |

**Constraints:**
- Does not fire in 1-on-1 chat contexts.
- Multiple users can appear in a single event if they were added at the same time.

---

### 4.13 memberLeft

Fires when one or more users **leave or are removed from a group chat or multi-person chat** that the bot is a member of.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#member-left-event

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

**Event-specific fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `left` | Object | Yes | Contains the list of users who left. |
| `left.members` | Array | Yes | Array of user objects. |
| `left.members[].type` | String | Yes | Always `user`. |
| `left.members[].userId` | String | Yes | User ID of the departing member. |

**Constraints:**
- No `replyToken` — cannot reply after a member leaves.
- Does not fire in 1-on-1 chat contexts.
- Multiple users can appear in a single event.

---

### 4.14 postback

Fires when a user triggers a **postback action** — for example by tapping a button in a template message, selecting a quick reply option, tapping a rich menu item, or using a datetime picker.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#postback-event

**Basic postback (button tap):**

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

**Datetime picker result:**

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

**Rich menu switch action result:**

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

**Event-specific fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `replyToken` | String | Yes | Token to reply to this postback action. |
| `postback` | Object | Yes | Contains the postback data. |
| `postback.data` | String | Yes | Data string defined in the postback action configuration. Up to 300 characters. |
| `postback.params` | Object | No | Present only for datetime picker or rich menu switch actions. |
| `postback.params.date` | String | No | Selected date in `YYYY-MM-DD` format. Present for date picker mode. |
| `postback.params.time` | String | No | Selected time in `HH:mm` format. Present for time picker mode. |
| `postback.params.datetime` | String | No | Selected datetime in `YYYY-MM-DDThh:mm` format. Present for datetime picker mode. |
| `postback.params.newRichMenuAliasId` | String | No | Alias ID of the rich menu that was switched to. Present for rich menu switch actions. |
| `postback.params.status` | String | No | Result of rich menu switch. Values: `SUCCESS`, `RICHMENU_ALIAS_ID_NOTFOUND`, `RICHMENU_NOTFOUND`, `FAILED`. |

**Constraints:**
- `postback.data` max length is 300 characters.
- A postback event can originate from 1-on-1 chats, group chats, and rooms.

---

### 4.15 beacon

Fires when a user's LINE app detects a **LINE Beacon** hardware device (a Bluetooth Low Energy transmitter).

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#beacon-event

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

**Event-specific fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `replyToken` | String | Yes | Token for sending a location-contextual reply. |
| `beacon` | Object | Yes | Contains beacon event data. |
| `beacon.hwid` | String | Yes | Hardware ID of the beacon device that was detected. |
| `beacon.type` | String | Yes | Beacon interaction type. Values: `enter` (user entered beacon range), `banner` (user tapped the beacon banner in the LINE app), `stay` (periodic event while user remains in range). |
| `beacon.dm` | String | No | Device message in hexadecimal — an optional string set by the beacon owner. Absent if the beacon does not support device messages. |

**Constraints:**
- Only fires in 1-on-1 chat context (`source.type: user`).
- Requires LINE Beacon hardware paired with the bot. For more information see: https://developers.line.biz/en/docs/messaging-api/using-beacons/

---

### 4.16 accountLink

Fires when a user completes or fails the **account linking** flow — the process of associating their LINE account with their account in your service.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#account-link-event

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

**Event-specific fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `replyToken` | String | No | Token for sending a linking confirmation message. May be absent on failure. |
| `link` | Object | Yes | Contains the account link result. |
| `link.result` | String | Yes | `ok` — linking succeeded. `failed` — linking failed (e.g., user cancelled, or an impersonation attempt was detected). |
| `link.nonce` | String | Yes | The nonce generated by your service at the start of the linking flow. Use this to correlate the event to the linking session and verify identity. |

**Constraints:**
- Only fires in 1-on-1 chat context (`source.type: user`).
- On failure, `source` and `replyToken` may be absent.
- For full account linking flow documentation: https://developers.line.biz/en/docs/messaging-api/linking-accounts/

---

### 4.17 things

Fires for **LINE Things** IoT device integration events — device link, device unlink, and scenario execution results. Note: As of the current API reference ToC (verified March 2026), this event type is not listed in the primary Webhook Event Objects section of the reference ToC. It may be deprecated or moved to a separate feature area. The payload specification below is sourced from the previous round of research; treat it as potentially out-of-date and verify against the current reference before implementing.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#things-event

**deviceLink — Device linked to user's LINE account:**

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

**deviceUnlink — Device unlinked from user's LINE account:**

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

**scenarioResult — Automated scenario execution result:**

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

**Event-specific fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `replyToken` | String | No | Present on `link` and `unlink` events; absent on `scenarioResult`. |
| `things` | Object | Yes | Contains the things event data. |
| `things.deviceId` | String | Yes | Unique ID of the LINE Things device. |
| `things.type` | String | Yes | Sub-type: `link`, `unlink`, or `scenarioResult`. |
| `things.result` | Object | No | Present only when `things.type` is `scenarioResult`. |
| `things.result.scenarioId` | String | Yes | ID of the scenario that was executed. |
| `things.result.revision` | Number | Yes | Revision number of the scenario. |
| `things.result.startTime` | Number | Yes | UNIX timestamp in milliseconds when the scenario started. |
| `things.result.endTime` | Number | Yes | UNIX timestamp in milliseconds when the scenario ended. |
| `things.result.resultCode` | String | Yes | Execution result. Values include: `success`, `gatt_error`, `runtime_error`. |
| `things.result.bleNotificationPayload` | String | No | Base64-encoded BLE notification payload, if received during the scenario. |
| `things.result.actionResults` | Array | No | Per-action results from the scenario execution. |
| `things.result.actionResults[].type` | String | Yes | Result type: `binary` or `void`. |
| `things.result.actionResults[].data` | String | No | Base64-encoded binary data. Present only when `type` is `binary`. |

---

### 4.18 unsend

Fires when a user **unsends (deletes) a previously sent message**. The original message content is no longer accessible after this event.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#unsend-event

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

**Event-specific fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `unsend` | Object | Yes | Contains the unsend data. |
| `unsend.messageId` | String | Yes | Message ID of the message that was unsent. Use this to locate and delete the corresponding message from your own storage. |

**Constraints:**
- No `replyToken` — cannot reply to an unsend event.
- The unsent message content cannot be retrieved via any API after this event fires.
- LINE documentation recommends that service providers respect the user's intent: delete or make inaccessible any stored copy of the unsent message.
- Fires in both 1-on-1 chats and group/room contexts.

---

### 4.19 videoPlayComplete

Fires when a user **finishes watching a video message** that was sent by the bot and that had a `trackingId` specified at send time. This event only fires for bot-sent videos with a `trackingId`, not for user-sent videos.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#video-viewing-complete

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

**Event-specific fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `replyToken` | String | Yes | Token to reply after viewing completes (e.g., a follow-up prompt or survey). |
| `videoPlayComplete` | Object | Yes | Contains the tracking data. |
| `videoPlayComplete.trackingId` | String | Yes | The `trackingId` you assigned when sending the original video message. Use this to identify which video was watched. |

**Constraints:**
- Only fires in 1-on-1 chat contexts (`source.type: user`). Does not fire in group or room contexts.
- Only fires if the video message was sent with a `trackingId` property.
- Fires when the video finishes playing completely (not on partial views).

---

### 4.20 membership

Fires when a user **joins, renews, or leaves** a membership plan offered by the LINE Official Account. This is a relatively recent addition to the webhook event types.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#membership-event
> **Reference:** https://developers.line.biz/en/docs/messaging-api/use-membership-features/

```json
{
  "type": "membership",
  "mode": "active",
  "timestamp": 1625665242211,
  "source": {
    "type": "user",
    "userId": "U4af4980629..."
  },
  "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
  "deliveryContext": { "isRedelivery": false },
  "replyToken": "nHuyWiB7yP5Zw52FIkcQobQuGDXCTA",
  "membership": {
    "type": "joined",
    "membershipId": 42
  }
}
```

**Event-specific fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `replyToken` | String | Yes | Token to reply to this membership event. |
| `membership` | Object | Yes | Contains the membership event data. |
| `membership.type` | String | Yes | Membership action. Values: `joined` (user joined the membership), `left` (user left/cancelled the membership), `renewed` (user renewed/continued the membership). |
| `membership.membershipId` | Number | Yes | Numeric ID of the membership plan that the user joined, left, or renewed. |

**Constraints:**
- Only fires in 1-on-1 chat contexts (`source.type: user`).
- Requires the LINE Official Account to have a membership plan configured. For setup instructions see: https://developers.line.biz/en/docs/messaging-api/use-membership-features/
- The membership feature and this webhook event type were confirmed present in the API reference as of March 2026.

---

## Appendix: Event Type Quick-Reference

| Event type | `type` value | Has `replyToken` | Fires in 1-on-1 | Fires in group/room |
|---|---|---|---|---|
| Message (text/image/video/audio/file/location/sticker) | `message` | Yes | Yes | Yes |
| Unsend | `unsend` | No | Yes | Yes |
| Follow | `follow` | Yes | Yes | No |
| Unfollow | `unfollow` | No | Yes | No |
| Join | `join` | Yes | No | Yes |
| Leave | `leave` | No | No | Yes |
| Member joined | `memberJoined` | Yes | No | Yes |
| Member left | `memberLeft` | No | No | Yes |
| Postback | `postback` | Yes | Yes | Yes |
| Video viewing complete | `videoPlayComplete` | Yes | Yes | No |
| Beacon | `beacon` | Yes | Yes | No |
| Account link | `accountLink` | No (on failure) / Yes (on success) | Yes | No |
| Things | `things` | Yes (link/unlink only) | Yes | No |
| Membership | `membership` | Yes | Yes | No |
