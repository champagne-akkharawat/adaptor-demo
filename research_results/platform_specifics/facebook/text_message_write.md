# Facebook Messenger Platform — Sending a Text Message

## Endpoint

```
POST https://graph.facebook.com/v25.0/{PAGE-ID}/messages
```

## Authentication

- Pass a **Page Access Token** as a query parameter: `?access_token={PAGE-ACCESS-TOKEN}`
- The `{PAGE-ID}` in the URL is the Facebook Page ID associated with your app.

## Request Headers

```
Content-Type: application/json
```

## Request Body

| Field            | Type   | Required | Description                                                                 |
|------------------|--------|----------|-----------------------------------------------------------------------------|
| `recipient.id`   | string | Yes      | Page-Scoped ID (PSID) of the recipient                                      |
| `messaging_type` | string | Yes      | Message type context (see values below)                                     |
| `message.text`   | string | Yes      | The text content to send (up to 2000 characters)                            |

### `messaging_type` Values

| Value            | Use Case                                                                 |
|------------------|--------------------------------------------------------------------------|
| `RESPONSE`       | Reply to a message received within the last 24 hours (most common)       |
| `UPDATE`         | Proactive non-promotional update initiated by the page                   |
| `MESSAGE_TAG`    | Message sent outside the 24-hour window using an approved message tag    |

## Example Request

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": { "id": "{PSID}" },
    "messaging_type": "RESPONSE",
    "message": { "text": "Hello, world!" }
  }' \
  "https://graph.facebook.com/v25.0/{PAGE-ID}/messages?access_token={PAGE-ACCESS-TOKEN}"
```

## Example Success Response

```json
{
  "recipient_id": "1008372609250235",
  "message_id": "m_AG5Hz2Uq7tuwNEhXfYYKj8mJEM_QPpz5jdBtHs5XKxSy0"
}
```

## Key Notes

- **24-Hour Window**: Once a user messages your page, you have 24 hours to reply using `messaging_type: RESPONSE`. Outside this window, only approved message tags or one-time notifications are permitted.
- **User Initiation**: You cannot send a message to a user who has not first contacted your page (without special permissions).
- **PSID**: The recipient ID must be a Page-Scoped ID — the identifier assigned to a user when they first message your page. It is not a global Facebook user ID.
- **Bot Disclosure**: Automated responses must identify themselves (e.g., "I'm the [Page Name] bot"), particularly for users in California and Germany.
- **Graph API Version**: The current documented version is `v25.0`. Pin your integration to a specific version.
