# Handover — adaptor-demo

> **Date:** 2026-03-18
> **Reason:** Permanent desktop change

---

## What This Project Is

A **multi-platform messaging adaptor** for the Aura Wellness CS (customer-service) hub. The goal is a single unified interface for sending and receiving messages across consumer and enterprise platforms, so the Aura application layer never needs to know which platform it is talking to.

**Platforms in scope:** LINE, Facebook Messenger, Instagram, Discord, Microsoft Teams, Slack, Twitter/X
**WhatsApp:** explicitly deferred — skipped for now.

---

## Repository Layout

```
adaptor-demo/
├── TODO.md                          ← primary task tracker — read this first
├── line-adaptor/                    ← the only adaptor built so far (Go)
│   ├── main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── handler/webhook.go       ← HTTP entry point
│   │   ├── line/
│   │   │   ├── signature.go         ← HMAC-SHA256 + Base64 verification
│   │   │   ├── events.go            ← webhook event parsing
│   │   │   ├── reply.go             ← replyToken + push send
│   │   │   ├── messages/            ← per-type handlers (text, image, video, audio, file, location, sticker)
│   │   │   └── content/             ← LINE Content API client + transcoding poller
│   │   └── logger/
│   ├── tests/
│   └── docker-compose.yml
└── research_results/
    ├── platforms.md                 ← high-level platform comparison
    ├── platform_specifics/          ← per-platform deep research
    │   ├── line/deep_research_2/    ← most current LINE docs (webhook.md, message_api.md)
    │   ├── facebook/deep_research/  ← webhook_reference.md, message_api.md
    │   └── instagram/deep_research/ ← webhook.md, message_api.md
    └── unified_message_schema/
        ├── overall_analysis.md      ← cross-platform comparison + schema recommendation ← READ THIS
        ├── features_to_support.md   ← per-platform feature list (in/out) for CS use-case
        ├── features_to_disregard.md
        ├── secondary_api_calls.md   ← what must be fetched after the webhook (media, quotes)
        ├── unique_platform_features.md
        ├── prototypes/
        │   ├── inbound/             ← 33 concrete JSON examples of the unified inbound schema
        │   │   ├── _inbound_overview.json   ← canonical field map with meta_data for all providers
        │   │   └── _inbound_plan.md         ← naming convention + special case notes
        │   └── db_schema/
        │       ├── chat.dbml        ← messages, sessions, attachments (DBML format)
        │       └── platform.dbml    ← companies, staffs, roles, permissions, auth (DBML format)
        └── proposals/unified_schema.md  ← 3 schema proposals with pros/cons
```

---

## Where Things Stand

### Research — Done

| Platform | Basic research | Deep research |
|---|---|---|
| LINE | done | done (`deep_research_2/`) — spot-checks still needed (see TODO) |
| Facebook Messenger | done | done — validation against official docs pending |
| Instagram | done | done |
| Discord, Teams, Slack, Twitter/X | done (basic) | not started |

### Schema — In Progress

The **unified inbound message shape** has been designed and prototyped:

- **33 concrete JSON examples** covering all message types × providers (LINE, Facebook, Instagram) are in `prototypes/inbound/`.
- The **database schema** (DBML) is designed — two PostgreSQL schemas: `platform` and `chat`.
- **A schema proposal has NOT been selected yet.** Three options were evaluated (see `proposals/unified_schema.md`). The recommended direction from `overall_analysis.md` is a **two-layer approach**:
  1. Normalized layer — platform-agnostic fields for CS inbox display
  2. Platform envelope — raw webhook payload preserved for rich replies

### The LINE Adaptor — Working

A real Go HTTP service that:
- Verifies LINE webhook signatures
- Parses all message types (text, image, video, audio, file, location, sticker)
- Fetches LINE-hosted media via the Content API (including transcoding poll for video)
- Replies via `replyToken` and falls back to push
- Logs raw + parsed events to `line-adaptor/logs/`

No Facebook, Instagram, or other adaptors have been built yet.

