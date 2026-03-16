# Twitter/X API — Send a Direct Message (Text)

## Endpoint

```
POST https://api.x.com/2/dm_conversations/with/{participant_id}/messages
```

| Field            | Value                                                               |
|------------------|---------------------------------------------------------------------|
| Method           | POST                                                                |
| Path parameter   | `participant_id` — numeric user ID of the recipient (1–19 digits)  |
| API version      | v2 (2.159)                                                          |

Alternative endpoints (when a conversation already exists):

```
POST https://api.x.com/2/dm_conversations/{dm_conversation_id}/messages
POST https://api.x.com/2/dm_conversations   (creates a new group conversation)
```

---

## Authentication

| Requirement      | Detail                                                      |
|------------------|-------------------------------------------------------------|
| Scheme           | OAuth 2.0 PKCE (User Access Token)                          |
| Header           | `Authorization: Bearer {USER_ACCESS_TOKEN}`                 |
| Required scopes  | `dm.write`, `dm.read`, `tweet.read`, `users.read`           |
| Account          | Approved X developer account with an active Project and App |

`dm.read` must be present alongside `dm.write` even when only sending.

---

## Request Headers

```
Authorization: Bearer {USER_ACCESS_TOKEN}
Content-Type: application/json
```

---

## Request Body

```json
{
  "text": "Your message text here."
}
```

| Field         | Type   | Required | Description                                       |
|---------------|--------|----------|---------------------------------------------------|
| `text`        | string | yes*     | Message body. Minimum 1 character.                |
| `attachments` | array  | yes*     | Required when no text. Array of attachment objects.|

*One of `text` or `attachments` must be present.

### With a media attachment

```json
{
  "text": "Check this out!",
  "attachments": [
    { "media_id": "1234567890123456789" }
  ]
}
```

Media must be uploaded via the X media upload API before use; `media_id` is the string ID returned from that upload.

---

## Example Request (cURL)

```bash
curl -X POST "https://api.x.com/2/dm_conversations/with/9876543210/messages" \
  -H "Authorization: Bearer $USER_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text": "Hello! This is a message from the API."}'
```

Replace `9876543210` with the recipient's numeric user ID.

---

## Response

**HTTP 201 Created**

```json
{
  "data": {
    "dm_conversation_id": "123456789-987654321",
    "dm_event_id": "128341038123123"
  }
}
```

| Field               | Description                                    |
|---------------------|------------------------------------------------|
| `dm_conversation_id`| ID of the conversation (numeric or `ID-ID` format) |
| `dm_event_id`       | ID of the specific message event               |

### Error response (RFC 7807)

```json
{
  "errors": [
    {
      "type": "...",
      "title": "...",
      "status": 403,
      "detail": "..."
    }
  ]
}
```

Common error types: `resource-not-found`, `invalid-request`, `client-forbidden`.

---

## Key Notes

- The recipient must be obtainable via the X User Lookup endpoint if only a username is known; the API requires the numeric user ID.
- You can only delete messages you sent yourself — not messages from other participants (`DELETE /2/dm_events/{id}`).
- Group conversations are created via `POST /2/dm_conversations` with participant ID arrays (excluding yourself).
- No explicit rate limit is published in the v2 DM docs; apply standard back-off on 429 responses.
- The token must be obtained via a 3-legged OAuth 2.0 PKCE flow (user-delegated), not app-only bearer tokens.
