# LINE Messaging API — Outbound (Sending Messages) Comprehensive Reference

> **Sources used throughout this document:**
> - [Messaging API reference](https://developers.line.biz/en/reference/messaging-api/)
> - [Channel access token](https://developers.line.biz/en/docs/basics/channel-access-token/)
> - [Issue channel access token v2.1 (JWT)](https://developers.line.biz/en/docs/messaging-api/generate-json-web-token/)
> - [Send messages](https://developers.line.biz/en/docs/messaging-api/sending-messages/)
> - [Message types](https://developers.line.biz/en/docs/messaging-api/message-types/)
> - [Build a bot](https://developers.line.biz/en/docs/messaging-api/building-bot/)
> - [Retry failed API requests](https://developers.line.biz/en/docs/messaging-api/retrying-api-request/)
> - [Use quick replies](https://developers.line.biz/en/docs/messaging-api/using-quick-reply/)
> - [Customize icon and display name (sender)](https://developers.line.biz/en/docs/messaging-api/icon-nickname-switch/)
> - [Flex Message elements](https://developers.line.biz/en/docs/messaging-api/flex-message-elements/)
> - [Messaging API pricing](https://developers.line.biz/en/docs/messaging-api/pricing/)
> - [Getting started with Messaging API](https://developers.line.biz/en/docs/messaging-api/getting-started/)

---

## Table of Contents

1. [Setup (Manual Steps)](#1-setup-manual-steps)
2. [Authentication](#2-authentication)
3. [All Send Endpoints](#3-all-send-endpoints)
4. [All Message Object Types](#4-all-message-object-types)

---

## 1. Setup (Manual Steps)

> **Assumes:** A Messaging API channel has already been created via the LINE Official Account Manager and linked to a provider in the LINE Developers Console. See the [getting-started guide](https://developers.line.biz/en/docs/messaging-api/getting-started/) for that prerequisite.

### 1.1 Retrieving a Channel Access Token

All three token types live under the same channel in the LINE Developers Console.

**Navigation path:**
1. Go to [https://developers.line.biz/console/](https://developers.line.biz/console/)
2. Select your **Provider**.
3. Click your **Messaging API channel**.
4. Open the **Messaging API** tab.

From there:

| Token Type | How to Get It |
|---|---|
| **Long-lived** | Scroll to **Channel access token (long-lived)** section → click **Issue**. The token is shown immediately and can be copied. Re-issuing invalidates the previous token. |
| **Short-lived** | Call the API endpoint `POST https://api.line.me/v2/oauth/accessToken` (see §2.2). There is no console button for short-lived tokens. |
| **Stateless** | Call the API endpoint `POST https://api.line.me/oauth2/v3/token` (see §2.4). No console button. |
| **v2.1 (user-specified expiry)** | Register an assertion signing key (public key) in the console under **Assertion Signing Key** → call `POST https://api.line.me/oauth2/v2.1/token` with a JWT (see §2.3). |

### 1.2 Token Type Comparison

| Type | Validity | Max Issued per Channel | Revocable | Recommended Use |
|---|---|---|---|---|
| Long-lived | Never expires | 1 (re-issuing revokes old) | Yes (console or API) | Quick prototyping only; not recommended for production |
| Short-lived | 30 days | 30 (oldest revoked when exceeded) | Yes | Simple server-side bots |
| v2.1 | Up to 30 days (developer sets `token_exp`) | 30 | Yes | Backends that want fine-grained expiry control |
| Stateless | 15 minutes | Unlimited | No (expires naturally) | High-throughput or serverless functions; no lifecycle management needed |

LINE's official recommendation is to use **short-lived**, **v2.1**, or **stateless** tokens. Long-lived tokens are discouraged on security grounds.

### 1.3 Bot Settings That Affect Message Sending

By default, a LINE Official Account has **auto-reply** and **greeting messages** enabled at the platform level. These will fire alongside any messages your webhook/API sends, causing duplicate responses.

**Where to disable them:**

1. In the LINE Developers Console, open your Messaging API channel → **Messaging API** tab.
2. Next to **Auto-reply messages**, click **Edit** — this opens the LINE Official Account Manager.
3. Set **Auto-reply messages** → **Disabled**.
4. Return to the Console and click **Edit** next to **Greeting messages**.
5. Set **Greeting messages** → **Disabled** (or customize as desired).

> Ref: [Build a bot — LINE Developers](https://developers.line.biz/en/docs/messaging-api/building-bot/)

### 1.4 Rate Limits and Quota Considerations

**Rate limits (requests per unit time):**

| Endpoint | Rate Limit |
|---|---|
| Send reply message | 2,000 req/sec |
| Send push message | 2,000 req/sec |
| Send multicast message | 200 req/sec (changed April 23, 2025; was higher) |
| Send broadcast message | 60 req/hour |
| Send narrowcast message | 60 req/hour |

> Rate limit changes: [April 2025 announcement](https://developers.line.biz/en/news/2025/04/23/messaging-api-rate-limit/)

**Monthly message quota:**

Message counts are based on the number of **recipients**, not the number of message objects in a request. Sending 4 message objects to 5 users = 5 messages charged.

- **Reply messages** are free and do not count.
- **Push, multicast, broadcast, narrowcast** count against the monthly quota.
- Blocked users and invalid IDs are not counted.

Check current usage:
```
GET https://api.line.me/v2/bot/message/quota          — total monthly limit
GET https://api.line.me/v2/bot/message/quota/consumption — amount used so far
```

Exceeding the quota returns an error response and messages are not sent. Upgrade plans via the LINE Official Account Manager.

> Ref: [Messaging API pricing](https://developers.line.biz/en/docs/messaging-api/pricing/)

---

## 2. Authentication

### 2.1 How to Authenticate API Requests

Every Messaging API request must include the channel access token as a Bearer token in the `Authorization` header:

```
Authorization: Bearer <channel_access_token>
Content-Type: application/json
```

There is no per-request signing; the token itself is the credential. Protect it accordingly.

### 2.2 Long-Lived and Short-Lived Tokens

**Long-lived token** — issued in the console, never expires. Only one can exist at a time per channel. Re-issuing immediately invalidates the previous one (with up to 24 hours of grace period).

**Short-lived token** — issued via API, valid 30 days, max 30 active per channel.

**Issue short-lived token:**

```
POST https://api.line.me/v2/oauth/accessToken
Content-Type: application/x-www-form-urlencoded
```

Request body (form-encoded):

```
grant_type=client_credentials&client_id={CHANNEL_ID}&client_secret={CHANNEL_SECRET}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `grant_type` | string | Yes | Always `client_credentials` |
| `client_id` | string | Yes | Channel ID (found on Basic settings tab) |
| `client_secret` | string | Yes | Channel secret (found on Basic settings tab) |

Successful response `200 OK`:

```json
{
  "access_token": "W1TeHCgfH2Liwa...",
  "expires_in": 2592000,
  "token_type": "Bearer"
}
```

| Field | Type | Description |
|---|---|---|
| `access_token` | string | The channel access token to use in requests |
| `expires_in` | integer | Seconds until expiry (2,592,000 = 30 days) |
| `token_type` | string | Always `"Bearer"` |

**Revoke short-lived or long-lived token:**

```
POST https://api.line.me/v2/oauth/revoke
Content-Type: application/x-www-form-urlencoded

access_token={TOKEN_TO_REVOKE}
```

**Verify short-lived or long-lived token:**

```
POST https://api.line.me/v2/oauth/verify
Content-Type: application/x-www-form-urlencoded

access_token={TOKEN}
```

### 2.3 Channel Access Token v2.1 (User-Specified Expiry)

This uses an **Assertion Signing Key** (RSA key pair) registered in the console. You generate a JWT signed with your private key, then exchange it for a channel access token.

**Step 1 — Register public key in the console:**
1. Console → Messaging API channel → **Messaging API** tab → **Assertion Signing Key** → **Register**.
2. Upload or paste your RSA public key (minimum 2048-bit, no `kid` property before registration).
3. The console returns a `kid` (key ID); store this.

**Step 2 — Create and sign a JWT:**

JWT header:
```json
{
  "alg": "RS256",
  "typ": "JWT",
  "kid": "536e453c-aa93-4449-8e90-add2608783c6"
}
```

JWT payload:
```json
{
  "iss": "1234567890",
  "sub": "1234567890",
  "aud": "https://api.line.me/",
  "exp": 1559702522,
  "token_exp": 86400
}
```

| Claim | Description |
|---|---|
| `iss` | Your Channel ID |
| `sub` | Your Channel ID (must equal `iss`) |
| `aud` | Always `"https://api.line.me/"` |
| `exp` | JWT expiry Unix timestamp; max 30 minutes from now |
| `token_exp` | Desired channel token validity in seconds; max 2,592,000 (30 days) |

Sign with RS256 using your RSA private key.

**Step 3 — Exchange JWT for a channel access token:**

```
POST https://api.line.me/oauth2/v2.1/token
Content-Type: application/x-www-form-urlencoded
```

Request body:

```
grant_type=client_credentials
&client_assertion_type=urn%3Aietf%3Aparams%3Aoauth%3Aclient-assertion-type%3Ajwt-bearer
&client_assertion={YOUR_SIGNED_JWT}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `grant_type` | string | Yes | Always `client_credentials` |
| `client_assertion_type` | string | Yes | Always `urn:ietf:params:oauth:client-assertion-type:jwt-bearer` (URL-encoded) |
| `client_assertion` | string | Yes | The signed JWT from Step 2 |

Successful response `200 OK`:

```json
{
  "access_token": "eyJhbGciOiJIUz...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "key_id": "sDTOzw5wIfxxxxxxx"
}
```

| Field | Description |
|---|---|
| `access_token` | Channel access token to use in API calls |
| `token_type` | Always `"Bearer"` |
| `expires_in` | Seconds until expiry (matches `token_exp` from JWT) |
| `key_id` | Unique key ID; store alongside the token for revocation |

**Revoke v2.1 token:**
```
POST https://api.line.me/oauth2/v2.1/revoke
Content-Type: application/x-www-form-urlencoded

client_id={CHANNEL_ID}&client_secret={CHANNEL_SECRET}&access_token={TOKEN}
```

**Get all valid v2.1 token key IDs:**
```
GET https://api.line.me/oauth2/v2.1/tokens/kid
  ?client_assertion_type=urn%3Aietf%3Aparams%3Aoauth%3Aclient-assertion-type%3Ajwt-bearer
  &client_assertion={JWT}
```

> Ref: [Issue channel access token v2.1](https://developers.line.biz/en/docs/messaging-api/generate-json-web-token/)

### 2.4 Stateless Channel Access Token

Stateless tokens are valid for **15 minutes** and cannot be revoked. They require the same JWT-based assertion as v2.1 tokens but are issued at a different endpoint and carry no lifecycle management burden — simply issue a fresh one per request or per short-lived operation.

```
POST https://api.line.me/oauth2/v3/token
Content-Type: application/x-www-form-urlencoded
```

Request body:

```
grant_type=client_credentials
&client_assertion_type=urn%3Aietf%3Aparams%3Aoauth%3Aclient-assertion-type%3Ajwt-bearer
&client_assertion={YOUR_SIGNED_JWT}
&client_id={CHANNEL_ID}
&client_secret={CHANNEL_SECRET}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `grant_type` | string | Yes | Always `client_credentials` |
| `client_assertion_type` | string | Yes | `urn:ietf:params:oauth:client-assertion-type:jwt-bearer` (URL-encoded) |
| `client_assertion` | string | Yes | Signed JWT (same structure as v2.1) |
| `client_id` | string | Yes | Channel ID |
| `client_secret` | string | Yes | Channel secret |

Successful response `200 OK`:

```json
{
  "access_token": "eyJhbGciOiJIUz...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

| Field | Description |
|---|---|
| `access_token` | Stateless token, valid 15 minutes |
| `token_type` | Always `"Bearer"` |
| `expires_in` | Always 900 (15 minutes) |

> Note: No `key_id` is returned. Stateless tokens cannot be revoked.

### 2.5 Best Practices for Backend Integration

- **Never expose** the channel access token or channel secret in client-side code.
- For **serverless / per-request** architectures: use stateless tokens — generate one at the start of each function invocation.
- For **long-running servers**: use v2.1 or short-lived tokens, refresh before expiry, store in a secrets manager (e.g., AWS Secrets Manager, HashiCorp Vault).
- For **prototyping only**: long-lived tokens are acceptable; rotate them before production.
- Use the `X-Line-Retry-Key` header on push/multicast/broadcast/narrowcast to prevent duplicate sends on network retry (see §3).

---

## 3. All Send Endpoints

**Base URL:** `https://api.line.me`

**Common headers for all send endpoints:**
```
Authorization: Bearer <channel_access_token>
Content-Type: application/json
```

**Optional retry header (push, multicast, broadcast, narrowcast only):**
```
X-Line-Retry-Key: 123e4567-e89b-12d3-a456-426614174000
```
The retry key is a UUID you generate. If you retry the exact same request with the same key within 24 hours, LINE executes it only once and returns `409 Conflict` with the original request ID in `x-line-accepted-request-id`. Omitting the header on the first request means the request can never be retried safely.

### 3.1 Reply Message

**Use when:** You need to respond to an incoming webhook event (message, postback, follow, etc.). The `replyToken` from the webhook event is required and is single-use.

```
POST https://api.line.me/v2/bot/message/reply
```

**Full request body:**

```json
{
  "replyToken": "nHuyWiB7yP5Zw52FIkcQobQuGDXCTA",
  "messages": [
    {
      "type": "text",
      "text": "Hello! How can I help you today?"
    }
  ],
  "notificationDisabled": false
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `replyToken` | string | Yes | One-time token from the webhook event. Valid for a short window (~60 s). |
| `messages` | array | Yes | 1–5 message objects. Each becomes a separate chat bubble. |
| `notificationDisabled` | boolean | No | If `true`, the message is sent silently (no push notification on the recipient's device). Default: `false`. |

**Successful response `200 OK`:**

```json
{
  "sentMessages": [
    {
      "id": "461704574915325955",
      "quoteToken": "I042asfA8..."
    }
  ]
}
```

| Field | Description |
|---|---|
| `sentMessages[].id` | LINE-assigned message ID |
| `sentMessages[].quoteToken` | Token usable to quote this message in a subsequent send |

**Error responses:**

| Status | Meaning |
|---|---|
| `400 Bad Request` | Malformed JSON, missing required field, invalid `replyToken`, or `replyToken` already used |
| `403 Forbidden` | Token lacks permission or wrong channel |
| `429 Too Many Requests` | Rate limit exceeded |

**Key notes:**
- Reply messages are **free** and do not count against the monthly quota.
- The `replyToken` expires quickly — call this endpoint immediately upon receiving the webhook event.
- Does **not** support `X-Line-Retry-Key`.

### 3.2 Push Message

**Use when:** You want to proactively send a message to a single user, group, or multi-person chat at any time.

```
POST https://api.line.me/v2/bot/message/push
```

**Full request body:**

```json
{
  "to": "U4af4980629f1c56b40adf3cfc6b1fc8a",
  "messages": [
    {
      "type": "text",
      "text": "Your appointment is confirmed for tomorrow at 10:00 AM."
    }
  ],
  "notificationDisabled": false,
  "customAggregationUnits": ["appointment_reminders"]
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `to` | string | Yes | Recipient ID: a user ID (`U…`), group ID (`C…`), or multi-person chat room ID (`R…`). |
| `messages` | array | Yes | 1–5 message objects. |
| `notificationDisabled` | boolean | No | Suppress device push notification. Default: `false`. |
| `customAggregationUnits` | array of string | No | Unit name(s) for message statistics aggregation. Max 1 element. Each string max 30 alphanumeric+underscore characters. |

**Successful response `200 OK`:**

```json
{
  "sentMessages": [
    {
      "id": "461704574915325952",
      "quoteToken": "q3Plxr4AgMd..."
    }
  ]
}
```

**Error responses:**

| Status | Meaning |
|---|---|
| `400` | Malformed body or invalid `to` value |
| `403` | Insufficient permissions |
| `409` | `X-Line-Retry-Key` already accepted (duplicate suppressed) |
| `429` | Rate limit or monthly quota exceeded |

**Key notes:**
- Counts against the monthly messaging quota.
- Supports `X-Line-Retry-Key` header for safe retries.
- Cannot send to blocked users; those sends are silently dropped and not counted against quota.

### 3.3 Multicast Message

**Use when:** You want to send the same message to multiple individual users in a single API call.

```
POST https://api.line.me/v2/bot/message/multicast
```

**Full request body:**

```json
{
  "to": [
    "U4af4980629f1c56b40adf3cfc6b1fc8a",
    "U1d8994c10d6fd2b3458b7e8b3e6a24d1",
    "Ucc4c0d4bf7f2b5f3be87a4a6c7d72e8d"
  ],
  "messages": [
    {
      "type": "text",
      "text": "Reminder: the clinic is closed this Sunday."
    }
  ],
  "notificationDisabled": false,
  "customAggregationUnits": ["clinic_broadcasts"]
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `to` | array of string | Yes | List of user IDs. Max 500 per request. Must be user IDs only (not group/room IDs). |
| `messages` | array | Yes | 1–5 message objects. |
| `notificationDisabled` | boolean | No | Suppress device notifications. Default: `false`. |
| `customAggregationUnits` | array of string | No | Statistics unit name. Max 1 element. |

**Successful response `200 OK`:**

```json
{}
```

Multicast returns an empty JSON object on success (unlike reply/push, no `sentMessages` array).

**Error responses:** Same codes as push (400, 403, 409, 429).

**Key notes:**
- Only user IDs are valid in `to`. Group IDs and room IDs are not supported.
- Max 500 user IDs per request. To send to more users, paginate into multiple requests.
- Quota cost = number of valid recipients reached.
- Supports `X-Line-Retry-Key`.

### 3.4 Broadcast Message

**Use when:** You want to send the same message to **all users** who have added your LINE Official Account as a friend.

```
POST https://api.line.me/v2/bot/message/broadcast
```

**Full request body:**

```json
{
  "messages": [
    {
      "type": "text",
      "text": "We have a special offer this weekend — check our app for details!"
    }
  ],
  "notificationDisabled": false
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `messages` | array | Yes | 1–5 message objects. |
| `notificationDisabled` | boolean | No | Suppress device notifications. Default: `false`. |

**Successful response `200 OK`:**

```json
{}
```

Returns an empty JSON object on success.

**Error responses:** 400, 403, 409, 429 as above.

**Key notes:**
- No `to` field — sends to the entire friend list automatically.
- Subject to the 60 req/hour rate limit.
- Quota cost = total active friends reached.
- Supports `X-Line-Retry-Key`.
- Does **not** send to users who have blocked the account.

### 3.5 Narrowcast Message

**Use when:** You want to send to a **segmented subset** of your audience, filtered by demographics (age, gender, OS, region) or by predefined audience groups, using logical operators.

```
POST https://api.line.me/v2/bot/message/narrowcast
```

Narrowcast is **processed asynchronously**. The endpoint returns immediately with a `requestId`; use the progress endpoint to track delivery:
```
GET https://api.line.me/v2/bot/message/progress/narrowcast?requestId={requestId}
```

**Full request body (all fields):**

```json
{
  "messages": [
    {
      "type": "text",
      "text": "Special offer for women aged 20-35 in Tokyo!"
    }
  ],
  "recipient": {
    "type": "operator",
    "and": [
      {
        "type": "audience",
        "audienceGroupId": 5614991017776
      },
      {
        "type": "operator",
        "not": {
          "type": "redelivery",
          "requestId": "5b59509c-c57b-11e9-aa8c-2a2ae2dbcce4"
        }
      }
    ]
  },
  "filter": {
    "demographic": {
      "type": "operator",
      "and": [
        {
          "type": "gender",
          "oneOf": ["female"]
        },
        {
          "type": "age",
          "gte": "age_20",
          "lt": "age_35"
        },
        {
          "type": "appType",
          "oneOf": ["ios", "android"]
        },
        {
          "type": "area",
          "oneOf": ["jp_13"]
        },
        {
          "type": "subscriptionPeriod",
          "gte": "day_7",
          "lt": "day_30"
        }
      ]
    }
  },
  "limit": {
    "max": 10000,
    "upToRemainingQuota": false,
    "forbidPartialDelivery": false
  },
  "notificationDisabled": false
}
```

**Top-level fields:**

| Field | Type | Required | Description |
|---|---|---|---|
| `messages` | array | Yes | 1–5 message objects. |
| `recipient` | object | No | Audience targeting via audience groups or redelivery objects combined with `operator` objects. If omitted, targets all friends. |
| `filter` | object | No | Demographic filter. If omitted, no demographic filtering. |
| `filter.demographic` | object | No | Demographic filter object (see below). |
| `limit` | object | No | Controls how many recipients can receive the message. |
| `limit.max` | integer | No | Maximum number of recipients. Recipients beyond this cap are chosen at random. |
| `limit.upToRemainingQuota` | boolean | No | If `true`, sends up to the remaining monthly quota and no more. Default: `false`. |
| `limit.forbidPartialDelivery` | boolean | No | If `true` and the cap cannot be met (e.g. too few eligible recipients), the entire delivery is cancelled. Default: `false`. |
| `notificationDisabled` | boolean | No | Suppress push notification. Default: `false`. |

**`recipient` object — audience targeting:**

```json
{
  "type": "audience",
  "audienceGroupId": 5614991017776
}
```

```json
{
  "type": "redelivery",
  "requestId": "5b59509c-c57b-11e9-aa8c-2a2ae2dbcce4"
}
```

Combine with `operator`:

```json
{
  "type": "operator",
  "and": [ ... ],
  "or":  [ ... ],
  "not": { ... }
}
```

| Recipient Type | Fields | Description |
|---|---|---|
| `audience` | `audienceGroupId` (integer) | Targets a specific audience group created via the audience API |
| `redelivery` | `requestId` (string) | Retargets users who received a previous narrowcast (by its request ID) |
| `operator` | `and` (array), `or` (array), `not` (object) | Boolean combinator; nest to express complex logic |

**`filter.demographic` object — demographic filter types:**

| Type | Fields | Values |
|---|---|---|
| `gender` | `oneOf` (array of string) | `"male"`, `"female"` |
| `age` | `gte` (string), `lt` (string) | `"age_15"`, `"age_20"`, `"age_25"`, `"age_30"`, `"age_35"`, `"age_40"`, `"age_45"`, `"age_50"`, `"age_55"`, `"age_60"`, `"age_65"`, `"age_70"` |
| `appType` | `oneOf` (array of string) | `"ios"`, `"android"` |
| `area` | `oneOf` (array of string) | Region codes, e.g. `"jp_13"` (Tokyo), `"jp_23"` (Aichi), `"jp_05"` (Akita), etc. |
| `subscriptionPeriod` | `gte` (string), `lt` (string) | `"day_7"`, `"day_30"`, `"day_90"`, `"day_180"`, `"day_365"` |
| `operator` | `and`, `or`, `not` | Boolean combinator for demographic conditions |

**Successful response `200 OK`:**

```json
{}
```

The `requestId` for tracking is returned in the response **header** `x-line-request-id`.

**Check progress:**
```
GET https://api.line.me/v2/bot/message/progress/narrowcast?requestId={id}
Authorization: Bearer {token}
```

Response:
```json
{
  "phase": "succeeded",
  "successCount": 9852,
  "failureCount": 12,
  "errorCode": 0,
  "acceptedTime": "2020-01-28T01:15:37Z",
  "completedTime": "2020-01-28T01:22:14Z"
}
```

**Key notes:**
- Asynchronous — do not assume delivery is complete when the API returns 200.
- Subject to a minimum recipient threshold (LINE does not publish the exact number).
- Rate limit: 60 req/hour.
- Supports `X-Line-Retry-Key`.

---

## 4. All Message Object Types

All message objects can appear in the `messages` array of any send endpoint. Every message type supports the optional `quickReply` and `sender` top-level properties (documented at the end of this section).

### 4.1 Text Message

Plain text with optional LINE emoji substitutions and optional quote.

```json
{
  "type": "text",
  "text": "Hello $ Have a great day!",
  "emojis": [
    {
      "index": 6,
      "productId": "5ac1bfd5040ab15980c9b435",
      "emojiId": "001"
    }
  ],
  "quoteToken": "IHQxOm9yaWdpbmFsLW1lc3NhZ2UtaWQ..."
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"text"` |
| `text` | string | Yes | Message content. Max 5,000 characters. Use `$` as a placeholder for each emoji in the `emojis` array. Newlines with `\n` render as line breaks. |
| `emojis` | array | No | LINE emoji substitutions. Each `$` in `text` is replaced in order by the corresponding emoji object. |
| `emojis[].index` | integer | Yes (if emojis) | Zero-based character index of the `$` placeholder in `text`. |
| `emojis[].productId` | string | Yes (if emojis) | Emoji product ID from the [LINE emoji list](https://developers.line.biz/en/docs/messaging-api/emoji-list/). |
| `emojis[].emojiId` | string | Yes (if emojis) | Specific emoji ID within the product. |
| `quoteToken` | string | No | Quote token of a previous message (obtained from `sentMessages[].quoteToken` or from a received webhook event's `message.quoteToken`). Renders a quote bubble above this message. |

### 4.2 Image Message

```json
{
  "type": "image",
  "originalContentUrl": "https://example.com/images/photo.jpg",
  "previewImageUrl": "https://example.com/images/photo_preview.jpg"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"image"` |
| `originalContentUrl` | string | Yes | HTTPS URL of the full-resolution image. Max 10 MB. Supported formats: JPEG, PNG. Max dimensions: 4096×4096 px. |
| `previewImageUrl` | string | Yes | HTTPS URL of the preview/thumbnail image shown in chat. Max 1 MB. Recommended: 240×240 px JPEG. |

**Notes:** Both URLs must use HTTPS with a valid CA certificate. URLs are fetched by LINE servers at delivery time; they must remain accessible.

### 4.3 Video Message

```json
{
  "type": "video",
  "originalContentUrl": "https://example.com/videos/clip.mp4",
  "previewImageUrl": "https://example.com/images/clip_thumb.jpg",
  "trackingId": "track-video-001"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"video"` |
| `originalContentUrl` | string | Yes | HTTPS URL of the MP4 video file. Max 200 MB. Max duration: 1 minute. |
| `previewImageUrl` | string | Yes | HTTPS URL of a thumbnail image displayed before playback. Max 1 MB. JPEG recommended. |
| `trackingId` | string | No | ID used to correlate video view events in webhooks (max 100 characters, alphanumeric + `_` + `-`). When specified, LINE sends a `videoPlayComplete` webhook event when the user finishes watching. |

### 4.4 Audio Message

```json
{
  "type": "audio",
  "originalContentUrl": "https://example.com/audio/greeting.m4a",
  "duration": 60000
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"audio"` |
| `originalContentUrl` | string | Yes | HTTPS URL of the audio file. Supported format: M4A (AAC). Max 200 MB. Max duration: 1 minute. |
| `duration` | integer | Yes | Length of the audio in milliseconds. Used to display the duration indicator in the chat UI. |

### 4.5 File Message — Not Sendable via API

The LINE Messaging API **does not support sending file messages**. The `"file"` type exists only in **incoming webhook events** (when a user sends a file to the bot). You can retrieve the file content using the content retrieval endpoint:

```
GET https://api-data.line.me/v2/bot/message/{messageId}/content
```

If you need to share a file with a user via the bot, the workaround is to upload the file to your own HTTPS storage and send it as a `"text"` message containing the URL, or use a `"flex"` message with a URI button linking to the download.

### 4.6 Location Message

```json
{
  "type": "location",
  "title": "Our Clinic",
  "address": "1-6-1 Marunouchi, Chiyoda-ku, Tokyo 100-0005",
  "latitude": 35.67966,
  "longitude": 139.76380
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"location"` |
| `title` | string | Yes | Location label (e.g., place name). Max 100 characters. |
| `address` | string | Yes | Street address. Max 300 characters. |
| `latitude` | number (float) | Yes | Geographic latitude. |
| `longitude` | number (float) | Yes | Geographic longitude. |

The message renders as a map pin in the LINE chat. Tapping opens a maps application.

### 4.7 Sticker Message

```json
{
  "type": "sticker",
  "packageId": "1",
  "stickerId": "1"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"sticker"` |
| `packageId` | string | Yes | ID of the sticker package. See the [sendable sticker list](https://developers.line.biz/en/docs/messaging-api/sticker-list/). |
| `stickerId` | string | Yes | ID of the specific sticker within the package. |

**Notes:** Only stickers explicitly listed in the [Messaging API sticker list](https://developers.line.biz/en/docs/messaging-api/sticker-list/) can be sent. Attempting to send unlisted stickers results in an error. The sticker list PDF is linked from that page.

### 4.8 Imagemap Message

An image with multiple independently tappable regions. Each region can trigger a URI action (open URL) or a message action (send text on behalf of the user). Optionally, a video can play over the image.

```json
{
  "type": "imagemap",
  "baseUrl": "https://example.com/images/imagemap/",
  "altText": "Tap an area to learn more",
  "baseSize": {
    "width": 1040,
    "height": 1040
  },
  "video": {
    "originalContentUrl": "https://example.com/videos/promo.mp4",
    "previewImageUrl": "https://example.com/images/promo_thumb.jpg",
    "area": {
      "x": 0,
      "y": 0,
      "width": 1040,
      "height": 585
    },
    "externalLink": {
      "linkUri": "https://example.com/promo",
      "label": "See more"
    }
  },
  "actions": [
    {
      "type": "uri",
      "label": "Open website",
      "linkUri": "https://example.com/product-a",
      "area": {
        "x": 0,
        "y": 0,
        "width": 520,
        "height": 1040
      }
    },
    {
      "type": "message",
      "label": "Learn more",
      "text": "Tell me more about Product B",
      "area": {
        "x": 520,
        "y": 0,
        "width": 520,
        "height": 1040
      }
    }
  ]
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"imagemap"` |
| `baseUrl` | string | Yes | Base HTTPS URL. LINE appends image size suffixes (`/240`, `/300`, `/460`, `/700`, `/1040`) to this URL to serve different device resolutions. All five sizes must be available. |
| `altText` | string | Yes | Alternative text shown in notifications and non-rendering contexts. Max 400 characters. |
| `baseSize.width` | integer | Yes | Width of the base image in pixels. Always `1040`. |
| `baseSize.height` | integer | Yes | Height of the base image in pixels. Must match the actual image height at 1040px width. |
| `video` | object | No | Video overlay on top of the image. |
| `video.originalContentUrl` | string | Yes (if video) | HTTPS URL of the MP4 video. Max 200 MB, max 1 minute. |
| `video.previewImageUrl` | string | Yes (if video) | HTTPS thumbnail URL shown before playback. |
| `video.area` | object | Yes (if video) | Coordinates and dimensions `{x, y, width, height}` within the 1040-unit grid where the video is displayed. |
| `video.externalLink.linkUri` | string | No | URL opened after video finishes playing. |
| `video.externalLink.label` | string | No | Label text shown on the post-playback button. Max 30 characters. |
| `actions` | array | Yes | Array of tappable area action objects (URI or message). |
| `actions[].type` | string | Yes | `"uri"` or `"message"` |
| `actions[].label` | string | Yes | Accessibility label for the area. Max 50 characters. |
| `actions[].linkUri` | string | Yes (URI type) | URL to open when tapped. |
| `actions[].text` | string | Yes (message type) | Text sent as the user's message when tapped. Max 400 characters. |
| `actions[].area` | object | Yes | `{x, y, width, height}` in the 1040-unit coordinate grid. |

### 4.9 Flex Message

Flex Messages use CSS Flexbox layout principles and offer the richest visual customization. The top-level message object wraps either a `BubbleContainer` (single bubble) or a `CarouselContainer` (scrollable multiple bubbles).

```json
{
  "type": "flex",
  "altText": "Appointment confirmation",
  "contents": {
    "type": "bubble",
    "size": "mega",
    "header": {
      "type": "box",
      "layout": "vertical",
      "contents": [
        {
          "type": "text",
          "text": "Appointment Confirmed",
          "weight": "bold",
          "color": "#FFFFFF",
          "size": "xl"
        }
      ],
      "backgroundColor": "#27ACB2"
    },
    "hero": {
      "type": "image",
      "url": "https://example.com/images/clinic.jpg",
      "size": "full",
      "aspectRatio": "3:2",
      "aspectMode": "cover"
    },
    "body": {
      "type": "box",
      "layout": "vertical",
      "spacing": "md",
      "contents": [
        {
          "type": "text",
          "text": "Dr. Smith — Tuesday, 10:00 AM",
          "wrap": true,
          "size": "md"
        },
        {
          "type": "separator"
        },
        {
          "type": "box",
          "layout": "horizontal",
          "contents": [
            {
              "type": "text",
              "text": "Location:",
              "size": "sm",
              "color": "#aaaaaa",
              "flex": 1
            },
            {
              "type": "text",
              "text": "3F, Main Building",
              "size": "sm",
              "flex": 2,
              "wrap": true
            }
          ]
        }
      ]
    },
    "footer": {
      "type": "box",
      "layout": "vertical",
      "spacing": "sm",
      "contents": [
        {
          "type": "button",
          "style": "primary",
          "action": {
            "type": "uri",
            "label": "Get Directions",
            "uri": "https://maps.example.com/clinic"
          }
        },
        {
          "type": "button",
          "style": "secondary",
          "action": {
            "type": "postback",
            "label": "Cancel Appointment",
            "data": "action=cancel&apptId=12345"
          }
        }
      ]
    }
  }
}
```

**Flex message top-level fields:**

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"flex"` |
| `altText` | string | Yes | Fallback text shown in notifications and on devices that don't support Flex Messages. Max 400 characters. |
| `contents` | object | Yes | A `BubbleContainer` or `CarouselContainer` object. |

**BubbleContainer fields:**

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"bubble"` |
| `size` | string | No | Bubble width: `"nano"`, `"micro"`, `"kilo"`, `"mega"` (default), `"giga"` |
| `direction` | string | No | Text direction: `"ltr"` (default) or `"rtl"` |
| `header` | Box component | No | Top section. Typically used for title or category label. |
| `hero` | Image/Video/Box component | No | Large visual area below the header. |
| `body` | Box component | No | Main content area. |
| `footer` | Box component | No | Bottom section, typically buttons. |
| `styles` | object | No | Background colors for header/hero/body/footer sections via `{ header: { backgroundColor: "#..." }, ... }` |
| `action` | Action object | No | Action triggered by tapping anywhere on the bubble (whole-bubble tap). |

**CarouselContainer fields:**

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"carousel"` |
| `contents` | array | Yes | Array of `BubbleContainer` objects. Max 12 bubbles. All bubbles must have the same height. |

**Key Flex components:**

| Component | `type` | Key Fields |
|---|---|---|
| Box | `"box"` | `layout` (`horizontal`/`vertical`/`baseline`), `contents` (array), `spacing`, `padding`, `backgroundColor`, `cornerRadius` |
| Text | `"text"` | `text`, `size`, `color`, `weight` (`bold`), `align`, `wrap`, `decoration`, `contents` (for spans) |
| Image | `"image"` | `url`, `size`, `aspectRatio`, `aspectMode` (`cover`/`fit`), `action` |
| Button | `"button"` | `action`, `style` (`primary`/`secondary`/`link`), `color`, `height` |
| Video | `"video"` | `url`, `previewUrl`, `aspectRatio`, `altContent` (fallback image) |
| Separator | `"separator"` | `margin`, `color` |
| Icon | `"icon"` | `url`, `size` — only valid inside a `baseline` box |
| Span | `"span"` | `text`, `color`, `size`, `weight`, `decoration` — used inside text `contents` for mixed styling |

> Design and preview Flex Messages interactively at [Flex Message Simulator](https://developers.line.biz/flex-simulator/).

### 4.10 Template Message

Template messages use pre-built layouts. They are simpler to implement than Flex Messages but offer less visual customization.

```json
{
  "type": "template",
  "altText": "Please select an option",
  "template": { ... }
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"template"` |
| `altText` | string | Yes | Text shown in notifications and on unsupported devices. Max 400 characters. |
| `template` | object | Yes | One of the four template type objects below. |

#### 4.10.1 Buttons Template

A single image (optional), title (optional), description text, and up to 4 action buttons.

```json
{
  "type": "buttons",
  "thumbnailImageUrl": "https://example.com/images/product.jpg",
  "imageAspectRatio": "rectangle",
  "imageSize": "cover",
  "imageBackgroundColor": "#FFFFFF",
  "title": "Product A",
  "text": "Choose an action below",
  "defaultAction": {
    "type": "uri",
    "label": "View product page",
    "uri": "https://example.com/product-a"
  },
  "actions": [
    {
      "type": "postback",
      "label": "Buy Now",
      "data": "action=buy&productId=A"
    },
    {
      "type": "uri",
      "label": "View Details",
      "uri": "https://example.com/product-a"
    },
    {
      "type": "message",
      "label": "Ask a Question",
      "text": "I have a question about Product A"
    }
  ]
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"buttons"` |
| `thumbnailImageUrl` | string | No | HTTPS image URL. JPEG or PNG. Max 1 MB. |
| `imageAspectRatio` | string | No | `"rectangle"` (1.51:1, default) or `"square"` (1:1) |
| `imageSize` | string | No | `"cover"` (default, fills area) or `"contain"` (letterboxed) |
| `imageBackgroundColor` | string | No | Hex color for the letterbox area when `imageSize` is `"contain"`. Default: `"#FFFFFF"`. |
| `title` | string | No | Bold title line. Max 40 characters. |
| `text` | string | Yes | Description text. Max 160 characters (60 if `thumbnailImageUrl` or `title` is set). |
| `defaultAction` | Action object | No | Action triggered by tapping the image, title, or text area (not a button). |
| `actions` | array | Yes | 1–4 action objects displayed as buttons. |

#### 4.10.2 Confirm Template

Two-button yes/no style template.

```json
{
  "type": "confirm",
  "text": "Are you sure you want to cancel your appointment?",
  "actions": [
    {
      "type": "postback",
      "label": "Yes, cancel",
      "data": "action=cancel_confirmed"
    },
    {
      "type": "message",
      "label": "No, keep it",
      "text": "Keep my appointment"
    }
  ]
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"confirm"` |
| `text` | string | Yes | Question or message text. Max 240 characters. |
| `actions` | array | Yes | Exactly 2 action objects. |

#### 4.10.3 Carousel Template

Multiple columns (cards) that users scroll through horizontally. Each column can have its own image, title, text, and up to 3 action buttons.

```json
{
  "type": "carousel",
  "imageAspectRatio": "rectangle",
  "imageSize": "cover",
  "columns": [
    {
      "thumbnailImageUrl": "https://example.com/images/service-a.jpg",
      "imageBackgroundColor": "#FFFFFF",
      "title": "Service A",
      "text": "Book a consultation",
      "defaultAction": {
        "type": "uri",
        "label": "Learn more",
        "uri": "https://example.com/service-a"
      },
      "actions": [
        {
          "type": "postback",
          "label": "Book Now",
          "data": "action=book&service=A"
        }
      ]
    },
    {
      "thumbnailImageUrl": "https://example.com/images/service-b.jpg",
      "title": "Service B",
      "text": "Our premium package",
      "actions": [
        {
          "type": "uri",
          "label": "View Package",
          "uri": "https://example.com/service-b"
        }
      ]
    }
  ]
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"carousel"` |
| `imageAspectRatio` | string | No | `"rectangle"` (default) or `"square"` — applies to all columns |
| `imageSize` | string | No | `"cover"` (default) or `"contain"` — applies to all columns |
| `columns` | array | Yes | 1–10 column objects. |
| `columns[].thumbnailImageUrl` | string | No | HTTPS image URL. Max 1 MB. |
| `columns[].imageBackgroundColor` | string | No | Background hex color for `"contain"` mode. |
| `columns[].title` | string | No | Bold title. Max 40 characters. |
| `columns[].text` | string | Yes | Description. Max 120 characters (60 if title/image set). |
| `columns[].defaultAction` | Action object | No | Action when tapping the image/title/text area. |
| `columns[].actions` | array | Yes | 1–3 action objects per column (all columns must have the same number of buttons). |

#### 4.10.4 Image Carousel Template

A carousel of images only, each with a single action. No text or title per card.

```json
{
  "type": "image_carousel",
  "columns": [
    {
      "imageUrl": "https://example.com/images/banner-1.jpg",
      "action": {
        "type": "uri",
        "label": "Visit page",
        "uri": "https://example.com/page-1"
      }
    },
    {
      "imageUrl": "https://example.com/images/banner-2.jpg",
      "action": {
        "type": "postback",
        "label": "Select",
        "data": "action=select&item=2"
      }
    }
  ]
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Always `"image_carousel"` |
| `columns` | array | Yes | 1–10 column objects. |
| `columns[].imageUrl` | string | Yes | HTTPS image URL. JPEG or PNG. Max 1 MB. Recommended: square images. |
| `columns[].action` | Action object | Yes | Single action triggered by tapping the image. |

---

### 4.11 The `quickReply` Object

The `quickReply` property can be appended to **any message type**. Quick reply buttons appear as a horizontally scrollable row at the bottom of the chat input area. Tapping a button dismisses the row and triggers the associated action.

```json
{
  "type": "text",
  "text": "How would you like to proceed?",
  "quickReply": {
    "items": [
      {
        "type": "action",
        "imageUrl": "https://example.com/icons/book.png",
        "action": {
          "type": "postback",
          "label": "Book Appointment",
          "data": "action=book",
          "displayText": "I want to book"
        }
      },
      {
        "type": "action",
        "action": {
          "type": "message",
          "label": "Talk to human",
          "text": "I want to talk to a person"
        }
      },
      {
        "type": "action",
        "action": {
          "type": "location",
          "label": "Send my location"
        }
      },
      {
        "type": "action",
        "action": {
          "type": "camera",
          "label": "Send photo"
        }
      },
      {
        "type": "action",
        "action": {
          "type": "cameraRoll",
          "label": "Send from album"
        }
      }
    ]
  }
}
```

**`quickReply` fields:**

| Field | Type | Required | Description |
|---|---|---|---|
| `quickReply` | object | No | Wrapper object on any message type. |
| `quickReply.items` | array | Yes | 1–13 quick reply button objects. |
| `items[].type` | string | Yes | Always `"action"` |
| `items[].imageUrl` | string | No | HTTPS URL of a 24×24px or larger PNG icon displayed on the button. Displayed at 24×24. Max 1 MB. |
| `items[].action` | Action object | Yes | The action triggered when this button is tapped. |

**Supported action types in quick reply:**

| Action Type | Exclusive to Quick Reply? | Description |
|---|---|---|
| `postback` | No | Sends a postback event to the webhook |
| `message` | No | Sends a text message as the user |
| `uri` | No | Opens a URL |
| `datetimepicker` | No | Opens a date/time picker |
| `clipboard` | No | Copies text to clipboard |
| `location` | Yes | Opens the location picker |
| `camera` | Yes | Opens the camera |
| `cameraRoll` | Yes | Opens the camera roll / photo library |

`richMenuSwitch` is **not supported** in quick reply.

**Behavior notes:**
- Buttons disappear after the user taps one, except for `camera`, `cameraRoll`, `datetimepicker`, and `location` actions (which persist until the user sends data).
- Buttons also disappear when any participant sends a new message in the chat.
- Supported on LINE for iOS and LINE for Android.

---

### 4.12 The `sender` Object

The `sender` property overrides the display name and icon for **a single message bubble**, without changing the official account profile. This is useful for multi-persona bots (e.g., different staff members or characters responding in context).

```json
{
  "type": "text",
  "text": "Hi! I'm Sarah from the support team. How can I help?",
  "sender": {
    "name": "Sarah",
    "iconUrl": "https://example.com/avatars/sarah.jpg"
  }
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `sender` | object | No | Overrides display name and icon for this message only. |
| `sender.name` | string | No | Display name shown on this message bubble. Max 20 characters. |
| `sender.iconUrl` | string | No | HTTPS URL of the avatar image. JPEG or PNG. Max 1 MB. Displayed at 30×30px in chat. |

**Notes:**
- Works with all message types and all five send endpoints.
- The chat room header at the top still shows the official account name.
- The chat will show `{sender.name} from {OfficialAccountName}` to clarify the source.
- Both `name` and `iconUrl` are individually optional; you can override one without the other.

> Ref: [Customize icon and display name](https://developers.line.biz/en/docs/messaging-api/icon-nickname-switch/)
