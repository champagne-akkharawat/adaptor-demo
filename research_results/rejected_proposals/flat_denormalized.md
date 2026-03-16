# Proposal 4: Flat Denormalized / Analytics-Optimized

**Inspired by**: Segment Track events, Mixpanel event schema, BigQuery/Redshift columnar storage conventions.

**Philosophy**: No nesting. Every field lives at the top level with a dot-notation prefix that encodes its domain (`sender_`, `conv_`, `content_`). Designed for direct ingestion into a data warehouse or analytics pipeline without any transformation. SQL `WHERE sender_platform = 'slack'` works out of the box. Avoids the overhead of JSON path traversal in columnar stores.

```json
{
  "id":                     "01HZAB1234XYZABC",
  "platform":               "slack",
  "platform_message_id":    "slack_msg_ts_1710582000.123456",
  "idempotency_key":        "slack:slack_msg_ts_1710582000.123456",
  "direction":              "inbound",
  "occurred_at":            "2026-03-16T10:00:00Z",
  "received_at":            "2026-03-16T10:00:00.201Z",

  "sender_id":              "U12345",
  "sender_name":            "Bob",
  "sender_type":            "user",

  "conv_id":                "C08AB1234",
  "conv_type":              "channel",
  "conv_platform_id":       "C08AB1234",

  "content_kind":           "image",
  "content_text":           null,
  "content_url":            "https://cdn.yourhub.com/media/slack/F08XYZ.jpg",
  "content_mime_type":      "image/jpeg",
  "content_width":          1280,
  "content_height":         720,
  "content_caption":        null,
  "content_duration_secs":  null,
  "content_filename":       null,
  "content_size_bytes":     null,
  "content_lat":            null,
  "content_lng":            null,
  "content_location_label": null,
  "content_sticker_id":     null,
  "content_template_id":    null,
  "content_template_vars":  null,

  "reply_to_id":            null,
  "adaptor":                "slack-adaptor",
  "adaptor_version":        "1.2.0",
  "schema_version":         "messages/v1"
}
```

**Field naming convention:**

| Prefix | Domain |
|---|---|
| `sender_` | Who sent the message |
| `conv_` | Which conversation it belongs to |
| `content_` | The message content |
| *(none)* | Top-level envelope / routing fields |

**Pros:**
- Zero transformation required to load into BigQuery, Redshift, Snowflake, or ClickHouse
- Every field is directly filterable with SQL — no JSON extraction functions needed
- Flat structure is trivially serialisable to CSV, Parquet, or Avro
- No null-traversal issues — missing nested objects become null scalar fields

**Cons:**
- Nullable fields are unavoidable — every content field exists on every message, most are null
- Adding a new content type means adding new columns to the table (schema migration)
- Verbose for non-analytics consumers — they see many null fields they don't care about
- No type enforcement within `content_*` — `content_width` could be populated on a text message
