# Proposal 4: ActivityStreams 2.0 / Semantic Object Model

**Inspired by**: [W3C ActivityStreams 2.0](https://www.w3.org/TR/activitystreams-core/), ActivityPub, JSON-LD.

**Philosophy**: Messages are *semantic objects* with a formal vocabulary. An inbound message is a `Create` activity performed by an `Actor` on an `Object` (the message content) within a `Place` (the conversation). Meaning is encoded in the structure, not just field names. Suited for federation, open standards compliance, or systems that need to interop with Mastodon/ActivityPub ecosystems.

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://internal/events/01HZ9K...",
  "type": "Create",
  "published": "2026-03-16T10:00:00Z",

  "instrument": {
    "type": "Service",
    "name": "line-adaptor",
    "version": "1.0.0"
  },

  "actor": {
    "type": "Person",
    "id": "https://internal/users/line:Uabc123",
    "name": "Jane Doe",
    "icon": { "type": "Image", "url": "https://..." }
  },

  "target": {
    "type": "Group",
    "id": "https://internal/conversations/line:Rxyz789",
    "name": null
  },

  "object": {
    "type": "Note",
    "id": "https://internal/messages/line:msg123",
    "content": "Hello world",
    "mediaType": "text/plain",
    "inReplyTo": null,
    "attachment": []
  },

  "x-platform": {
    "name": "line",
    "message_id": "msg123",
    "direction": "inbound"
  }
}
```

**ActivityStreams type mappings for `object.type`:**

| Canonical content | AS2 type | Notes |
|---|---|---|
| Text message | `Note` | `content` holds the text |
| Image | `Image` | `url` holds the media URL |
| Video | `Video` | `url` + `duration` |
| Audio | `Audio` | `url` + `duration` |
| File | `Document` | `url` + `mediaType` + `name` |
| Location | `Place` | `latitude`, `longitude`, `name` |
| Sticker | `Emoji` | Non-standard; falls back to `Note` with `mediaType: image/webp` |
| Template/rich | `Article` | `content` is structured HTML or JSON summary |
| Unsupported | `Object` | Generic fallback |

**`x-platform` extension fields** (prefixed to avoid namespace collision):

| Field | Type | Description |
|---|---|---|
| `x-platform.name` | enum | Source platform |
| `x-platform.message_id` | string | Platform's original message ID (for dedup) |
| `x-platform.direction` | enum | `inbound` \| `outbound` |

**Pros:**
- Fully standardised vocabulary — zero ambiguity about what `actor`, `object`, `target` mean
- `id` is a URI, which gives globally unique, resolvable identifiers for free
- Natively extensible via JSON-LD `@context` without breaking existing consumers
- Future-proof: directly compatible with ActivityPub federation if needed

**Cons:**
- Verbose — `@context`, URI-based IDs, and nested objects add significant overhead
- AS2 vocabulary doesn't map cleanly to all message types (stickers, templates need workarounds)
- URI-based IDs are awkward for internal SQS/DB use without a resolver layer
- Steep learning curve; most backend engineers are not familiar with AS2