---

## The Most Important Outstanding Decision

**Select a schema proposal.** Everything downstream (adaptor output format, DB write logic, queue contract) depends on this. See [research_results/proposals/unified_schema.md](research_results/proposals/unified_schema.md) and [research_results/unified_message_schema/overall_analysis.md](research_results/unified_message_schema/overall_analysis.md).

The recommended inbound shape from `overall_analysis.md`:

```
UnifiedMessage {
  id, platform, channel_id, sender_id, timestamp, direction
  type, text, attachments[]
  reply_token (LINE only), reply_to_mid
  raw (original webhook event)
}
```

---

## Critical Implementation Rules (Do Not Forget)

| Rule | Detail |
|---|---|
| **Download FB/IG media immediately** | CDN URLs are pre-signed and expire (~1 hour). Treat the webhook as a trigger to enqueue a download job. |
| **Check LINE video transcoding before download** | Poll `GET /v2/bot/message/{id}/content/transcoding` until `status: "succeeded"` before fetching the binary. |
| **LINE media has no URL in the webhook** | Only a `messageId` is provided. Must call `GET /v2/bot/message/{id}/content`. |
| **LINE quoted messages are not re-fetchable** | Store original message content on first receipt — there is no API to retrieve past messages by ID. |
| **IG story media must not be stored** | Per Meta policy, only the URL reference may be stored, not the binary. |
| **Track `last_user_message_at` per conversation** | FB/IG enforce a 24-hour reply window. `HUMAN_AGENT` extends this to 7 days (human-sent only). LINE has no window. |
| **Signature verification needs raw body bytes** | Read body before JSON parsing. LINE: `x-line-signature` HMAC-SHA256 + Base64. FB/IG: `X-Hub-Signature-256` HMAC-SHA256 + `sha256=` hex. |
| **Webhook batching** | A single POST may carry multiple events. LINE: `events[]`. FB/IG: `entry[].messaging[]`. Iterate all arrays. |
| **Reactions and read receipts are not new messages** | Update status/metadata on the stored outbound message; do not create a new conversation event. |

---

## Webhook Signature Differences

| Platform | Header | Encoding |
|---|---|---|
| LINE | `x-line-signature` | HMAC-SHA256(channel_secret, body) → Base64 |
| Facebook | `X-Hub-Signature-256` | HMAC-SHA256(app_secret, body) → `sha256=` + hex |
| Instagram | `X-Hub-Signature-256` | Same as Facebook |

---

## Database Schema Summary (DBML)

Two PostgreSQL schemas. Cross-schema foreign keys are stored IDs only (enforced at application level, not DB level).

**`platform` schema** — org structure and auth
`companies` → `business_units` → `staffs`, `roles`, `staff_business_units`, `permission_groups`, `permissions`, `role_permissions`, `refresh_tokens`, `channel_groups`

**`chat` schema** — messaging
`providers` → `channels` (linked to `platform.channel_groups`) → `sessions` → `session_messages` → `messages`
`customers` ← `provider_customers` (stores platform-scoped IDs: LINE `userId`, FB `PSID`, IG `IGSID`)
`messages` → `message_attachments`
`provider_messages` (1-to-1 with `messages`) — stores raw webhook payload + provider-native field breakdown
`provider_message_attachments` — provider-native attachment metadata
`agent_channel_permissions` — access control per agent per channel

---

## Tooling

- **Language:** Go 1.24+
- **Schema format:** DBML (use [dbdiagram.io](https://dbdiagram.io) to visualise)
- **LINE adaptor:** plain `net/http`, no framework

---

## Next Actions (from TODO.md)

1. **Select a schema proposal** — unblocks all adaptor implementation work
2. **LINE deep_research_2:** resolve outstanding ⚠️ spot-check items
3. **Facebook deep_research:** validate findings against official docs; resolve ⚠️ items
4. **Instagram:** complete `message_api.md` deep research
5. **Build Facebook and Instagram adaptors** using the LINE adaptor as a reference
