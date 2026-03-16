# Proposal 6: OpenTelemetry Semantic Conventions-Aligned

**Inspired by**: OpenTelemetry Semantic Conventions, structured logging conventions (Datadog, Honeycomb), the Elastic Common Schema (ECS).

**Philosophy**: Treat every message event as an OTel span with structured attributes. Field names follow OTel dot-notation conventions (`messaging.system`, `messaging.destination`, `enduser.id`). The schema is designed to be ingested directly by an OTel collector as a log record or span event — meaning distributed traces, metrics, and message events live in the same observability backend with zero glue code.

```json
{
  "trace_id":    "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id":     "00f067aa0ba902b7",
  "timestamp":   "2026-03-16T10:00:00.312Z",
  "severity":    "INFO",
  "body":        "inbound message received from slack",

  "attributes": {
    "messaging.system":              "slack",
    "messaging.operation":           "receive",
    "messaging.message.id":          "slack_msg_ts_1710582000.123456",
    "messaging.message.body.size":   42,
    "messaging.destination.name":    "C08AB1234",
    "messaging.destination.kind":    "channel",

    "enduser.id":                    "U12345",
    "enduser.name":                  "Bob",
    "enduser.role":                  "user",

    "hub.message.id":                "01HZAB1234XYZABC",
    "hub.message.direction":         "inbound",
    "hub.message.idempotency_key":   "slack:slack_msg_ts_1710582000.123456",
    "hub.content.kind":              "image",
    "hub.content.url":               "https://cdn.yourhub.com/media/slack/F08XYZ.jpg",
    "hub.content.mime_type":         "image/jpeg",
    "hub.content.width":             1280,
    "hub.content.height":            720,

    "hub.adaptor.name":              "slack-adaptor",
    "hub.adaptor.version":           "1.2.0",
    "hub.reply_to_id":               null
  },

  "resource": {
    "service.name":    "adaptor-slack",
    "service.version": "1.2.0",
    "deployment.environment": "production"
  }
}
```

**Attribute namespaces:**

| Namespace | Source | Description |
|---|---|---|
| `messaging.*` | OTel Semantic Conventions | Standard messaging attributes — interops with any OTel consumer |
| `enduser.*` | OTel Semantic Conventions | Sender identity |
| `hub.*` | Custom (this system) | Hub-specific fields not covered by OTel conventions |
| `resource.*` | OTel Resource | Which service/deployment emitted this |

**Pros:**
- Spans, logs, and message events are all in the same format — one query in Honeycomb/Grafana covers all
- `trace_id` / `span_id` give native distributed tracing with zero extra work
- `messaging.*` attributes are already understood by Datadog, Jaeger, Zipkin, and Tempo
- Alerting and dashboards can be built on message volume, latency, and error rates without a separate analytics schema
- `resource.*` gives deployment context — which adaptor version processed this message

**Cons:**
- Flat `attributes` map loses type information — all values are strings/numbers, no nested objects
- OTel conventions don't cover CS-specific concepts (reply windows, template messages) — falls back to `hub.*` namespace
- `trace_id` / `span_id` require an active OTel SDK in the adaptor — adds a dependency
- Not a natural fit for a message queue payload — OTel is primarily a telemetry format, not a data format
