# Microsoft Teams Bot Framework — Receiving a Text Message

## Overview

When a Teams user sends a message to a bot, the Bot Connector service delivers an **Activity** object to the bot's registered messaging endpoint via an HTTP POST. The bot must expose a public HTTPS endpoint, verify the incoming JWT token, and process the Activity.

---

## 1. Setup Steps

### 1.1 Register the Bot

1. Create an **Azure Bot resource** in the Azure Portal (or use the Teams Developer Portal).
2. Note the **Microsoft App ID** and **App Password** (client secret) that are generated.
3. The registration creates a Bot Channel Registration that links your app credentials to the Bot Connector service.

### 1.2 Configure the Messaging Endpoint

In the Azure Bot resource, set the **Messaging endpoint** to your bot's public HTTPS URL:

```
https://<your-domain>/api/messages
```

This is the URL the Bot Connector service will POST inbound Activities to.

### 1.3 Enable the Microsoft Teams Channel

In the Azure Bot resource, go to **Channels** and enable **Microsoft Teams**. This allows Teams to route messages through the Bot Connector to your endpoint.

### 1.4 Create and Publish a Teams App Manifest

Create a Teams app manifest (`manifest.json`) referencing your Bot App ID. Package it as a `.zip` and sideload it into Teams (for development) or publish to your org's app catalog.

---

## 2. Incoming Activity Payload Structure

When a user sends a text message, the Bot Connector POSTs a JSON body to your endpoint matching the **Activity** schema. Below is the full structure for a typical text message in Teams.

```json
{
    "type": "message",
    "id": "<activityId>",
    "timestamp": "2024-01-15T10:23:45.123Z",
    "localTimestamp": "2024-01-15T10:23:45.123Z",
    "serviceUrl": "https://smba.trafficmanager.net/teams/",
    "channelId": "msteams",
    "from": {
        "id": "29:1AbCdEfGhIjKlMnOpQrStUvWx",
        "name": "Alice Smith",
        "aadObjectId": "aaaabbbb-0000-cccc-1111-dddd2222eeee"
    },
    "conversation": {
        "isGroup": false,
        "conversationType": "personal",
        "tenantId": "ffffeeeee-dddd-cccc-bbbb-aaaaaaaaaaaa",
        "id": "a:1qhNLqpUtmuI6U35gzjsJn7uRnCkW8NiZALHfN8AMxdbprS1uta2aT",
        "name": null
    },
    "recipient": {
        "id": "28:12345678-abcd-efgh-ijkl-mnopqrstuvwx",
        "name": "My Bot"
    },
    "textFormat": "plain",
    "locale": "en-US",
    "text": "Hello bot, what's the weather?",
    "replyToId": null,
    "entities": [
        {
            "locale": "en-US",
            "country": "US",
            "platform": "Web",
            "timezone": "America/New_York",
            "type": "clientInfo"
        }
    ],
    "channelData": {
        "tenant": {
            "id": "ffffeeeee-dddd-cccc-bbbb-aaaaaaaaaaaa"
        }
    }
}
```

---

## 3. Key Activity Fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `"message"` for text messages. Other types: `"conversationUpdate"`, `"invoke"`, `"event"`, `"messageReaction"` |
| `id` | string | Unique ID of this activity. Use as `replyToId` in responses. |
| `timestamp` | ISO 8601 | UTC time the message was sent. |
| `serviceUrl` | string | Base URL for the Bot Connector. **Must be used for all reply requests.** Store per-conversation. |
| `channelId` | string | Always `"msteams"` for Teams. |
| `from.id` | string | The sender's channel account ID. Prefixed with `29:` for users. |
| `from.name` | string | Display name of the sender. |
| `from.aadObjectId` | string | The sender's Microsoft Entra (AAD) object ID. Unique per user across tenants. Useful for Graph API calls. |
| `conversation.id` | string | The conversation (thread) ID. Use in POST endpoints to reply or send new messages into this conversation. |
| `conversation.isGroup` | bool | `true` for channel/group chat, `false` for 1:1. |
| `conversation.conversationType` | string | `"personal"` (1:1), `"groupChat"`, or `"channel"`. |
| `conversation.tenantId` | string | The M365 tenant ID. Also available in `channelData.tenant.id`. |
| `recipient.id` | string | The bot's channel account ID. Prefixed with `28:`. |
| `recipient.name` | string | The bot's display name. |
| `text` | string | The plain text content of the message. This is the primary field for text messages. |
| `textFormat` | string | `"plain"` or `"markdown"`. Defaults to `"markdown"` if omitted. |
| `locale` | string | BCP-47 locale tag of the user's Teams client (e.g. `"en-US"`). |
| `replyToId` | string | If set, this Activity is a reply to another Activity with this ID (threaded conversation). |
| `entities` | array | Metadata about the client. The `clientInfo` entity contains platform, timezone, locale. |
| `channelData` | object | Teams-specific data. Always contains `tenant.id`. For channel messages also contains `channel.id`, `team.id`. |
| `attachments` | array | Present when the message includes files, images, or Adaptive Cards. Empty or absent for plain text. |

---

## 4. How to Extract Sender and Text

