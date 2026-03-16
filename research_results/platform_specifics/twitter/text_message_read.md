# Twitter/X API — Receiving Direct Messages (Incoming DM Events)

## Two approaches

X provides two complementary ways to receive incoming DMs:

| Approach                  | Mechanism       | Best for                              |
|---------------------------|-----------------|---------------------------------------|
| Account Activity API      | Webhook (push)  | Real-time, production integrations    |
| DM Lookup endpoints       | Polling (pull)  | Simple integrations, 30-day history   |

---

## Approach 1 — Webhook via Account Activity API (recommended)

### Setup Steps

1. **Register a webhook URL** using the V2 Webhooks API:
   ```
   POST https://api.twitter.com/1.1/account_activity/all/{env_name}/webhooks.json
   ```
   The URL must be public HTTPS, no custom port numbers.

2. **Pass the CRC challenge** — X immediately sends a GET request to your endpoint; your server must respond with a valid HMAC-SHA256 token (see Verification section below).

3. **Add a user subscription** so X knows which account's activity to stream to your webhook:
   ```
   POST https://api.twitter.com/1.1/account_activity/all/{env_name}/subscriptions.json
   ```
   This step uses OAuth 1.0a (3-legged) authenticated as the user being subscribed.

4. **Receive events** — X sends POST requests with JSON payloads to your registered URL whenever activity occurs for subscribed users.

5. **Acknowledge receipt** — Return HTTP `200 OK` within 10 seconds or X will retry.

### Authentication for Setup

| Action                           | Auth method                          |
|----------------------------------|--------------------------------------|
| Register/manage webhook URL      | OAuth 2.0 App-Only Bearer Token      |
| Add/remove user subscriptions    | OAuth 1.0a (3-legged, user context)  |

---

## Incoming DM Event Payload

X delivers a POST to your webhook with the following JSON structure:

```json
{
  "for_user_id": "4337869213",
  "direct_message_events": [
    {
      "type": "message_create",
      "id": "954491830116155396",
      "created_timestamp": "1516403560557",
      "message_create": {
        "target": {
          "recipient_id": "4337869213"
        },
        "sender_id": "3001969357",
        "source_app_id": "13090192",
        "message_data": {
          "text": "Hello World!",
          "entities": {
            "hashtags": [],
            "urls": [],
            "user_mentions": []
          }
        }
      }
    }
  ],
  "users": {
    "3001969357": { /* sender user object */ },
    "4337869213": { /* recipient user object */ }
  }
}
```

### Extracting Sender and Text

| Value          | JSON path                                                          |
|----------------|--------------------------------------------------------------------|
| Sender ID      | `direct_message_events[n].message_create.sender_id`               |
| Message text   | `direct_message_events[n].message_create.message_data.text`        |
| Recipient ID   | `direct_message_events[n].message_create.target.recipient_id`      |
| Subscribed user| `for_user_id` (the account whose activity triggered the event)     |
| Event ID       | `direct_message_events[n].id`                                      |
| Timestamp (ms) | `direct_message_events[n].created_timestamp`                       |

The `users` map in the payload provides full user objects keyed by user ID, so sender details (name, username, etc.) can be resolved without an additional API call.

### Event Types

| `type` value       | Meaning                              |
|--------------------|--------------------------------------|
| `message_create`   | A message was sent in the conversation |
| `ParticipantsJoin` | A user joined a group conversation   |
| `ParticipantsLeave`| A user left a group conversation     |

---

## Webhook Verification (CRC Challenge)

X periodically sends a GET request to your webhook URL to confirm you control the endpoint. Your server must respond correctly or the webhook will be deactivated.

**Request from X:**
```
GET {your_webhook_url}?crc_token={random_token}
```

**Required response (HTTP 200):**
```json
{
  "response_token": "sha256={HMAC_SHA256_hash}"
}
```

**Computing the hash (pseudocode):**
```
response_token = "sha256=" + Base64( HMAC-SHA256( consumer_secret, crc_token ) )
```

- `consumer_secret` is your app's API Key Secret from the X Developer Portal.
- The response must arrive within **10 seconds**.

**Signature verification on incoming POST events:**

Each POST from X includes the header:
```
x-twitter-webhooks-signature: sha256={HMAC_SHA256_of_payload}
```

Verify this against your consumer secret and the raw POST body before processing.

---

## Approach 2 — Polling via DM Lookup Endpoints

For simpler integrations, poll for new DM events directly.

| Purpose                     | Endpoint                                                              |
|-----------------------------|-----------------------------------------------------------------------|
| All DM events for the user  | `GET https://api.x.com/2/dm_events`                                  |
| Events in a 1:1 conversation| `GET https://api.x.com/2/dm_conversations/with/{participant_id}/dm_events` |
| Events by conversation ID   | `GET https://api.x.com/2/dm_conversations/{dm_conversation_id}/dm_events` |

- **Authentication**: OAuth 2.0 PKCE User Access Token with `dm.read`, `tweet.read`, `users.read` scopes.
- **Retention**: Events from the last **30 days** are available.
- Use `since_id` / `until_id` pagination parameters to page through events and avoid re-processing.

---

## Key Notes

- A single webhook can serve multiple subscribed users; `for_user_id` disambiguates whose event arrived.
- The Account Activity API is the only way to receive DMs in real time — the REST lookup endpoints require polling.
- Webhooks require your endpoint to be publicly reachable on standard HTTPS (port 443); local development requires a tunnel (e.g., ngrok).
- Always verify the `x-twitter-webhooks-signature` header before acting on a payload to prevent spoofing.
- Media attachments in incoming DMs are referenced via `message_data.attachment.media` and require a separate media fetch if content is needed.
