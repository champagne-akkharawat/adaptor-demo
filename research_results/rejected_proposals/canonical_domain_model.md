# Proposal 2: Canonical Domain Model (Rich Object Graph)

**Inspired by**: ActivityPub (W3C), XMPP message stanzas, enterprise canonical data model (CDM) pattern.

**Philosophy**: Model the *domain* richly. Every entity (User, Conversation, Message, Content) is a first-class object with its own identity. Favours completeness and queryability over minimalism. Suited for direct DB writes.

```json
{
  "message": {
    "id": "msg_01HZ9K...",
    "platform": "line",
    "platform_message_id": "abc123",
    "created_at": "2026-03-16T10:00:00Z",
    "received_at": "2026-03-16T10:00:01Z",
    "direction": "inbound",

    "sender": {
      "id": "usr_01HZ...",
      "platform_user_id": "Uabc123",
      "display_name": "Jane Doe",
      "avatar_url": "https://...",
      "account_type": "user"
    },

    "conversation": {
      "id": "conv_01HZ...",
      "platform_conversation_id": "Rxyz789",
      "type": "direct",
      "participants": []
    },

    "body": {
      "type": "text",
      "text": "Hello world",
      "language": "en",
      "media": null,
      "location": null,
      "template": null
    },

    "attachments": [],

    "thread": {
      "reply_to_id": null,
      "thread_id": null
    },

    "metadata": {
      "raw_platform_event": null,
      "adaptor_version": "1.0.0",
      "tags": []
    }
  }
}
```

**Key fields:**

| Field | Type | Description |
|---|---|---|
| `message.id` | ULID | Internal canonical ID |
| `message.platform` | enum | `line`, `slack`, `discord`, `whatsapp`, etc. |
| `message.platform_message_id` | string | Original ID from the platform (for dedup/idempotency) |
| `message.created_at` | RFC3339 | When the platform says the message was sent |
| `message.received_at` | RFC3339 | When the adaptor processed it |
| `message.direction` | enum | `inbound` \| `outbound` |
| `sender.account_type` | enum | `user`, `bot`, `system` |
| `body.type` | enum | `text`, `image`, `video`, `audio`, `file`, `location`, `sticker`, `template`, `unsupported` |
| `body.language` | BCP47? | Detected or declared language |
| `body.media` | object? | `{ url, mime_type, size_bytes, duration_secs, thumbnail_url }` |
| `body.location` | object? | `{ lat, lng, label }` |
| `body.template` | object? | `{ template_id, variables }` — for structured/rich messages |
| `metadata.raw_platform_event` | object? | Optional: store the original raw payload for debugging |

**Pros:**
- Both `id` (internal) and `platform_message_id` (external) are explicit — clean dedup story
- `created_at` vs `received_at` distinction is critical for ordering and latency monitoring
- `direction` field makes the same schema usable for outbound
- `metadata.raw_platform_event` allows lossless storage

**Cons:**
- Heavier object — more fields to populate/null out per adaptor
- Nested structure is more verbose to serialise onto a queue message