```
sender name   = activity.from.name
sender userId = activity.from.id           // channel-scoped, bot-specific
sender aadId  = activity.from.aadObjectId  // stable AAD object ID, use for Graph API
message text  = activity.text
conversation  = activity.conversation.id   // use for replying
tenant        = activity.conversation.tenantId  // or activity.channelData.tenant.id
serviceUrl    = activity.serviceUrl        // base URL for sending replies
```

For channel messages, Teams may mention the bot by name. The raw `text` field will contain the mention markup (e.g. `<at>BotName</at> hello`). Strip the mention to get the actual user text; the Bot Framework SDK's `TurnContext.removeMentionText()` helper does this automatically.

---

## 5. Inbound Authentication — Verifying the Request

The Bot Connector service signs every inbound request with a JWT token in the `Authorization` header. Your bot **must** validate this token before processing. Failure to do so exposes the bot to spoofed requests.

### 5.1 Extract and Parse the Token

```
Authorization: Bearer <JWT>
```

### 5.2 Obtain the OpenID Metadata

```http
GET https://login.botframework.com/v1/.well-known/openidconfiguration
```

This returns a JSON document with a `jwks_uri` pointing to the current public signing keys. Cache this document and refresh the keys at least once every 24 hours.

```json
{
    "issuer": "https://api.botframework.com",
    "jwks_uri": "https://login.botframework.com/v1/.well-known/keys",
    "id_token_signing_alg_values_supported": ["RS256"]
}
```

### 5.3 Fetch the Signing Keys

```http
GET https://login.botframework.com/v1/.well-known/keys
```

Returns a JWK Set. Each key also has an `endorsements` array listing channel IDs it endorses (e.g. `"msteams"`).

### 5.4 Validate the JWT Token

All seven of these checks are required:

| # | Check |
|---|-------|
| 1 | Token is in the `Authorization` header with the `Bearer` scheme. |
| 2 | Token is valid JSON conforming to the JWT standard. |
| 3 | `iss` (issuer) claim equals `"https://api.botframework.com"`. |
| 4 | `aud` (audience) claim equals your bot's **Microsoft App ID**. |
| 5 | Token is within its validity period (`nbf` / `exp`). Allow 5 minutes clock-skew. |
| 6 | Token signature is valid against a key from the JWK Set fetched in 5.3, using RS256. |
| 7 | `serviceUrl` claim in the token matches the `serviceUrl` field in the incoming Activity body. |

If the `channelId` in the Activity is `"msteams"`, also verify that the signing key's `endorsements` array includes `"msteams"`. Reject with HTTP 403 if not present.

### 5.5 JWT Claims Example (Connector → Bot)

```json
{
  "aud": "<YOUR-MICROSOFT-APP-ID>",
  "iss": "https://api.botframework.com",
  "nbf": 1481049243,
  "exp": 1481053143,
  "serviceurl": "https://smba.trafficmanager.net/teams/"
}
```

### 5.6 Bot Framework SDK Handles This Automatically

If you use the Bot Framework SDK (C#, JavaScript, or Python), JWT validation is handled for you by the `BotFrameworkAdapter` / `CloudAdapter`. You only need to provide the App ID and password in configuration.

---

## 6. Conversation Types and Context

| `conversationType` | `channelId` | Description |
|--------------------|-------------|-------------|
| `"personal"` | `"msteams"` | 1:1 direct message between user and bot |
| `"groupChat"` | `"msteams"` | Group chat where bot has been added |
| `"channel"` | `"msteams"` | Post in a Teams channel; `channelData.channel.id` and `channelData.team.id` are also present |

For channel messages, the Activity will also contain:

```json
{
    "channelData": {
        "tenant":  { "id": "<tenantId>" },
        "channel": { "id": "<channelId>", "name": "<channelName>" },
        "team":    { "id": "<teamId>", "name": "<teamName>" }
    }
}
```

---

## 7. Activity Types to Handle

| `type` value | When it arrives |
|---|---|
| `"message"` | User sent a text message (or attachment). Main message handler. |
| `"conversationUpdate"` | Bot or member added/removed from conversation. Used to store conversation references for proactive messaging. |
| `"invoke"` | Triggered by Adaptive Card actions, task modules, or message extensions. |
| `"messageReaction"` | User added or removed a reaction to a message. |
| `"event"` | Custom events from Teams client (e.g. meeting lifecycle). |

---

## 8. Storing Conversation References

To enable proactive messaging later, store the conversation reference when you first receive an Activity:

```json
{
    "bot": {
        "id": "<recipient.id>",
        "name": "<recipient.name>"
    },
    "conversation": {
        "id": "<conversation.id>"
    },
    "serviceUrl": "<serviceUrl>"
}
```

Use this reference (plus the tenant ID) to POST to `/v3/conversations/{id}/activities` at a later time without a user-initiated turn.

---

## Sources

- https://learn.microsoft.com/en-us/azure/bot-service/rest-api/bot-framework-rest-connector-send-and-receive-messages
- https://learn.microsoft.com/en-us/microsoftteams/platform/bots/how-to/conversations/send-proactive-messages
- https://learn.microsoft.com/en-us/azure/bot-service/rest-api/bot-framework-rest-connector-authentication
- https://learn.microsoft.com/en-us/microsoftteams/platform/bots/build-a-bot
