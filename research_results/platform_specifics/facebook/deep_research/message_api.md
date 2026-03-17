# Facebook Messenger — Send Text Reply API

**Scope**: Sending a plain text message as a reply to a customer-initiated conversation via Messenger (Graph API).

---

## 1. Authentication

> **Reference**: https://developers.facebook.com/docs/facebook-login/guides/access-tokens/get-long-lived

### Required Permission
- `pages_messaging` — must be granted on the Page whose token you use

### Page Access Token

All Send API calls require a **Page Access Token** passed as `access_token` in the request.

#### Step 1 — Short-lived User Token (via Graph API Explorer or OAuth)
Obtained from the Meta developer console or your OAuth flow. Valid for ~1–2 hours.

#### Step 2 — Exchange for Long-lived User Token (~60 days)
```http
GET https://graph.facebook.com/v25.0/oauth/access_token
  ?grant_type=fb_exchange_token
  &client_id={APP_ID}
  &client_secret={APP_SECRET}
  &fb_exchange_token={SHORT_LIVED_USER_TOKEN}
```

Response:
```json
{ "access_token": "<long-lived-user-token>", "expires_in": 5183944 }
```

#### Step 3 — Get Long-lived Page Access Token (never expires under normal conditions)
```http
GET https://graph.facebook.com/v25.0/{USER_ID}/accounts
  ?access_token={LONG_LIVED_USER_TOKEN}
```

Response:
```json
{
  "data": [
    {
      "id": "<PAGE_ID>",
      "name": "My Page",
      "access_token": "<long-lived-page-token>",
      "perms": ["MODERATE", "ADVERTISE", ...]
    }
  ]
}
```

> **Security**: Exchange calls use the `APP_SECRET` — perform these server-side only, never from a client.

---

## 2. Send Text Reply

> **Reference**: https://developers.facebook.com/docs/messenger-platform/reference/send-api

### Endpoint
```
POST https://graph.facebook.com/v25.0/{PAGE_ID}/messages
```

### When You Can Reply
You can only send a `RESPONSE` message if the customer sent your Page a message **within the last 24 hours** (the standard messaging window). Outside this window, you cannot send unprompted text.

### Request

**Headers**
```
Content-Type: application/json
```

**Query parameter**
```
?access_token={PAGE_ACCESS_TOKEN}
```

