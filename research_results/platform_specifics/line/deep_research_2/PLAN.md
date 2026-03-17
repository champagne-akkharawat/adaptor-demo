# Research Plan: LINE Messaging API — Full Redo (Round 2)

**Output directory:** `research_results/platform_specifics/line/deep_research_2/`

**Output files (3):**
- `webhook.md` — full webhook reference (setup, auth, all event types)
- `message_api.md` — full send reference (setup, auth, all endpoints + message types)
- `supplementary.md` — gaps not in round 1 (content retrieval, profiles, reliability, routing, errors)

The existing files in `deep_research/` are **not modified**. This is a clean redo.

---

## Format Reference Files

- [facebook/deep_research/message_api.md](../../facebook/deep_research/message_api.md) — primary format model (section structure, field tables, JSON style, error table, cURL section, production checklist)
- [facebook/deep_research/webhook_reference.md](../../facebook/deep_research/webhook_reference.md) — primary webhook format model

---

## Execution Steps

### Step A — Webhook Foundation (run all in parallel)

| # | Type | Resource | Purpose |
|---|---|---|---|
| A.1 | Fetch | `https://developers.line.biz/en/reference/messaging-api/index.html.md` | Raw markdown dump of full API reference — use as fallback if anchor fetches return nav-only |
| A.2 | Fetch | `https://developers.line.biz/en/docs/messaging-api/getting-started/` | Account creation and Messaging API activation steps |
| A.3 | Fetch | `https://developers.line.biz/en/docs/messaging-api/building-bot/` | Webhook URL setup, verification, TLS requirements, disabling auto-reply |
| A.4 | Fetch | `https://developers.line.biz/en/docs/messaging-api/verify-webhook-url/` | Webhook URL verification flow |
| A.5 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#signature-validation` | Signature verification — algorithm, header name, HMAC-SHA256 details |
| A.6 | Fetch | `https://developers.line.biz/en/docs/messaging-api/receiving-messages/` | Webhook envelope, event batching, delivery behavior, timeout, retry/redelivery |
| A.7 | Search | `LINE Messaging API webhook event types all 2025 site:developers.line.biz` | Confirm current complete list of webhook event types |

---

### Step B — All Webhook Event Payloads (run all in parallel, after A)

One fetch per event type group. Use A.1 raw markdown dump as fallback if any returns nav-only.

| # | Type | Resource | Purpose |
|---|---|---|---|
| B.1 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#message-event` | message event — envelope + source object |
| B.2 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#text-message` | message.text payload |
| B.3 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#image-message` | message.image payload |
| B.4 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#video-message` | message.video payload |
| B.5 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#audio-message` | message.audio payload |
| B.6 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#file-message` | message.file payload |
| B.7 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#location-message` | message.location payload |
| B.8 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#sticker-message` | message.sticker payload |
| B.9 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#follow-event` | follow event |
| B.10 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#unfollow-event` | unfollow event |
| B.11 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#join-event` | join event |
| B.12 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#leave-event` | leave event |
| B.13 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#member-join-event` | memberJoined event |
| B.14 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#member-leave-event` | memberLeft event |
| B.15 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#postback-event` | postback event |
| B.16 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#beacon-event` | beacon event |
| B.17 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#account-link-event` | accountLink event |
| B.18 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#things-event` | things event |
| B.19 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#unsend-event` | unsend event |
| B.20 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#video-viewing-complete` | videoPlayComplete event |
| B.21 | Search | `LINE Messaging API new webhook event types added 2024 2025 site:developers.line.biz` | Catch any new event types not in round 1 |

---

### Step C — Message API Auth and Send Endpoints (run all in parallel, after A)

Can run concurrently with Step B.

| # | Type | Resource | Purpose |
|---|---|---|---|
| C.1 | Fetch | `https://developers.line.biz/en/docs/basics/channel-access-token/` | All 4 token types overview — long-lived, short-lived, v2.1, stateless |
| C.2 | Fetch | `https://developers.line.biz/en/docs/messaging-api/generate-json-web-token/` | v2.1 JWT assertion signing key flow |
| C.3 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#send-reply-message` | Reply endpoint — full request/response spec |
| C.4 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#send-push-message` | Push endpoint |
| C.5 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#send-multicast-message` | Multicast endpoint |
| C.6 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#send-broadcast-message` | Broadcast endpoint |
| C.7 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#send-narrowcast-message` | Narrowcast endpoint + audience/recipient objects |
| C.8 | Fetch | `https://developers.line.biz/en/docs/messaging-api/sending-messages/` | Rate limits, monthly quota, notificationDisabled, retryKey behavior |

