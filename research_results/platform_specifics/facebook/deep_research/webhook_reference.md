# Facebook Messenger Platform — Webhook API Reference

**Sources:** Meta for Developers official documentation (developers.facebook.com)
**Last verified:** March 2026
**API version note:** Always use the latest Graph API version to receive the most complete webhook data.

---

## Table of Contents

1. [Webhook Setup — Manual Steps](#1-webhook-setup--manual-steps)
   - 1.1 [Create a Meta Developer Account](#11-create-a-meta-developer-account)
   - 1.2 [Create a Meta App](#12-create-a-meta-app)
   - 1.3 [Connect a Facebook Page](#13-connect-a-facebook-page)
   - 1.4 [Configure the Webhook (Callback URL + Verify Token)](#14-configure-the-webhook-callback-url--verify-token)
   - 1.5 [Select Subscribed Fields](#15-select-subscribed-fields)
   - 1.6 [Subscribe the Page via Graph API](#16-subscribe-the-page-via-graph-api)
   - 1.7 [Required Permissions](#17-required-permissions)
   - 1.8 [App Review Requirements for Production](#18-app-review-requirements-for-production)
2. [Webhook Authentication — X-Hub-Signature-256](#2-webhook-authentication--x-hub-signature-256)
   - 2.1 [Overview](#21-overview)
   - 2.2 [Verification Algorithm](#22-verification-algorithm)
   - 2.3 [Python Example](#23-python-example)
   - 2.4 [Go Example](#24-go-example)
3. [Outer Envelope Structure](#3-outer-envelope-structure)
4. [Webhook Event Payloads](#4-webhook-event-payloads)
   - 4.1 [messages — text](#41-messages--text)
   - 4.2 [messages — image](#42-messages--image)
   - 4.3 [messages — video](#43-messages--video)
   - 4.4 [messages — audio](#44-messages--audio)
   - 4.5 [messages — file](#45-messages--file)
   - 4.6 [messages — location](#46-messages--location)
   - 4.7 [messages — sticker](#47-messages--sticker)
   - 4.8 [messages — reply (mid reference)](#48-messages--reply-mid-reference)
   - 4.9 [messages — quick reply](#49-messages--quick-reply)
   - 4.10 [message_reactions — reaction / unreaction](#410-message_reactions--reaction--unreaction)
   - 4.11 [messages — unsend (is_unsent)](#411-messages--unsend-is_unsent)
   - 4.12 [message_deliveries](#412-message_deliveries)
   - 4.13 [message_reads](#413-message_reads)
   - 4.14 [messaging_postbacks](#414-messaging_postbacks)
   - 4.15 [messaging_optins — Recurring Notifications](#415-messaging_optins--recurring-notifications)
   - 4.16 [messaging_referrals](#416-messaging_referrals)
   - 4.17 [messaging_handovers — pass_thread_control](#417-messaging_handovers--pass_thread_control)
   - 4.18 [messaging_handovers — take_thread_control](#418-messaging_handovers--take_thread_control)
   - 4.19 [messaging_handovers — request_thread_control](#419-messaging_handovers--request_thread_control)
   - 4.20 [messaging_handovers — app_roles](#420-messaging_handovers--app_roles)
   - 4.21 [messaging_policy_enforcement](#421-messaging_policy_enforcement)
   - 4.22 [standby channel](#422-standby-channel)
5. [All Subscribable Webhook Fields (Reference Table)](#5-all-subscribable-webhook-fields-reference-table)

---

## 1. Webhook Setup — Manual Steps

### 1.1 Create a Meta Developer Account

1. Navigate to **https://developers.facebook.com/** and click **Get Started**.
2. Log in with a personal Facebook account (used as the developer identity).
3. Accept the Meta Platform Terms and Developer Policies.
4. Your account is now registered as a Meta developer. You gain access to the App Dashboard.

**Ref:** https://developers.facebook.com/docs/development/register

---

### 1.2 Create a Meta App

Meta now uses **use cases** rather than plain app types to determine what APIs are available.

**Steps:**

1. Navigate to **https://developers.facebook.com/apps/creation/**.
2. Enter your **app name** and a **contact email address**. Click **Next**.
3. Select a use case. For Messenger bots, select:
   - **"Engage with customers on Messenger from Meta"** — this grants access to the Messenger Platform APIs, `pages_messaging` permission, and the Webhooks product.
4. Optionally connect a **business portfolio** (required for production apps accessing data you don't own). Click **Next**.
5. Review the listed requirements (App Review may be needed). Click **Next**.
6. Review the summary and click **Go to dashboard**.

You are redirected to your App Dashboard. Note your **App ID** and **App Secret** (found under **Settings > Basic**).

**Notes:**
- You may have administrator or developer roles on a maximum of 15 apps not connected to a verified business portfolio.
- Use cases cannot be removed after app creation.
- The `pages_messaging` permission and Messenger product are automatically added when the Messenger use case is selected.

**Ref:** https://developers.facebook.com/docs/development/create-an-app/

---

### 1.3 Connect a Facebook Page

The Messenger Platform operates through Facebook Pages — your bot speaks as the Page.

**Steps:**

1. In the App Dashboard, go to **Products > Messenger > Settings**.
2. In the **Access Tokens** section, click **Add or Remove Pages** and select the Facebook Page to connect. Grant the requested permissions.
3. In the **Token Generation** section, select the Page from the dropdown. An **Page Access Token** will appear.
4. Copy the Page Access Token and store it securely (e.g., as an environment variable `PAGE_ACCESS_TOKEN`). This token is not persisted in the UI — a new token is generated each time you select the page, but previously generated tokens remain valid.

**Important:** Until the app has passed App Review, the Page Access Token only allows interaction with Facebook accounts that have **Administrator**, **Developer**, or **Tester** roles on the app.

---

### 1.4 Configure the Webhook (Callback URL + Verify Token)

Your webhook server must be reachable over **HTTPS** with a **valid TLS/SSL certificate** (self-signed certificates are not supported).

**Steps:**

1. In the App Dashboard, go to **Products > Messenger > Settings**.
2. In the **Webhooks** section, click **Add Callback URL**.
3. Enter your endpoint URL in the **Callback URL** field (e.g., `https://your-domain.com/webhook`).
4. Enter a string of your choosing in the **Verify Token** field. This is a shared secret you define — Meta will send it back to confirm ownership. Store it as `VERIFY_TOKEN`.
5. Click **Verify and Save**.

Meta will immediately send a **GET verification request** to your endpoint. Your server must respond correctly (see Section 2 for verification logic).

**Verification Request Format:**

```
GET https://your-domain.com/webhook?hub.mode=subscribe&hub.verify_token=<YOUR_VERIFY_TOKEN>&hub.challenge=1158201444
```

| Query Parameter   | Description                                                              |
|-------------------|--------------------------------------------------------------------------|
| `hub.mode`        | Always `subscribe`                                                       |
| `hub.verify_token`| The verify token you entered in the dashboard                            |
| `hub.challenge`   | An integer your server must echo back in the response body               |

**Required server behavior:**
- If `hub.mode == "subscribe"` and `hub.verify_token` matches your stored token → respond `200 OK` with `hub.challenge` as the response body (plain text integer).
- Otherwise → respond `403 Forbidden`.

---

### 1.5 Select Subscribed Fields

After the webhook endpoint is verified, subscribe to the specific event types you need.

In the **Webhooks** section of Messenger Settings, click **Edit** next to your Page subscription and check the fields you want (e.g., `messages`, `message_deliveries`, `message_reads`, `messaging_postbacks`, etc.).

See [Section 5](#5-all-subscribable-webhook-fields-reference-table) for the full list of available fields.

**Note:** You can also manage subscriptions programmatically (see Section 1.6).

---

### 1.6 Subscribe the Page via Graph API

In addition to the dashboard, you can subscribe a Page to webhook fields programmatically:

```http
POST /v21.0/<PAGE_ID>/subscribed_apps
  ?subscribed_fields=messages,messaging_postbacks,message_deliveries,message_reads
  &access_token=<PAGE_ACCESS_TOKEN>
```

To list current subscriptions:

```http
GET /v21.0/<PAGE_ID>/subscribed_apps?access_token=<PAGE_ACCESS_TOKEN>
```

To remove subscriptions:

```http
DELETE /v21.0/<PAGE_ID>/subscribed_apps?access_token=<PAGE_ACCESS_TOKEN>
```

You can also configure at the app level using the `/app/subscriptions` Graph API endpoint.

---

### 1.7 Required Permissions

| Permission                  | Purpose                                                            |
|-----------------------------|--------------------------------------------------------------------|
| `pages_messaging`           | Send and receive messages on behalf of a Page (required)          |
| `pages_manage_metadata`     | Subscribe and receive webhooks for a Page (required)              |
| `pages_read_engagement`     | Read Page content and subscriber counts (optional)                |
| `pages_show_list`           | List Pages managed by a user (often required for token generation)|
| `catalog_management`        | Receive product details in `messages` webhooks (for Shops)        |

For development (roles on the app), these permissions work in Standard Access mode. For public production use, they must be approved through App Review.

---

### 1.8 App Review Requirements for Production

When your app will be used by people who do **not** have a role on the app (i.e., the general public), you must submit it for App Review.

**Pre-submission checklist:**
1. Ensure the app abides by all [Messenger Platform Policies](https://developers.facebook.com/policy#messengerplatform).
2. Ensure the app follows [Community Standards](https://www.facebook.com/communitystandards).
3. Complete the [pre-launch checklist](https://developers.facebook.com/docs/messenger-platform/product-overview/launch).
4. **Publish** the Facebook Page associated with the app.
5. Ensure your webhook returns `200 OK` to all events within **20 seconds**.
6. If the app has gated content, provide reviewer credentials or a trigger phrase.

**Submitting:**
- Go to **App Dashboard > App Review**.
- Request the specific permissions your app needs (e.g., `pages_messaging`).
- Each permission requires a written description of use, screen recordings, and test credentials.
- Meta's review team will test the app.

**After approval:**
- The app moves from Development Mode to Live Mode.
- Standard Access to `pages_messaging` is granted automatically when the app is live and the Page is connected.
- Advanced permissions (e.g., `pages_read_engagement`) require separate review.

**Ref:**
- https://developers.facebook.com/docs/messenger-platform/app-review/
- https://developers.facebook.com/docs/resp-plat-initiatives/individual-processes/app-review/

---

## 2. Webhook Authentication — X-Hub-Signature-256

### 2.1 Overview

Every POST event notification from Meta includes an **`X-Hub-Signature-256`** header of the form:

```
X-Hub-Signature-256: sha256=<hex_digest>
```

This is an **HMAC-SHA256** signature computed over the raw request body using your app's **App Secret** as the key. Validating this header proves:
- The payload originated from Meta (not a spoofed request).
- The payload was not tampered with in transit.

**Important Unicode escaping note:** Meta generates the signature against an *escaped unicode* version of the payload, with lowercase hex digits. For example, the character `ä` (U+00E4) is represented as `\u00e4` in the signed form. If you compute HMAC against the raw decoded UTF-8 bytes you will get a different result. **You must compute the HMAC against the raw request body bytes exactly as received** (before any JSON parsing) — this works correctly as long as you do not re-encode the body.

**App Secret:** Found in **App Dashboard > Settings > Basic > App Secret**.

---

### 2.2 Verification Algorithm

1. Read the raw request body bytes (before JSON parsing).
2. Read the `X-Hub-Signature-256` header value. Strip the `sha256=` prefix to get the expected hex digest.
3. Compute `HMAC-SHA256(key=APP_SECRET, message=raw_body)`.
4. Compare your computed hex digest to the expected hex digest using a **constant-time comparison** (to prevent timing attacks).
5. If they match → payload is authentic. If they don't → reject with `403`.

---

### 2.3 Python Example

```python
import hashlib
import hmac
from flask import Flask, request, abort

app = Flask(__name__)
APP_SECRET = "your_app_secret_here"  # load from environment variable

@app.route("/webhook", methods=["POST"])
def webhook():
    signature_header = request.headers.get("X-Hub-Signature-256", "")
    if not signature_header.startswith("sha256="):
        abort(403, "Missing signature")

    expected_signature = signature_header[7:]  # strip "sha256="

    # Compute HMAC against the raw body bytes
    computed = hmac.new(
        APP_SECRET.encode("utf-8"),
        msg=request.get_data(),  # raw body, do NOT use request.json
        digestmod=hashlib.sha256,
    ).hexdigest()

    # Constant-time comparison to prevent timing attacks
    if not hmac.compare_digest(computed, expected_signature):
        abort(403, "Invalid signature")

    payload = request.get_json()
    # ... process payload ...
    return "EVENT_RECEIVED", 200
```

---

### 2.4 Go Example

```go
package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "io"
    "log"
    "net/http"
    "os"
    "strings"
)

var appSecret = os.Getenv("APP_SECRET")

func verifySignature(body []byte, signatureHeader string) bool {
    if !strings.HasPrefix(signatureHeader, "sha256=") {
        return false
    }
    expectedHex := signatureHeader[7:] // strip "sha256="

    mac := hmac.New(sha256.New, []byte(appSecret))
    mac.Write(body)
    computed := hex.EncodeToString(mac.Sum(nil))

    // hmac.Equal does constant-time comparison
    expectedBytes, err := hex.DecodeString(expectedHex)
    if err != nil {
        return false
    }
    computedBytes, _ := hex.DecodeString(computed)
    return hmac.Equal(computedBytes, expectedBytes)
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    sigHeader := r.Header.Get("X-Hub-Signature-256")
    if !verifySignature(body, sigHeader) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // body is verified — parse and process
    log.Printf("Verified webhook payload: %s", string(body))
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("EVENT_RECEIVED"))
}

func main() {
    http.HandleFunc("/webhook", webhookHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

---

## 3. Outer Envelope Structure

All Messenger webhook POST payloads share the same outer JSON envelope:

```json
{
  "object": "page",
  "entry": [
    {
      "id": "<PAGE_ID>",
      "time": 1458692752478,
      "messaging": [
        {
          "sender": {
            "id": "<PSID>"
          },
          "recipient": {
            "id": "<PAGE_ID>"
          },
          "timestamp": 1458692752478,
          "...event-specific fields...": {}
        }
      ]
    }
  ]
}
```

| Field              | Type      | Description                                                                                   |
|--------------------|-----------|-----------------------------------------------------------------------------------------------|
| `object`           | String    | Always `"page"` for Messenger webhooks.                                                       |
| `entry`            | Array     | Array of entry objects. Meta may batch multiple events in one request.                        |
| `entry[].id`       | String    | The Facebook Page ID.                                                                         |
| `entry[].time`     | Number    | Unix timestamp (milliseconds) of when the entry was sent.                                     |
| `entry[].messaging`| Array     | Array of messaging event objects. **Will usually contain exactly one item**, but can be multiple if batched. |
| `sender.id`        | String    | The Page-scoped ID (PSID) of the user. Unique per user per Page.                              |
| `recipient.id`     | String    | Your Facebook Page ID.                                                                        |
| `timestamp`        | Number    | Unix timestamp (milliseconds) of the event.                                                   |

**PSID:** Every user has a unique Page-Scoped ID (PSID) for each Facebook Page they message. The same user has a different PSID on each Page.

**Delivery requirements:**
- Your server must respond `200 OK` within **5 seconds** (20 seconds maximum for App Review compliance).
- Respond with `200 OK` even if you cannot process the event — you can queue for async processing.
- If your server fails repeatedly, Meta will retry 2–3 times immediately, then alert you after 15 minutes. After 1 hour of continuous failure, the app is unsubscribed from webhooks.

**For standby channel events**, the outer envelope uses `"standby"` instead of `"messaging"` as the array key (see Section 4.22).

---

## 4. Webhook Event Payloads

### 4.1 `messages` — text

**Subscription field:** `messages`
**Triggered by:** A user sending a plain text message to your Page.

```json
{
  "sender": {
    "id": "<PSID>"
  },
  "recipient": {
    "id": "<PAGE_ID>"
  },
  "timestamp": 1458692752478,
  "message": {
    "mid": "mid.1457764197618:41d102a3e1ae206a38",
    "text": "hello, world!"
  }
}
```

**`message` object fields:**

| Field         | Type    | Description                                                                              |
|---------------|---------|------------------------------------------------------------------------------------------|
| `mid`         | String  | Unique message ID. Use for deduplication.                                                |
| `text`        | String  | The text content of the message.                                                         |
| `quick_reply` | Object  | Present if the user tapped a Quick Reply button. Contains `payload` (developer-defined). |
| `reply_to`    | Object  | Present if the user replied to a specific message. Contains `mid` of the referenced message. |

---

### 4.2 `messages` — image

**Subscription field:** `messages`
**Triggered by:** A user sending an image (including GIF) to your Page.

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1518479195308,
  "message": {
    "mid": "mid.$cAAJdkrCd2ORnva8ErFhjGm0X_Q_c",
    "attachments": [
      {
        "type": "image",
        "payload": {
          "url": "https://example.com/image.jpg"
        }
      }
    ]
  }
}
```

**`attachments[].payload` fields for `image`:**

| Field       | Type   | Description                                                                     |
|-------------|--------|---------------------------------------------------------------------------------|
| `url`       | String | The CDN URL of the image. URLs expire and should be downloaded promptly.        |
| `sticker_id`| Number | Present if the image is a sticker (see Section 4.7). Omitted for regular images.|

---

### 4.3 `messages` — video

**Subscription field:** `messages`
**Triggered by:** A user sending a video to your Page.

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1518479195308,
  "message": {
    "mid": "mid.$cAAJdkrCd2ORnva8ErFhjGm0X_Q_c",
    "attachments": [
      {
        "type": "video",
        "payload": {
          "url": "https://example.com/video.mp4"
        }
      }
    ]
  }
}
```

**`attachments[].payload` fields for `video`:**

| Field | Type   | Description                                                         |
|-------|--------|---------------------------------------------------------------------|
| `url` | String | CDN URL of the video file. Download promptly as URLs expire.        |

**Note:** Facebook Reels shared into a conversation arrive as type `reel` or `ig_reel` with a `payload.reel_video_id` in addition to `url`.

---

### 4.4 `messages` — audio

**Subscription field:** `messages`
**Triggered by:** A user sending an audio file or voice message.

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1518479195308,
  "message": {
    "mid": "mid.$cAAJdkrCd2ORnva8ErFhjGm0X_Q_c",
    "attachments": [
      {
        "type": "audio",
        "payload": {
          "url": "https://example.com/audio.m4a"
        }
      }
    ]
  }
}
```

**`attachments[].payload` fields for `audio`:**

| Field | Type   | Description                                                         |
|-------|--------|---------------------------------------------------------------------|
| `url` | String | CDN URL of the audio file. Download promptly as URLs expire.        |

---

### 4.5 `messages` — file

**Subscription field:** `messages`
**Triggered by:** A user sending a generic file attachment (PDF, document, etc.).

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1518479195308,
  "message": {
    "mid": "mid.$cAAJdkrCd2ORnva8ErFhjGm0X_Q_c",
    "attachments": [
      {
        "type": "file",
        "payload": {
          "url": "https://example.com/document.pdf"
        }
      }
    ]
  }
}
```

**`attachments[].payload` fields for `file`:**

| Field | Type   | Description                                              |
|-------|--------|----------------------------------------------------------|
| `url` | String | CDN URL of the file. Download promptly as URLs expire.   |

---

### 4.6 `messages` — location

**Subscription field:** `messages`
**Triggered by:** A user sharing their location via Messenger.

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1518479195308,
  "message": {
    "mid": "mid.$cAAJdkrCd2ORnva8ErFhjGm0X_Q_c",
    "attachments": [
      {
        "type": "location",
        "title": "Joe's Location",
        "payload": {
          "coordinates": {
            "lat": 37.331684,
            "long": -122.030271
          }
        }
      }
    ]
  }
}
```

**`attachments[]` fields for `location`:**

| Field                         | Type   | Description                                    |
|-------------------------------|--------|------------------------------------------------|
| `type`                        | String | `"location"`                                   |
| `title`                       | String | Label for the location (e.g., user's name).    |
| `payload.coordinates.lat`     | Number | Latitude as a decimal number.                  |
| `payload.coordinates.long`    | Number | Longitude as a decimal number.                 |

---

### 4.7 `messages` — sticker

**Subscription field:** `messages`
**Triggered by:** A user sending a sticker (including the default "Like" thumbs-up sticker).

Stickers arrive as `type: "image"` attachments with an additional `sticker_id` field in the payload.

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1518479195308,
  "message": {
    "mid": "mid.1457764197618:41d102a3e1ae206a38",
    "attachments": [
      {
        "type": "image",
        "payload": {
          "url": "https://example.fbcdn.net/sticker.png",
          "sticker_id": 369239263222822
        }
      }
    ]
  }
}
```

**`attachments[].payload` fields for sticker:**

| Field       | Type   | Description                                                                                  |
|-------------|--------|----------------------------------------------------------------------------------------------|
| `url`       | String | CDN URL of the sticker image.                                                                |
| `sticker_id`| Number | Persistent identifier for the sticker. For example, `369239263222822` is the "Like" sticker. |

**Detecting a sticker:** Check `attachments[0].type == "image"` and `attachments[0].payload.sticker_id` is present.

---

### 4.8 `messages` — reply (mid reference)

**Subscription field:** `messages`
**Triggered by:** A user replying to a specific message in the conversation thread.

The `reply_to` object references the message being replied to.

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1458692752478,
  "message": {
    "mid": "m_1457764197618:41d102a3e1ae206a38",
    "text": "hello, world!",
    "reply_to": {
      "mid": "m_1fTq8oLumEyIp3Q2MR-aY7IfLZDamVrALniheU",
      "is_self_reply": false
    }
  }
}
```

**`message.reply_to` fields:**

| Field          | Type    | Description                                                                 |
|----------------|---------|-----------------------------------------------------------------------------|
| `mid`          | String  | The message ID of the message being replied to.                             |
| `is_self_reply`| Boolean | `true` if the user is replying to their own message; `false` otherwise.     |

---

### 4.9 `messages` — quick reply

**Subscription field:** `messages`
**Triggered by:** A user tapping a Quick Reply button.

Quick Reply buttons are sent by your bot to present a short menu of options. When the user taps one, the same `messages` webhook fires with a `quick_reply` field.

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1458692752478,
  "message": {
    "mid": "mid.1457764197618:41d102a3e1ae206a38",
    "text": "Yes",
    "quick_reply": {
      "payload": "DEVELOPER_DEFINED_PAYLOAD_YES"
    }
  }
}
```

**`message.quick_reply` fields:**

| Field     | Type   | Description                                                                        |
|-----------|--------|------------------------------------------------------------------------------------|
| `payload` | String | The developer-defined payload string attached to the Quick Reply button (max 1000 chars). |

---

### 4.10 `message_reactions` — reaction / unreaction

**Subscription field:** `message_reactions`
**Triggered by:** A user adding or removing a reaction to/from a message your Page sent.

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1458668856463,
  "reaction": {
    "reaction": "love",
    "emoji": "\u2764\uFE0F",
    "action": "react",
    "mid": "<MID_OF_REACTED_TO_MESSAGE>"
  }
}
```

**Unreaction example:**

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1458668856463,
  "reaction": {
    "reaction": "love",
    "emoji": "\u2764\uFE0F",
    "action": "unreact",
    "mid": "<MID_OF_MESSAGE>"
  }
}
```

**`reaction` object fields:**

| Field      | Type   | Description                                                                                    |
|------------|--------|------------------------------------------------------------------------------------------------|
| `reaction` | String | Text name of the reaction. Values: `smile`, `angry`, `sad`, `wow`, `love`, `like`, `dislike`, `other`. |
| `emoji`    | String | The UTF-8 emoji character corresponding to the reaction.                                       |
| `action`   | String | `"react"` (reaction added) or `"unreact"` (reaction removed).                                 |
| `mid`      | String | The message ID of the message that was reacted to.                                             |

---

### 4.11 `messages` — unsend (is_unsent)

**Subscription field:** `messages`
**Triggered by:** A user "unsending" (deleting for everyone) a message they previously sent.

When a message is unsent, the `messages` webhook fires with `message.is_deleted: true`. The original message content is no longer accessible.

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1458692752478,
  "message": {
    "mid": "mid.1457764197618:41d102a3e1ae206a38",
    "is_deleted": true
  }
}
```

**`message` fields for unsend:**

| Field        | Type    | Description                                                         |
|--------------|---------|---------------------------------------------------------------------|
| `mid`        | String  | The message ID of the message that was unsent.                      |
| `is_deleted` | Boolean | `true` when the message has been unsent by the user.                |

**Note:** When `is_deleted: true` is present, no `text` or `attachments` field is included — the content has been deleted.

---

### 4.12 `message_deliveries`

**Subscription field:** `message_deliveries`
**Triggered by:** A message your Page sent has been delivered to the user's device.

```json
{
  "sender": {
    "id": "<PSID>"
  },
  "recipient": {
    "id": "<PAGE_ID>"
  },
  "delivery": {
    "mids": [
      "mid.1458668856218:ed81099e15d3f4f233"
    ],
    "watermark": 1458668856253
  }
}
```

**`delivery` object fields:**

| Field       | Type   | Description                                                                                                   |
|-------------|--------|---------------------------------------------------------------------------------------------------------------|
| `mids`      | Array  | Array of message IDs that were delivered. **May not be present** for older Messenger clients (use `watermark` as fallback). |
| `watermark` | Number | Unix timestamp (milliseconds). All messages sent **before or at** this timestamp have been delivered. Always present. |

**Usage note:** `watermark` is always reliable. `mids` provides per-message granularity but may be absent for backward-compatibility reasons with older clients.

**Doc:** https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/message-deliveries

---

### 4.13 `message_reads`

**Subscription field:** `message_reads`
**Triggered by:** A user reads (opens) a message your Page sent.

```json
{
  "sender": {
    "id": "<PSID>"
  },
  "recipient": {
    "id": "<PAGE_ID>"
  },
  "timestamp": 1458668856463,
  "read": {
    "watermark": 1458668856253
  }
}
```

**`read` object fields:**

| Field       | Type   | Description                                                                                               |
|-------------|--------|-----------------------------------------------------------------------------------------------------------|
| `watermark` | Number | Unix timestamp (milliseconds). All messages sent **before or at** this timestamp have been read by the user. |

**Doc:** https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/message-reads

---

### 4.14 `messaging_postbacks`

**Subscription field:** `messaging_postbacks`
**Triggered by:**
- User clicks a **Postback Button** in a message template.
- User taps the **Get Started** button (first time opening the conversation).
- User clicks a **Persistent Menu** item.

```json
{
  "sender": {
    "id": "<PSID>"
  },
  "recipient": {
    "id": "<PAGE_ID>"
  },
  "timestamp": "1527459824",
  "postback": {
    "mid": "m_MESSAGE-ID",
    "title": "TITLE-FOR-THE-CTA",
    "payload": "USER-DEFINED-PAYLOAD",
    "referral": {
      "ref": "USER-DEFINED-REFERRAL-PARAM",
      "source": "SHORT-URL",
      "type": "OPEN_THREAD"
    }
  }
}
```

**`postback` object fields:**

| Field              | Type   | Required | Description                                                                                                              |
|--------------------|--------|----------|--------------------------------------------------------------------------------------------------------------------------|
| `mid`              | String | Yes      | The message ID.                                                                                                          |
| `title`            | String | Yes      | The text label of the button or menu item the user clicked.                                                              |
| `payload`          | String | Yes      | Developer-defined string (up to 1000 chars) attached to the CTA. Only sent to the app that originally sent the button.  |
| `referral`         | Object | No       | Present when the user entered the conversation via an m.me link, Click-to-Messenger ad, QR code, or Welcome Screen.     |
| `referral.ref`     | String | No       | Arbitrary data from the `ref` param of the m.me link. Alphanumeric + `-`, `_`, `=` only.                                |
| `referral.source`  | String | No       | URL for the referral. `"SHORTLINK"` for m.me links, `"ADS"` for Messenger Conversation Ads.                             |
| `referral.type`    | String | No       | Always `"OPEN_THREAD"` for m.me link referrals.                                                                          |

**Note on standby channel:** Postback events delivered via the Standby channel do **not** include the `postback.payload`. The primary app that sent the button receives the full payload.

**Doc:** https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messaging_postbacks/

---

### 4.15 `messaging_optins` — Recurring Notifications

**Subscription field:** `messaging_optins`
**Triggered by:**
- A user opts in to receive Marketing / Recurring Notification Messages.
- A user re-opts in (clicks "Continue messages" before token expiration).
- A user changes their opt-in status (stops or resumes notifications).

**Note:** The legacy "Send to Messenger" plugin optin (`optin.ref` from `data-ref`) is also delivered through this event. For the current Notification Messages flow, `optin.type` is always `"notification_messages"`.

```json
{
  "sender": {
    "id": "<PSID>"
  },
  "recipient": {
    "id": "<PAGE_ID>"
  },
  "timestamp": "<TIMESTAMP>",
  "optin": {
    "type": "notification_messages",
    "payload": "ADDITIONAL-INFORMATION",
    "notification_messages_token": "NOTIFICATION-MESSAGES-TOKEN",
    "notification_messages_frequency": "DAILY",
    "notification_messages_timezone": "America/Los_Angeles",
    "token_expiry_timestamp": "<UNIX_TIMESTAMP>",
    "user_token_status": "NOT_REFRESHED",
    "notification_messages_status": "STOP NOTIFICATIONS",
    "title": "Weekly News Update"
  }
}
```

**`optin` object fields:**

| Field                             | Type   | Description                                                                                                                     |
|-----------------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------|
| `type`                            | String | Always `"notification_messages"` for the current flow.                                                                         |
| `payload`                         | String | Additional developer-defined information included in the webhook.                                                               |
| `title`                           | String | The title of the notification template the user opted into.                                                                     |
| `notification_messages_token`     | String | The token used to send Marketing Messages to this user for this topic + frequency combination.                                  |
| `notification_messages_frequency` | Enum   | `DAILY` (1 per 24h, 6 months), `WEEKLY` (1 per week, 9 months), `MONTHLY` (1 per month, 12 months). Removed in API v16.       |
| `notification_messages_timezone`  | String | IANA timezone identifier for the opted-in user (e.g., `"America/New_York"`).                                                   |
| `token_expiry_timestamp`          | Number | Unix timestamp of when the `notification_messages_token` expires.                                                               |
| `user_token_status`               | Enum   | `"REFRESHED"` — user re-opted in after expiration. `"NOT_REFRESHED"` — default, token not yet refreshed.                      |
| `notification_messages_status`    | Enum   | **Only present when user changes status.** `"STOP NOTIFICATIONS"` or `"RESUME NOTIFICATIONS"`.                                 |

**Doc:** https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messaging_optins

---

### 4.16 `messaging_referrals`

**Subscription field:** `messaging_referrals`
**Triggered by:** An existing Messenger thread is re-entered via an m.me link with a `ref` param, or a Click-to-Messenger ad (when the user already has an open thread). For new thread referrals, the data appears in the `messaging_postbacks` event.

**m.me link example:**

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1458692752478,
  "referral": {
    "ref": "my-promo-code-123",
    "source": "SHORTLINK",
    "type": "OPEN_THREAD"
  }
}
```

**Ad referral example:**

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1458692752478,
  "referral": {
    "ref": "optional-ref-from-ad",
    "ad_id": "<ID_OF_THE_AD>",
    "source": "ADS",
    "type": "OPEN_THREAD",
    "ads_context_data": {
      "ad_title": "Summer Sale — 50% Off",
      "photo_url": "https://example.com/ad-image.jpg",
      "video_url": "https://example.com/ad-thumb.jpg",
      "post_id": "<POST_ID>",
      "product_id": "<PRODUCT_ID>",
      "flow_id": "<WELCOME_MESSAGE_FLOW_ID>"
    }
  }
}
```

**`referral` object fields:**

| Field                              | Type   | Description                                                                                             |
|------------------------------------|--------|---------------------------------------------------------------------------------------------------------|
| `ref`                              | String | Arbitrary data from the m.me link `ref` param or ad configuration. Alphanumeric + `-`, `_`, `=` only.  |
| `source`                           | String | `"SHORTLINK"` for m.me links, `"ADS"` for Messenger Conversation Ads.                                  |
| `type`                             | String | Always `"OPEN_THREAD"`.                                                                                 |
| `referer_uri`                      | String | (Optional) URI of the website from which the message was sent (Chat Plugin contexts).                  |
| `ad_id`                            | String | (Ads only) The ID of the Click-to-Messenger advertisement.                                              |
| `ads_context_data`                 | Object | (Ads only) Context from the ad that triggered the conversation.                                         |
| `ads_context_data.ad_title`        | String | Title of the advertisement.                                                                             |
| `ads_context_data.photo_url`       | String | URL of the image from the ad.                                                                           |
| `ads_context_data.video_url`       | String | Thumbnail URL of the video from the ad.                                                                 |
| `ads_context_data.post_id`         | String | ID of the Facebook post associated with the ad.                                                         |
| `ads_context_data.product_id`      | String | (Optional) ID of the product the user showed interest in.                                               |
| `ads_context_data.flow_id`         | String | (Optional) ID of the partner app Welcome Message flow.                                                  |

**Doc:** https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messaging_referrals

---

### 4.17 `messaging_handovers` — pass_thread_control

**Subscription field:** `messaging_handovers`
**Triggered by:** Thread control has been **passed to your application** from another app (or from Page Inbox).

This is part of the Messenger **Handover Protocol**, which allows multiple apps (Primary Receiver and Secondary Receivers) to take turns controlling a conversation.

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1458692752478,
  "pass_thread_control": {
    "new_owner_app_id": "123456789",
    "previous_owner_app_id": "987654321",
    "metadata": "context from the previous app"
  }
}
```

**`pass_thread_control` fields:**

| Field                  | Type   | Description                                                                              |
|------------------------|--------|------------------------------------------------------------------------------------------|
| `new_owner_app_id`     | String | App ID that is now in control of the thread (your app's ID when receiving this event).   |
| `previous_owner_app_id`| String | App ID that previously held control. `null` if the thread was previously in idle mode.   |
| `metadata`             | String | Optional custom string passed by the app that initiated the handover.                    |

**Known app IDs:**
- `263902037430900` — Page Inbox (Facebook's native inbox)

---

### 4.18 `messaging_handovers` — take_thread_control

**Subscription field:** `messaging_handovers`
**Triggered by:** Thread control has been **taken away from your application** by another app (typically the Primary Receiver).

```json
{
  "sender": { "id": "<PSID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1458692752478,
  "take_thread_control": {
    "previous_owner_app_id": "123456789",
    "new_owner_app_id": "987654321",
    "metadata": "taking back control"
  }
}
```

**`take_thread_control` fields:**

| Field                  | Type   | Description                                                                         |
|------------------------|--------|-------------------------------------------------------------------------------------|
| `previous_owner_app_id`| String | App ID that was in control (your app's ID when receiving this event). Can be `null` if thread was in idle mode. |
| `new_owner_app_id`     | String | App ID that now holds control.                                                      |
| `metadata`             | String | Optional custom string passed by the app that initiated the takeover.               |

---

### 4.19 `messaging_handovers` — request_thread_control

**Subscription field:** `messaging_handovers`
**Triggered by:** A Secondary Receiver app has called the Request Thread Control API, asking the Primary Receiver to pass control. This event is delivered **only to the Primary Receiver app**. The Primary Receiver may choose to honor or ignore the request.

```json
{
  "sender": { "id": "<USER_ID>" },
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1458692752478,
  "request_thread_control": {
    "requested_owner_app_id": 123456789,
    "metadata": "customer wants live agent"
  }
}
```

**`request_thread_control` fields:**

| Field                    | Type   | Description                                                                          |
|--------------------------|--------|--------------------------------------------------------------------------------------|
| `requested_owner_app_id` | Number | App ID of the Secondary Receiver requesting thread control.                          |
| `metadata`               | String | Optional custom string passed by the Secondary Receiver.                             |

---

### 4.20 `messaging_handovers` — app_roles

**Subscription field:** `messaging_handovers`
**Triggered by:** A Page administrator changes the role assigned to your application.

```json
{
  "recipient": { "id": "<PAGE_ID>" },
  "timestamp": 1458692752478,
  "app_roles": {
    "123456789": ["primary_receiver"]
  }
}
```

**`app_roles` fields:**

| Field                   | Type            | Description                                                                         |
|-------------------------|-----------------|-------------------------------------------------------------------------------------|
| `app_roles`             | Object          | A map from App ID (string) to an array of role strings.                             |
| `app_roles["<APP_ID>"]` | Array of String | Role(s) assigned to the app. Values: `"primary_receiver"` or `"secondary_receiver"`. |

**Note:** Unlike other `messaging_handovers` events, `app_roles` does not have a `sender` field.

**Doc:** https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messaging_handovers/

---

### 4.21 `messaging_policy_enforcement`

**Subscription field:** `messaging_policy_enforcement`
**Triggered by:** Meta takes a policy enforcement action against the Page/app for policy violations. Common violations include excessive spam, inappropriate content (pornography, self-harm content), abuse of message tags, etc.

```json
{
  "recipient": {
    "id": "<PAGE_ID>"
  },
  "timestamp": 1458692752478,
  "policy_enforcement": {
    "action": "block",
    "reason": "The bot violated our Platform Policies (https://developers.facebook.com/devpolicy/#messengerplatform). Common violations include sending out excessive spammy messages or being non-functional."
  }
}
```

**`policy_enforcement` object fields:**

| Field    | Type   | Description                                                                                          |
|----------|--------|------------------------------------------------------------------------------------------------------|
| `action` | String | The enforcement action taken. Values: `"warning"`, `"block"`, `"unblock"`.                           |
| `reason` | String | Human-readable explanation of the violation. **Absent when `action` is `"unblock"`.**                |

**Note:** Unlike most webhook events, `messaging_policy_enforcement` does not include a `sender` field — only `recipient` (your Page ID) is present.

**Policy reference:**
- https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messaging_policy_enforcement/
- https://developers.facebook.com/docs/messenger-platform/policy/

---

### 4.22 `standby` channel

**Subscription field:** `standby`
**Triggered by:** A message, read event, or delivery event occurs in a conversation where your application is **not the current thread owner** (i.e., another app or Page Inbox currently controls the thread). This is part of the Handover Protocol.

The standby channel delivers the same event types as the normal `messaging` channel — but because you are not in control, you should not respond to the user. You can use standby events to monitor the conversation for context.

**Outer envelope difference:** Uses `"standby"` array instead of `"messaging"`:

```json
{
  "object": "page",
  "entry": [
    {
      "id": "<PAGE_ID>",
      "time": 1458692752478,
      "standby": [
        {
          "sender": { "id": "<USER_ID>" },
          "recipient": { "id": "<PAGE_ID>" },
          "timestamp": 1458692752478,
          "message": {
            "mid": "mid.1457764197618:41d102a3e1ae206a38",
            "text": "hello"
          }
        }
      ]
    }
  ]
}
```

**Supported event types in standby channel:**

| Event Type        | Description                                                          |
|-------------------|----------------------------------------------------------------------|
| `message`         | Text messages and attachments received while not in control.         |
| `read`            | Read receipts for messages sent by your Page.                        |
| `delivery`        | Delivery receipts for messages sent by your Page.                    |

**`standby` envelope properties:**

| Field          | Type      | Description                                          |
|----------------|-----------|------------------------------------------------------|
| `id`           | String    | The Page ID.                                         |
| `time`         | Timestamp | Unix timestamp (ms) of the event.                    |
| `standby`      | Array     | Array of messaging event objects (same schema as `messaging` array). |

**Doc:** https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/standby/

---

## 5. All Subscribable Webhook Fields (Reference Table)

These are the field names you select when configuring your webhook subscription.

| Field                            | Messenger Only | Description                                                                                                    |
|----------------------------------|:--------------:|----------------------------------------------------------------------------------------------------------------|
| `messages`                       | No             | Incoming messages from users (text, attachments, quick replies, replies, unsend, referral-enriched messages).  |
| `message_deliveries`             | Yes            | Delivery confirmations for messages sent by your Page.                                                         |
| `message_echoes`                 | Yes            | Echo of messages sent by your Page (outgoing messages).                                                        |
| `message_edits`                  | Yes            | Notification when a user edits a previously sent message.                                                      |
| `message_reactions`              | No             | When a user reacts to or unreacts from a message your Page sent.                                               |
| `message_reads`                  | Yes            | Read receipts — when a user reads a message your Page sent.                                                    |
| `messaging_account_linking`      | Yes            | User links or unlinks their Messenger account with your business account.                                      |
| `messaging_feedback`             | Yes            | User submits feedback via a Customer Feedback template.                                                        |
| `messaging_game_plays`           | Yes            | User plays a round of an Instant Game (Beta).                                                                  |
| `messaging_handovers`            | No             | Handover Protocol events: pass/take/request thread control, app role changes.                                  |
| `messaging_optins`               | Yes            | User opts in (or modifies opt-in) for Recurring Notification Messages; legacy Send to Messenger plugin events. |
| `messaging_payments`             | Yes            | Payment transactions (Beta).                                                                                   |
| `messaging_policy_enforcement`   | Yes            | Policy violation warnings, blocks, or unblocks on your Page/app.                                               |
| `messaging_postbacks`            | No             | User clicks Postback Button, Get Started button, or Persistent Menu item.                                      |
| `messaging_referrals`            | No             | User re-enters an existing thread via m.me link or Click-to-Messenger ad.                                      |
| `messaging_seen`                 | No (Instagram) | Read receipts for Instagram Messaging (equivalent of `message_reads` for Messenger).                          |
| `messenger_template_status_update`| Yes           | Status change on a Utility Message template review.                                                            |
| `response_feedback`              | No             | User provides feedback by clicking feedback buttons.                                                           |
| `send_cart`                      | Yes            | User sends a cart/order message.                                                                               |
| `standby`                        | No             | Events delivered to the standby channel when your app is not the thread owner (Handover Protocol).            |

---

## Appendix: Reference URLs

| Topic                            | URL                                                                                                              |
|----------------------------------|------------------------------------------------------------------------------------------------------------------|
| Messenger Platform Overview      | https://developers.facebook.com/docs/messenger-platform                                                          |
| Webhooks Setup                   | https://developers.facebook.com/docs/messenger-platform/webhooks                                                 |
| Webhook Events Overview          | https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/                               |
| `messages` Event                 | https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messages/                      |
| `message_deliveries` Event       | https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/message-deliveries             |
| `message_reads` Event            | https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/message-reads                  |
| `message_reactions` Event        | https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/message-reactions              |
| `messaging_postbacks` Event      | https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messaging_postbacks/           |
| `messaging_optins` Event         | https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messaging_optins               |
| `messaging_referrals` Event      | https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messaging_referrals            |
| `messaging_handovers` Event      | https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messaging_handovers/           |
| `messaging_policy_enforcement`   | https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/messaging_policy_enforcement/  |
| `standby` Event                  | https://developers.facebook.com/docs/messenger-platform/reference/webhook-events/standby/                       |
| Handover Protocol API Reference  | https://developers.facebook.com/docs/messenger-platform/reference/handover-protocol                             |
| Create an App                    | https://developers.facebook.com/docs/development/create-an-app/                                                 |
| App Review (Messenger)           | https://developers.facebook.com/docs/messenger-platform/app-review/                                             |
| App Review (General)             | https://developers.facebook.com/docs/resp-plat-initiatives/individual-processes/app-review/                     |
| Quick Start                      | https://developers.facebook.com/docs/messenger-platform/getting-started/quick-start/                            |
| Messenger Platform Policies      | https://developers.facebook.com/policy#messengerplatform                                                         |
