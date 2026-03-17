# LINE Messaging API — Outbound (Sending Messages) Comprehensive Reference

> **Sources:**
> - https://developers.line.biz/en/reference/messaging-api/ (primary API reference, all send endpoints, message objects, rate limits, error responses)
> - https://developers.line.biz/en/reference/messaging-api/index.html.md (raw markdown dump — primary fallback)
> - https://developers.line.biz/en/docs/basics/channel-access-token/ (all 4 token types overview)
> - https://developers.line.biz/en/docs/messaging-api/generate-json-web-token/ (v2.1 JWT assertion signing key flow)
> - https://developers.line.biz/en/docs/messaging-api/sending-messages/ (sending messages overview, quota counting)
> - https://developers.line.biz/en/docs/messaging-api/retrying-api-request/ (X-Line-Retry-Key behavior, 409 Conflict)
> - https://developers.line.biz/en/reference/messaging-api/#issue-channel-access-token-v2-1 (v2.1 token issue/verify/revoke endpoints)
> - https://developers.line.biz/en/reference/messaging-api/#issue-shortlived-channel-access-token (short-lived token endpoint)
> - https://developers.line.biz/en/reference/messaging-api/#issue-stateless-channel-access-token (stateless token endpoint)

---

## Table of Contents

1. [Setup (Manual Steps)](#1-setup-manual-steps)
   - 1.1 [Retrieving a Channel Access Token](#11-retrieving-a-channel-access-token)
   - 1.2 [Token Type Comparison](#12-token-type-comparison)
   - 1.3 [Bot Settings That Affect Message Sending](#13-bot-settings-that-affect-message-sending)
   - 1.4 [Rate Limits and Quota Considerations](#14-rate-limits-and-quota-considerations)
2. [Authentication](#2-authentication)
   - 2.1 [Bearer Token Header](#21-bearer-token-header)
   - 2.2 [Long-lived and Short-lived Tokens](#22-long-lived-and-short-lived-tokens)
   - 2.3 [Channel Access Token v2.1 (JWT Assertion)](#23-channel-access-token-v21-jwt-assertion)
   - 2.4 [Stateless Token](#24-stateless-token)
3. [Send Endpoints](#3-send-endpoints)
   - 3.1 [Reply Message](#31-reply-message)
   - 3.2 [Push Message](#32-push-message)
   - 3.3 [Multicast Message](#33-multicast-message)
   - 3.4 [Broadcast Message](#34-broadcast-message)
   - 3.5 [Narrowcast Message](#35-narrowcast-message)
4. [Message Object Types](#4-message-object-types)
   - 4.1 [Text](#41-text)
   - 4.2 [Image](#42-image)
   - 4.3 [Video](#43-video)
   - 4.4 [Audio](#44-audio)
   - 4.5 [File](#45-file)
   - 4.6 [Location](#46-location)
   - 4.7 [Sticker](#47-sticker)
   - 4.8 [Template (Buttons / Confirm / Carousel / Image Carousel)](#48-template)
   - 4.9 [Imagemap](#49-imagemap)
   - 4.10 [Flex](#410-flex)
   - 4.11 [Quick Reply](#411-quick-reply)

---

## 1. Setup (Manual Steps)

### 1.1 Retrieving a Channel Access Token

All four token types require a **Messaging API channel** created in the [LINE Developers Console](https://developers.line.biz/console/). There is no OAuth flow involving end-users — the token authenticates the bot server itself to the LINE Platform.

**Steps:**

1. Log in to the LINE Developers Console.
2. Select your provider and open (or create) a **Messaging API** channel.
3. Navigate to the **Messaging API** tab within the channel settings.
4. Choose the token type appropriate for your use case (see §1.2):
   - **Long-lived**: Issue directly from the console UI. No API call needed.
   - **Short-lived**: Call `POST https://api.line.me/v2/oauth/accessToken` with `client_id` and `client_secret`.
   - **v2.1**: Generate a key pair, register the public key in **Basic settings**, then call `POST https://api.line.me/oauth2/v2.1/token` with a JWT assertion.
   - **Stateless**: Call `POST https://api.line.me/oauth2/v3/token` with either `client_id`/`client_secret` or a JWT assertion.

> **Reference:** https://developers.line.biz/en/docs/basics/channel-access-token/

---

### 1.2 Token Type Comparison

| Token Type | Validity Period | Max per Channel | Revocable | Issue Method | Recommended Use |
|---|---|---|---|---|---|
| **Long-lived** | Indefinite (no expiry) | 1 | Yes | Console UI only | Development, simple deployments; not recommended for production due to single-token risk |
| **Short-lived** | 30 days | 30 | Yes (auto-revokes oldest when limit hit) | API (`POST /v2/oauth/accessToken`) | Production with rotation; auto-revokes oldest token when the 30-token cap is hit |
| **v2.1 (user-specified expiry)** | Up to 30 days (developer-set) | 30 | Yes (explicit revoke; denied if at cap) | API (`POST /oauth2/v2.1/token`) with JWT | Production with controlled rotation; allows per-team token scoping |
| **Stateless** | 15 minutes | Limitless | No — cannot be revoked once issued | API (`POST /oauth2/v3/token`) | High-frequency serverless or ephemeral environments; issue per-request |

**Key notes:**
- Expired tokens do not count against the per-channel maximum for short-lived and v2.1 tokens.
- Reissuing a long-lived token invalidates the current one (optionally with a 24-hour grace period overlap).
- Short-lived and v2.1 tokens should **not** be re-issued on every API call — reuse within validity period. Stateless tokens are specifically designed for per-call issuance.
- If you suspect a revocable token has been compromised, revoke it immediately. A compromised token could allow a third party to send broadcast messages to all friends.

---

### 1.3 Bot Settings That Affect Message Sending

These settings are configured in the LINE Official Account Manager (not the Developers Console):

| Setting | Effect on Sending |
|---|---|
| **Auto-reply** | If enabled, LINE's built-in auto-reply fires alongside your bot's webhook reply. Disable to avoid duplicate responses. |
| **Greeting message** | Fires automatically on friend add. Disable if your bot handles follow events with a custom welcome push. |
| **Webhook** | Must be **enabled** for webhooks to be delivered to your server. Without this, the bot receives no `replyToken`s. |
| **Response mode** | Set to **Bot** (not Chat). In Chat mode, the Messaging API is disabled for that account. |

> **Reference:** https://developers.line.biz/en/docs/messaging-api/building-bot/

---

### 1.4 Rate Limits and Quota Considerations

#### Per-Endpoint Rate Limits

| Endpoint(s) | Rate Limit |
|---|---|
| Send reply message | 2,000 requests/second |
| Send push message | 2,000 requests/second |
| Send multicast message | 200 requests/second |
| Send narrowcast message | 60 requests/hour |
| Send broadcast message | 60 requests/hour |
| Issue short-lived channel access token | 370 requests/second |
| All other API endpoints | 2,000 requests/second |

Rate limits are enforced **per channel**, not per IP address or per sending method variant. If you use the same LINE Official Account from multiple channels, each channel gets independent limits.

Exceeding the rate limit returns `429 Too Many Requests` with the message `"The API rate limit has been exceeded. Try again later."`

#### Monthly Message Quota

The number of messages counted against your monthly quota equals the **number of recipients**, not the number of message objects. A single request containing 5 message objects sent to 100 users counts as 100 messages against the quota.

Quota is shared across push, multicast, broadcast, and narrowcast sends. Reply messages are **free and do not consume monthly quota**.

When the monthly quota is exhausted, any further push/multicast/broadcast/narrowcast attempt returns `429 Too Many Requests` with `"You have reached your monthly limit."`

Specific free tier and paid tier message limits are documented in the [Messaging API pricing page](https://developers.line.biz/en/docs/messaging-api/pricing/) and are subject to change — they are not reproduced here to avoid documenting stale figures.

#### `notificationDisabled`

Setting `"notificationDisabled": true` in a push, multicast, broadcast, or narrowcast request suppresses the push notification on the recipient's device. The message is still delivered and visible in the chat; the user simply does not receive a notification sound/badge. Defaults to `false`.

---

## 2. Authentication

### 2.1 Bearer Token Header

Every Messaging API request (except token issuance) requires a channel access token in the `Authorization` header:

```
Authorization: Bearer {channel access token}
```

The token is an opaque string. It must be transmitted over HTTPS. All `api.line.me` endpoints enforce TLS.

---

### 2.2 Long-lived and Short-lived Tokens

#### Long-lived Token

Issued from the LINE Developers Console → Messaging API tab → **Channel access token (long-lived)**. There is no programmatic issue endpoint for long-lived tokens.

**Revoke:**
```
POST https://api.line.me/v2/oauth/revoke
Content-Type: application/x-www-form-urlencoded

access_token={long-lived or short-lived channel access token}
```

Response: `200` with empty body. No error is returned for invalid tokens.

#### Short-lived Token — Issue

```
POST https://api.line.me/v2/oauth/accessToken
Content-Type: application/x-www-form-urlencoded
```

**Request body (form-encoded):**

| Field | Type | Required | Notes |
|---|---|---|---|
| `grant_type` | string | Yes | Must be `client_credentials` |
| `client_id` | string | Yes | Channel ID from the console |
| `client_secret` | string | Yes | Channel secret from the console |

**Example cURL:**
```bash
curl -v -X POST https://api.line.me/v2/oauth/accessToken \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data-urlencode 'grant_type=client_credentials' \
  --data-urlencode 'client_id={channel ID}' \
  --data-urlencode 'client_secret={channel secret}'
```

**Success response (200):**
```json
{
  "access_token": "W1TeHCgfH2Liwa.....",
  "expires_in": 2592000,
  "token_type": "Bearer"
}
```

| Field | Notes |
|---|---|
| `access_token` | The token string to use in `Authorization: Bearer` header |
| `expires_in` | Seconds until expiry — 2592000 = 30 days |
| `token_type` | Always `"Bearer"` |

**Error responses:**

| Code | Reason |
|---|---|
| `400` | Invalid `client_id` or `client_secret` |

**Note:** If the 30-token limit is reached, the oldest token is automatically revoked to make room for the new one.

---

### 2.3 Channel Access Token v2.1 (JWT Assertion)

This token type uses RSA-signed JWT for authentication. The flow has three phases: key pair generation, public key registration, then token issuance.

> **Reference:** https://developers.line.biz/en/docs/messaging-api/generate-json-web-token/

#### Phase 1: Generate an RSA Key Pair

The assertion signing key must be an RSA-2048 JWK with these properties:

| Property | Value | Notes |
|---|---|---|
| `kty` | `RSA` | Cryptographic algorithm family |
| `alg` | `RS256` | RSASSA-PKCS1-v1_5 with SHA-256 |
| `use` | `sig` | OR use `key_ops: ["verify"]` — specify one |
| `e` | base64url | Public exponent |
| `n` | base64url | Modulus |

The public key must **not** have a `kid` property before registration — `kid` is assigned by LINE upon registration.

Example public key (JWK format):
```json
{
  "alg": "RS256",
  "e": "AQAB",
  "kty": "RSA",
  "n": "_RzHf7cgG_i6Pdo...",
  "use": "sig"
}
```

#### Phase 2: Register the Public Key

In the LINE Developers Console: **Basic settings** tab → **Register a public key** (next to "Assertion signing key"). Paste the public key JSON and click **Register**. The console returns a `kid` string — store this value.

#### Phase 3: Generate the JWT

The JWT has three parts: header, payload, signature.

**JWT Header:**

```json
{
  "alg": "RS256",
  "typ": "JWT",
  "kid": "536e453c-aa93-4449-8e90-add2608783c6"
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `alg` | string | Yes | Always `RS256` |
| `typ` | string | Yes | Always `JWT` |
| `kid` | string | Yes | The key ID returned from registering the public key |

**JWT Payload:**

```json
{
  "iss": "1234567890",
  "sub": "1234567890",
  "aud": "https://api.line.me/",
  "exp": 1559702522,
  "token_exp": 86400
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `iss` | string | Yes | Your channel ID. Must equal `sub`. |
| `sub` | string | Yes | Your channel ID. Must equal `iss`. |
| `aud` | string | Yes | Always `https://api.line.me/` |
| `exp` | number | Yes | JWT expiry as Unix epoch seconds. Max 30 minutes from issue time. |
| `token_exp` | number | Yes | Desired channel token validity in seconds. Max 2,592,000 (30 days). |

Sign the Base64url-encoded header + Base64url-encoded payload with your **private key** using RS256. The resulting JWT string is your assertion.

#### Phase 4: Issue the Channel Access Token

```
POST https://api.line.me/oauth2/v2.1/token
Content-Type: application/x-www-form-urlencoded
```

**Request body:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `grant_type` | string | Yes | Must be `client_credentials` |
| `client_assertion_type` | string | Yes | Must be `urn:ietf:params:oauth:client-assertion-type:jwt-bearer` |
| `client_assertion` | string | Yes | The JWT you generated above |

**Example cURL:**
```bash
curl -v -X POST https://api.line.me/oauth2/v2.1/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=client_credentials' \
  --data-urlencode 'client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer' \
  --data-urlencode 'client_assertion={JWT}'
```

**Success response (200):**
```json
{
  "access_token": "eyJhbGciOiJIUz.....",
  "token_type": "Bearer",
  "expires_in": 2592000,
  "key_id": "sDTOzw5wIfxxxxPEzcmeQA"
}
```

| Field | Notes |
|---|---|
| `access_token` | Use in `Authorization: Bearer` |
| `expires_in` | Seconds remaining (set by `token_exp` in the JWT) |
| `key_id` | The `kid` associated with this token; use to track which key issued which token |

**Error responses:**

| Code | Reason |
|---|---|
| `400` | JWT assertion verification failed; JWT expired; max 30 tokens per channel already issued |
| `404` | Signing key (kid) not registered in the channel |

**Revoke a v2.1 token:**
```bash
curl -X POST https://api.line.me/oauth2/v2.1/revoke \
  --data-urlencode 'client_id={channel ID}' \
  --data-urlencode 'client_secret={channel secret}' \
  --data-urlencode 'access_token={access token}'
```

Returns `200` with empty body. No error for an already-invalid token.

---

### 2.4 Stateless Token

Stateless tokens are valid for 15 minutes, cannot be revoked, and have no per-channel issuance limit. They are designed to be issued fresh for every API call or short-lived operation.

```
POST https://api.line.me/oauth2/v3/token
Content-Type: application/x-www-form-urlencoded
```

**Option A — Issue from channel ID and channel secret:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `grant_type` | string | Yes | `client_credentials` |
| `client_id` | string | Yes | Channel ID |
| `client_secret` | string | Yes | Channel secret |

**Option B — Issue from JWT assertion:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `grant_type` | string | Yes | `client_credentials` |
| `client_assertion_type` | string | Yes | `urn:ietf:params:oauth:client-assertion-type:jwt-bearer` |
| `client_assertion` | string | Yes | A JWT signed with the registered private key |

**Example cURL (Option A):**
```bash
curl -v -X POST https://api.line.me/oauth2/v3/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=client_credentials' \
  --data-urlencode 'client_id={channel ID}' \
  --data-urlencode 'client_secret={channel secret}'
```

**Success response (200):**
```json
{
  "token_type": "Bearer",
  "access_token": "ey....",
  "expires_in": 900
}
```

`expires_in: 900` = 15 minutes, always.

**Error responses:**

| Code | Reason |
|---|---|
| `400` | Invalid `client_id`, invalid `client_secret`, JWT verification failed, or JWT expired |
| `404` | Signing key not registered (JWT method only) |

---

## 3. Send Endpoints

All send endpoints share these request headers:

```
Content-Type: application/json
Authorization: Bearer {channel access token}
```

---

### 3.1 Reply Message

Reply messages are **free** (do not consume monthly quota). They require a one-time-use `replyToken` from a webhook event. Use this endpoint as the default response to user-initiated interactions.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#send-reply-message

**Endpoint:**
```
POST https://api.line.me/v2/bot/message/reply
```

**Rate limit:** 2,000 requests/second

**Full request body:**
```json
{
  "replyToken": "nHuyWiB7yP5Zw52FIkcQobQuGDXCTA",
  "messages": [
    {
      "type": "text",
      "text": "Hello, user"
    },
    {
      "type": "text",
      "text": "May I help you?"
    }
  ],
  "notificationDisabled": false
}
```

**Field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `replyToken` | string | Yes | One-time token from the webhook event object |
| `messages` | array | Yes | 1–5 message objects |
| `notificationDisabled` | boolean | No | `true` suppresses push notification; default `false` |

**Success response (200):**
```json
{
  "sentMessages": [
    {
      "id": "461230966842064897",
      "quoteToken": "IStG5h1Tz7b..."
    }
  ]
}
```

| Field | Notes |
|---|---|
| `sentMessages[].id` | Message ID of the sent message |
| `sentMessages[].quoteToken` | Token to quote this message in a future message |

**Error responses:**

| Status | Meaning |
|---|---|
| `400` | Invalid or expired reply token; invalid message object |
| `401` | Missing or invalid channel access token |
| `403` | Insufficient permissions |
| `429` | Rate limit exceeded |
| `500` | Internal server error |

**Key notes:**
- Reply tokens are **single-use** — once used, the token is consumed.
- Reply tokens must be used within **approximately 1 minute** of receiving the webhook. Use beyond one minute is not guaranteed. The hard cutoff is 20 minutes from the time the event occurred — after that, the token is invalid regardless.
- Reply tokens from **redelivered webhooks** are valid for 1 minute from the redelivery, unless the original token was already used or 20 minutes have elapsed from the original event.
- Do not rely on the specific time limit — treat reply tokens as very short-lived and use them immediately.
- `X-Line-Retry-Key` is **not supported** for reply message. Do not include it.

**Example cURL:**
```bash
curl -v -X POST https://api.line.me/v2/bot/message/reply \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer {channel access token}' \
  -d '{
    "replyToken": "nHuyWiB7yP5Zw52FIkcQobQuGDXCTA",
    "messages": [
      { "type": "text", "text": "Hello, user" }
    ]
  }'
```

---

### 3.2 Push Message

Sends a message to a single user, group chat, or multi-person chat at any time (not triggered by an event). Consumes monthly quota. Use when you need to send a proactive or asynchronous message to a known recipient.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#send-push-message

**Endpoint:**
```
POST https://api.line.me/v2/bot/message/push
```

**Rate limit:** 2,000 requests/second

**Full request body:**
```json
{
  "to": "U4af4980629...",
  "messages": [
    {
      "type": "text",
      "text": "Hello, world1"
    },
    {
      "type": "text",
      "text": "Hello, world2"
    }
  ],
  "notificationDisabled": false
}
```

**Field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `to` | string | Yes | A user ID (`U...`), group ID (`C...`), or multi-person chat ID (`R...`) |
| `messages` | array | Yes | 1–5 message objects |
| `notificationDisabled` | boolean | No | Suppresses push notification; default `false` |

**Request header (optional for idempotency):**

| Header | Type | Notes |
|---|---|---|
| `X-Line-Retry-Key` | string (UUID) | Hexadecimal UUID. If provided, the same UUID retried within 24 hours returns `409` instead of sending again. See §3.2 Key Notes. |

**Success response (200):**
```json
{
  "sentMessages": [
    {
      "id": "461230966842064897",
      "quoteToken": "IStG5h1Tz7b..."
    }
  ]
}
```

**Error responses:**

| Status | Meaning |
|---|---|
| `400` | User ID not in this channel; non-existent group/room; invalid message object |
| `401` | Invalid channel access token |
| `403` | Insufficient permissions |
| `409` | Same retry key already accepted (see X-Line-Retry-Key notes) |
| `429` | Rate limit or monthly quota exceeded |
| `500` | Internal server error |

**Key notes:**
- Push messages are delivered silently to users who **blocked** your account — `200` is returned but the message is never received. This is by design.
- Users who sent a message to your account within the **last 7 days** (without being friends) can also receive push messages.
- `X-Line-Retry-Key` should be specified on the **first** request. Requests without a retry key cannot be retried safely.
- A retry key remains valid for **24 hours** after the first accepted request. After 24 hours, the same key can be used again for a different message.
- Retrying with the same key after a `200` returns `409 Conflict` with `x-line-accepted-request-id` header pointing to the original accepted request.

**Example cURL:**
```bash
curl -v -X POST https://api.line.me/v2/bot/message/push \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer {channel access token}' \
  -H 'X-Line-Retry-Key: 123e4567-e89b-12d3-a456-426614174000' \
  -d '{
    "to": "U4af4980629...",
    "messages": [
      { "type": "text", "text": "Hello!" }
    ]
  }'
```

---

### 3.3 Multicast Message

Sends the same message to multiple user IDs in a single request. More efficient than sending individual push messages when the content is identical. Cannot target groups or multi-person chats — only individual user IDs.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#send-multicast-message

**Endpoint:**
```
POST https://api.line.me/v2/bot/message/multicast
```

**Rate limit:** 200 requests/second

**Full request body:**
```json
{
  "to": [
    "U4af4980629...",
    "U0c229f96c4..."
  ],
  "messages": [
    {
      "type": "text",
      "text": "Hello, world1"
    },
    {
      "type": "text",
      "text": "Hello, world2"
    }
  ],
  "notificationDisabled": false
}
```

**Field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `to` | array of strings | Yes | Array of user IDs. Max 500 per request (not explicitly documented — see notes). |
| `messages` | array | Yes | 1–5 message objects |
| `notificationDisabled` | boolean | No | Suppresses push notification; default `false` |

**Request header (optional):**

| Header | Type | Notes |
|---|---|---|
| `X-Line-Retry-Key` | string (UUID) | Same idempotency behavior as push |

**Success response (200):**
```json
{}
```

Empty JSON object.

**Error responses:**

| Status | Meaning |
|---|---|
| `400` | User ID not in this channel; non-user ID (e.g., group ID) in `to` array; invalid message object |
| `401` | Invalid channel access token |
| `409` | Same retry key already accepted |
| `429` | Rate limit or monthly quota exceeded |
| `500` | Internal server error |

**Key notes:**
- If any user in the `to` array is invalid (wrong channel, blocked, etc.), those users are silently skipped — the other recipients still receive the message and `200` is returned.
- If an error occurs, **no users** receive the message. The entire batch is atomic.
- When sending to a single user, prefer push message (lower latency); use multicast for batch efficiency.
- Group IDs (`C...`) and room IDs (`R...`) are not accepted in the `to` array.

**Example cURL:**
```bash
curl -v -X POST https://api.line.me/v2/bot/message/multicast \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer {channel access token}' \
  -H 'X-Line-Retry-Key: 123e4567-e89b-12d3-a456-426614174001' \
  -d '{
    "to": ["U4af4980629...", "U0c229f96c4..."],
    "messages": [
      { "type": "text", "text": "Batch announcement" }
    ]
  }'
```

---

### 3.4 Broadcast Message

Sends a message to **all users** who have ever added your LINE Official Account as a friend, regardless of whether they are currently following. Does not target groups or rooms. The highest-reach send method — use for account-wide announcements.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#send-broadcast-message

**Endpoint:**
```
POST https://api.line.me/v2/bot/message/broadcast
```

**Rate limit:** 60 requests/hour

**Full request body:**
```json
{
  "messages": [
    {
      "type": "text",
      "text": "Announcement to all users"
    }
  ],
  "notificationDisabled": false
}
```

**Field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `messages` | array | Yes | 1–5 message objects |
| `notificationDisabled` | boolean | No | Suppresses push notification; default `false` |

**Request header (optional):**

| Header | Type | Notes |
|---|---|---|
| `X-Line-Retry-Key` | string (UUID) | Same idempotency behavior as push/multicast |

**Success response (200):**
```json
{
  "sentMessages": [
    {
      "id": "461230966842064897",
      "quoteToken": "IStG5h1Tz7b..."
    }
  ]
}
```

**Error responses:**

| Status | Meaning |
|---|---|
| `400` | Invalid message object |
| `401` | Invalid channel access token |
| `403` | Insufficient permissions |
| `409` | Same retry key already accepted |
| `429` | Rate limit exceeded (60/hour) or monthly quota exhausted |
| `500` | Internal server error |

**Key notes:**
- Broadcast counts against the monthly quota at the rate of one message per recipient (the account's total friend count).
- Users who blocked the account are excluded from delivery but are still counted toward the quota reservation.
- The 60 requests/hour rate limit is low — plan bulk sends accordingly.

**Example cURL:**
```bash
curl -v -X POST https://api.line.me/v2/bot/message/broadcast \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer {channel access token}' \
  -H 'X-Line-Retry-Key: 123e4567-e89b-12d3-a456-426614174002' \
  -d '{
    "messages": [
      { "type": "text", "text": "Hello everyone!" }
    ]
  }'
```

---

### 3.5 Narrowcast Message

Sends a message to a filtered subset of users using audience targeting (uploaded user lists, retargeting audiences) or demographic filters (age, gender, OS, region). Delivered asynchronously — the API accepts the request and returns `202`, then delivers in the background.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#send-narrowcast-message

**Endpoint:**
```
POST https://api.line.me/v2/bot/message/narrowcast
```

**Rate limit:** 60 requests/hour

**Full request body:**
```json
{
  "messages": [
    {
      "type": "text",
      "text": "Targeted message"
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
          "type": "audience",
          "audienceGroupId": 4389303728991
        }
      }
    ]
  },
  "filter": {
    "demographic": {
      "type": "operator",
      "or": [
        {
          "type": "operator",
          "and": [
            {
              "type": "gender",
              "oneOf": ["male", "female"]
            },
            {
              "type": "age",
              "gte": "age_20",
              "lt": "age_25"
            },
            {
              "type": "appType",
              "oneOf": ["android", "ios"]
            },
            {
              "type": "area",
              "oneOf": ["jp_23", "jp_05"]
            },
            {
              "type": "subscriptionPeriod",
              "gte": "day_7",
              "lt": "day_30"
            }
          ]
        }
      ]
    }
  },
  "limit": {
    "max": 10000,
    "upToRemainingQuota": true
  },
  "notificationDisabled": false
}
```

**Top-level field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `messages` | array | Yes | 1–5 message objects |
| `recipient` | object | No | Audience/operator targeting. If omitted, sends to all friends (same as broadcast). |
| `filter` | object | No | Demographic filter. `filter.demographic` is an operator or leaf demographic object. |
| `limit` | object | No | Caps on number of messages sent |
| `notificationDisabled` | boolean | No | Suppresses push notification; default `false` |

**Recipient object (audience):**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | `"audience"` |
| `audienceGroupId` | number | Yes | ID of a pre-created audience group |

**Recipient object (operator — combine audiences with AND / OR / NOT):**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | `"operator"` |
| `and` | array | One of | Array of recipient objects — all must match |
| `or` | array | One of | Array of recipient objects — any must match |
| `not` | object | One of | Single recipient object to negate |

Max **10** audience/redelivery objects per request (combined across all operator nesting).

**Demographic filter leaf objects:**

| `type` | Key Fields | Notes |
|---|---|---|
| `gender` | `oneOf`: `["male"]`, `["female"]`, or both | |
| `age` | `gte`, `lt`: e.g., `"age_20"`, `"age_25"` | |
| `appType` | `oneOf`: `["android"]`, `["ios"]`, or both | |
| `area` | `oneOf`: region codes, e.g., `"jp_23"` | |
| `subscriptionPeriod` | `gte`, `lt`: e.g., `"day_7"`, `"day_30"` | How long user has followed |

**Limit object:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `max` | number | No | Maximum number of messages to send |
| `upToRemainingQuota` | boolean | No | `true` = cap delivery at remaining monthly quota; prevents quota overrun |

**Request header (optional):**

| Header | Type | Notes |
|---|---|---|
| `X-Line-Retry-Key` | string (UUID) | Same idempotency behavior as push |

**Success response (202 Accepted):**
```json
{
  "requestId": "5b59509c-c57b-11e9-aa8c-2a86e4085a59",
  "acceptedTime": "2019-08-26T01:05:09Z"
}
```

Delivery happens asynchronously after this response. Use `GET /v2/bot/message/narrowcast/progress?requestId=...` to check progress.

**Error responses:**

| Status | Meaning |
|---|---|
| `400` | Invalid message object; redelivery request ID doesn't satisfy conditions |
| `401` | Invalid channel access token |
| `403` | Target reach < 100 (demographic filter used); insufficient permissions |
| `409` | Same retry key already accepted |
| `429` | Rate limit exceeded or monthly quota insufficient |
| `500` | Internal server error |

**Key notes:**
- Narrowcast reserves monthly quota equal to the account's **full target reach** at send start, not the actual number of recipients. This temporary reservation can cause `429` for concurrent sends even if quota is technically available.
- Final recipient count must be ≥ 50 when demographic attributes or audiences are used. Requests with fewer recipients succeed with `202` but fail during delivery.
- Each individual audience in a multi-audience request must have ≥ 50 members (unless it's a user-upload or chat-tag audience).
- Users under age 20 in Thailand are always excluded from demographically filtered sends.
- Use `limit.upToRemainingQuota: true` to prevent exhausting the monthly quota during delivery.

---

## X-Line-Retry-Key — Full Behavior Reference

> **Reference:** https://developers.line.biz/en/docs/messaging-api/retrying-api-request/

Supported endpoints: push, multicast, narrowcast, broadcast. **Not** supported for reply message (returns `400` if included).

**Behavior:**

1. Generate a UUID (`123e4567-e89b-12d3-a456-426614174000`) and include it in the first request.
2. If the request succeeds (`200`/`202`), the retry key is consumed. Retrying with the same key returns `409`.
3. If the request fails with a `500` or times out, retry with the **same UUID** and **same request body**. The LINE Platform deduplicates on the key and executes the request only once.
4. A retry key is valid for **24 hours** from the first accepted request.

**Retry decision table:**

| Status Received | Action |
|---|---|
| `500 Internal Server Error` | Retry with same UUID |
| Timeout / network failure | Retry with same UUID |
| `200` / `202` | Do NOT retry — request accepted |
| `409 Conflict` | Do NOT retry — already accepted |
| `4xx` (except `409`) | Do NOT retry — fix the request |

**409 response body:**
```json
{
  "message": "The retry key is already accepted"
}
```

Headers accompanying `409`:
```
x-line-request-id: {new request ID for this retry attempt}
x-line-accepted-request-id: {request ID of the originally accepted request}
```

For push messages, `409` also returns the original `sentMessages` array (with `id` and `quoteToken`).

---

## 4. Message Object Types

All message objects share two optional common properties:

| Property | Type | Notes |
|---|---|---|
| `quickReply` | object | Quick reply buttons to display with this message. Only the `quickReply` of the **last** message object in the array is shown if multiple messages are sent. See §4.11. |
| `sender` | object | Override the display name and icon for this message. `sender.name` (string, max 20 chars) and `sender.iconUrl` (HTTPS URL to an icon image). |

---

### 4.1 Text

Plain text message. The most common message type.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#text-message

**JSON example:**
```json
{
  "type": "text",
  "text": "Hello, world"
}
```

**With LINE emoji:**
```json
{
  "type": "text",
  "text": "$ LINE emoji $",
  "emojis": [
    {
      "index": 0,
      "productId": "5ac1bfd5040ab15980c9b435",
      "emojiId": "001"
    },
    {
      "index": 13,
      "productId": "5ac1bfd5040ab15980c9b435",
      "emojiId": "002"
    }
  ]
}
```

**With quote (referencing a prior message):**
```json
{
  "type": "text",
  "text": "Yes, you can.",
  "quoteToken": "yHAz4Ua2wx7..."
}
```

**With sender override:**
```json
{
  "type": "text",
  "text": "Hello, I am Cony!!",
  "sender": {
    "name": "Cony",
    "iconUrl": "https://line.me/conyprof"
  }
}
```

**Field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | Always `"text"` |
| `text` | string | Yes | Message content. Max 5,000 characters. |
| `emojis` | array | No | LINE emoji substitutions within the text |
| `emojis[].index` | number | Yes (if emojis) | Position of `$` placeholder in `text`, measured in UTF-16 code units |
| `emojis[].productId` | string | Yes (if emojis) | Product ID of the emoji set |
| `emojis[].emojiId` | string | Yes (if emojis) | Emoji ID within the product |
| `quoteToken` | string | No | Quote token from a prior message's `quoteToken` field in a webhook or send response |
| `sender` | object | No | Override bot name and icon |
| `quickReply` | object | No | Quick reply buttons |

**Constraints:**
- `text` max: **5,000 characters**
- Emoji `index` positions are counted in **UTF-16 code units** — surrogate pairs (e.g., many emoji) count as 2 units, not 1. Use `$` as a placeholder at each emoji position; the `$` is replaced by the LINE emoji at render time.
- The `text` field must contain exactly as many `$` placeholders as there are entries in `emojis`.

#### Text message v2 (textV2)

A newer variant that uses `{key}` substitution syntax for mentions and emojis.

```json
{
  "type": "textV2",
  "text": "Welcome, {user1}! {laugh}\n{everyone} There is a newcomer!",
  "substitution": {
    "user1": {
      "type": "mention",
      "mentionee": {
        "type": "user",
        "userId": "U49585cd0d5..."
      }
    },
    "laugh": {
      "type": "emoji",
      "productId": "5a8555cfe6256cc92ea23c2a",
      "emojiId": "002"
    },
    "everyone": {
      "type": "mention",
      "mentionee": {
        "type": "all"
      }
    }
  }
}
```

**Constraints for textV2:**
- Supports up to **20 mentions** and **20 emojis** per message.
- `mentionee.type: "user"` only works in **reply** or **push** messages, not multicast/broadcast/narrowcast.
- The destination must be a **group chat or multi-person chat** for user mentions.
- All mentioned users must be members of the destination chat.

---

### 4.2 Image

Sends an image file via HTTPS URL.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#image-message

**JSON example:**
```json
{
  "type": "image",
  "originalContentUrl": "https://example.com/original.jpg",
  "previewImageUrl": "https://example.com/preview.jpg"
}
```

**Field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | Always `"image"` |
| `originalContentUrl` | string | Yes | HTTPS URL to the full-size image |
| `previewImageUrl` | string | Yes | HTTPS URL to the thumbnail/preview image |
| `quickReply` | object | No | Quick reply buttons |
| `sender` | object | No | Override bot name and icon |

**Constraints:**
- Both URLs must use **HTTPS**.
- Max URL length: **2,000 characters** each.
- Supported formats: **JPEG**, **PNG**.
- Max file size: **10 MB** for `originalContentUrl`; **1 MB** for `previewImageUrl`.
- URLs must be publicly accessible at send time (no auth required).

---

### 4.3 Video

Sends a video file. The preview image is shown before playback.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#video-message

**JSON example:**
```json
{
  "type": "video",
  "originalContentUrl": "https://example.com/original.mp4",
  "previewImageUrl": "https://example.com/preview.jpg",
  "trackingId": "track-id"
}
```

**Field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | Always `"video"` |
| `originalContentUrl` | string | Yes | HTTPS URL to the video file |
| `previewImageUrl` | string | Yes | HTTPS URL to the preview/thumbnail image |
| `trackingId` | string | No | Custom ID for tracking video views via the `videoPlayComplete` webhook event. Max 100 chars, alphanumeric, `-`, `_`. |
| `quickReply` | object | No | Quick reply buttons |
| `sender` | object | No | Override bot name and icon |

**Constraints:**
- Both URLs must use **HTTPS**.
- Max URL length: **2,000 characters** each.
- Supported video format: **MP4**.
- Max video file size: **200 MB**.
- Preview image: JPEG or PNG, max **1 MB**.
- The aspect ratio of `originalContentUrl` and `previewImageUrl` should match — mismatched ratios cause the preview to display behind the video.
- Very wide or tall videos may be cropped in some client environments.

---

### 4.4 Audio

Sends an audio file.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#audio-message

**JSON example:**
```json
{
  "type": "audio",
  "originalContentUrl": "https://example.com/original.m4a",
  "duration": 60000
}
```

**Field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | Always `"audio"` |
| `originalContentUrl` | string | Yes | HTTPS URL to the audio file |
| `duration` | number | Yes | Duration of the audio in milliseconds |
| `quickReply` | object | No | Quick reply buttons |
| `sender` | object | No | Override bot name and icon |

**Constraints:**
- URL must use **HTTPS**. Max **2,000 characters**.
- Supported format: **M4A** (AAC encoded).
- Max file size: **200 MB**.
- `duration` is displayed in the chat UI before the user plays the audio.

---

### 4.5 File

Sends a file to users. Only available in one-to-one chats (not group chats or multi-person chats). Note: this is the **send** direction only; users can also send files to bots via webhook.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#file-message

**Note on availability:** The file message type for **sending** (bot → user) is documented in the reference but its support in the send API has constraints. As of the current documentation, file messages sent via the bot can only be received in **one-to-one chats**. Verify current availability via the LINE Developers documentation — this constraint may change.

**JSON example:**
```json
{
  "type": "file",
  "originalContentUrl": "https://example.com/document.pdf",
  "fileName": "document.pdf",
  "fileSize": 1024000
}
```

**Field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | Always `"file"` |
| `originalContentUrl` | string | Yes | HTTPS URL to the file |
| `fileName` | string | Yes | File name with extension, displayed in chat |
| `fileSize` | number | Yes | File size in bytes |

**Constraints:**
- Max file size: **200 MB**.
- URL must use HTTPS.

---

### 4.6 Location

Sends a location pin with coordinates and address text.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#location-message

**JSON example:**
```json
{
  "type": "location",
  "title": "my location",
  "address": "1-3 Kioicho, Chiyoda-ku, Tokyo, 102-8282, Japan",
  "latitude": 35.67966,
  "longitude": 139.73669
}
```

**Field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | Always `"location"` |
| `title` | string | Yes | Label for the location pin. Max 100 characters. |
| `address` | string | Yes | Address text. Max 100 characters. |
| `latitude` | number | Yes | Latitude in decimal degrees |
| `longitude` | number | Yes | Longitude in decimal degrees |
| `quickReply` | object | No | Quick reply buttons |
| `sender` | object | No | Override bot name and icon |

**Constraints:**
- `title` and `address` are display-only; they do not affect the map pin position (only `latitude`/`longitude` do).

---

### 4.7 Sticker

Sends a LINE sticker from the official sticker store.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#sticker-message

**JSON example:**
```json
{
  "type": "sticker",
  "packageId": "446",
  "stickerId": "1988"
}
```

**With quote:**
```json
{
  "type": "sticker",
  "packageId": "789",
  "stickerId": "10855",
  "quoteToken": "yHAz4Ua2wx7..."
}
```

**Field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | Always `"sticker"` |
| `packageId` | string | Yes | Sticker package ID |
| `stickerId` | string | Yes | Individual sticker ID within the package |
| `quoteToken` | string | No | Quote a prior message |
| `quickReply` | object | No | Quick reply buttons |

**Constraints:**
- Only stickers from the **Messaging API sticker list** can be sent. The full list is available at https://developers.line.biz/en/docs/messaging-api/sticker-list/
- Sending a sticker with an invalid `packageId`/`stickerId` combination returns a `400` error.

---

### 4.8 Template

Template messages use predefined layouts. They require an `altText` fallback for LINE clients that do not support templates. Four template types are available.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#template-messages

All templates share this outer envelope:

```json
{
  "type": "template",
  "altText": "Fallback text for unsupported clients",
  "template": { ... }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | Always `"template"` |
| `altText` | string | Yes | Fallback text shown when templates are not supported. Max 400 characters. |
| `template` | object | Yes | Template-specific object with `type` field |

---

#### 4.8.1 Buttons Template

An image with a title, body text, and up to 4 action buttons.

```json
{
  "type": "template",
  "altText": "This is a buttons template",
  "template": {
    "type": "buttons",
    "thumbnailImageUrl": "https://example.com/bot/images/image.jpg",
    "imageAspectRatio": "rectangle",
    "imageSize": "cover",
    "imageBackgroundColor": "#FFFFFF",
    "title": "Menu",
    "text": "Please select",
    "defaultAction": {
      "type": "uri",
      "label": "View detail",
      "uri": "http://example.com/page/123"
    },
    "actions": [
      {
        "type": "postback",
        "label": "Buy",
        "data": "action=buy&itemid=123"
      },
      {
        "type": "postback",
        "label": "Add to cart",
        "data": "action=add&itemid=123"
      },
      {
        "type": "uri",
        "label": "View detail",
        "uri": "http://example.com/page/123"
      }
    ]
  }
}
```

**Field table (template object):**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | `"buttons"` |
| `thumbnailImageUrl` | string | No | HTTPS URL; JPEG or PNG; max 10 MB |
| `imageAspectRatio` | string | No | `"rectangle"` (1.51:1) or `"square"` (1:1); default `"rectangle"` |
| `imageSize` | string | No | `"cover"` or `"contain"`; default `"cover"` |
| `imageBackgroundColor` | string | No | Hex color for letterbox areas (e.g., `"#FFFFFF"`); default `"#FFFFFF"` |
| `title` | string | No | Title text. Max 40 characters. |
| `text` | string | Yes | Body text. Max 160 characters (60 if thumbnail is set). |
| `defaultAction` | action object | No | Action triggered when the image area is tapped |
| `actions` | array | Yes | 1–4 action objects |

---

#### 4.8.2 Confirm Template

A text message with exactly two action buttons (e.g., Yes/No).

```json
{
  "type": "template",
  "altText": "this is a confirm template",
  "template": {
    "type": "confirm",
    "text": "Are you sure?",
    "actions": [
      {
        "type": "message",
        "label": "Yes",
        "text": "yes"
      },
      {
        "type": "message",
        "label": "No",
        "text": "no"
      }
    ]
  }
}
```

**Field table (template object):**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | `"confirm"` |
| `text` | string | Yes | Body text. Max 240 characters. |
| `actions` | array | Yes | Exactly **2** action objects |

---

#### 4.8.3 Carousel Template

A horizontally scrollable series of column objects, each with an image, title, text, and up to 3 buttons.

```json
{
  "type": "template",
  "altText": "this is a carousel template",
  "template": {
    "type": "carousel",
    "imageAspectRatio": "rectangle",
    "imageSize": "cover",
    "columns": [
      {
        "thumbnailImageUrl": "https://example.com/bot/images/item1.jpg",
        "imageBackgroundColor": "#FFFFFF",
        "title": "Product 1",
        "text": "Description of product 1",
        "defaultAction": {
          "type": "uri",
          "label": "View detail",
          "uri": "http://example.com/page/1"
        },
        "actions": [
          {
            "type": "postback",
            "label": "Buy",
            "data": "action=buy&itemid=1"
          },
          {
            "type": "postback",
            "label": "Add to cart",
            "data": "action=add&itemid=1"
          },
          {
            "type": "uri",
            "label": "View detail",
            "uri": "http://example.com/page/1"
          }
        ]
      },
      {
        "thumbnailImageUrl": "https://example.com/bot/images/item2.jpg",
        "title": "Product 2",
        "text": "Description of product 2",
        "actions": [
          {
            "type": "postback",
            "label": "Buy",
            "data": "action=buy&itemid=2"
          }
        ]
      }
    ]
  }
}
```

**Field table (template object):**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | `"carousel"` |
| `columns` | array | Yes | 1–10 column objects |
| `imageAspectRatio` | string | No | `"rectangle"` or `"square"`; applies to all columns; default `"rectangle"` |
| `imageSize` | string | No | `"cover"` or `"contain"`; applies to all columns; default `"cover"` |

**Column object fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `thumbnailImageUrl` | string | No | HTTPS image URL; JPEG or PNG; max 10 MB |
| `imageBackgroundColor` | string | No | Hex background color |
| `title` | string | No | Max 40 characters |
| `text` | string | Yes | Max 120 characters (60 if thumbnail present) |
| `defaultAction` | action object | No | Action on image area tap |
| `actions` | array | Yes | 1–3 action objects. All columns must have the same number of actions. |

---

#### 4.8.4 Image Carousel Template

A horizontally scrollable series of images, each with a single action.

```json
{
  "type": "template",
  "altText": "this is an image carousel template",
  "template": {
    "type": "image_carousel",
    "columns": [
      {
        "imageUrl": "https://example.com/bot/images/item1.jpg",
        "action": {
          "type": "postback",
          "label": "Buy",
          "data": "action=buy&itemid=1"
        }
      },
      {
        "imageUrl": "https://example.com/bot/images/item2.jpg",
        "action": {
          "type": "message",
          "label": "Yes",
          "text": "yes"
        }
      }
    ]
  }
}
```

**Field table (template object):**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | `"image_carousel"` |
| `columns` | array | Yes | 1–10 image carousel column objects |

**Image carousel column fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `imageUrl` | string | Yes | HTTPS URL; JPEG or PNG; aspect ratio 1:1; max 10 MB |
| `action` | action object | Yes | Exactly one action per column |

**Template action object types (all templates):**

| `type` | Key Fields | Notes |
|---|---|---|
| `postback` | `label`, `data`, `displayText` | Fires postback event; `data` max 300 chars |
| `message` | `label`, `text` | Sends a message as the user; `text` max 300 chars |
| `uri` | `label`, `uri`, `altUri` | Opens a URL; `uri` max 1,000 chars |
| `datetimepicker` | `label`, `data`, `mode`, `initial`, `max`, `min` | Opens date/time picker; `mode`: `date`, `time`, `datetime` |
| `camera` | `label` | Opens camera |
| `cameraRoll` | `label` | Opens camera roll |
| `location` | `label` | Opens location picker |

`label` max: **20 characters** for all action types.

---

### 4.9 Imagemap

An image with one or more tappable areas (hotspots), each triggering a URI, message, or clipboard copy action. Optionally includes a video overlay.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#imagemap-message

**JSON example:**
```json
{
  "type": "imagemap",
  "baseUrl": "https://example.com/bot/images/rm001",
  "altText": "This is an imagemap",
  "baseSize": {
    "width": 1040,
    "height": 1040
  },
  "video": {
    "originalContentUrl": "https://example.com/video.mp4",
    "previewImageUrl": "https://example.com/video_preview.jpg",
    "area": {
      "x": 0,
      "y": 0,
      "width": 1040,
      "height": 585
    },
    "externalLink": {
      "linkUri": "https://example.com/see_more.html",
      "label": "See More"
    }
  },
  "actions": [
    {
      "type": "uri",
      "label": "https://example.com/",
      "linkUri": "https://example.com/",
      "area": {
        "x": 0,
        "y": 586,
        "width": 520,
        "height": 454
      }
    },
    {
      "type": "message",
      "label": "hello",
      "text": "Hello",
      "area": {
        "x": 520,
        "y": 586,
        "width": 520,
        "height": 454
      }
    }
  ]
}
```

**Field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | Always `"imagemap"` |
| `baseUrl` | string | Yes | HTTPS base URL for images (without size suffix); must be accessible at `{baseUrl}/{width}` |
| `altText` | string | Yes | Fallback text. Max 400 characters. |
| `baseSize` | object | Yes | Coordinate reference size. `width` must be 1040. |
| `baseSize.width` | number | Yes | Always `1040` |
| `baseSize.height` | number | Yes | Height in px, corresponding to the 1040-wide reference frame |
| `video` | object | No | Optional video overlay on the imagemap |
| `video.originalContentUrl` | string | Yes (if video) | HTTPS URL to the video (MP4) |
| `video.previewImageUrl` | string | Yes (if video) | HTTPS URL to the video preview image |
| `video.area` | object | Yes (if video) | `{x, y, width, height}` in base coordinate space |
| `video.externalLink` | object | No | Link label shown after video ends. `{linkUri, label}` |
| `actions` | array | Yes | 1 or more imagemap action objects |

**Imagemap action fields:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | `"uri"`, `"message"`, or `"clipboard"` |
| `area` | object | Yes | `{x, y, width, height}` in base coordinate space |
| `linkUri` | string | For `uri` | HTTPS URL to open |
| `text` | string | For `message` | Message text sent as user. Max 300 characters. |
| `clipboardText` | string | For `clipboard` | Text to copy. Max 1,000 characters. Available LINE v14.0.0+. |
| `label` | string | No | Accessibility label |

**Image requirements for imagemap:**
- Formats: **JPEG** or **PNG**
- Required widths: **240px, 300px, 460px, 700px, 1040px** — all five must be accessible at `{baseUrl}/{width}`
- Max file size: **10 MB** per image
- Do not include a file extension in the `baseUrl` path

---

### 4.10 Flex

Flex messages are highly customizable messages with a CSS Flexible Box-inspired layout engine. They support complex multi-section layouts not possible with template messages.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#flex-message

**JSON example (minimal bubble):**
```json
{
  "type": "flex",
  "altText": "Flex message alt text",
  "contents": {
    "type": "bubble",
    "hero": {
      "type": "image",
      "url": "https://example.com/image.jpg",
      "size": "full",
      "aspectRatio": "20:13"
    },
    "body": {
      "type": "box",
      "layout": "vertical",
      "contents": [
        {
          "type": "text",
          "text": "Product Name",
          "weight": "bold",
          "size": "xl"
        },
        {
          "type": "text",
          "text": "Description text here",
          "wrap": true,
          "color": "#666666"
        }
      ]
    },
    "footer": {
      "type": "box",
      "layout": "vertical",
      "contents": [
        {
          "type": "button",
          "style": "primary",
          "action": {
            "type": "uri",
            "label": "Buy Now",
            "uri": "https://example.com"
          }
        }
      ]
    }
  }
}
```

**Top-level field table:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | Always `"flex"` |
| `altText` | string | Yes | Fallback text shown on clients that do not support Flex messages. Max 400 characters. |
| `contents` | object | Yes | Either a **bubble** container or a **carousel** container |

**Bubble container (single card):**

| Field | Notes |
|---|---|
| `type: "bubble"` | Single scrollable card |
| `header` | Optional box component at the top |
| `hero` | Optional image/video at the top of the card |
| `body` | Main content area (box component) |
| `footer` | Bottom area for actions (box component) |
| `styles` | Optional style overrides for header/hero/body/footer |
| `size` | Card width: `nano`, `micro`, `kilo`, `mega`, `giga` |

**Carousel container (multiple bubbles):**
```json
{
  "type": "carousel",
  "contents": [ { "type": "bubble", ... }, { "type": "bubble", ... } ]
}
```

Max **12 bubbles** per carousel.

**Key component types:**

| Component | `type` value | Purpose |
|---|---|---|
| Box | `"box"` | Layout container; `layout`: `horizontal`, `vertical`, `baseline` |
| Text | `"text"` | Text with rich styling (`weight`, `color`, `size`, `wrap`) |
| Button | `"button"` | Tappable button with an action |
| Image | `"image"` | Image with sizing and alignment options |
| Video | `"video"` | Video component |
| Icon | `"icon"` | Small inline icon (baseline layout only) |
| Separator | `"separator"` | Horizontal rule |
| Filler | `"filler"` | Flexible space |
| Span | `"span"` | Inline text with different styling (within a text component) |

**Constraints:**
- `altText` max: **400 characters**
- Maximum raw JSON size: **100,000 characters** per flex message body.
- Full Flex Message specification is extensive. See the dedicated [Flex Message reference](https://developers.line.biz/en/reference/messaging-api/#flex-message) for all component properties.
- Use the [Flex Message Simulator](https://developers.line.biz/flex-simulator/) in the developer tools to design and preview Flex messages.

---

### 4.11 Quick Reply

Quick reply is not a standalone message type — it is a property added to **any message object**. Quick reply buttons appear as tappable chips above the keyboard after the message is delivered. They disappear when tapped or when the user sends any other message.

> **Reference:** https://developers.line.biz/en/reference/messaging-api/#quick-reply

**JSON example (attached to a text message):**
```json
{
  "type": "text",
  "text": "Which service would you like?",
  "quickReply": {
    "items": [
      {
        "type": "action",
        "imageUrl": "https://example.com/icons/service1.png",
        "action": {
          "type": "postback",
          "label": "Service A",
          "data": "SERVICE_A"
        }
      },
      {
        "type": "action",
        "action": {
          "type": "camera",
          "label": "Open camera"
        }
      },
      {
        "type": "action",
        "action": {
          "type": "cameraRoll",
          "label": "Send photo"
        }
      },
      {
        "type": "action",
        "action": {
          "type": "location",
          "label": "Send location"
        }
      }
    ]
  }
}
```

**`quickReply` object:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `items` | array | Yes | 1–13 quick reply button objects |

**Quick reply button object:**

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | string | Yes | Always `"action"` |
| `imageUrl` | string | No | HTTPS URL to a small icon shown with the button. JPEG or PNG. 1:1 aspect ratio. Max 1 MB. |
| `action` | action object | Yes | The action to trigger when tapped |

**Supported action types for quick reply:**

| Action Type | Behavior |
|---|---|
| `postback` | Fires a postback event to your webhook |
| `message` | Sends a text message as the user |
| `uri` | Opens a URL |
| `datetimepicker` | Opens a date/time picker |
| `camera` | Opens the camera |
| `cameraRoll` | Opens the camera roll / photo library |
| `location` | Opens the location picker |

**Constraints:**
- Max **13** quick reply items per message.
- Button `label` max: **20 characters**.
- Quick replies are displayed only for the **last** message object in a multi-message send. If you send 3 messages in one request, only the third message's `quickReply` is shown.
- Quick reply buttons disappear immediately after the user taps one or sends any message — they are single-use and ephemeral.
- If a LINE client version does not support quick reply, only the message text is shown; the buttons are silently omitted.

---

## Common Error Response Structure

All Messaging API errors return a JSON body:

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

| Field | Type | Notes |
|---|---|---|
| `message` | string | Top-level error description |
| `details` | array | Present only when the error has sub-errors (e.g., validation failures) |
| `details[].message` | string | Description of the specific sub-error |
| `details[].property` | string | JSON path of the invalid property |

**Named error messages:**

| `message` value | Meaning |
|---|---|
| `"Invalid reply token"` | Reply token is expired or already used |
| `"Failed to send messages"` | Message send failed — e.g., user ID does not exist |
| `"You have reached your monthly limit."` | Monthly quota exhausted |
| `"The API rate limit has been exceeded. Try again later."` | Per-endpoint rate limit hit |
| `"The retry key is already accepted"` | Same `X-Line-Retry-Key` UUID already processed (`409`) |
| `"The request body has X error(s)"` | Validation error; see `details` array |
| `"Authentication failed due to the following reason: XXX"` | Auth failure; reason in message text |
| `"Access to this API is not available for your account"` | API not permitted for this account/plan |

**HTTP status code summary:**

| Status | Meaning |
|---|---|
| `200 OK` | Request successful |
| `202 Accepted` | Narrowcast accepted for asynchronous delivery |
| `400 Bad Request` | Malformed request, invalid parameters, validation error |
| `401 Unauthorized` | Missing or invalid channel access token |
| `403 Forbidden` | Not authorized; account plan restriction; target reach too low for demographic filter |
| `404 Not Found` | User/resource not found or user consent not given |
| `409 Conflict` | Retry key already accepted |
| `410 Gone` | Resource no longer available |
| `413 Payload Too Large` | Request body exceeds 2 MB |
| `415 Unsupported Media Type` | Unsupported file MIME type |
| `429 Too Many Requests` | Rate limit exceeded or monthly quota exhausted |
| `500 Internal Server Error` | LINE Platform internal error — safe to retry with the same `X-Line-Retry-Key` |

**Response headers:**

| Header | Notes |
|---|---|
| `X-Line-Request-Id` | Unique ID assigned to every request; useful for support inquiries |
| `X-Line-Accepted-Request-Id` | Included in `409` responses; the request ID of the originally accepted request |

---

## Production Checklist

- [ ] Channel access token type selected and stored securely (server-side only — never expose in client code or logs)
- [ ] For v2.1 tokens: private key stored in a secure secret manager; public key registered in the console; `kid` stored alongside the private key
- [ ] For stateless tokens: issuance happens per request or per short session; no revocation needed
- [ ] All send requests use `https://api.line.me/...` (not `http://`)
- [ ] `X-Line-Retry-Key` included on first attempt for push, multicast, broadcast, narrowcast (use a freshly generated UUID per logical operation)
- [ ] Reply tokens are consumed immediately on receipt — no queuing or delayed use
- [ ] `notificationDisabled` behavior tested and confirmed correct for use case
- [ ] Monthly quota tracking implemented — monitor `/v2/bot/message/quota/consumption` to avoid surprise `429`s
- [ ] Rate limits respected (especially narrowcast/broadcast at 60/hour)
- [ ] All media URLs are HTTPS and publicly accessible (no authentication, no cookies required)
- [ ] Template `altText` and Flex `altText` set to meaningful fallback text
- [ ] Error handling covers `400`, `401`, `403`, `409`, `429`, `500` — with retry only on `500`/timeout using the same retry key
- [ ] Webhook auto-reply and greeting message disabled in LINE Official Account Manager if bot handles these events
- [ ] Bot response mode set to **Bot** (not Chat) in LINE Official Account Manager
