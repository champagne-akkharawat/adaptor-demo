# Microsoft Teams Bot Framework — Sending a Text Message

## Overview

Teams bots communicate through the **Bot Connector service** using the Activity schema over HTTPS REST. There are two sending patterns: replying to an existing message (reactive) and sending into a conversation without a user prompt (non-reply / proactive).

---

## 1. Reply to an Incoming Message

### Endpoint

```
POST {serviceUrl}/v3/conversations/{conversationId}/activities/{activityId}
```

- `{serviceUrl}` — taken from the `serviceUrl` field of the incoming Activity (e.g. `https://smba.trafficmanager.net/teams`). Never hardcode this for replies; always use the value from the inbound request.
- `{conversationId}` — the `conversation.id` field from the incoming Activity.
- `{activityId}` — the `id` field of the Activity you are replying to (same as `replyToId` in your outbound body).

### Method

`POST`

### Authentication

`Authorization: Bearer <ACCESS_TOKEN>`

The token is a JWT obtained from the Microsoft Entra ID login service using the OAuth 2.0 client credentials flow (see Authentication section below).

### Request Headers

```
Authorization: Bearer <ACCESS_TOKEN>
Content-Type: application/json
```

### Request Body

Minimum body for a plain-text reply:

```json
{
    "type": "message",
    "from": {
        "id": "<bot-channel-account-id>",
        "name": "<bot-display-name>"
    },
    "conversation": {
        "id": "<conversationId>"
    },
    "recipient": {
        "id": "<user-channel-account-id>",
        "name": "<user-display-name>"
    },
    "text": "Hello, this is the bot's reply.",
    "replyToId": "<activityId-of-original-message>"
}
```

Full example from the official docs:

```http
POST https://smba.trafficmanager.net/teams/v3/conversations/abcd1234/activities/5d5cdc723
Authorization: Bearer ACCESS_TOKEN
Content-Type: application/json
```

```json
{
    "type": "message",
    "from": {
        "id": "12345678",
        "name": "Pepper's News Feed"
    },
    "conversation": {
        "id": "abcd1234",
        "name": "Convo1"
    },
    "recipient": {
        "id": "1234abcd",
        "name": "SteveW"
    },
    "text": "My bot's reply",
    "replyToId": "5d5cdc723"
}
```

### Success Response

HTTP 200 with a body containing the new activity ID:

```json
{
    "id": "<new-activityId>"
}
```

---

## 2. Non-Reply Message (Send into Existing Conversation)

When you need to send a message that is not a direct reply to a user's message (e.g. a notification):

### Endpoint

```
POST {serviceUrl}/v3/conversations/{conversationId}/activities
```

### Request Body

Same Activity schema as above, but without `replyToId`.

---

## 3. Update an Existing Message

```
PUT {serviceUrl}/v3/conversations/{conversationId}/activities/{activityId}
```

```json
{
    "type": "message",
    "text": "This message has been updated"
}
```

You must have cached the `activityId` from the original POST response.

---

## 4. Proactive Message (New Conversation or Thread)

### Step 1 — Create the Conversation

```
POST {serviceUrl}/v3/conversations
Authorization: Bearer <ACCESS_TOKEN>
Content-Type: application/json
```

```json
{
    "bot": {
        "id": "28:10j12ou0d812-2o1098-c1mjojzldxcj-1098028n",
        "name": "The Bot"
    },
    "members": [
        {
            "id": "29:012d20j1cjo20211"
        }
    ],
    "channelData": {
        "tenant": {
            "id": "197231joe-1209j01821-012kdjoj"
        }
    }
}
```

Response returns `{ "id": "<conversationId>" }`. Store this for subsequent sends.

### Step 2 — Send the Message

Use the non-reply endpoint above with the returned `conversationId`.

### Global serviceUrl Fallbacks (for proactive only, when no inbound activity is available)

| Environment | serviceUrl |
|-------------|-----------|
| Public      | `https://smba.trafficmanager.net/teams/` |
| GCC         | `https://smba.infra.gcc.teams.microsoft.com/teams` |
| GCC High    | `https://smba.infra.gov.teams.microsoft.us/teams` |
| DoD         | `https://smba.infra.dod.teams.microsoft.us/teams` |

---

## 5. Authentication — Obtaining the Bearer Token

Use the OAuth 2.0 client credentials flow against Microsoft Entra ID.

### Multi-tenant bot (most common)

```http
POST https://login.microsoftonline.com/botframework.com/oauth2/v2.0/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
&client_id=<MICROSOFT-APP-ID>
&client_secret=<MICROSOFT-APP-PASSWORD>
&scope=https%3A%2F%2Fapi.botframework.com%2F.default
```

### Single-tenant bot

```http
POST https://login.microsoftonline.com/<TENANT-ID>/oauth2/v2.0/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
&client_id=<MICROSOFT-APP-ID>
&client_secret=<MICROSOFT-APP-PASSWORD>
&scope=https%3A%2F%2Fapi.botframework.com%2F.default
```

### Token Response

```json
{
    "token_type": "Bearer",
    "expires_in": 3600,
    "ext_expires_in": 3600,
    "access_token": "eyJhbGciOiJIUzI1Ni..."
}
```

- Tokens are valid for 3600 seconds. Cache and proactively refresh them.
- Use the exact `access_token` value as-is in the `Authorization` header — do not escape or encode it.

---

## 6. Key Notes

- **serviceUrl is dynamic.** For replies, always use the `serviceUrl` from the incoming Activity. Hardcoding it will break in non-public clouds and when Microsoft updates routing.
- **Bot registration required.** The bot must be registered in Azure Bot Service (via Azure Portal or Developer Portal for Teams) to obtain an App ID and password.
- **App must be installed.** For proactive messages to a group chat or channel, the app containing the bot must already be installed in that context. For personal scope, the bot must be installed for that user. A 403 is returned if these conditions are not met.
- **403 with `MessageWritesBlocked`.** Indicates the user has blocked or uninstalled the bot. You can use this to build a report of opted-out users.
- **Proactive messages to users via `aadObjectId`** are only supported in personal scope.
- **Teams doesn't support sending proactive messages using email or UPN** — use `userId`, `channelId`, or `teamId`.
- **Throttling.** Microsoft imposes no hard limit on message count, but individual channels (Teams) enforce their own throttling. Messages sent in quick succession may arrive out of order.
- **Bot Framework SDK** (C#, JS, Python) handles token acquisition and Activity routing automatically. Direct REST is the raw layer underneath.

---

## Sources

- https://learn.microsoft.com/en-us/azure/bot-service/rest-api/bot-framework-rest-connector-send-and-receive-messages
- https://learn.microsoft.com/en-us/microsoftteams/platform/bots/how-to/conversations/send-proactive-messages
- https://learn.microsoft.com/en-us/azure/bot-service/rest-api/bot-framework-rest-connector-authentication
