# Instagram Messaging API — Send / Graph API Reference

> **Scope:** CS integration hub sending outbound messages to Instagram users via the Meta Messenger Platform (Facebook Login path) using the Graph API Send endpoint.
>
> **Primary docs:**
> - Messenger Platform (Instagram): https://developers.facebook.com/docs/messenger-platform/instagram/
> - Send Messages: https://developers.facebook.com/docs/messenger-platform/instagram/features/send-messages
> - Send API Reference: https://developers.facebook.com/docs/messenger-platform/reference/send-api/
> - Access Tokens & Permissions: https://developers.facebook.com/docs/instagram-platform/instagram-api-with-instagram-login/get-access-tokens-and-permissions
> - System User Tokens: https://developers.facebook.com/docs/business-management-apis/system-users/install-apps-and-generate-tokens/
> - Graph API Changelog: https://developers.facebook.com/docs/graph-api/changelog/versions/

---

## Table of Contents

1. [Setup — Manual Steps](#1-setup--manual-steps)
2. [Authentication / Authorization](#2-authentication--authorization)
3. [Send API — Endpoint Overview](#3-send-api--endpoint-overview)
4. [Text Messages](#4-text-messages)
5. [Attachments](#5-attachments)
   - 5.1 [Image](#51-image)
   - 5.2 [Video](#52-video)
   - 5.3 [Audio](#53-audio)
   - 5.4 [File Attachments — NOT SUPPORTED](#54-file-attachments--not-supported)
6. [Templates](#6-templates)
   - 6.1 [Generic Template](#61-generic-template)
   - 6.2 [Product Template](#62-product-template-catalog-linked)
7. [Quick Replies](#7-quick-replies)
8. [Sender Actions](#8-sender-actions)
9. [Profile Setup — Ice Breakers](#9-profile-setup--ice-breakers)
10. [Profile Setup — Persistent Menu](#10-profile-setup--persistent-menu)
11. [Handover Protocol](#11-handover-protocol)
    - 11.1 [Pass Thread Control](#111-pass-thread-control)
    - 11.2 [Take Thread Control](#112-take-thread-control)
    - 11.3 [Conversation Routing (Current Replacement)](#113-conversation-routing-current-replacement)
12. [Private Replies](#12-private-replies)
    - 12.1 [Reply to a Comment](#121-reply-to-a-comment)
    - 12.2 [Reply to a Story Mention](#122-reply-to-a-story-mention)
13. [Common Errors](#13-common-errors)
14. [Minimal cURL Example](#14-minimal-curl-example)
15. [Production Requirements](#15-production-requirements)

---

## 1. Setup — Manual Steps

> **Reference**: https://developers.facebook.com/docs/messenger-platform/instagram/

### 1.1 Prerequisites

- **Meta Developer Account** at https://developers.facebook.com/
- **Meta Business Manager** account at https://business.facebook.com/
- **Instagram Professional Account** — must be type **Business** or **Creator**; Personal accounts are NOT supported

### 1.2 Create a Meta App

- Go to https://developers.facebook.com/ → **My Apps** → **Create App**
- Select app type: **Business**
- Select use case: **"Manage messaging & content on Instagram"**
- Add the **Messenger Platform** product to the app

### 1.3 Link Instagram Professional Account to a Facebook Page

- In Meta Business Manager → **Accounts** → **Instagram Accounts** → **Add**
- The Instagram Business/Creator account must be connected to a Facebook Page
- This linkage is required for the Facebook Login (Messenger Platform) path used by this integration

### 1.4 Enable Instagram Messaging in the App Dashboard

- In the app dashboard, go to **Messenger** → **Settings**
- Under Instagram, connect the Facebook Page that is linked to your Instagram account
- Generate a Page Access Token (or use a System User Token — see Section 2)

### 1.5 Required Permissions

> **Reference**: https://developers.facebook.com/docs/instagram-platform/app-review/

| Field | Type | Required | Notes |
|---|---|---|---|
| `instagram_basic` | Permission | No App Review (Standard Access sufficient) | Foundational Instagram account access |
| `instagram_manage_messages` | Permission | Yes — Advanced Access required | Send and receive DMs; core permission for this integration |
| `pages_manage_metadata` | Permission | Yes — Advanced Access required | Webhook subscriptions and comment access |
| `pages_show_list` | Permission | Yes — Advanced Access required | List Pages connected to the authorised user |

### 1.6 App Review — instagram_manage_messages Advanced Access

> **Reference**: https://developers.facebook.com/docs/instagram-platform/app-review/

Submit via **developers.facebook.com → App Review → Permissions and Features**. Required submission elements:

1. Privacy Policy URL (live and publicly accessible)
2. App icon
3. App category
4. Data Handling questionnaire (describe pre-processing of message data)
5. Written use case description (explain the CS hub use case)
6. Screen recording / video demo of the full DM flow end-to-end
7. Test credentials (your platform dashboard login — NOT Instagram credentials)
8. Step-by-step reviewer instructions

Submit `instagram_manage_messages`, `pages_manage_metadata`, and `pages_show_list` together in one submission.

**Review timeline:** Officially 10 business days; real-world expectation is 2–6 weeks.

### 1.7 Business Verification Requirement

> **Reference**: https://www.facebook.com/business/help/1095661473946872

Business Verification is a prerequisite for Advanced Access App Review. Without it, App Review submissions for advanced permissions will be declined.

- **Process:** Meta Business Manager → **Security Center** → **Start Verification**
- **Accepted documents:** Articles of incorporation, business license, tax registration, business bank statement, utility bill
- **Timeline:** 1–5 business days after document submission

### 1.8 Rate Limit Tiers

> **Reference**: https://creatorflow.so/blog/instagram-api-rate-limits-explained/

- **Standard rate limit:** 200 automated DMs per hour per Instagram Business account (rolling 60-minute window)
- **Implied daily cap:** ~4,800 DMs/day
- No published high-volume tier for Instagram DMs (unlike Facebook Messenger Platform which has tiered escalation)
- Pre-2024 limit was ~5,000 API calls/hour; reduced by ~96% in 2024

---

## 2. Authentication / Authorization

> **Reference**: https://developers.facebook.com/docs/instagram-platform/instagram-api-with-instagram-login/get-access-tokens-and-permissions

### 2.1 Instagram-Scoped User ID (IGSID)

The **IGSID** is the unique identifier assigned to an Instagram user in the context of messaging with a specific Instagram Business account.

- Equivalent to PSID (Page-Scoped User ID) in Facebook Messenger, but scoped to the Instagram Business Account rather than the Facebook Page
- Appears in webhook payloads as `entry[].messaging[].sender.id`
- **Stable** for a given user–business pair; different per Instagram Business account (the same user has different IGSIDs at different businesses)
- CS hub implication: store IGSIDs mapped to the Instagram Business Account ID, not globally

### 2.2 Short-lived User Access Token

- Obtained via Graph API Explorer (manual/test) or an OAuth 2.0 web login flow
- Validity: approximately 1–2 hours
- Cannot exchange an already-expired token

### 2.3 Long-lived User Access Token

Exchange a valid short-lived token:

```http
GET https://graph.facebook.com/v23.0/oauth/access_token
  ?grant_type=fb_exchange_token
  &client_id={app-id}
  &client_secret={app-secret}
  &fb_exchange_token={short-lived-user-token}
```

Response:

```json
{
  "access_token": "EAAGm...",
  "token_type": "bearer",
  "expires_in": 5183944
}
```

Validity: 60 days. Exchange calls use the `app-secret` — perform these server-side only, never from a client.

### 2.4 Long-lived Page Access Token

Use the long-lived User Access Token to retrieve Page tokens:

```http
GET https://graph.facebook.com/v23.0/{user-id}/accounts
  ?access_token={long-lived-user-access-token}
```

Response includes Page objects, each with an `access_token` field. The Page Access Token returned from a long-lived User Token does **not expire** under normal conditions (provided the user does not revoke, the app is not removed from the Page, and the Page role is retained).

### 2.5 Never-expiring System User Token (Recommended for CS Hub)

> **Reference**: https://developers.facebook.com/docs/business-management-apis/system-users/install-apps-and-generate-tokens/

Steps in Meta Business Manager:

1. **Business Settings → Users → System Users → Add** (set a name and role: Admin)
2. Under the System User → **Add Assets** → **Apps** tab → select your app, grant full control
3. Click **Generate New Token** → select your app → check permissions: `instagram_manage_messages`, `instagram_basic`, `pages_manage_metadata`
4. Set token expiry to **Never**
5. Copy and store the token immediately (it is shown only once)

**Programmatic generation:**

```http
POST https://graph.facebook.com/v23.0/{SYSTEM-USER-ID}/access_tokens
```

Parameters: `business_app`, `scope`, `access_token` (admin token), `appsecret_proof`

`appsecret_proof` = HMAC-SHA256(key=app_secret, message=access_token)

Pass `set_token_expires_in_60_days=true` to force token rotation if desired. By default the token is non-expiring.

### 2.6 access_token Param vs. Authorization: Bearer Header

Both methods are technically supported:

| Method | Recommendation | Notes |
|---|---|---|
| `Authorization: Bearer <token>` | Recommended | Required by OAuth 2.0 RFC 6750; shown as primary in all Meta documentation |
| `?access_token=<token>` | Avoid | Legacy; still functional but not formally deprecated. Token appears in server logs, Referer headers, and browser history |

All new CS hub integrations should use the `Authorization: Bearer` header exclusively.

### 2.7 Business Verification Gate

Without completed Business Verification, App Review for Advanced Access permissions will be declined. `instagram_basic` at Standard Access does not require Business Verification. All other CS hub permissions require it as a precondition.

---

## 3. Send API — Endpoint Overview

> **Reference**: https://developers.facebook.com/docs/messenger-platform/reference/send-api/

**Endpoint:**

```
POST https://graph.facebook.com/v23.0/me/messages
```

**Current API version:** v23.0 (released May 29, 2025; valid until June 9, 2026; minimum accepted version v22.0 as of September 9, 2025)

**Important restriction:** Businesses cannot initiate a conversation on Instagram. The user must message first before the business can reply.

### Top-level Request Body Fields

| Field | Type | Required | Notes |
|---|---|---|---|
| `recipient` | Object | Yes | `{ "id": "<IGSID>" }`. Instagram only supports IGSID; no phone-number or user_ref alternatives |
| `message` | Object | Yes* | *Required unless `sender_action` is set. Contains `text` or `attachment` |
| `messaging_type` | String | Yes | `RESPONSE`, `UPDATE`, or `MESSAGE_TAG` |
| `sender_action` | String | No | `typing_on`, `typing_off`, `mark_seen`. Mutually exclusive with `message` |
| `tag` | String | Conditional | Required when `messaging_type` is `MESSAGE_TAG` |
| `notification_type` | String | No | `REGULAR` (default), `SILENT_PUSH`, `NO_PUSH` |

### messaging_type Values

| Field | Type | Required | Notes |
|---|---|---|---|
| `RESPONSE` | String | — | Replying to a user-initiated message; within 24 hours of user's last message |
| `UPDATE` | String | — | Proactive non-promotional update; within 24 hours of user's last message |
| `MESSAGE_TAG` | String | — | Out-of-window send with an approved tag; window depends on tag |

### MESSAGE_TAG Values on Instagram

Only two tags are valid on Instagram:

| Field | Type | Required | Notes |
|---|---|---|---|
| `HUMAN_AGENT` | String | — | 7-day window after user's last message. Must be sent by a live human — NOT automated bots |
| `NOTIFICATION_MESSAGE` | String | — | Outside all standard windows; requires notification opt-in token as `recipient.id` |

**Tags NOT supported on Instagram** (Facebook Messenger only): `CONFIRMED_EVENT_UPDATE`, `POST_PURCHASE_UPDATE`, `ACCOUNT_UPDATE`

### Success Response Envelope

```json
{
  "recipient_id": "1234567890123456", // IGSID of the recipient
  "message_id": "m_AbCdEfGhIjKlMnOpQrStUv"
}
```

---

## 4. Text Messages

> **Reference**: https://developers.facebook.com/docs/messenger-platform/instagram/features/send-messages

- **Character limit:** 1,000 characters (vs 2,000 for Facebook Messenger)
- URLs render as tappable links; plain Unicode only — no markdown, no HTML
- Emojis are supported and count against the 1,000 character limit

**Within the 24-hour window (RESPONSE):**

```json
{
  "recipient": { "id": "<IGSID>" },    // Instagram-Scoped User ID from webhook sender.id
  "messaging_type": "RESPONSE",
  "message": {
    "text": "Hello! How can we help you today?"
  }
}
```

**Out-of-window with HUMAN_AGENT tag (human-sent only):**

```json
{
  "recipient": { "id": "<IGSID>" },
  "messaging_type": "MESSAGE_TAG",
  "tag": "HUMAN_AGENT",
  "message": {
    "text": "Following up on your earlier question — here is the resolution."
  }
}
```

---

## 5. Attachments

> **Reference**: https://developers.facebook.com/docs/messenger-platform/instagram/features/send-messages

**URL requirements for all attachment types:**
- HTTPS only (HTTP is rejected)
- Publicly accessible (no authentication, no IP allowlisting)
- Server must return HTTP 200 with the correct `Content-Type` header
- File extension should be present and correct in the URL path

**Reusable attachments:** Set `is_reusable: true` in the payload to receive an `attachment_id` in the response. Use the `attachment_id` in future messages instead of re-uploading the file.

Success response with `is_reusable: true`:

```json
{
  "recipient_id": "1234567890123456",
  "message_id": "m_AbCdEfGhIjKlMnOpQrStUv",
  "attachment_id": "987654321098765"
}
```

Reuse in subsequent messages:

```json
{
  "recipient": { "id": "<IGSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "image",
      "payload": { "attachment_id": "987654321098765" }
    }
  }
}
```

The Attachment Upload API (`POST /me/message_attachments`) is available; behavior is consistent with the Messenger Platform documentation (verify against live Meta docs before shipping).

### 5.1 Image

| Field | Type | Required | Notes |
|---|---|---|---|
| `attachment.type` | String | Yes | `"image"` |
| `payload.url` | String | Yes (if no attachment_id) | HTTPS URL to the image |
| `payload.is_reusable` | Boolean | No | `true` returns an `attachment_id` for reuse |

- **Supported formats:** JPEG, PNG, GIF (static and animated)
- **Max file size:** 8 MB

```json
{
  "recipient": { "id": "<IGSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "image",
      "payload": {
        "url": "https://example.com/image.jpg",
        "is_reusable": true
      }
    }
  }
}
```

### 5.2 Video

| Field | Type | Required | Notes |
|---|---|---|---|
| `attachment.type` | String | Yes | `"video"` |
| `payload.url` | String | Yes (if no attachment_id) | HTTPS URL to the video |
| `payload.is_reusable` | Boolean | No | `true` returns an `attachment_id` for reuse |

- **Supported formats:** MP4 (recommended), MOV, 3GP
- **Max file size:** 16 MB (enforce a lower limit in your implementation for Instagram safety)
- **Additional requirements:** No edit lists in the container, moov atom at the front, AAC audio codec (max 48 kHz)

```json
{
  "recipient": { "id": "<IGSID>" },
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

### 5.3 Audio

| Field | Type | Required | Notes |
|---|---|---|---|
| `attachment.type` | String | Yes | `"audio"` |
| `payload.url` | String | Yes (if no attachment_id) | HTTPS URL to the audio file |
| `payload.is_reusable` | Boolean | No | `true` returns an `attachment_id` for reuse |

- **Supported formats:** MP3, OGG, WAV, AAC, AMR, OPUS
- **Max file size:** 16 MB

```json
{
  "recipient": { "id": "<IGSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "attachment": {
      "type": "audio",
      "payload": {
        "url": "https://example.com/audio.mp3",
        "is_reusable": true
      }
    }
  }
}
```

### 5.4 File Attachments — NOT SUPPORTED

> **Reference**: https://developers.cm.com/messaging/docs/instagram-messaging

The `type: "file"` attachment type is **not available for Instagram**. Attempting to send `"type": "file"` will be rejected by the API.

**Workaround:** Send a text message containing a public HTTPS link to the hosted document.

```json
{
  "recipient": { "id": "<IGSID>" },
  "messaging_type": "RESPONSE",
  "message": {
    "text": "Here is your document: https://example.com/files/invoice-2026.pdf"
  }
}
```

---

## 6. Templates

> **Reference**: https://developers.facebook.com/docs/messenger-platform/instagram/features/generic-template

All templates share the same outer envelope — `message.attachment.type` is always `"template"`, and `payload.template_type` selects the variant.

### 6.1 Generic Template

A structured card or horizontally scrollable carousel.

- **Max elements:** 10 (min 2 for carousel)
- **Max buttons per element:** 3
- **Supported button types:** `web_url` and `postback` only (`phone_number`, `element_share`, `account_link` are NOT supported on Instagram)
- Desktop is not supported; mobile/app only

> **Reference**: https://developers.facebook.com/docs/messenger-platform/instagram/features/generic-template

**Element fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `title` | String | Yes | Card headline |
| `image_url` | String | No* | *Required if no `subtitle` and no `buttons` |
| `subtitle` | String | No* | *Required if no `image_url` and no `buttons` |
| `default_action` | Object | No | Card body tap action; `type` must be `web_url` |
| `buttons` | Array | No* | Up to 3; *required if no `image_url` and no `subtitle` |

**Button fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | String | Yes | `web_url` or `postback` only on Instagram |
| `title` | String | Yes | Button label |
| `url` | String | Conditional | Required for `web_url` type |
| `payload` | String | Conditional | Required for `postback` type; returned in webhook |

**Single-card example:**

```json
{
  "recipient": { "id": "<IGSID>" },
  "message": {
    "attachment": {
      "type": "template",
      "payload": {
        "template_type": "generic",
        "elements": [
          {
            "title": "Product Name",
            "image_url": "https://example.com/image.jpg",
            "subtitle": "Short description",
            "default_action": {
              "type": "web_url",
              "url": "https://example.com/product",
              "webview_height_ratio": "tall"
            },
            "buttons": [
              {
                "type": "web_url",
                "url": "https://example.com/buy",
                "title": "Shop Now"
              },
              {
                "type": "postback",
                "title": "Learn More",
                "payload": "LEARN_MORE"
              }
            ]
          }
        ]
      }
    }
  }
}
```

### 6.2 Product Template (Catalog-Linked)

> **Reference**: https://developers.facebook.com/docs/messenger-platform/send-messages/template/product/

Displays products directly from a linked catalog. Product image, title, and price are pulled automatically from the catalog — no manual fields needed.

**Prerequisites:**
- Commerce Manager account
- Product catalog linked to the Instagram account in Commerce Manager
- Products must be active and commerce-eligible

| Field | Type | Required | Notes |
|---|---|---|---|
| `template_type` | String | Yes | `"product"` |
| `elements` | Array | Yes | Each element identifies one catalog product |
| `elements[].id` | String | Yes | Retailer ID from the catalog (Commerce Manager → Catalog → Items → Retailer ID column). Use `elements[].id`, NOT `product_elements[].id` |

**Max elements:** Up to 10 (unverified — no Instagram-specific figure published; consistent with generic template limit).

```json
{
  "recipient": { "id": "<IGSID>" },
  "message": {
    "attachment": {
      "type": "template",
      "payload": {
        "template_type": "product",
        "elements": [
          { "id": "12345678901234" },  // Retailer ID from catalog
          { "id": "98765432109876" }
        ]
      }
    }
  }
}
```

---

## 7. Quick Replies

> **Reference**: https://developers.facebook.com/docs/messenger-platform/instagram/features/quick-replies/

Quick replies render as tap-to-select chips above the message composer. They disappear once tapped and cannot be re-selected. Available to Instagram Business profiles only.

When tapped, quick replies fire `messages` webhook events (NOT `messaging_postbacks`).

| Field | Type | Required | Notes |
|---|---|---|---|
| `content_type` | String | Yes | `text`, `user_phone_number`, or `user_email` |
| `title` | String | Conditional | Required for `content_type: text`. Max 20 chars |
| `payload` | String | Conditional | Required for `content_type: text`. Max 256 chars |
| `image_url` | String | No | NOT supported on Instagram (silently ignored); Facebook Messenger only |

**Full example (all three content_types):**

```json
{
  "recipient": { "id": "<IGSID>" },
  "message": {
    "text": "How would you like to proceed?",
    "quick_replies": [
      {
        "content_type": "text",
        "title": "Chat with agent",
        "payload": "QR_CHAT_AGENT"
      },
      {
        "content_type": "user_phone_number"  // Unconfirmed on Instagram — may be Messenger only
      },
      {
        "content_type": "user_email"         // Unconfirmed on Instagram — may be Messenger only
      }
    ]
  }
}
```

**Constraints:**
- Max 13 quick replies per message
- Max `title` length: 20 chars (truncated if exceeded)
- `image_url` is silently ignored on Instagram; do not include it
- `user_phone_number` and `user_email`: defined in the spec but **not confirmed on Instagram** — treat as Facebook Messenger-only until Meta explicitly documents Instagram support
- Quick replies are only visible on the most recent message in the thread
- Cannot be sent in the same message as an attachment (text body only)
- Visible in the Instagram app only, not Instagram web

**Webhook payload when tapped (content_type: text):**

```json
{
  "sender": { "id": "<IGSID>" },
  "message": {
    "mid": "mid.xxx",
    "text": "Chat with agent",
    "quick_reply": { "payload": "QR_CHAT_AGENT" }
  }
}
```

---

## 8. Sender Actions

> **Reference**: https://developers.facebook.com/docs/messenger-platform/send-messages/sender-actions

Sender actions control the chat UI state — showing a typing indicator or marking messages as read.

`sender_action` is a **top-level field**, NOT inside the `message` object. The request must contain ONLY `recipient` and `sender_action` (no `message` field).

| Field | Type | Required | Notes |
|---|---|---|---|
| `recipient` | Object | Yes | `{ "id": "<IGSID>" }` |
| `sender_action` | String | Yes | `typing_on`, `typing_off`, or `mark_seen`. Top-level — NOT inside `message` |

**typing_on:**

```json
{
  "recipient": { "id": "<IGSID>" },
  "sender_action": "typing_on"
}
```

**typing_off:**

```json
{
  "recipient": { "id": "<IGSID>" },
  "sender_action": "typing_off"
}
```

**mark_seen:**

```json
{
  "recipient": { "id": "<IGSID>" },
  "sender_action": "mark_seen"
}
```

**Response:** `{ "recipient_id": "<IGSID>" }` (no `message_id` is returned)

**Constraints:**
- `typing_on` auto-expires after **20 seconds** if no follow-up message or `typing_off` is sent
- No `message_id` is returned in the response
- Do not batch `typing_on` and `typing_off` in the same API batch call
- `mark_seen` can be sent proactively

---

## 9. Profile Setup — Ice Breakers

> **Reference**: https://developers.facebook.com/docs/instagram-platform/instagram-api-with-instagram-login/messaging-api/ice-breakers/

Ice breakers are conversation starters that appear **only the first time** a user opens the DM conversation window. Confirmed supported on Instagram.

**Endpoint:** `POST /{ig-user-id}/messenger_profile`

| Field | Type | Required | Notes |
|---|---|---|---|
| `ice_breakers` | Array | Yes | Max 4 items |
| `ice_breakers[].question` | String | Yes | Shown to the user. Max 80 chars |
| `ice_breakers[].payload` | String | Yes | Returned in `messaging_postbacks` webhook when tapped |

**Set ice breakers:**

```json
{
  "ice_breakers": [
    { "question": "What are your opening hours?", "payload": "OPENING_HOURS" },
    { "question": "Where are you located?",       "payload": "LOCATION" },
    { "question": "Do you offer returns?",        "payload": "RETURNS" },
    { "question": "How can I track my order?",    "payload": "TRACK_ORDER" }
  ]
}
```

**Delete ice breakers:**

```
DELETE /{ig-user-id}/messenger_profile
```

Body:

```json
{ "fields": ["ice_breakers"] }
```

---

## 10. Profile Setup — Persistent Menu

> **Reference**: https://developers.facebook.com/docs/instagram-platform/instagram-api-with-instagram-login/messaging-api/persistent-menu/

A persistent hamburger-style menu accessible from the conversation at any time. Confirmed supported on Instagram.

**Endpoint:** `POST /{ig-user-id}/messenger_profile`

- Max 3 top-level items
- Max 5 sub-items per nested menu
- Max 30 chars per title

| Field | Type | Required | Notes |
|---|---|---|---|
| `persistent_menu` | Array | Yes | Locale-specific menu configurations; must include `"default"` |
| `locale` | String | Yes | `"default"` required; additional locale codes optional |
| `composer_input_disabled` | Boolean | Yes | `false` = user can type; `true` = disables free text input |
| `call_to_actions` | Array | Yes | Top-level items. Max 3 on Instagram |
| `call_to_actions[].type` | String | Yes | `postback`, `web_url`, or `nested` |
| `call_to_actions[].title` | String | Yes | Max 30 chars |
| `call_to_actions[].payload` | String | Conditional | Required for `postback` type |
| `call_to_actions[].url` | String | Conditional | Required for `web_url` type |
| `call_to_actions[].call_to_actions` | Array | Conditional | Required for `nested` type; max 5 sub-items |

**Set persistent menu:**

```json
{
  "persistent_menu": [
    {
      "locale": "default",
      "composer_input_disabled": false,
      "call_to_actions": [
        {
          "type": "postback",
          "title": "Get Started",
          "payload": "GET_STARTED"
        },
        {
          "type": "nested",
          "title": "My Account",
          "call_to_actions": [
            { "type": "postback", "title": "Order History",   "payload": "ORDER_HISTORY" },
            { "type": "postback", "title": "Contact Support", "payload": "CONTACT_SUPPORT" }
          ]
        },
        {
          "type": "web_url",
          "title": "Visit Website",
          "url": "https://www.example.com"
        }
      ]
    }
  ]
}
```

**Delete persistent menu:**

```
DELETE /{ig-user-id}/messenger_profile
```

Body:

```json
{ "fields": ["persistent_menu"] }
```

---

## 11. Handover Protocol

> **Reference**: https://developers.facebook.com/docs/messenger-platform/instagram/features/handover-protocol/

> **IMPORTANT: Handover Protocol is deprecated for Instagram.** Meta has replaced it with **Conversation Routing**. Migration from Handover Protocol to Conversation Routing is one-way and **irreversible**. For apps that have NOT yet migrated, `pass_thread_control` and `take_thread_control` remain functional (no published hard sunset date found at time of research).

### 11.1 Pass Thread Control

> **Reference**: https://developers.facebook.com/docs/messenger-platform/instagram/features/handover-protocol/

**Endpoint:** `POST https://graph.facebook.com/v23.0/me/pass_thread_control`

| Field | Type | Required | Notes |
|---|---|---|---|
| `recipient.id` | String | Yes | IGSID of the user |
| `target_app_id` | String | Yes | App ID of the receiving app |
| `metadata` | String | No | Arbitrary string (max 1000 chars) forwarded to the receiving app via `messaging_handovers` webhook |

**Well-known `target_app_id` values:**
- Instagram-native Inbox: `1217981644879628`
- Facebook Page Inbox: `263902037430900`

```json
{
  "recipient": { "id": "<IGSID>" },
  "target_app_id": "1217981644879628",  // Instagram-native Inbox
  "metadata": "Escalating to human agent"
}
```

Response: `{ "success": true }`

### 11.2 Take Thread Control

**Endpoint:** `POST https://graph.facebook.com/v23.0/me/take_thread_control`

Only the **Primary Receiver** can call this endpoint.

| Field | Type | Required | Notes |
|---|---|---|---|
| `recipient.id` | String | Yes | IGSID of the user |
| `metadata` | String | No | Context string passed to the outgoing app |

```json
{
  "recipient": { "id": "<IGSID>" },
  "metadata": "Taking control to handle escalated case"
}
```

Response: `{ "success": true }`

### 11.3 Conversation Routing (Current Replacement)

> **Reference**: https://developers.facebook.com/docs/messenger-platform/instagram/features/conversation-routing/

Conversation Routing is the current Meta-recommended replacement for the Handover Protocol on Instagram:

- Routes conversations automatically based on entry point: organic DMs go to the Default App; ad/sponsored message entries go to the Marketing App
- Configuration is done in Meta Business Manager (static, per-app assignment) — it is not API-driven at runtime
- Migration from Handover Protocol to Conversation Routing is **irreversible**

---

## 12. Private Replies

> **Reference**: https://developers.facebook.com/docs/instagram-platform/private-replies/

Private replies allow a business to initiate a DM conversation from a public interaction (a comment on a post, or a story mention). These are distinct from standard DMs in that the initiating trigger is a public event rather than a direct inbound message.

### 12.1 Reply to a Comment

Sends a private DM in response to a public comment on an Instagram post.

- **Recipient field:** `recipient.comment_id` — NOT `recipient.id` (IGSID)
- **Source of comment_id:** Webhook `comments` subscription → event `id` field; or `GET /{ig-media-id}/comments`
- **Required permissions:** `instagram_manage_messages` + `pages_manage_metadata`
- **Reply window:** 7 days from comment creation (Instagram Live comments: active broadcast only)
- **Supported message type:** Text only for the initial private reply

| Field | Type | Required | Notes |
|---|---|---|---|
| `recipient.comment_id` | String | Yes | From webhook comment event `id` or Pages API. Do NOT use an IGSID here |
| `message.text` | String | Yes | Plain text, max 1000 chars |
| `messaging_type` | String | Yes | Must be `RESPONSE` |

```json
{
  "recipient": { "comment_id": "17858893269000001" },  // comment ID, NOT an IGSID
  "message": {
    "text": "Thanks for your comment! We've sent you more details via DM."
  },
  "messaging_type": "RESPONSE"
}
```

### 12.2 Reply to a Story Mention

Sends a reply when a user mentions the business's Instagram account in their story.

- **Trigger:** Inbound webhook message event with `message.attachments[].type = "story_mention"`
- **Recipient:** IGSID from `sender.id` of the story mention event (uses `recipient.id`, NOT `comment_id`)
- **Reply window:** 24 hours from the story mention (standard messaging window)
- **Supported message types:** Text (safest); attachments available within the open messaging window

| Field | Type | Required | Notes |
|---|---|---|---|
| `recipient.id` | String | Yes | IGSID from `sender.id` of the inbound story mention webhook event |
| `message.text` | String | Yes | Plain text, max 1000 chars |
| `messaging_type` | String | Yes | `RESPONSE` within 24h; use `MESSAGE_TAG` + `HUMAN_AGENT` for 7-day window (human-sent only) |

```json
{
  "recipient": { "id": "<IGSID_FROM_STORY_MENTION_EVENT>" },  // IGSID, not comment_id
  "message": {
    "text": "Thanks for mentioning us! We'd love to chat."
  },
  "messaging_type": "RESPONSE"
}
```

---

## 13. Common Errors

> **Reference**: https://developers.facebook.com/docs/messenger-platform/reference/send-api/

| Code | HTTP Status | Message | Likely Cause | Fix |
|---|---|---|---|---|
| 10 | 403 | "Application does not have permission for this action" | Missing `instagram_manage_messages` permission, Standard Access only, or app in Development mode | Obtain Advanced Access; switch app to Live mode |
| 100 | 400 | "Invalid parameter" / "The parameter recipient is required" | Malformed payload, wrong IGSID format, invalid `comment_id`, or cold outreach attempt (user never contacted you) | Validate IGSID/`comment_id`; ensure user initiated the conversation first |
| 190 | 401 | "Invalid OAuth access token" | Token expired, revoked, or scopes changed | Regenerate token; for System User tokens, verify Business Manager permissions |
| 200 | 403 | "Requires instagram_basic permission" | Permission declared but not at Advanced Access level, or no Business Verification completed | Complete Business Verification; resubmit App Review for Advanced Access |
| 551 | 400 | "This person isn't available right now" | User blocked the account, opted out of messages, or has restrictive privacy settings | Do not retry; only contact users who initiated the conversation |
| 613 | 400 | "Calls to this API have exceeded the rate limit" | Exceeded 200 automated DMs/hour | Implement a queue with exponential backoff; rate limit resets after 60 minutes |
| 10 / subcode 2534022 | 403 | "This message is sent outside of allowed window" | Automated message sent more than 24h after the user's last message without a valid `MESSAGE_TAG` | Reply within the 24h window; use `HUMAN_AGENT` tag for human-sent follow-ups within 7 days |
| Tag validation | 400 | "OAuthException — message tag invalid" | Using a Facebook-only tag (`CONFIRMED_EVENT_UPDATE`, etc.) on Instagram, or using `HUMAN_AGENT` for automated messages | Only `HUMAN_AGENT` and `NOTIFICATION_MESSAGE` are valid on Instagram; `HUMAN_AGENT` must be sent by a human |

---

## 14. Minimal cURL Example

> **Reference**: https://developers.facebook.com/docs/messenger-platform/reference/send-api/

```bash
curl -X POST "https://graph.facebook.com/v23.0/me/messages" \
  -H "Authorization: Bearer <SYSTEM_USER_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": { "id": "<RECIPIENT_IGSID>" },
    "message": { "text": "Hello! Thank you for reaching out. How can we help you today?" },
    "messaging_type": "RESPONSE"
  }'
```

---

## 15. Production Requirements

> **Reference**: https://developers.facebook.com/docs/instagram-platform/app-review/

- [ ] Instagram account is Business or Creator type (Personal accounts are NOT supported)
- [ ] Instagram account linked to a Facebook Page
- [ ] Facebook Page connected to Meta Business Manager
- [ ] Business Verification complete in Meta Business Manager Security Center
- [ ] Meta App type is Business, Messenger Platform product added
- [ ] App switched to Live mode (not Development mode)
- [ ] App Review approved for `instagram_manage_messages` (Advanced Access)
- [ ] App Review approved for `pages_manage_metadata` (Advanced Access)
- [ ] App Review approved for `pages_show_list` (Advanced Access)
- [ ] Privacy Policy URL live and linked in app settings
- [ ] System User created (Admin role), token generated with all required permissions, stored securely
- [ ] Webhook configured with verified callback URL and verify token
- [ ] Webhook subscriptions active: `messages`, `messaging_postbacks`, `comments` (for private replies)
- [ ] Rate limit handling implemented (queue + backoff for error 613)
- [ ] 24-hour window logic implemented (no automated sends outside window without a valid tag)
- [ ] `HUMAN_AGENT` tag restricted to genuine human-sent messages only
- [ ] End-to-end flow tested in Development mode before going live