**Body**
```json
{
  "recipient": {
    "id": "<PSID>"
  },
  "messaging_type": "RESPONSE",
  "message": {
    "text": "Hi! Thanks for reaching out. How can I help you today?"
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `recipient.id` | string | Yes | Page-Scoped User ID (PSID) from the incoming webhook event |
| `messaging_type` | enum | Yes | `RESPONSE` for replies within the 24-hour window |
| `message.text` | string | Yes | UTF-8, max 2000 characters |

### Success Response
```json
{
  "recipient_id": "<PSID>",
  "message_id": "m_AG5Hz2U..."
}
```

### Where to Get the PSID
The PSID comes from the incoming webhook `messaging` event:
```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "message": { "text": "Hello!" }
}
```
Store `sender.id` from the webhook payload and use it as `recipient.id` in your reply.

---

## 3. Send Image Attachment

> **Reference**: https://developers.facebook.com/docs/messenger-platform/send-messages/saving-assets

Same endpoint and auth as text replies. Replace `message.text` with `message.attachment`.

### Option A — Send by URL (no pre-upload)

```json
{
  "recipient": { "id": "<PSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "image",
      "payload": {
        "url": "https://example.com/photo.jpg",
        "is_reusable": true
      }
    }
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `attachment.type` | string | Yes | `"image"` |
| `payload.url` | string | Yes | Publicly accessible URL to the image |
| `payload.is_reusable` | bool | No | `true` saves the image on Meta's servers and returns an `attachment_id` in the response for reuse later |

Response when `is_reusable: true`:
```json
{
  "recipient_id": "<PSID>",
  "message_id": "m_AG5Hz2U...",
  "attachment_id": "1857777774821032"
}
```

### Option B — Pre-upload, then send by attachment_id

Use this when you want to reuse the same image across multiple sends without re-fetching the URL each time.

**Step 1 — Upload the image**

```bash
POST https://graph.facebook.com/v25.0/me/message_attachments?access_token={PAGE_ACCESS_TOKEN}
Content-Type: application/json

{
  "message": {
    "attachment": {
      "type": "image",
      "payload": {
        "url": "https://example.com/photo.jpg",
        "is_reusable": true
      }
    }
  }
}
```

Response:
```json
{ "attachment_id": "1857777774821032" }
```

> The `attachment_id` is private to the uploading Page — other Pages cannot use it.

**Step 2 — Send using the attachment_id**

```json
{
  "recipient": { "id": "<PSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "image",
      "payload": {
        "attachment_id": "1857777774821032"
      }
    }
  }
}
```

### Constraints
- Max file size: **25 MB**
- The image URL must be publicly accessible at send time (for Option A)
- Supported formats: JPEG, PNG, GIF, WebP (Meta accepts standard web image formats; JPEG/PNG are safest)

---

## 4. Send Video Attachment

> **Reference**: https://developers.facebook.com/docs/messenger-platform/send-messages/saving-assets

Identical structure to image. Change `type` to `"video"`.

### Option A — Send by URL

```json
{
  "recipient": { "id": "<PSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "video",
      "payload": {
        "url": "https://example.com/clip.mp4",
        "is_reusable": true
      }
    }
  }
}
```

### Option B — Send by attachment_id (pre-uploaded)

**Step 1 — Upload**
```
POST https://graph.facebook.com/v25.0/me/message_attachments?access_token={PAGE_ACCESS_TOKEN}
Content-Type: application/json

{
  "message": {
    "attachment": {
      "type": "video",
      "payload": {
        "url": "https://example.com/clip.mp4",
        "is_reusable": true
      }
    }
  }
}
```

Response: `{ "attachment_id": "..." }`

**Step 2 — Send**
```json
{
  "recipient": { "id": "<PSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "video",
      "payload": { "attachment_id": "<ATTACHMENT_ID>" }
    }
  }
}
```

### Constraints
- Max file size: **25 MB**
- Recommended format: **MP4** (H.264 video, AAC audio) — Meta does not publish a formal list of accepted codecs, but MP4 is the only format confirmed working in practice
- `is_reusable` is supported (same as image)

---

## 5. Send Audio Attachment

> **Reference**: https://developers.facebook.com/docs/messenger-platform/send-messages/saving-assets

Same structure as image/video. Change `type` to `"audio"`.

> **Note**: When uploading via the Attachment Upload API, you must set the `Content-Type` header to match the audio MIME type (e.g. `audio/mp3`). This is the one audio-specific difference from image/video uploads.

### Option A — Send by URL

```json
{
  "recipient": { "id": "<PSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "audio",
      "payload": {
        "url": "https://example.com/clip.mp3",
        "is_reusable": true
      }
    }
  }
}
```

### Option B — Send by attachment_id (pre-uploaded)

**Step 1 — Upload** (note the explicit `Content-Type` header)
```
POST https://graph.facebook.com/v25.0/me/message_attachments?access_token={PAGE_ACCESS_TOKEN}
Content-Type: audio/mp3

{
  "message": {
    "attachment": {
      "type": "audio",
      "payload": {
        "url": "https://example.com/clip.mp3",
        "is_reusable": true
      }
    }
  }
}
```

Response: `{ "attachment_id": "..." }`

**Step 2 — Send**
```json
{
  "recipient": { "id": "<PSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "audio",
      "payload": { "attachment_id": "<ATTACHMENT_ID>" }
    }
  }
}
```

### Constraints
- Max file size: **25 MB**
- Recommended format: **MP3** — Meta does not publish a formal list, but MP3 is the only format explicitly referenced in official docs
- `is_reusable` is supported

---

## 6. Send File Attachment

> **Reference**: https://developers.facebook.com/docs/messenger-platform/send-messages/saving-assets

Same structure as other attachment types. Change `type` to `"file"`.

### Option A — Send by URL

```json
{
  "recipient": { "id": "<PSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "file",
      "payload": {
        "url": "https://example.com/document.pdf",
        "is_reusable": true
      }
    }
  }
}
```

### Option B — Send by attachment_id (pre-uploaded)

**Step 1 — Upload**
```
POST https://graph.facebook.com/v25.0/me/message_attachments?access_token={PAGE_ACCESS_TOKEN}
Content-Type: application/json

{
  "message": {
    "attachment": {
      "type": "file",
      "payload": {
        "url": "https://example.com/document.pdf",
        "is_reusable": true
      }
    }
  }
}
```

Response: `{ "attachment_id": "..." }`

**Step 2 — Send**
```json
{
  "recipient": { "id": "<PSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "file",
      "payload": { "attachment_id": "<ATTACHMENT_ID>" }
    }
  }
}
```

### Constraints
- Max file size: **25 MB**
- Meta does not publish a formal allowlist of file types; common document formats (PDF, DOCX, XLSX, ZIP) are accepted in practice
- `is_reusable` is supported
- The file URL must be publicly accessible at send time (for Option A)

---

## 7. Send Template Messages

> **Reference**: https://developers.facebook.com/docs/messenger-platform/send-messages/templates

All templates share the same outer envelope — `message.attachment.type` is always `"template"`, and `payload.template_type` selects the variant.

```
POST https://graph.facebook.com/v25.0/{PAGE_ID}/messages?access_token={PAGE_ACCESS_TOKEN}
```

---

### 7.1 Button Template

> **Reference**: https://developers.facebook.com/docs/messenger-platform/reference/template/button

Displays a text message with 1–3 action buttons below it.

```json
{
  "recipient": { "id": "<PSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "template",
      "payload": {
        "template_type": "button",
        "text": "What would you like to do?",
        "buttons": [
          {
            "type": "web_url",
            "url": "https://example.com/order/123",
            "title": "View Order"
          },
          {
            "type": "postback",
            "title": "Track Shipment",
            "payload": "TRACK_ORDER_123"
          },
          {
            "type": "phone_number",
            "title": "Call Support",
            "payload": "+15551234567"
          }
        ]
      }
    }
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `text` | string | Yes | UTF-8, max 640 characters |
| `buttons` | array | Yes | 1–3 buttons |
| `buttons[].type` | enum | Yes | `web_url`, `postback`, `phone_number` |
| `buttons[].title` | string | Yes | Button label |
| `buttons[].url` | string | For `web_url` | Destination URL |
| `buttons[].payload` | string | For `postback` / `phone_number` | Developer string or E.164 phone number |

---

### 7.2 Generic Template

> **Reference**: https://developers.facebook.com/docs/messenger-platform/reference/template/generic

A horizontally scrollable card (or single card) with image, title, subtitle, and buttons.

```json
{
  "recipient": { "id": "<PSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "template",
      "payload": {
        "template_type": "generic",
        "sharable": false,
        "elements": [
          {
            "title": "Your Order #1234",
            "subtitle": "Estimated delivery: Mar 20",
            "image_url": "https://example.com/product.jpg",
            "default_action": {
              "type": "web_url",
              "url": "https://example.com/order/1234"
            },
            "buttons": [
              {
                "type": "postback",
                "title": "Track Order",
                "payload": "TRACK_1234"
              },
              {
                "type": "web_url",
                "url": "https://example.com/order/1234",
                "title": "View Details"
              }
            ]
          }
        ]
      }
    }
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `elements` | array | Yes | Max **1 element** (carousel removed in current API) |
| `elements[].title` | string | Yes | Max 80 characters |
| `elements[].subtitle` | string | No | Max 80 characters |
| `elements[].image_url` | string | No | Public image URL |
| `elements[].default_action` | object | No | URL button fields (minus `title`) — opens on card tap |
| `elements[].buttons` | array | No | Max 3 buttons; same button types as Button Template |
| `sharable` | bool | No | Shows native Messenger share button |

---

### 7.3 Media Template

> **Reference**: https://developers.facebook.com/docs/messenger-platform/reference/template/media

Sends a full-width image or video with optional buttons.

```json
{
  "recipient": { "id": "<PSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "template",
      "payload": {
        "template_type": "media",
        "sharable": true,
        "elements": [
          {
            "media_type": "image",
            "attachment_id": "1857777774821032",
            "buttons": [
              {
                "type": "web_url",
                "url": "https://example.com/product",
                "title": "Shop Now"
              }
            ]
          }
        ]
      }
    }
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `elements[].media_type` | enum | Yes | `image` or `video` |
| `elements[].attachment_id` | string | One of | ID from Attachment Upload API |
| `elements[].url` | string | One of | Public URL — mutually exclusive with `attachment_id` |
| `elements[].buttons` | array | No | Max 3 buttons |
| Max elements | — | — | 1 |

---

### 7.4 Receipt Template

> **Reference**: https://developers.facebook.com/docs/messenger-platform/reference/template/receipt

Sends an order confirmation / receipt. No buttons.

```json
{
  "recipient": { "id": "<PSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "template",
      "payload": {
        "template_type": "receipt",
        "recipient_name": "Jane Doe",
        "merchant_name": "Aura Wellness",
        "order_number": "ORDER-20260317-001",
        "currency": "USD",
        "payment_method": "Visa 4242",
        "order_url": "https://example.com/orders/001",
        "timestamp": "1742198400",
        "elements": [
          {
            "title": "Wellness Plan - Monthly",
            "subtitle": "Digital subscription",
            "quantity": 1,
            "price": 29.99,
            "currency": "USD",
            "image_url": "https://example.com/plan.jpg"
          }
        ],
        "address": {
          "street_1": "123 Main St",
          "city": "San Francisco",
          "postal_code": "94105",
          "state": "CA",
          "country": "US"
        },
        "summary": {
          "subtotal": 29.99,
          "shipping_cost": 0.00,
          "total_tax": 2.70,
          "total_cost": 32.69
        },
        "adjustments": [
          { "name": "Promo AURA10", "amount": -3.00 }
        ]
      }
    }
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `recipient_name` | string | Yes | Displayed on receipt |
| `order_number` | string | Yes | Must be unique |
| `currency` | string | Yes | ISO 4217 code (e.g. `USD`) |
| `payment_method` | string | Yes | Free text (e.g. `"Visa 4242"`) |
| `summary.total_cost` | number | Yes | |
| `merchant_name` | string | No | Defaults to Page name |
| `timestamp` | string | No | Unix epoch seconds |
| `elements` | array | No | Max 100 items |
| `address` | object | No | Shipping address |
| `summary.subtotal` | number | No | |
| `summary.shipping_cost` | number | No | |
| `summary.total_tax` | number | No | |
| `adjustments` | array | No | Discounts / fees |

---

## 8. Quick Replies

> **Reference**: https://developers.facebook.com/docs/messenger-platform/send-messages/quick-replies

Quick replies appear as tappable chips above the message input bar. They disappear once the user taps one or sends their own message.

Quick replies can be attached to **any message type** (text, image, template, etc.) by adding a `quick_replies` array to the `message` object.

### With a text message

```json
{
  "recipient": { "id": "<PSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "text": "Which topic can I help you with?",
    "quick_replies": [
      {
        "content_type": "text",
        "title": "Order Status",
        "payload": "QR_ORDER_STATUS"
      },
      {
        "content_type": "text",
        "title": "Billing",
        "payload": "QR_BILLING",
        "image_url": "https://example.com/icons/billing.png"
      },
      {
        "content_type": "user_phone_number"
      },
      {
        "content_type": "user_email"
      }
    ]
  }
}
```

### Fields

| Field | Type | Required | Notes |
|---|---|---|---|
| `content_type` | enum | Yes | `text`, `user_phone_number`, `user_email` |
| `title` | string | For `text` | Button label; max 20 characters |
| `payload` | string | For `text` | Developer-defined string returned in the postback webhook; max 1000 characters |
| `image_url` | string | No | Icon displayed next to the title; only for `content_type: text` |

### `content_type` behaviour

| Value | What the user sees | What you receive in the webhook |
|---|---|---|
| `text` | Custom label chip | `messaging.quick_reply.payload` with your `payload` string |
| `user_phone_number` | Phone number chip (pre-filled with user's number) | The user's phone number as the payload |
| `user_email` | Email chip (pre-filled with user's email) | The user's email address as the payload |

### Constraints
- Max **13** quick replies per message
- `title` max 20 characters; truncated beyond that
- `payload` max 1000 characters
- Quick replies vanish after the user taps one or sends any other message — they are one-shot

---

## 9. Sender Actions

> **Reference**: https://developers.facebook.com/docs/messenger-platform/reference/send-api

Sender actions control the chat UI state — showing a typing indicator or marking messages as seen. They are sent to the same endpoint but use `sender_action` instead of `message`.

```
POST https://graph.facebook.com/v25.0/{PAGE_ID}/messages?access_token={PAGE_ACCESS_TOKEN}
```

### typing_on — show typing indicator

```json
{
  "recipient": { "id": "<PSID>" },
  "sender_action": "typing_on"
}
```

### typing_off — hide typing indicator

```json
{
  "recipient": { "id": "<PSID>" },
  "sender_action": "typing_off"
}
```

### mark_seen — mark last message as read

```json
{
  "recipient": { "id": "<PSID>" },
  "sender_action": "mark_seen"
}
```

### Fields

| Field | Type | Required | Notes |
|---|---|---|---|
| `recipient.id` | string | Yes | PSID |
| `sender_action` | enum | Yes | `typing_on`, `typing_off`, `mark_seen` |

> `messaging_type` and `message` are **not included** when sending a sender action — they are mutually exclusive with `sender_action`.

### Constraints
- `typing_on` automatically times out after ~20 seconds if not cancelled with `typing_off`
- Sender actions do not return a `message_id` — the response is `{ "recipient_id": "<PSID>" }` only
- Rate limits apply the same as regular messages

---

## 10. Persona API (Custom Sender Name & Avatar)

> **Reference**: https://developers.facebook.com/docs/graph-api/reference/page/personas/

The Persona API lets a message appear as if sent by a named agent (e.g. "Aura Support — Dr. Lee") with a custom avatar, instead of the generic Page identity.

### Step 1 — Create a persona (one-time per agent)

```
POST https://graph.facebook.com/v25.0/{PAGE_ID}/personas?access_token={PAGE_ACCESS_TOKEN}
Content-Type: application/json

{
  "name": "Aura Support — Dr. Lee",
  "profile_picture_url": "https://example.com/avatars/dr-lee.jpg"
}
```

Response:
```json
{ "id": "1234567890" }
```

Store this `id` — it is stable and reusable across conversations.

### Step 2 — Send a message as the persona

Add `persona_id` to any standard send request:

```json
{
  "recipient": { "id": "<PSID>" },
  "messaging_type": "RESPONSE",
  "persona_id": "1234567890",
  "message": {
    "text": "Hi! I'm Dr. Lee from Aura Support. How can I help you today?"
  }
}
```

`persona_id` works with all message types (text, attachments, templates, quick replies).

### Delete a persona

```
DELETE https://graph.facebook.com/v25.0/{PERSONA_ID}?access_token={PAGE_ACCESS_TOKEN}
```

Response: `{ "success": true }`

### Fields

| Field | Type | Required | Notes |
|---|---|---|---|
| `name` | string | Yes (create) | Display name shown in the conversation |
| `profile_picture_url` | string | Yes (create) | Public URL; recommended square image |
| `persona_id` | string | Yes (send) | ID returned from persona creation |

### Constraints
- Personas are scoped to the Page — not visible across other Pages
- The `profile_picture_url` must be publicly accessible at persona creation time
- There is no documented hard limit on number of personas per Page

---

## 11. Handover Protocol — Pass / Take / Request Thread Control

> **Reference**: https://developers.facebook.com/docs/graph-api/reference/page/pass_thread_control/ · https://developers.facebook.com/docs/graph-api/reference/page/take_thread_control/ · https://developers.facebook.com/docs/graph-api/reference/page/request_thread_control/

The Handover Protocol lets multiple apps (e.g. a bot and a human agent inbox) share control of the same Messenger conversation. One app is the **Primary Receiver** (holds thread control by default); others are **Secondary Receivers**.

All three endpoints share the same auth and response shape: `{ "success": true }`.

```
POST https://graph.facebook.com/v25.0/{PAGE_ID}/<endpoint>?access_token={PAGE_ACCESS_TOKEN}
Content-Type: application/json
```

---

### 11.1 Pass Thread Control

The current thread owner (Primary or Secondary Receiver) hands control to another app.

```json
{
  "recipient": { "id": "<PSID>" },
  "target_app_id": "123456789",
  "metadata": "Transferring to live agent — billing issue"
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `recipient.id` | string | Yes | PSID of the conversation to hand over |
| `target_app_id` | string | No | App ID of the receiving Secondary Receiver; defaults to Page Inbox (`263902037430900`) if omitted |
| `metadata` | string | No | Free-text context passed to the receiving app; visible in the `messaging_handovers` webhook event |

**Page Inbox app ID**: `263902037430900` — use this to hand off to a human agent in Meta Business Suite / Inbox.

Response:
```json
{ "success": true }
```

---

### 11.2 Take Thread Control

The Primary Receiver forcibly reclaims thread control from a Secondary Receiver.

```json
{
  "recipient": { "id": "<PSID>" },
  "metadata": "Bot resuming after agent resolved the issue"
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `recipient.id` | string | Yes | PSID |
| `metadata` | string | No | Context string passed to the outgoing app |

> Only the **Primary Receiver** can call this endpoint. Secondary Receivers must use `request_thread_control` instead.

---

### 11.3 Request Thread Control

A Secondary Receiver signals to the Primary Receiver that it wants control. The Primary Receiver decides whether to pass it.

```json
{
  "recipient": { "id": "<PSID>" },
  "metadata": "Human agent wants to take over"
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `recipient.id` | string | Yes | PSID |
| `metadata` | string | No | Reason for the request; delivered to the Primary Receiver via webhook |

The Primary Receiver receives a `messaging_handovers` webhook event with `request_thread_control` and can then call `pass_thread_control` to grant it.

---

### Webhook events triggered

All three actions fire a `messaging_handovers` event to the affected apps:

| Action | Event received by |
|---|---|
| pass_thread_control | New thread owner (target app) |
| take_thread_control | App losing control |
| request_thread_control | Primary Receiver |

---

## 12. Common Errors (all message types)

| Code | Message | Fix |
|---|---|---|
| 10 | `Application does not have permission for this action` | App lacks `pages_messaging` permission or hasn't passed App Review for production |
| 100 | `Invalid parameter` | Malformed `recipient.id` or `messaging_type` |
| 200 | `Permissions error` | Page token doesn't have `MESSAGE` task capability on the Page |
| 551 | `This person isn't available right now` | User opted out of messages or blocked the Page |
| 613 | `Calls to this API have exceeded the rate limit` | Slow down; standard tier is 200 calls/hour per PSID |

---

## 13. Minimal cURL Example

```bash
curl -X POST "https://graph.facebook.com/v25.0/{PAGE_ID}/messages?access_token={PAGE_ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": { "id": "<PSID>" },
    "messaging_type": "RESPONSE",
    "message": { "text": "Hello! How can I help you?" }
  }'
```

---

## 14. Production Requirements

Before going live beyond app admins/testers:
- Submit your app for **App Review** and request `pages_messaging` advanced access
- Your business must complete **Meta Business Verification**
- The Page must be published (not in draft mode)
