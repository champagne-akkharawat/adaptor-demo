# Proposal 5: Protobuf-Aligned / Wire-Format Schema

**Inspired by**: Protocol Buffers 3 field design, gRPC service contracts, Twirp conventions.

**Philosophy**: The JSON schema is a direct shadow of a `.proto` definition. Every field has an implied number (by position), every type maps to a proto scalar or message type, and `oneof` semantics are enforced by the `content` variant. Designed for systems that need to transition from JSON to binary Protobuf serialisation without a schema redesign, or that use JSON as a human-readable debug representation of proto messages.

```json
{
  "id":                   "01HZAB1234XYZABC",
  "platform":             2,
  "platform_message_id":  "slack_msg_ts_1710582000.123456",
  "idempotency_key":      "slack:slack_msg_ts_1710582000.123456",
  "direction":            1,
  "timestamp_ms":         1710582000312,

  "sender": {
    "id":   "U12345",
    "name": "Bob",
    "type": 1
  },

  "conversation": {
    "id":   "C08AB1234",
    "type": 3
  },

  "content": {
    "text":  null,
    "image": {
      "url":       "https://cdn.yourhub.com/media/slack/F08XYZ.jpg",
      "mime_type": "image/jpeg",
      "width":     1280,
      "height":    720,
      "caption":   ""
    },
    "video":    null,
    "audio":    null,
    "file":     null,
    "location": null,
    "sticker":  null,
    "template": null
  },

  "reply_to_id": ""
}
```

**Enum values (integer wire format):**

| Field | Values |
|---|---|
| `platform` | `1=discord, 2=slack, 3=teams, 4=line, 5=facebook, 6=instagram, 7=twitter, 8=whatsapp` |
| `direction` | `1=inbound, 2=outbound` |
| `sender.type` | `1=user, 2=bot, 3=system` |
| `conversation.type` | `1=direct, 2=group, 3=channel` |

**Corresponding `.proto` definition:**

```proto
syntax = "proto3";

message CanonicalMessage {
  string id                   = 1;
  Platform platform           = 2;
  string platform_message_id  = 3;
  string idempotency_key      = 4;
  Direction direction         = 5;
  int64 timestamp_ms          = 6;
  Sender sender               = 7;
  Conversation conversation   = 8;
  Content content             = 9;
  string reply_to_id          = 10;
}

message Content {
  oneof kind {
    TextContent     text     = 1;
    ImageContent    image    = 2;
    VideoContent    video    = 3;
    AudioContent    audio    = 4;
    FileContent     file     = 5;
    LocationContent location = 6;
    StickerContent  sticker  = 7;
    TemplateContent template = 8;
  }
}
```

**Pros:**
- 1:1 correspondence between JSON and `.proto` — switching to binary serialisation requires no schema redesign
- Integer enums reduce payload size and eliminate string comparison in hot paths
- `oneof` in proto enforces exactly-one content variant at the serialiser level, not just by convention
- gRPC service contracts can be generated directly from the same `.proto`
- Compact: proto3 binary is typically 3–10× smaller than equivalent JSON

**Cons:**
- Integer enums are opaque in raw JSON logs — requires a lookup table to interpret
- `timestamp_ms` (int64 epoch) is less human-readable than RFC3339
- Empty string `""` as proto3's zero-value for strings is semantically ambiguous vs. `null`
- Consumer teams unfamiliar with proto conventions will find the integer enums surprising
