# Instagram Messaging API — Sending a Text Message

## Endpoint

```
POST https://graph.facebook.com/{api-version}/{page-id}/messages
```

- `{api-version}`: Use the current stable version (e.g. `v25.0`). The Instagram Messaging API is accessed through the Graph API versioned base URL.
- `{page-id}`: The Facebook Page ID that is connected to the Instagram Professional account. You may also use `me` as a shorthand when authenticating as the page (resolves to the page the token represents).

---

## HTTP Method

`POST`

---

## Authentication

Pass a **Page Access Token (PAT)** as a query parameter:

```
?access_token={PAGE_ACCESS_TOKEN}
```

How to obtain a PAT:
1. The user completes Facebook Login and grants `instagram_basic`, `instagram_manage_messages`, and `pages_manage_metadata` permissions.
2. Exchange the resulting User Access Token for a Page Access Token by calling `GET /{page-id}?fields=access_token`.
3. If a **long-lived** User Access Token is used, the resulting Page Access Token has **no expiration date**.

Required permissions:
- `instagram_basic`
- `instagram_manage_messages`
- `pages_manage_metadata`

---

## Request Body

Send as either `application/json` (with `Content-Type: application/json` header) or as form-encoded query parameters.

### Minimal body for a text message

```json
{
  "recipient": {
    "id": "{IGSID}"
  },
  "message": {
    "text": "{MESSAGE_TEXT}"
  }
}
```

### With explicit messaging type (recommended)

```json
{
  "recipient": {
    "id": "{IGSID}"
  },
  "messaging_type": "RESPONSE",
  "message": {
    "text": "{MESSAGE_TEXT}"
  }
}
```

### Field reference

| Field | Type | Required | Description |
|---|---|---|---|
| `recipient.id` | string | Yes | The Instagram-Scoped ID (IGSID) of the recipient. Obtained from an incoming webhook `sender.id`. |
| `message.text` | string | Yes | The text content. Must be UTF-8 encoded and fewer than 1000 characters. |
| `messaging_type` | string | Recommended | `RESPONSE` — reply within the 24-hour window. `MESSAGE_TAG` — send outside window with an approved tag. `UPDATE` — general non-promotional update. |

---

## Example Request (cURL)

```bash
curl -X POST \
  "https://graph.facebook.com/v25.0/{PAGE-ID}/messages?access_token={PAGE_ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": { "id": "{IGSID}" },
    "messaging_type": "RESPONSE",
    "message": { "text": "Hello! How can I help you today?" }
  }'
```

Form-encoded alternative (also accepted):

```bash
curl -i -X POST \
  "https://graph.facebook.com/v25.0/me/messages?access_token={PAGE_ACCESS_TOKEN}" \
  --data 'recipient={"id":"{IGSID}"}&message={"text":"Hello! How can I help you today?"}'
```

---

## Success Response

```json
{
  "recipient_id": "{IGSID}",
  "message_id": "AG5Hz2U..."
}
```

| Field | Description |
|---|---|
| `recipient_id` | The IGSID of the recipient the message was sent to. |
| `message_id` | Unique identifier for the sent message. |

---

## Key Notes

### 24-Hour Messaging Window
- Standard replies must be sent within **24 hours** of the customer's last message.
- Outside this window, use `"messaging_type": "MESSAGE_TAG"` with an approved tag to re-engage.
- Message requests (from users who haven't interacted before) expire after **30 days** of inactivity.

### Recipient ID (IGSID)
- The `recipient.id` must be an **Instagram-Scoped ID (IGSID)** — a per-user, per-business identifier.
- The IGSID is provided as `sender.id` in every inbound webhook event; capture and store it when a user first messages you.

### App Access Level
- Apps with **Standard Access** can only message users who have roles on the app (developers/testers).
- **Advanced Access** (requires App Review) is required to message general Instagram users.

### Text Constraints
- Maximum length: **1000 characters** (UTF-8).
- Links embedded in the text must be properly formatted URLs; they will render as clickable links.

### No Group Messaging
- The Instagram Messaging API does not support group conversations; each thread is one customer per conversation.

### Connected Tools Setting
- The Instagram account must have the **"Connected Tools"** messaging setting enabled, and the account must be linked to a Facebook Page.