---

### Step D — All Message Object Types (run all in parallel, after C)

| # | Type | Resource | Purpose |
|---|---|---|---|
| D.1 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#text-message` | Text message object — full field spec |
| D.2 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#image-message` | Image message object |
| D.3 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#video-message` | Video message object |
| D.4 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#audio-message` | Audio message object |
| D.5 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#file-message` | File message object (send direction) |
| D.6 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#location-message` | Location message object |
| D.7 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#sticker-message` | Sticker message object |
| D.8 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#template-messages` | Template message overview — button, confirm, carousel, image carousel |
| D.9 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#imagemap-message` | Imagemap message object |
| D.10 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#flex-message` | Flex message object |
| D.11 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#quick-reply` | Quick reply object |
| D.12 | Search | `LINE Messaging API new message types added 2024 2025 site:developers.line.biz` | Catch any new message types not in round 1 |

---

### Step E — Supplementary Topics (run all in parallel, after A)

Can run concurrently with B, C, D.

| # | Type | Resource | Purpose |
|---|---|---|---|
| E.1 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#get-content` | GET /v2/bot/message/{messageId}/content — full spec |
| E.2 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#get-profile` | GET /v2/bot/profile/{userId} — full response schema |
| E.3 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#get-group-member-profile` | Group member profile endpoint |
| E.4 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#get-room-member-profile` | Room member profile endpoint |
| E.5 | Fetch | `https://developers.line.biz/en/reference/messaging-api/#error-responses` | Error response JSON structure and status codes |
| E.6 | Fetch | `https://developers.line.biz/en/docs/messaging-api/check-webhook-error-statistics/` | Webhook error cause codes (request_timeout etc.) |
| E.7 | Fetch | `https://developers.line.biz/en/docs/messaging-api/group-chats/` | Source-type routing — group vs room vs user, valid endpoints per type |
| E.8 | Search | `LINE Messaging API get message content expiry size limit preview thumbnail site:developers.line.biz` | Content expiry window, size constraints, preview endpoint |
| E.9 | Search | `LINE Messaging API source type group room push "to" field groupId userId bot private DM site:developers.line.biz` | Routing: correct `to` field, whether bot can DM group-met user |
| E.10 | Search | `LINE Messaging API error response body fields "message" "details" named error codes 400 403 429` | Named error codes / message strings in error body |

---

### Step F — Write the Three Documents (sequential, after all research steps)

Write in this order:
1. `webhook.md` — use Steps A + B results
2. `message_api.md` — use Steps C + D results
3. `supplementary.md` — use Steps A (reliability) + E results

---

## Document Structures

### webhook.md

```
# LINE Messaging API Webhook Integration

> Sources: [list]

## Table of Contents
1. Setup (Manual Steps)
   1.1 Create a LINE Official Account
   1.2 Enable the Messaging API
   1.3 Configure the Webhook URL
   1.4 Disable Auto-reply and Greeting Messages
2. Authentication — Signature Verification
   2.1 X-Line-Signature header
   2.2 HMAC-SHA256 verification algorithm
   2.3 Code example
3. Webhook Envelope
   3.1 Common envelope fields
   3.2 Source object (user / group / room variants)
4. Webhook Event Types and Payloads
   4.1  message — text
   4.2  message — image
   4.3  message — video
   4.4  message — audio
   4.5  message — file
   4.6  message — location
   4.7  message — sticker
   4.8  follow
   4.9  unfollow
   4.10 join
   4.11 leave
   4.12 memberJoined
   4.13 memberLeft
   4.14 postback
   4.15 beacon
   4.16 accountLink
   4.17 things
   4.18 unsend
   4.19 videoPlayComplete
```

### message_api.md

```
# LINE Messaging API — Outbound (Sending Messages) Comprehensive Reference

> Sources: [list]

## Table of Contents
1. Setup (Manual Steps)
   1.1 Retrieving a Channel Access Token
   1.2 Token Type Comparison
   1.3 Bot Settings That Affect Message Sending
   1.4 Rate Limits and Quota
2. Authentication
   2.1 Bearer token header
   2.2 Long-lived and short-lived tokens
   2.3 Channel access token v2.1 (JWT)
   2.4 Stateless token
3. Send Endpoints
   3.1 Reply message
   3.2 Push message
   3.3 Multicast message
   3.4 Broadcast message
   3.5 Narrowcast message
4. Message Object Types
   4.1  Text
   4.2  Image
   4.3  Video
   4.4  Audio
   4.5  File
   4.6  Location
   4.7  Sticker
   4.8  Template (button / confirm / carousel / image carousel)
   4.9  Imagemap
   4.10 Flex
   4.11 Quick reply
```

