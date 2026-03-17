# LINE Messaging API — Supplementary Reference

> **Scope:** CS integration hub — media retrieval, user profiles, webhook reliability,
>   source-type routing, and error handling. Companion to `webhook.md` and `message_api.md`.
> **Primary docs:** https://developers.line.biz/en/reference/messaging-api/
>   https://developers.line.biz/en/docs/messaging-api/

---

## Table of Contents

1. [Media / File Content Retrieval](#1-media--file-content-retrieval)
   - 1.1 [GET /v2/bot/message/{messageId}/content](#11-get-v2botmessagemessageidcontent)
   - 1.2 [Preview / thumbnail endpoint](#12-preview--thumbnail-endpoint)
   - 1.3 [Content types, size limits, and expiry window](#13-content-types-size-limits-and-expiry-window)
   - 1.4 [Auth requirements](#14-auth-requirements)

2. [User & Profile Retrieval APIs](#2-user--profile-retrieval-apis)
   - 2.1 [GET /v2/bot/profile/{userId}](#21-get-v2botprofileuserid)
   - 2.2 [GET /v2/bot/group/{groupId}/member/{userId}](#22-get-v2botgroupgroupidmemberuserid)
   - 2.3 [GET /v2/bot/room/{roomId}/member/{userId}](#23-get-v2botroomroomidmemberuserid)
   - 2.4 [GET /v2/bot/followers/ids — pagination and rate limits](#24-get-v2botfollowersids)
   - 2.5 [Privacy settings — what is redacted or absent](#25-privacy-settings--what-is-redacted-or-absent)

3. [Webhook Reliability & Delivery Behavior](#3-webhook-reliability--delivery-behavior)
   - 3.1 [Expected bot server response](#31-expected-bot-server-response-status-code-body-timeout)
   - 3.2 [Retry / redelivery policy](#32-retry--redelivery-policy)
   - 3.3 [Event batching — multiple events in one POST](#33-event-batching--multiple-events-in-one-post)
   - 3.4 [Ordering guarantees](#34-ordering-guarantees)
   - 3.5 [Idempotency and duplicate delivery (webhookEventId)](#35-idempotency-and-duplicate-delivery-webhookeventid)

4. [Source Type Routing (User / Group / Room)](#4-source-type-routing-user--group--room)
   - 4.1 [Source object schema — all three variants](#41-source-object-schema--all-three-variants)
   - 4.2 [Send endpoint decision table](#42-send-endpoint-decision-table)
   - 4.3 [Can the bot DM a user it met in a group?](#43-can-the-bot-dm-a-user-it-met-in-a-group)
   - 4.4 [Endpoint restrictions by source type](#44-endpoint-restrictions-by-source-type)

5. [Error Response Schema and Handling](#5-error-response-schema-and-handling)
   - 5.1 [Error response body — full JSON structure](#51-error-response-body--full-json-structure)
   - 5.2 [HTTP status codes — all values and meanings](#52-http-status-codes--all-values-and-meanings)
   - 5.3 [Named error codes / message strings](#53-named-error-codes--message-strings)
   - 5.4 [Retry strategy — transient vs terminal errors](#54-retry-strategy--transient-vs-terminal-errors)
   - 5.5 [Webhook error cause codes](#55-webhook-error-cause-codes)

---

## 1. Media / File Content Retrieval

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#get-content

### 1.1 GET /v2/bot/message/{messageId}/content

This endpoint retrieves images, videos, audio, and files sent by users to your LINE Official Account. The `messageId` is obtained from the webhook event object delivered to your bot server when the user sends a message.

**Important domain note:** This endpoint uses `api-data.line.me`, not `api.line.me`. Using the wrong domain will result in a connection failure.

**Precondition:** This endpoint only works when the `contentProvider.type` property of the webhook message object equals `line`. If `contentProvider.type` is `external`, the content is hosted on the sender's own CDN; follow the `contentProvider.originalContentUrl` instead and do not call this endpoint.

**Response format:** The response body is a **binary stream** of the file content. The API does **not** return an HTTP redirect (3xx). The client receives the raw bytes directly in the `200 OK` response body. The `Content-Type` response header indicates the media format (e.g., `image/jpeg`, `video/mp4`, `audio/m4a`, `application/octet-stream` for files).

**Large video/audio files:** For very large video or audio files that have not finished processing on LINE's servers, the API may return `202 Accepted`. The client should wait and retry the request.

#### Full HTTP request example

```
GET /v2/bot/message/325708{messageId}325708/content HTTP/1.1
Host: api-data.line.me
Authorization: Bearer {channel access token}
```

#### Full HTTP response example

```
HTTP/1.1 200 OK
Content-Type: image/jpeg
Content-Length: 143823

<binary image data>
```

#### Path parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `messageId` | String | Yes | Message ID from the webhook event object |

---

### 1.2 Preview / thumbnail endpoint

**No dedicated preview or thumbnail endpoint exists** in the LINE Messaging API for user-sent message content.

The web search of `developers.line.biz` surfaces a note that preview images exist for image and video messages (as a smaller version of the content), but there is no separate REST endpoint at a path like `/v2/bot/message/{messageId}/content/preview` documented in the official API reference. The `content` endpoint retrieves the full binary; LINE may internally compress or resize the content before serving it ("Content sent by users may be transformed internally, such as shrinking"), but this transformation is automatic and not accessible via a separate endpoint.

Sticker images cannot be retrieved via this endpoint at all.

---

### 1.3 Content types, size limits, and expiry window

**Content expiry:** Content sent by users is **automatically deleted after a certain period from when the message was sent**. The official documentation states this but does **not specify the exact duration** — the specific timeframe is not documented in the public API reference. Implementations must retrieve content promptly after receiving the webhook event.

**Request body size limit:** The Messaging API enforces a maximum request body size of **2 MB** (returns `413 Payload Too Large` if exceeded). This limit applies to API call request bodies, not to the downloaded content itself.

**Content transformation:** LINE may internally transform user-sent content before it is available via this endpoint. For example, images may be compressed or resized. The delivered binary may differ from what the user originally sent.

**Supported content types retrieved by this endpoint:**

| Message type | Typical `Content-Type` response header |
|---|---|
| Image | `image/jpeg` |
| Video | `video/mp4` |
| Audio | `audio/m4a` |
| File | Varies; `application/octet-stream` is common |

No file size limits for individual message content downloads are documented in the public API reference beyond the general 2 MB request body limit.

---

### 1.4 Auth requirements

All calls to this endpoint require a valid channel access token in the `Authorization` header:

```
Authorization: Bearer {channel access token}
```

The token must be issued for the channel (LINE Official Account) that received the original message. Using a token from a different channel will result in a `403 Forbidden` response.

Rate limit: The `Get content` endpoint falls under the "Other API endpoints" category at **2,000 requests per second** per channel.

---

## 2. User & Profile Retrieval APIs

### 2.1 GET /v2/bot/profile/{userId}

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#get-profile

Gets the profile of a user who has added your LINE Official Account as a friend, or who sent a message to the account.

#### Full HTTP request example

```
GET /v2/bot/profile/U4af4980629... HTTP/1.1
Host: api.line.me
Authorization: Bearer {channel access token}
```

#### Full JSON response example

```json
{
  "displayName": "LINE taro",
  "userId": "U4af4980629...",
  "pictureUrl": "https://example.com/picture.jpg",
  "statusMessage": "Hello, LINE",
  "language": "en"
}
```

#### Field table

| Field | Type | Always Present | Notes |
|---|---|---|---|
| `displayName` | String | Yes | The user's display name as set in their LINE profile. **This is the field to show in a CS agent UI as the contact's name.** |
| `userId` | String | Yes | Unique user ID. Begins with `U` followed by 32 hex characters. |
| `pictureUrl` | String | No | URL of the user's profile picture. Absent if the user has not set a profile picture. |
| `statusMessage` | String | No | The user's status message. Absent if not set. |
| `language` | String | No | The user's language setting in LINE. Only included when the user has agreed to provide this information. May be absent. |

**Display name for CS agent UI:** Use `displayName` as the primary human-readable identifier for the contact. `userId` is the stable programmatic key for linking records.

---

### 2.2 GET /v2/bot/group/{groupId}/member/{userId}

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#get-group-member-profile

Gets the profile of a member of a group chat. The bot must currently be a member of the group chat. This endpoint returns a profile even if the user has **not** added the bot as a friend, making it more broadly available than the `/v2/bot/profile/{userId}` endpoint for users encountered in group contexts.

#### Full HTTP request example

```
GET /v2/bot/group/Ca56f94637c.../member/U4af4980629... HTTP/1.1
Host: api.line.me
Authorization: Bearer {channel access token}
```

#### Full JSON response example

```json
{
  "displayName": "LINE taro",
  "userId": "U4af4980629...",
  "pictureUrl": "https://example.com/picture.jpg",
  "statusMessage": "Hello, LINE"
}
```

#### Field table

| Field | Type | Always Present | Notes |
|---|---|---|---|
| `displayName` | String | Yes | User's display name. Use this in CS agent UI. |
| `userId` | String | Yes | Unique user ID. |
| `pictureUrl` | String | No | Profile picture URL. Absent if not set. |
| `statusMessage` | String | No | Status message. Absent if not set. |

**Note:** The `language` field is not returned by this endpoint (only by `/v2/bot/profile/{userId}`). The profile returned here reflects what the user has set in LINE; the response does not include friendship status.

---

### 2.3 GET /v2/bot/room/{roomId}/member/{userId}

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#get-room-member-profile

Gets the profile of a member of a multi-person chat. Functionally equivalent to the group member profile endpoint but scoped to multi-person chats (rooms).

**Deprecation context:** From LINE version 10.17.0, multi-person chats were merged into group chats. New chats created on LINE 10.17.0 or later are always group chats. Existing multi-person chats remain accessible via room endpoints, but no new rooms will be created.

#### Full HTTP request example

```
GET /v2/bot/room/Ra8dbf4673c.../member/U4af4980629... HTTP/1.1
Host: api.line.me
Authorization: Bearer {channel access token}
```

#### Full JSON response example

```json
{
  "displayName": "LINE taro",
  "userId": "U4af4980629...",
  "pictureUrl": "https://example.com/picture.jpg",
  "statusMessage": "Hello, LINE"
}
```

#### Field table

| Field | Type | Always Present | Notes |
|---|---|---|---|
| `displayName` | String | Yes | User's display name. Use in CS agent UI. |
| `userId` | String | Yes | Unique user ID. |
| `pictureUrl` | String | No | Profile picture URL. Absent if not set. |
| `statusMessage` | String | No | Status message. Absent if not set. |

---

### 2.4 GET /v2/bot/followers/ids

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#get-follower-ids

Gets a list of user IDs for users who have added your LINE Official Account as a friend. Only users who have consented to share their profile information with LINE Official Accounts are included (see §2.5).

**Account type restriction:** This endpoint is only available on verified or premium accounts. It is **not available** on unverified LINE Official Accounts.

#### Full HTTP request example

```
GET /v2/bot/followers/ids?limit=1000 HTTP/1.1
Host: api.line.me
Authorization: Bearer {channel access token}
```

#### Query parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `start` | String | No | Cursor value from the `next` field of the previous response. Omit for the first page. |
| `limit` | Integer | No | Maximum number of user IDs to return per page. Default and maximum is `1000`. |

#### Full JSON response example

```json
{
  "userIds": [
    "U4af4980629...",
    "U0c29a5bdf...",
    "U30ca6d474..."
  ],
  "next": "yANU9IA..."
}
```

#### Response field table

| Field | Type | Always Present | Notes |
|---|---|---|---|
| `userIds` | Array of String | Yes | List of user IDs for this page. May be empty if no more users exist. |
| `next` | String | No | Cursor for the next page of results. Absent when this is the last page (no more results). |

#### Pagination

This endpoint uses **cursor-based pagination**. To retrieve all follower IDs:

1. Call the endpoint without a `start` parameter to get the first page.
2. If the response includes a `next` field, call the endpoint again with `start={next value}` to get the next page.
3. Continue until the response does not include a `next` field.

#### Rate limit

Falls under "Other API endpoints" — **2,000 requests per second** per channel.

---

### 2.5 Privacy settings — what is redacted or absent

LINE requires user consent for profile information to be shared with LINE Official Accounts. Users who have never used LINE for iOS or LINE for Android (e.g., accounts created exclusively on LINE for PC, which has been discontinued since April 2020) have never had the opportunity to consent.

**Effects of missing consent:**

- The `userId` field is absent from `source` objects in webhook events for non-consenting users in group and room chats.
- Non-consenting users are excluded from the responses of:
  - `GET /v2/bot/followers/ids`
  - `GET /v2/bot/group/{groupId}/member/ids`
  - `GET /v2/bot/room/{roomId}/member/ids`
- The `mention` object within text message webhook events will not include the `userId` of non-consenting mentioned users.
- Calling `GET /v2/bot/profile/{userId}` for a user who has not consented, has not friended the account, or has blocked the account returns `404 Not Found`.

**Fields that may be absent even for consenting users:**

- `pictureUrl` — absent if the user has not set a profile picture
- `statusMessage` — absent if the user has not set a status message
- `language` — absent on the group/room member profile endpoints; conditionally present on `/v2/bot/profile/{userId}` depending on the user's settings

---

## 3. Webhook Reliability & Delivery Behavior

> **Reference:** https://developers.line.biz/en/docs/messaging-api/receiving-messages/
>   https://developers.line.biz/en/reference/messaging-api/#webhooks

### 3.1 Expected bot server response (status code, body, timeout)

**Required status code:** The bot server must return HTTP status code **`200`** after receiving the webhook POST request from the LINE Platform. Any response other than a `2xx` status code is treated as a delivery failure.

**Response body:** The response body does not matter and is ignored by LINE. It is conventional to return an empty body or `{}`.

**Timeout:** The webhook error statistics documentation explicitly states that the `request_timeout` cause code is triggered when "the bot server didn't return a response within **2 seconds**." The bot server must therefore acknowledge the webhook within **2 seconds** of receiving the POST request.

**Recommended pattern:** Because business logic (database writes, downstream API calls, reply message sends) often takes longer than 2 seconds, LINE's documentation recommends processing webhook events **asynchronously**. The bot server should immediately return `200 OK` and queue the event for processing by a background worker.

**Empty event confirmation:** LINE may send a POST with an empty `events` array to confirm the webhook URL is reachable. The bot server must still return `200` in this case:

```json
{
  "destination": "xxxxxxxxxx",
  "events": []
}
```

---

### 3.2 Retry / redelivery policy

LINE **does** automatically retry failed webhook deliveries. This feature is called **Webhook Redelivery**.

**Opt-in required:** Webhook redelivery is **disabled by default** and must be manually enabled. To enable it: LINE Developers Console → channel → Messaging API tab → enable "Webhook redelivery". A caution notice must be acknowledged before enabling.

**Trigger condition:** Redelivery is triggered when the bot server does not return a `2xx` response to the original webhook delivery.

**Redelivery count and interval:** The LINE Platform redelivers failed webhooks "for a certain period of time." The exact number of retry attempts and the interval between retries are **not disclosed** by LINE and are **subject to change without notice**.

**Redelivery flag:** Each redelivered event includes `deliveryContext.isRedelivery: true`. Original (first-delivery) events have `deliveryContext.isRedelivery: false`.

**Ordering impact:** When webhook redelivery is enabled, the order in which events arrive at the bot server can differ significantly from the order in which they occurred. Use `timestamp` to reconstruct chronological order if needed.

---

### 3.3 Event batching — multiple events in one POST

A single webhook POST from the LINE Platform **may contain multiple webhook event objects** in the `events` array. The bot server must be prepared to handle an arbitrary number of events in a single request.

There is no documented maximum number of events per webhook POST.

Events from different users may be batched together. For example, a message event from user A and a follow event from user B may arrive in the same webhook POST:

```json
{
  "destination": "xxxxxxxxxx",
  "events": [
    {
      "type": "message",
      "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
      "source": { "type": "user", "userId": "U80696558e1aa831..." },
      "timestamp": 1625665242211,
      "mode": "active",
      "deliveryContext": { "isRedelivery": false },
      "replyToken": "757913772c4646b784d4b7ce46d12671",
      "message": { "type": "text", "id": "14353798921116", "text": "Hello" }
    },
    {
      "type": "follow",
      "webhookEventId": "01FZ74ASS536FW97EX38NKCZQK",
      "source": { "type": "user", "userId": "Ufc729a925b3abef..." },
      "timestamp": 1625665242214,
      "mode": "active",
      "deliveryContext": { "isRedelivery": false },
      "replyToken": "bb173f4d9cf64aed9d408ab4e36339ad"
    }
  ]
}
```

---

### 3.4 Ordering guarantees

**No guaranteed ordering.** LINE explicitly states: "the order of webhook events you receive can be different from the order the events occurred." This applies especially when webhook redelivery is enabled.

To reconstruct event chronology, use the `timestamp` field present on every webhook event object. The `timestamp` value is the Unix time in milliseconds of when the event occurred, **not** when it was delivered or redelivered.

---

### 3.5 Idempotency and duplicate delivery (webhookEventId)

Each webhook event object includes a `webhookEventId` field that uniquely identifies the event. This ID is a string in **ULID format** (Universally Unique Lexicographically Sortable Identifier).

**Use for deduplication:** When webhook redelivery is enabled, the same logical event may be delivered more than once. Use `webhookEventId` as the deduplication key. Before processing an event, check whether a record with that `webhookEventId` already exists in your store; if it does, skip processing.

**Where it appears:** `webhookEventId` is a top-level field on every event object within the `events` array:

```json
{
  "destination": "xxxxxxxxxx",
  "events": [
    {
      "type": "message",
      "webhookEventId": "01FZ74A0TDDPYRVKNK77XKC3ZR",
      "deliveryContext": {
        "isRedelivery": false
      },
      "timestamp": 1692251666727,
      "source": {
        "type": "user",
        "userId": "U80696558e1aa831..."
      },
      "mode": "active",
      "replyToken": "757913772c4646b784d4b7ce46d12671",
      "message": {
        "type": "text",
        "id": "14353798921116",
        "text": "Hello, world"
      }
    }
  ]
}
```

**`deliveryContext.isRedelivery`** is the companion field: `true` means this is a redelivered copy of a previously attempted delivery. Even when `isRedelivery` is `false`, duplicate delivery is theoretically possible under network fault conditions; therefore, `webhookEventId`-based deduplication is the authoritative check.

---

## 4. Source Type Routing (User / Group / Room)

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#webhook-event-objects
>   https://developers.line.biz/en/docs/messaging-api/group-chats/

### 4.1 Source object schema — all three variants

Every webhook event object contains a `source` property that identifies where the event originated. There are three variants.

#### Source user

```json
"source": {
  "type": "user",
  "userId": "U4af4980629..."
}
```

| Field | Type | Always Present | Notes |
|---|---|---|---|
| `type` | String | Yes | Always `"user"` |
| `userId` | String | Yes | ID of the user who triggered the event |

#### Source group chat

```json
"source": {
  "type": "group",
  "groupId": "Ca56f94637c...",
  "userId": "U4af4980629..."
}
```

| Field | Type | Always Present | Notes |
|---|---|---|---|
| `type` | String | Yes | Always `"group"` |
| `groupId` | String | Yes | ID of the group chat |
| `userId` | String | No | ID of the user who triggered the event. Only included in message events. Only present for users of LINE for iOS/Android who have consented. |

#### Source multi-person chat (room)

```json
"source": {
  "type": "room",
  "roomId": "Ra8dbf4673c...",
  "userId": "U4af4980629..."
}
```

| Field | Type | Always Present | Notes |
|---|---|---|---|
| `type` | String | Yes | Always `"room"` |
| `roomId` | String | Yes | ID of the multi-person chat |
| `userId` | String | No | ID of the user who triggered the event. Only included in message events. Only present for consenting iOS/Android users. |

---

### 4.2 Send endpoint decision table

When sending a proactive (push) message in response to or following a webhook event, use the appropriate ID from the source object as the `to` value:

| `source.type` | `to` field value | Recommended send endpoint |
|---|---|---|
| `user` | `source.userId` | push |
| `group` | `source.groupId` | push |
| `room` | `source.roomId` | push |

**Reply messages** (`POST /v2/bot/message/reply`) use the `replyToken` from the webhook event rather than the `to` field, and are available for all three source types whenever a reply token is present.

**Multicast messages** (`POST /v2/bot/message/multicast`) accept a list of `userId` values but **cannot** be used to send to group chats or multi-person chats. Multicast is user-only.

**Push message endpoint:** `POST https://api.line.me/v2/bot/message/push`

Example push message to a group:

```json
{
  "to": "Ca56f94637c...",
  "messages": [
    {
      "type": "text",
      "text": "Hello, group!"
    }
  ]
}
```

---

### 4.3 Can the bot DM a user it met in a group?

A bot **can** send a private push message to a user it encountered only in a group chat, **provided** the user's `userId` is available from the webhook event's `source.userId` field.

To send a private DM to that user:
- Use `POST /v2/bot/message/push` with `"to": "{source.userId}"` (the user's ID, not the groupId).
- The user does **not** need to have added the bot as a friend for push messages to be delivered.

**Critical constraint:** `source.userId` is only present in group/room source objects for message events, and only for users of LINE for iOS or LINE for Android who have consented to profile information sharing. If `userId` is absent from the source object (due to non-consent), there is no way to address that specific user for a private message.

In practice: if a user interacts in a group chat and their `userId` is captured from the webhook, the bot can push a private message to that user. There is no requirement for a prior one-to-one friendship.

---

### 4.4 Endpoint restrictions by source type

**Group-chat-specific endpoints** (require `groupId`; the bot must be a member of the group):

| Endpoint | Method | Description |
|---|---|---|
| `/v2/bot/group/{groupId}/summary` | GET | Get group chat name and picture |
| `/v2/bot/group/{groupId}/members/count` | GET | Get number of members |
| `/v2/bot/group/{groupId}/members/ids` | GET | Get list of member user IDs |
| `/v2/bot/group/{groupId}/member/{userId}` | GET | Get a single member's profile |
| `/v2/bot/group/{groupId}/leave` | POST | Bot leaves the group |

**Multi-person-chat-specific endpoints** (require `roomId`):

| Endpoint | Method | Description |
|---|---|---|
| `/v2/bot/room/{roomId}/members/count` | GET | Get number of members |
| `/v2/bot/room/{roomId}/members/ids` | GET | Get list of member user IDs |
| `/v2/bot/room/{roomId}/member/{userId}` | GET | Get a single member's profile |
| `/v2/bot/room/{roomId}/leave` | POST | Bot leaves the room |

**User-specific endpoints** (require `userId`; user must be a friend or have consented):

| Endpoint | Method | Description |
|---|---|---|
| `/v2/bot/profile/{userId}` | GET | Get full user profile including `language` |
| `/v2/bot/followers/ids` | GET | List all follower user IDs (verified accounts only) |

**Cross-type endpoints** (work with user, group, or room IDs via the `to` field):

| Endpoint | Method | Description |
|---|---|---|
| `/v2/bot/message/push` | POST | Send push message to user, group, or room |
| `/v2/bot/message/reply` | POST | Reply via reply token (source-agnostic) |

---

## 5. Error Response Schema and Handling

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#error-responses
>   https://developers.line.biz/en/reference/messaging-api/#status-codes

### 5.1 Error response body — full JSON structure

When an API call fails, LINE returns a JSON body in the following structure. The HTTP status code reflects the error category; the JSON body provides details.

```json
{
  "message": "The request body has 2 error(s)",
  "details": [
    {
      "message": "May not be empty",
      "property": "messages[0].text"
    },
    {
      "message": "Must be one of the following values: [text, image, video, audio, location, sticker, template, imagemap]",
      "property": "messages[1].type"
    }
  ]
}
```

#### Error response field table

| Field | Type | Always Present | Notes |
|---|---|---|---|
| `message` | String | Yes | Top-level human-readable error description. See §5.3 for the enumerated message strings. |
| `details` | Array | No | Array of per-field error detail objects. Omitted entirely if empty (not returned as `[]`). |
| `details[].message` | String | No | Description of the specific sub-error. Not included in all error responses. |
| `details[].property` | String | No | The JSON field name or query parameter name in the request that caused the error. Not included in all error responses. |

---

### 5.2 HTTP status codes — all values and meanings

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#status-codes

| Status Code | Name | Meaning |
|---|---|---|
| 200 | OK | Request successful. |
| 400 | Bad Request | Problem with the request. Validation errors in the JSON body, invalid reply token, malformed JSON, or unsupported content type. See `details` array for specifics. |
| 401 | Unauthorized | Valid channel access token not specified in the `Authorization` header. |
| 403 | Forbidden | Not authorized to access the resource. The account or plan does not have permission for the requested operation. |
| 404 | Not Found | The requested resource does not exist. For profile endpoints: the user ID does not exist, the user has not consented to profile sharing, the user has not added the account as a friend, or the user has blocked the account. |
| 409 | Conflict | An API request with the same retry key has already been accepted. Used with the idempotency retry-key mechanism. |
| 410 | Gone | The resource existed but is no longer available (e.g., expired content). |
| 413 | Payload Too Large | The request body exceeds the maximum size of 2 MB. |
| 415 | Unsupported Media Type | The media type of the uploaded file is not supported. |
| 429 | Too Many Requests | Rate limit exceeded, concurrent operation limit exceeded, monthly free message quota exceeded, or additional message allowance exceeded. |
| 500 | Internal Server Error | Error on the LINE internal server. |

---

### 5.3 Named error codes / message strings

The `message` field in the error response body uses one of the following documented string values:

| `message` value | Trigger condition |
|---|---|
| `"The request body has X error(s)"` | Validation error(s) found in the request JSON. `X` is the count. The `details` array provides per-field breakdowns. |
| `"Invalid reply token"` | The `replyToken` used in a reply message send has expired or was already used. |
| `"The property, XXX, in the request body is invalid (line: XXX, column: XXX)"` | An invalid property name was specified. |
| `"The request body could not be parsed as JSON (line: XXX, column: XXX)"` | Malformed JSON in the request body. |
| `"The content type, XXX, is not supported"` | The `Content-Type` of the request is not accepted by the endpoint. |
| `"Authentication failed due to the following reason: XXX"` | Authentication failed. The reason replaces `XXX`. |
| `"Access to this API is not available for your account"` | The account plan or role does not have permission for this endpoint. |
| `"Failed to send messages"` | Message send failed. May indicate that the target `userId` does not exist. |
| `"You have reached your monthly limit."` | Monthly free message quota or additional message allowance has been exceeded. |
| `"The API rate limit has been exceeded. Try again later."` | Per-channel rate limit for this endpoint exceeded. |
| `"Not found"` | Profile information could not be retrieved. User may not exist, may not have consented, may not be a friend, or may have blocked the account. |

---

### 5.4 Retry strategy — transient vs terminal errors

| Error condition | Retry safe? |
|---|---|
| `500 Internal Server Error` | Yes — transient server-side error. Use exponential backoff. |
| `429 Too Many Requests` (rate limit exceeded) | Yes — back off and retry after the rate limit window resets. |
| `429 Too Many Requests` (monthly quota exceeded) | No — retrying will not succeed until the quota resets at the start of the next billing month. |
| `202 Accepted` on content download (content not yet ready) | Yes — wait and retry; content is still being processed. |
| `409 Conflict` (duplicate retry key) | No — the request was already accepted. Do not resend; retrieve the original result using `X-Line-Accepted-Request-Id`. |
| `400 Bad Request` (validation error) | No — fix the request payload before retrying. |
| `400 Bad Request` (invalid reply token) | No — the reply token has expired or been used. Reply tokens are single-use and time-limited. |
| `401 Unauthorized` | No — refresh or reissue the channel access token before retrying. |
| `403 Forbidden` | No — the account/plan does not have permission. Retrying with the same credentials will not help. |
| `404 Not Found` (profile/resource) | No — the user or resource does not exist or consent is not granted. Retrying is futile unless the underlying condition changes (e.g., user re-adds the account). |
| `410 Gone` (content expired) | No — the content has been permanently deleted. |
| `413 Payload Too Large` | No — reduce the request body size before retrying. |
| Network timeout / connection error (no response) | Yes — use exponential backoff with jitter. |

**Retry-key mechanism:** For idempotent message-send operations, include a `Retry-Key` header with a UUID. If a request times out before a response is received, resend with the same `Retry-Key` to avoid duplicate message delivery. If LINE already accepted a request with that key, it returns `409` and sets `X-Line-Accepted-Request-Id` to the original request ID.

---

### 5.5 Webhook error cause codes

> **Reference:** https://developers.line.biz/en/docs/messaging-api/check-webhook-error-statistics/

LINE records webhook delivery errors that can be viewed in the LINE Developers Console (Messaging API tab → Webhook errors tab) and exported as TSV. The error statistics feature must be enabled separately under "Error statistics aggregation."

There are four defined cause codes:

| Cause code | Meaning |
|---|---|
| `could_not_connect` | The LINE Platform attempted to send a webhook to the bot server but could not establish a connection. Typically indicates the bot server is down, unreachable, or the URL is misconfigured. |
| `request_timeout` | The bot server did not return a response within **2 seconds**. The request was treated as failed. |
| `error_status_code` | The bot server returned an HTTP response with a status code outside the `2xx` range. |
| `unclassified` | An error occurred that does not fit the above categories (unknown error). |

**Notes:**
- Error statistics do not include requests made to verify the webhook URL.
- No alert thresholds or automatic notifications are documented; monitoring is manual via the console or TSV export.
- Error count and rate data can be used to detect bot server degradation patterns (e.g., a spike in `request_timeout` indicates the bot is too slow; `could_not_connect` indicates an outage).

---

*Document compiled from:*
- *https://developers.line.biz/en/reference/messaging-api/*
- *https://developers.line.biz/en/docs/messaging-api/receiving-messages/*
- *https://developers.line.biz/en/docs/messaging-api/group-chats/*
- *https://developers.line.biz/en/docs/messaging-api/check-webhook-error-statistics/*
- *https://developers.line.biz/en/docs/messaging-api/user-consent/*
