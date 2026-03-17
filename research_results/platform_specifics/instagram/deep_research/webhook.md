# Instagram Messaging API — Webhook (Inbound) Integration Research

> **Scope:** CS integration hub receiving inbound messages from Instagram professional accounts via the Meta Messenger Platform / Instagram Messaging API.
>
> **Primary docs:**
> - Messenger Platform (Instagram): https://developers.facebook.com/docs/messenger-platform/instagram/
> - Instagram Messaging Webhooks: https://developers.facebook.com/docs/messenger-platform/instagram/features/webhook/
> - Instagram Platform Webhooks: https://developers.facebook.com/docs/instagram-platform/webhooks/
> - Graph API Webhooks (Getting Started): https://developers.facebook.com/docs/graph-api/webhooks/getting-started/
> - Graph API Webhooks Reference — Instagram: https://developers.facebook.com/docs/graph-api/webhooks/reference/instagram

---

## Table of Contents

1. [Setup — Manual Steps](#1-setup--manual-steps)
2. [Authentication / Signature Verification](#2-authentication--signature-verification)
3. [Webhook Event Payload Types](#3-webhook-event-payload-types)
   - 3.1 [Top-Level Envelope](#31-top-level-envelope)
   - 3.2 [`messages` field](#32-messages-field)
   - 3.3 [`messaging_seen`](#33-messaging_seen)
   - 3.4 [`messaging_postbacks`](#34-messaging_postbacks)
   - 3.5 [`messaging_optins`](#35-messaging_optins)
   - 3.6 [`messaging_referrals`](#36-messaging_referrals)
   - 3.7 [`messaging_handovers`](#37-messaging_handovers)
   - 3.8 [Standby Channel](#38-standby-channel)
4. [Known Limitations](#4-known-limitations)

---

## 1. Setup — Manual Steps

### 1.1 Create a Meta Developer Account

- Go to https://developers.facebook.com/ and click **Get Started**.
- Log in with a personal Facebook account.
- Accept Meta's developer policies and verify your account (phone or email).

**Official doc:** https://developers.facebook.com/docs/development/register/

---

### 1.2 Create a Meta App (Business Type)

- Navigate to **My Apps → Create App** in the App Dashboard: https://developers.facebook.com/apps/
- When asked for a use case, choose **"Other"**, then proceed.
- On the App Type screen, select **"Business"** — this is required to add the Instagram product.
- Provide your app name and contact email.
- Optionally link a Meta Business Account (can be done later, but required before going live).

**Official doc:** https://developers.facebook.com/docs/instagram-platform/create-an-instagram-app/
https://developers.facebook.com/docs/development/create-an-app/

---

### 1.3 Connect an Instagram Professional Account

Instagram Professional Accounts must be of type **Creator** or **Business**.

Two paths exist depending on the login flow your app uses:

| Path | Base URL | Login Flow | Relevant Permission Names |
|------|----------|------------|--------------------------|
| Instagram Login (newer) | `graph.instagram.com` | Instagram OAuth | `instagram_business_basic`, `instagram_business_manage_messages` |
| Facebook Login (legacy) | `graph.facebook.com` | Facebook OAuth (Page token) | `instagram_basic`, `instagram_manage_messages`, `pages_manage_metadata` |

For the Facebook Login path the Instagram account must be linked to a Facebook Page. In the App Dashboard, under **Instagram → Basic Display or Messenger**, add a test Instagram account or use Business Login to let external accounts authorise your app.

**Official doc:** https://developers.facebook.com/docs/instagram-platform/instagram-api-with-instagram-login/
https://developers.facebook.com/docs/messenger-platform/instagram/

---

### 1.4 Enable Instagram Messaging in the App Dashboard

- In **App Dashboard → Add Products**, locate **Instagram** (or **Messenger**) and click **Set Up**.
- Under the Instagram product page, enable **"Instagram Messaging"**.
- This unlocks the Messenger Platform webhook fields (`messages`, `messaging_seen`, `messaging_postbacks`, etc.) for Instagram objects.

**Official doc:** https://developers.facebook.com/docs/messenger-platform/instagram/

---

### 1.5 Configure the Webhook Subscription

#### 1.5.1 Callback URL Requirements

- Must be an **HTTPS** endpoint with a valid TLS/SSL certificate from a trusted CA.
- Self-signed certificates are **not** accepted.
- Must respond to both GET (verification) and POST (event) requests.

#### 1.5.2 GET Verification Handshake

When you save a webhook in the App Dashboard, Meta sends a one-time `GET` request to your callback URL with the following query parameters:

| Parameter | Type | Description |
|-----------|------|-------------|
| `hub.mode` | String | Always `"subscribe"` |
| `hub.challenge` | Integer | Random integer; your endpoint must echo this value back |
| `hub.verify_token` | String | The token string you configured in the App Dashboard |

Your endpoint must:
1. Confirm `hub.mode == "subscribe"`.
2. Confirm `hub.verify_token` matches your stored secret token.
3. Respond with HTTP `200` and the plain-text body of `hub.challenge`.

If you return anything else (or respond too slowly), Meta will reject the subscription.

**Official doc:** https://developers.facebook.com/docs/graph-api/webhooks/getting-started/

#### 1.5.3 Subscribing via the App Dashboard

- In **App Dashboard → Instagram → Webhooks** (or **Messenger → Webhooks**), enter:
  - **Callback URL** — your HTTPS endpoint.
  - **Verify Token** — a secret string you define (e.g., a UUID).
- Click **Verify and Save**.
- After verification, enable the individual webhook fields (see §1.5.4).

#### 1.5.4 Selected Webhook Fields

Subscribe to these fields under the `instagram` webhook object:

| Field | Trigger |
|-------|---------|
| `messages` | Inbound text, media, story replies/mentions, reactions, deletes, shares, quick replies |
| `messaging_seen` | Read receipts — recipient has seen a message |
| `messaging_postbacks` | User taps an Icebreaker, Generic Template button, or persistent menu item |
| `messaging_optins` | User opts in to recurring (marketing) notifications |
| `messaging_referrals` | User enters conversation via an `ig.me` link with a `ref` param |
| `messaging_handovers` | Thread control passed between Primary and Secondary Receiver apps |
| `standby` | Events received while your app is NOT the thread owner |

⚠️ The field name used internally by the Instagram webhook object is `messaging_referral` (singular) in some contexts. The subscription field name shown in the dashboard and Graph API Webhooks reference may differ by one letter. Verify against the dashboard at time of implementation.

#### 1.5.5 Programmatic Subscription (Page-level)

For the Facebook Login path, after OAuth you must also call:

```
POST /{page-id}/subscribed_apps
  ?subscribed_fields=messages,messaging_seen,messaging_postbacks,messaging_optins,messaging_referrals,messaging_handovers,standby
  &access_token={page-access-token}
```

For the Instagram Login path, subscribe via:

```
POST /me/subscribed_apps
  ?subscribed_fields=messages,messaging_seen,messaging_postbacks,messaging_optins,messaging_referrals,messaging_handovers
  &access_token={instagram-user-access-token}
```

**Official doc:** https://developers.facebook.com/docs/instagram-platform/webhooks/

---

### 1.6 Required Permissions

#### Facebook Login path

| Permission | Purpose |
|------------|---------|
| `instagram_basic` | Read basic Instagram account info |
| `instagram_manage_messages` | Send and receive Instagram DMs |
| `pages_manage_metadata` | Subscribe to Page-level webhooks |
| `pages_show_list` | List Pages the user manages |

#### Instagram Login path

| Permission | Purpose |
|------------|---------|
| `instagram_business_basic` | Read basic Instagram business account info |
| `instagram_business_manage_messages` | Send and receive Instagram DMs |

**Official doc:** https://developers.facebook.com/docs/permissions/
https://developers.facebook.com/docs/instagram-platform/instagram-api-with-instagram-login/messaging-api/

---

### 1.7 App Review for Production Access

In **development mode**, webhooks are only delivered to users who have a role on the app (Administrator, Developer, Tester).

To receive events from any Instagram user, you must:

1. **Switch the app to Live mode** in the App Dashboard.
2. **Complete App Review** for each permission requiring Advanced Access.

#### Permissions requiring Advanced Access

| Permission | Access Level Required for Production |
|------------|--------------------------------------|
| `instagram_manage_messages` | **Advanced Access** (for serving accounts you don't own) |
| `instagram_business_manage_messages` | **Advanced Access** |
| `pages_manage_metadata` | Advanced Access (varies) |

#### App Review Submission Requirements

- **Business Verification:** Your Meta Business Account must be verified.
- **Successful API calls:** Make at least 1 successful API call using each permission within 30 days of submitting.
- **Use-case description:** Explain why your app needs the permission, step-by-step user workflow, and how usage aligns with Meta's policies.
- **Screen recording:** Upload a video demonstrating the permission flow in your app.
- **App accessibility:** Your app must be publicly accessible, or you must provide tester credentials.
- **Data handling questions:** May be required for sensitive permissions.
- **Privacy policy URL:** Mandatory.

⚠️ Meta has deprecated old scope names as of January 27, 2025. Ensure you are using the current permission names listed above.

**Official doc:** https://developers.facebook.com/docs/instagram-platform/app-review/
https://developers.facebook.com/docs/resp-plat-initiatives/individual-processes/app-review/submission-guide

---

## 2. Authentication / Signature Verification

### 2.1 The `X-Hub-Signature-256` Header

Every inbound webhook `POST` from Meta includes the header:

```
X-Hub-Signature-256: sha256=<hex-digest>
```

The digest is an **HMAC-SHA256** of the raw request body, keyed with your **App Secret** (found in App Dashboard → Settings → Basic).

### 2.2 Verification Algorithm

1. Retrieve the raw (unparsed) request body as bytes.
2. Retrieve your App Secret.
3. Compute `HMAC-SHA256(key=app_secret, message=raw_body)` — produce a lowercase hex string.
4. Prepend `"sha256="` to form the expected signature.
5. Compare the expected signature to the value of `X-Hub-Signature-256` using a **constant-time** (timing-safe) comparison to prevent timing attacks.
6. If they match → process the event. If they differ → respond with `HTTP 403` and discard.

> **Critical:** You must read the **raw body bytes before** any JSON parsing. Parsing may normalise whitespace or key ordering, which would change the computed hash.

**Official doc:** https://developers.facebook.com/docs/graph-api/webhooks/getting-started/

### 2.3 Python Example

```python
import hashlib
import hmac
from flask import Flask, request, abort

app = Flask(__name__)
APP_SECRET = "YOUR_APP_SECRET"

@app.route("/webhook", methods=["POST"])
def webhook():
    signature_header = request.headers.get("X-Hub-Signature-256", "")
    if not signature_header.startswith("sha256="):
        abort(403)

    received_sig = signature_header[len("sha256="):]
    raw_body = request.get_data()  # raw bytes, before any parsing

    expected_sig = hmac.new(
        APP_SECRET.encode("utf-8"),
        raw_body,
        hashlib.sha256
    ).hexdigest()

    if not hmac.compare_digest(expected_sig, received_sig):
        abort(403)

    payload = request.get_json()
    # process payload ...
    return "OK", 200
```

### 2.4 Node.js Example

```javascript
const express = require("express");
const crypto = require("crypto");

const app = express();
const APP_SECRET = process.env.APP_SECRET;

// Must use raw body parser — NOT express.json() — before this middleware
app.use(express.raw({ type: "application/json" }));

app.post("/webhook", (req, res) => {
  const sigHeader = req.headers["x-hub-signature-256"] || "";
  if (!sigHeader.startsWith("sha256=")) {
    return res.sendStatus(403);
  }

  const receivedSig = sigHeader.slice("sha256=".length);
  const rawBody = req.body; // Buffer from express.raw()

  const expectedSig = crypto
    .createHmac("sha256", APP_SECRET)
    .update(rawBody)
    .digest("hex");

  // Timing-safe comparison
  const trusted = crypto.timingSafeEqual(
    Buffer.from(expectedSig, "hex"),
    Buffer.from(receivedSig, "hex")
  );

  if (!trusted) {
    return res.sendStatus(403);
  }

  const payload = JSON.parse(rawBody.toString("utf-8"));
  // process payload ...
  res.sendStatus(200);
});
```

> ⚠️ In Node.js, if you have already applied `express.json()` globally, the raw body will be consumed and unavailable. Apply `express.raw()` specifically to your webhook route, or capture the raw body via the `verify` option of `express.json`.

### 2.5 Response Requirements

- Respond with **HTTP 200** within a few seconds for every webhook `POST`.
- If processing takes longer, respond immediately with 200 and handle the event asynchronously.
- Failure to respond with 200 triggers retries: Meta retries immediately, then at decreasing intervals over **36 hours** before dropping the event.
- Implement idempotency — Meta may deliver the same event more than once.

---

## 3. Webhook Event Payload Types

### 3.1 Top-Level Envelope

All Instagram webhook notifications share the same outer structure:

```json
{
  "object": "instagram",
  "entry": [
    {
      "id": "<IG_USER_ID>",
      "time": 1569262486134,
      "messaging": [
        {
          "sender":    { "id": "<IGSID>" },
          "recipient": { "id": "<IG_USER_ID>" },
          "timestamp": 1569262485349,
          "<event_type>": { ... }
        }
      ]
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `object` | String | Always `"instagram"` for Instagram events |
| `entry` | Array | Array of change entries; typically one per notification |
| `entry[].id` | String | The Instagram professional account ID (IG User ID) receiving the event |
| `entry[].time` | Unix ms | Server-side timestamp of the notification batch |
| `entry[].messaging` | Array | Array of messaging events; **typically contains one item** |
| `messaging[].sender.id` | String | **IGSID** — the Instagram-scoped ID of the person who sent/triggered the event |
| `messaging[].recipient.id` | String | IG User ID of your professional account |
| `messaging[].timestamp` | Unix ms | Client-side timestamp of the event |

> ⚠️ For **standby channel** events the outer array key is `"standby"` instead of `"messaging"` (see §3.8).

**Official doc:** https://developers.facebook.com/docs/messenger-platform/instagram/features/webhook/
https://developers.facebook.com/docs/graph-api/webhooks/reference/instagram

---

### 3.2 `messages` Field

Triggered when an Instagram user sends a message to your professional account. The `message` object nested inside the outer envelope carries the event detail.

**Required webhook field subscription:** `messages`
**Required permissions:** `instagram_manage_messages` (or `instagram_business_manage_messages`)

#### 3.2.1 Text Message

```json
{
  "object": "instagram",
  "entry": [{
    "id": "17841400008460056",
    "time": 1569262486134,
    "messaging": [{
      "sender":    { "id": "1234567890" },
      "recipient": { "id": "17841400008460056" },
      "timestamp": 1569262485349,
      "message": {
        "mid": "m_ABCDEFG123456",
        "text": "Hello, I need help with my order"
      }
    }]
  }]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `message.mid` | String | Yes | Unique message ID |
| `message.text` | String | Yes (for text) | The message text content |

#### 3.2.2 Image Attachment

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "attachments": [
      {
        "type": "image",
        "payload": {
          "url": "https://cdn.example.com/image.jpg"
        }
      }
    ]
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `message.attachments` | Array | Yes | List of attachments |
| `attachments[].type` | String | Yes | `"image"` |
| `attachments[].payload.url` | String | Yes | CDN URL of the image |

#### 3.2.3 Video Attachment

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "attachments": [
      {
        "type": "video",
        "payload": {
          "url": "https://cdn.example.com/video.mp4"
        }
      }
    ]
  }
}
```

Same field structure as image; `type` is `"video"`.

#### 3.2.4 Audio Attachment

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "attachments": [
      {
        "type": "audio",
        "payload": {
          "url": "https://cdn.example.com/audio.m4a"
        }
      }
    ]
  }
}
```

Same field structure; `type` is `"audio"`.

> ⚠️ Voice messages sent via Instagram's microphone button are delivered as `audio` attachments. GIFs and stickers do **not** trigger webhook events.

#### 3.2.5 File Attachment

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "attachments": [
      {
        "type": "file",
        "payload": {
          "url": "https://cdn.example.com/document.pdf"
        }
      }
    ]
  }
}
```

Same field structure; `type` is `"file"`.

#### 3.2.6 Story Mention

Sent when the Instagram user mentions your professional account in their Story. The `payload.url` is a CDN link to the story media.

```json
{
  "object": "instagram",
  "entry": [{
    "id": "17841400008460056",
    "time": 1569262486134,
    "messaging": [{
      "sender":    { "id": "1234567890" },
      "recipient": { "id": "17841400008460056" },
      "timestamp": 1569262485349,
      "message": {
        "mid": "m_ABCDEFG123456",
        "attachments": [
          {
            "type": "story_mention",
            "payload": {
              "url": "https://cdn.example.com/story.jpg"
            }
          }
        ]
      }
    }]
  }]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `attachments[].type` | String | Yes | Always `"story_mention"` |
| `attachments[].payload.url` | String | Yes | CDN URL of the story; do NOT cache or persist the media content on your server |

> ⚠️ You may store the `url` string (reference) but Meta's policy prohibits storing the media content itself. The URL expires.

You can retrieve additional story context via:

```
GET /{message-id}?fields=story
```

**Official doc:** https://developers.facebook.com/docs/messenger-platform/instagram/features/story-mention/

#### 3.2.7 Story Reply

Sent when the user replies directly to one of your Stories (or you reply to theirs). The reply context is in `message.reply_to.story`.

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "text": "Love this!",
    "reply_to": {
      "story": {
        "url": "https://cdn.example.com/story.jpg",
        "id": "17858893269000001"
      }
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `message.reply_to` | Object | Yes | Present when message is a reply |
| `reply_to.story.url` | String | Yes | CDN URL of the story being replied to |
| `reply_to.story.id` | String | Yes | Media ID of the story |
| `message.text` | String | No | The reply text (if any) |

> ⚠️ Story reply webhooks do **not** trigger for GIF or sticker replies.

#### 3.2.8 Reaction Add

Sent when a user reacts to a message. Supported reactions: `angry`, `sad`, `wow`, `love`, `like`, `laugh`, `other`.

```json
{
  "sender":    { "id": "1234567890" },
  "recipient": { "id": "17841400008460056" },
  "timestamp": 1569262485349,
  "reaction": {
    "mid": "m_ABCDEFG123456",
    "action": "react",
    "reaction": "love",
    "emoji": "❤️"
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `reaction.mid` | String | Yes | ID of the message that received the reaction |
| `reaction.action` | String | Yes | `"react"` for adding a reaction |
| `reaction.reaction` | String | Yes | Reaction name: `angry`, `sad`, `wow`, `love`, `like`, `laugh`, `other` |
| `reaction.emoji` | String | Yes | Unicode emoji character for the reaction |

> ⚠️ Reaction events are delivered under the `messages` webhook field subscription but the event key inside the messaging object is `reaction`, not `message`.

#### 3.2.9 Reaction Remove

```json
{
  "reaction": {
    "mid": "m_ABCDEFG123456",
    "action": "unreact"
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `reaction.mid` | String | Yes | ID of the message whose reaction was removed |
| `reaction.action` | String | Yes | Always `"unreact"` |
| `reaction.reaction` | String | No | ⚠️ May or may not be present on unreact events — handle gracefully |

#### 3.2.10 Unsend (Message Deleted)

Sent when a user deletes (unsends) a message they previously sent.

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "is_deleted": true
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `message.mid` | String | Yes | ID of the message that was deleted |
| `message.is_deleted` | Boolean | Yes | Always `true` |

#### 3.2.11 Share — Link

When a user shares a URL/link into a DM with your account, it arrives as a `fallback` attachment type (unsupported generic share).

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "attachments": [
      {
        "type": "fallback",
        "payload": {
          "title": "Example Article Title",
          "url": "https://example.com/article"
        }
      }
    ]
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `attachments[].type` | String | Yes | `"fallback"` for generic link shares |
| `attachments[].payload.title` | String | No | Preview title of the shared content |
| `attachments[].payload.url` | String | No | URL of the shared link |

> ⚠️ Not all link shares produce a URL in the payload. Some may only include a `title`. Treat both fields as optional.

#### 3.2.12 Share — Media (Instagram Post/Reel)

When a user shares an Instagram post or Reel into the conversation:

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "attachments": [
      {
        "type": "ig_reel",
        "payload": {
          "url": "https://cdn.example.com/reel-thumbnail.jpg",
          "reel_video_id": "17858893269000001"
        }
      }
    ]
  }
}
```

For a shared Instagram image post, the type is `"image"` or `"reel"` depending on content type. Field structure mirrors §3.2.2.

> ⚠️ Media shared from private accounts does NOT include the media URL — the payload may be empty or the webhook may not fire. Instagram Stories, view-once (ephemeral) media, and disappearing content are not supported for media share.

#### 3.2.13 Ephemeral / View-Once Media

When a user sends a view-once (disappearing) media item:

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "attachments": [
      {
        "type": "ephemeral"
      }
    ]
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `attachments[].type` | String | Yes | `"ephemeral"` — no URL is provided |

#### 3.2.14 Product (Instagram Shop)

When a user shares a product from an Instagram Shop or a message is initiated from a product page:

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "referral": {
      "product": {
        "id": "1234567890"
      }
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `message.referral.product.id` | String | Yes | The Instagram catalog product ID that triggered the conversation |

#### 3.2.15 Referral Inside Message (Click-to-DM Ad)

When a message is initiated from a Click-to-DM (CTD) ad, the first inbound message includes a `referral` object:

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "text": "Hi",
    "referral": {
      "ref": "promo_summer_2024",
      "ad_id": "23843694560100001",
      "source": "ADS",
      "type": "OPEN_THREAD",
      "ads_context_data": {
        "ad_title": "Summer Sale",
        "photo_url": "https://cdn.example.com/ad.jpg",
        "video_url": "https://cdn.example.com/ad-thumb.jpg",
        "post_id": "17858893269000001",
        "product_id": "1234567890"
      }
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `message.referral.ref` | String | No | Custom ref data set in the ad |
| `message.referral.ad_id` | String | Yes | Meta ad ID |
| `message.referral.source` | String | Yes | `"ADS"` |
| `message.referral.type` | String | Yes | `"OPEN_THREAD"` |
| `message.referral.ads_context_data.ad_title` | String | No | Title of the ad |
| `message.referral.ads_context_data.photo_url` | String | No | Image URL from the ad |
| `message.referral.ads_context_data.video_url` | String | No | Video thumbnail URL from the ad |
| `message.referral.ads_context_data.post_id` | String | No | Post ID of the ad |
| `message.referral.ads_context_data.product_id` | String | No | Product ID if ad features a product |

#### 3.2.16 Quick Reply

When a user taps a quick reply button:

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "text": "Yes",
    "quick_reply": {
      "payload": "QUICK_REPLY_YES"
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `message.quick_reply.payload` | String | Yes | The developer-defined payload string attached to the quick reply button |
| `message.text` | String | No | The button label text (what the user tapped) |

#### 3.2.17 Inline Reply (Reply to Message)

When a user replies to a specific message within the thread (not a story reply):

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "text": "That sounds great",
    "reply_to": {
      "mid": "m_ORIGINAL_MESSAGE_ID"
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `message.reply_to.mid` | String | Yes | Message ID of the message being replied to |

#### 3.2.18 Unsupported Message

When a message type is not supported by the API:

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "is_unsupported": true
  }
}
```

#### 3.2.19 Echo Message (Sent by Your Business)

When your app or another agent sends a message from the professional account, an echo event is fired (only if `message_echoes` field is also subscribed):

```json
{
  "message": {
    "mid": "m_ABCDEFG123456",
    "text": "Thank you for contacting us",
    "is_echo": true
  }
}
```

---

### 3.3 `messaging_seen`

Sent when the Instagram user reads (opens) a conversation and sees your messages.

**Required webhook field subscription:** `messaging_seen`

```json
{
  "object": "instagram",
  "entry": [{
    "id": "17841400008460056",
    "time": 1569262486134,
    "messaging": [{
      "sender":    { "id": "1234567890" },
      "recipient": { "id": "17841400008460056" },
      "timestamp": 1569262485349,
      "read": {
        "mid": "m_ABCDEFG123456"
      }
    }]
  }]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `read` | Object | Yes | Read receipt object |
| `read.mid` | String | Yes | ID of the last message that was read |

> ⚠️ `messaging_seen` reflects that the user **saw** the message. It does not necessarily indicate that the user read every word — just that the conversation was opened.

**Official doc:** https://developers.facebook.com/docs/messenger-platform/instagram/features/webhook/

---

### 3.4 `messaging_postbacks`

Sent when a user taps a button in an Icebreaker (conversation starter) or a Generic Template button.

**Required webhook field subscription:** `messaging_postbacks`

```json
{
  "object": "instagram",
  "entry": [{
    "id": "17841400008460056",
    "time": 1569262486134,
    "messaging": [{
      "sender":    { "id": "1234567890" },
      "recipient": { "id": "17841400008460056" },
      "timestamp": 1569262485349,
      "postback": {
        "mid": "m_ABCDEFG123456",
        "title": "Track My Order",
        "payload": "POSTBACK_TRACK_ORDER"
      }
    }]
  }]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `postback` | Object | Yes | Postback event object |
| `postback.mid` | String | Yes | Message ID of the postback event |
| `postback.title` | String | Yes | Text label of the button the user tapped |
| `postback.payload` | String | Yes | Developer-defined payload string attached to the button |
| `postback.referral` | Object | No | Referral data if the postback was triggered via a referral source |

**Postback with referral (from a persistent menu or m.me link):**

```json
{
  "postback": {
    "mid": "m_ABCDEFG123456",
    "title": "Get Started",
    "payload": "GET_STARTED",
    "referral": {
      "ref": "campaign_abc",
      "source": "SHORTLINK",
      "type": "OPEN_THREAD"
    }
  }
}
```

**Official doc:** https://developers.facebook.com/docs/messenger-platform/instagram/features/webhook/

---

### 3.5 `messaging_optins`

Sent when a user opts in to receive recurring (marketing) notification messages. This fires when the user taps **"Allow"** on an opt-in prompt your app sends.

**Required webhook field subscription:** `messaging_optins`

```json
{
  "object": "instagram",
  "entry": [{
    "id": "17841400008460056",
    "time": 1569262486134,
    "messaging": [{
      "sender":    { "id": "1234567890" },
      "recipient": { "id": "17841400008460056" },
      "timestamp": 1569262485349,
      "optin": {
        "type": "notification_messages",
        "payload": "PROMO_CODE_SUBSCRIBE",
        "notification_messages_token": "ABCDEFGHIJ1234567890",
        "notification_messages_frequency": "WEEKLY",
        "notification_messages_timezone": "America/New_York",
        "token_expiry_timestamp": 1735689600,
        "user_token_status": "REFRESHED",
        "notification_messages_status": "RESUME NOTIFICATIONS",
        "title": "Weekly Deals"
      }
    }]
  }]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `optin` | Object | Yes | Opt-in event object |
| `optin.type` | String | Yes | Always `"notification_messages"` |
| `optin.payload` | String | No | Custom data you embedded in the opt-in request |
| `optin.notification_messages_token` | String | Yes | Marketing message token; use this instead of the PSID to send recurring messages to this user |
| `optin.notification_messages_frequency` | Enum | Yes | `DAILY` (token valid 6 months), `WEEKLY` (9 months), or `MONTHLY` (12 months) |
| `optin.notification_messages_timezone` | String | Yes | IANA timezone identifier for the recipient |
| `optin.token_expiry_timestamp` | Unix s | Yes | When the `notification_messages_token` expires |
| `optin.user_token_status` | Enum | No | `REFRESHED` (user renewed consent) or `NOT_REFRESHED` (token expired without renewal) |
| `optin.notification_messages_status` | Enum | No | `RESUME NOTIFICATIONS` or `STOP NOTIFICATIONS` — only present when user changes preference |
| `optin.title` | String | No | The notification topic title the user opted into |

> ⚠️ Store `notification_messages_token` — it is required to send outbound marketing messages. It is **different** from the user IGSID and is scoped to the specific opt-in topic and frequency.

**Official doc:** https://developers.facebook.com/docs/messenger-platform/send-messages/recurring-notifications

---

### 3.6 `messaging_referrals`

Sent when a user with an existing conversation thread enters via an `ig.me` link that includes a `ref` parameter (e.g. `https://ig.me/m/youraccount?ref=promo`). This is distinct from the referral embedded in the first inbound message from a new conversation (§3.2.15).

**Required webhook field subscription:** `messaging_referrals`

#### 3.6.1 ig.me Link Referral

```json
{
  "object": "instagram",
  "entry": [{
    "id": "17841400008460056",
    "time": 1569262486134,
    "messaging": [{
      "sender":    { "id": "1234567890" },
      "recipient": { "id": "17841400008460056" },
      "timestamp": 1569262485349,
      "referral": {
        "ref": "promo_summer_2024",
        "source": "SHORTLINK",
        "type": "OPEN_THREAD"
      }
    }]
  }]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `referral` | Object | Yes | Referral event object |
| `referral.ref` | String | Yes | Custom ref string from the link parameter (alphanumeric, `-`, `_`, `=`) |
| `referral.source` | String | Yes | Origin: `"SHORTLINK"` for `ig.me` links |
| `referral.type` | String | Yes | Currently always `"OPEN_THREAD"` |

#### 3.6.2 Click-to-DM Ad Referral (Existing Thread)

```json
{
  "referral": {
    "ref": "campaign_123",
    "ad_id": "23843694560100001",
    "source": "ADS",
    "type": "OPEN_THREAD",
    "ads_context_data": {
      "ad_title": "Summer Sale",
      "photo_url": "https://cdn.example.com/ad.jpg",
      "video_url": "https://cdn.example.com/ad-thumb.jpg",
      "post_id": "17858893269000001",
      "product_id": "1234567890",
      "flow_id": "PARTNER_FLOW_ID"
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `referral.source` | String | Yes | `"ADS"` for ad-originated referrals |
| `referral.ad_id` | String | Yes | Meta ad ID |
| `referral.ads_context_data.ad_title` | String | No | Ad title |
| `referral.ads_context_data.photo_url` | String | No | Ad image URL |
| `referral.ads_context_data.video_url` | String | No | Ad video thumbnail URL |
| `referral.ads_context_data.post_id` | String | No | Post ID of the associated post |
| `referral.ads_context_data.product_id` | String | No | Product ID if ad features a product |
| `referral.ads_context_data.flow_id` | String | No | Partner app flow identifier |

> ⚠️ A **Story CTA** (Story with a swipe-up or DM button) may also generate a referral event. The `source` value for story-originated referrals is not explicitly documented in the currently available Meta docs at the time of research — expected to be `"STORY"` or `"STORY_MENTION"` but treat this as unverified. Monitor the `source` field at runtime.

**Official doc:** https://developers.facebook.com/docs/messenger-platform/instagram/features/webhook/
https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messaging_referrals

---

### 3.7 `messaging_handovers`

Sent to apps participating in the **Handover Protocol** — a system that routes a conversation between a Primary Receiver app (main bot/CRM) and one or more Secondary Receiver apps (human agent inbox, etc.).

**Required webhook field subscription:** `messaging_handovers`

**Official doc:** https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messaging_handovers/

#### 3.7.1 `pass_thread_control`

Sent to the **new owner** (Secondary Receiver) when another app passes thread control to it.

```json
{
  "object": "instagram",
  "entry": [{
    "id": "17841400008460056",
    "time": 1458692752000,
    "messaging": [{
      "sender":    { "id": "1234567890" },
      "recipient": { "id": "17841400008460056" },
      "timestamp": 1458692752478,
      "pass_thread_control": {
        "new_owner_app_id": "123456789",
        "previous_owner_app_id": "987654321",
        "metadata": "Escalated to human agent"
      }
    }]
  }]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `pass_thread_control` | Object | Yes | Thread control transfer object |
| `pass_thread_control.new_owner_app_id` | String | Yes | App ID of the app receiving thread control |
| `pass_thread_control.previous_owner_app_id` | String | Yes | App ID of the app giving up thread control; `null` if previously idle |
| `pass_thread_control.metadata` | String | No | Custom string passed in the `pass_thread_control` API request |

#### 3.7.2 `take_thread_control`

Sent to the **losing app** (previous owner) when the Primary Receiver takes back thread control.

```json
{
  "object": "instagram",
  "entry": [{
    "id": "17841400008460056",
    "time": 1458692752000,
    "messaging": [{
      "sender":    { "id": "1234567890" },
      "recipient": { "id": "17841400008460056" },
      "timestamp": 1458692752478,
      "take_thread_control": {
        "previous_owner_app_id": "123456789",
        "new_owner_app_id": "987654321",
        "metadata": "Bot resuming control"
      }
    }]
  }]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `take_thread_control` | Object | Yes | Thread takeover object |
| `take_thread_control.previous_owner_app_id` | String | Yes | App ID of the app losing control |
| `take_thread_control.new_owner_app_id` | String | Yes | App ID of the app gaining control |
| `take_thread_control.metadata` | String | No | Custom string from the API request |

#### 3.7.3 `request_thread_control`

Sent to the **Primary Receiver** when a Secondary Receiver app requests thread ownership.

```json
{
  "object": "instagram",
  "entry": [{
    "id": "17841400008460056",
    "time": 1458692752000,
    "messaging": [{
      "sender":    { "id": "1234567890" },
      "recipient": { "id": "17841400008460056" },
      "timestamp": 1458692752478,
      "request_thread_control": {
        "requested_owner_app_id": 123456789,
        "metadata": "Agent requesting control"
      }
    }]
  }]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `request_thread_control` | Object | Yes | Thread control request object |
| `request_thread_control.requested_owner_app_id` | Number | Yes | App ID of the Secondary Receiver requesting control |
| `request_thread_control.metadata` | String | No | Custom string from the request |

#### 3.7.4 `app_roles`

Sent when a Page admin changes your app's role assignment.

```json
{
  "object": "instagram",
  "entry": [{
    "id": "17841400008460056",
    "time": 1458692752000,
    "messaging": [{
      "recipient": { "id": "17841400008460056" },
      "timestamp": 1458692752478,
      "app_roles": {
        "123456789": ["primary_receiver"]
      }
    }]
  }]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `app_roles` | Object | Yes | Map of `app_id` → array of role strings |
| `app_roles[app_id]` | Array | Yes | Roles assigned; values: `"primary_receiver"` or `"secondary_receiver"` |

---

### 3.8 Standby Channel

When your app is a **Secondary Receiver** (not the current thread owner), events are delivered to it via the `standby` channel rather than the normal `messaging` channel. This allows secondary apps to observe conversations without interrupting the primary receiver.

**Required webhook field subscription:** `standby`

**Official doc:** https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/standby/

#### Envelope Structure

```json
{
  "object": "instagram",
  "entry": [{
    "id": "17841400008460056",
    "time": 1458692752000,
    "standby": [
      {
        "sender":    { "id": "1234567890" },
        "recipient": { "id": "17841400008460056" },
        "timestamp": 1458692752478,
        "message": {
          "mid": "m_ABCDEFG123456",
          "text": "I need help"
        }
      }
    ]
  }]
}
```

The key difference is **`"standby"` replaces `"messaging"`** at the entry level.

#### Supported Events in Standby

| Event Type | Notes |
|------------|-------|
| `messages` | Full message payload as described in §3.2 |
| `messaging_postbacks` | Postback payload — **note: the `payload` field is omitted** for standby postbacks |
| `messaging_seen` / `read` | Read receipts |

> ⚠️ For `messaging_postbacks` delivered via the Standby channel, the `postback.payload` field is **not included** in the webhook. Only the app that originally sent the button receives the full payload. Your standby receiver will see `title` but not `payload`.

#### Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `entry[].standby` | Array | Contains messaging event objects for the secondary/standby receiver |
| `standby[].sender.id` | String | IGSID of the user who triggered the event |
| `standby[].recipient.id` | String | IG User ID of your professional account |
| `standby[].timestamp` | Unix ms | Event timestamp |

---

## 4. Known Limitations

| Limitation | Detail |
|------------|--------|
| GIFs and stickers | Do **not** trigger webhook events |
| Disappearing / view-once media | Delivered as `type: "ephemeral"` — no URL provided |
| Group messaging | Not supported |
| Media from private accounts | Media URL is not included in webhook payloads |
| Story replies with GIFs/stickers | Do not trigger webhooks |
| Creator accounts | Require explicit API consent before webhooks are delivered |
| Development mode | Webhooks only sent to users with app roles (Administrator, Developer, Tester) |
| Standard Access in Live mode | Webhooks still only go to role users; **Advanced Access required** for all users |
| Inactive request conversations | Conversations 30+ days old do not appear in API calls |
| Batch size | Maximum 1,000 updates per webhook POST |
| Retry window | Failed webhooks retried over 36 hours then dropped |
| Carousel reactions | Only the first image reaction is captured |
| Old scope deprecation | Legacy permission scope names deprecated January 27, 2025 |

---

*Research compiled 2026-03-17. All payload structures and field names should be verified against the live Meta developer documentation before production deployment, as Meta periodically updates the Messenger Platform API without backwards-compatibility guarantees.*