### supplementary.md

```
# LINE Messaging API — Supplementary Reference

> **Scope:** Media retrieval, user profiles, webhook reliability,
>   source-type routing, and error handling.
> **Primary docs:** https://developers.line.biz

## Table of Contents
1. Media / File Content Retrieval
   1.1 GET /v2/bot/message/{messageId}/content
   1.2 Preview / thumbnail endpoint (if exists)
   1.3 Content types, size limits, expiry window
   1.4 Auth requirements
2. User & Profile Retrieval APIs
   2.1 GET /v2/bot/profile/{userId}
   2.2 GET /v2/bot/group/{groupId}/member/{userId}
   2.3 GET /v2/bot/room/{roomId}/member/{userId}
   2.4 GET /v2/bot/followers/ids
   2.5 Privacy settings — what is redacted or absent
3. Webhook Reliability & Delivery Behavior
   3.1 Expected bot server response (status, body, timeout)
   3.2 Retry / redelivery policy
   3.3 Event batching — multiple events in one POST
   3.4 Ordering guarantees
   3.5 Idempotency and duplicate delivery (webhookEventId)
4. Source Type Routing (User / Group / Room)
   4.1 Source object schema — all three variants
   4.2 Send endpoint decision table
   4.3 Can the bot DM a user it met in a group?
   4.4 Endpoint restrictions by source type
5. Error Response Schema and Handling
   5.1 Error response body — full JSON structure
   5.2 HTTP status codes
   5.3 Named error codes / message strings
   5.4 Retry strategy — transient vs terminal
   5.5 Webhook error cause codes
```

---

## Per-Section Content Notes

**webhook.md §4 (event payloads):** Each event follows: brief description → `> **Reference**: URL` → full JSON example → field table (Field / Type / Required / Notes) → constraints.

**message_api.md §3 (send endpoints):** Each endpoint: use-case summary → full request body JSON → field table → response JSON → error table → key notes.

**message_api.md §4 (message types):** Each type: full JSON object → field table → constraints (character limits, file size limits, etc.).

**supplementary.md §1:** State explicitly whether response is binary stream or redirect. State content expiry window or "not documented."

**supplementary.md §3:** State explicitly "LINE **does / does not** retry." Document `webhookEventId` as deduplication key.

**supplementary.md §4:** Must include decision table:

| `source.type` | `to` field value | Recommended endpoint |
|---|---|---|
| `user` | `source.userId` | push |
| `group` | `source.groupId` | push |
| `room` | `source.roomId` | push |

Explicitly answer in prose: "A bot **can / cannot** send a private DM to a user it met only in a group."

**supplementary.md §5:** Error body as JSON first, then field table, then status code table (must cover 400, 401, 403, 404, 409, 429, 500). Retry strategy as two-column table.

---

## Pre-Save Checklist

### webhook.md
- [ ] Opens with sources block
- [ ] Setup steps include doc URL references per sub-step
- [ ] Signature verification includes HMAC-SHA256 pseudocode or code example
- [ ] All ~19 event types are documented
- [ ] Each event type has: JSON example + field table (Field / Type / Required / Notes)
- [ ] Source object documented for all three variants (user / group / room)

### message_api.md
- [ ] Opens with sources block
- [ ] Token comparison table covers all 4 types
- [ ] v2.1 JWT section includes JWT header + payload examples
- [ ] All 5 send endpoints documented with full request body + response
- [ ] `X-Line-Retry-Key` behavior documented under send endpoints
- [ ] All 11 message object types documented
- [ ] Each message type has field table + constraints
- [ ] Rate limits and quota tables present

### supplementary.md
- [ ] Opens with `> **Scope:**` and `> **Primary docs:**` block
- [ ] Every section has at least one `> **Reference**: URL` from `developers.line.biz`
- [ ] §1 states content expiry window (or "not documented")
- [ ] §1 states whether response is binary stream or redirect
- [ ] §2 field tables: 4 columns (Field / Type / Always Present / Notes)
- [ ] §3 explicitly answers yes/no on webhook redelivery + whether opt-in
- [ ] §4 contains source-type → endpoint decision table
- [ ] §4 explicitly answers yes/no on private DM to group-met user
- [ ] §5 error body shown as JSON example
- [ ] §5 status code table covers 400, 401, 403, 404, 409, 429, 500
- [ ] No unsourced factual claims in any section
