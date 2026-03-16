# Facebook Messenger Platform — Receiving a Text Message (Webhook)

## Webhook Event Structure

Meta delivers incoming messages as a `POST` request to your webhook endpoint with this JSON structure:

```json
{
  "object": "page",
  "entry": [
    {
      "id": "{PAGE_ID}",
      "time": 1458692752478,
      "messaging": [
        {
          "sender":    { "id": "{PSID}" },
          "recipient": { "id": "{PAGE_ID}" },
          "timestamp": 1458692752478,
          "message": {
            "mid":  "mid.1457764197618:41d102a3e1ae206a38",
            "text": "Hello, world!"
          }
        }
      ]
    }
  ]
}
```

## Extracting Sender and Text

| Data Point  | JSON Path                                  | Notes                                      |
|-------------|--------------------------------------------|--------------------------------------------|
| Sender ID   | `entry[0].messaging[0].sender.id`          | Page-Scoped ID (PSID) of the sending user  |
| Page ID     | `entry[0].messaging[0].recipient.id`       | Your Facebook Page ID                      |
| Message text| `entry[0].messaging[0].message.text`       | Plain text content of the message          |
| Message ID  | `entry[0].messaging[0].message.mid`        | Unique message identifier                  |
| Timestamp   | `entry[0].messaging[0].timestamp`          | Unix ms timestamp; use for ordering        |

> Note: `entry` and `messaging` are arrays. Always iterate over both — Meta may batch multiple entries or messaging events in a single delivery.

## Signature Verification

Meta signs every webhook payload using your **App Secret** and includes the signature in the request header:

```
X-Hub-Signature-256: sha256=<hex_digest>
```

### Verification Steps

1. Read the raw request body (before any JSON parsing).
2. Compute `HMAC-SHA256` of the raw body using your App Secret as the key.
3. Hex-encode the result (lowercase).
4. Compare it to the value in `X-Hub-Signature-256` (strip the `sha256=` prefix).
5. If they match, the payload is authentic; otherwise reject it with `403`.

### Important Detail

Meta generates the signature over an **escaped unicode** version of the payload (e.g., `ä` is represented as `\u00e4`). Ensure your comparison uses the same raw bytes that Meta signed.

### Example (Go)

```go
mac := hmac.New(sha256.New, []byte(appSecret))
mac.Write(rawBody)
expected := hex.EncodeToString(mac.Sum(nil))

received := strings.TrimPrefix(r.Header.Get("X-Hub-Signature-256"), "sha256=")
if !hmac.Equal([]byte(expected), []byte(received)) {
    http.Error(w, "invalid signature", http.StatusForbidden)
    return
}
```

## Webhook Setup

1. **Create an endpoint** that accepts both `GET` (verification) and `POST` (event delivery) requests.
2. **GET verification**: Meta sends `hub.mode`, `hub.verify_token`, and `hub.challenge` as query params. Confirm `hub.mode == "subscribe"` and `hub.verify_token` matches your configured token, then respond `200 OK` with the value of `hub.challenge`.
3. **Register the webhook** in the Meta App Dashboard under Messenger > Webhooks.
4. **Subscribe to fields**: Select the `messages` field to receive incoming text messages. Other available fields include `message_echoes`, `message_reads`, etc.
5. **HTTPS required**: Your endpoint must use a valid TLS certificate (self-signed certificates are not accepted).

## Responding to Webhook Events

- Return `200 OK` within **5 seconds** of receiving the payload.
- Perform signature verification and queue the payload for async processing if needed — do not block the response on downstream work.
- If Meta does not receive a `200 OK`, it will retry delivery. After 1 hour of failed retries, the webhook subscription is automatically disabled.

## Key Notes

- **Message ordering**: If multiple messages arrive out of sequence, use the `timestamp` field to reconstruct correct order.
- **`messages` vs `message_echoes`**: Subscribe to `messages` for customer-sent messages. `message_echoes` captures messages sent by your page, and is a separate subscription field.
- **Retry handling**: Implement idempotency using `message.mid` to avoid processing the same message twice on retried deliveries.
