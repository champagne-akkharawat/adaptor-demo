# Slack — Receiving a Text Message (Events API)

## Overview

Incoming messages are delivered via the **Events API**. Slack sends an HTTP POST request containing a JSON payload to your configured endpoint each time a subscribed event occurs.

---

## Event Payload Structure

Every Events API delivery shares a common outer envelope:

```json
{
  "token": "<verification_token>",
  "team_id": "T123ABC456",
  "api_app_id": "A123ABC456",
  "event": { ... },
  "type": "event_callback",
  "event_id": "Ev123ABC456",
  "event_time": 1609459200,
  "authorizations": [
    {
      "enterprise_id": null,
      "team_id": "T123ABC456",
      "user_id": "U123ABC456",
      "is_bot": true
    }
  ]
}
```

### Inner `event` object for a channel message

```json
{
  "type": "message",
  "channel": "C123ABC456",
  "user": "U987XYZ321",
  "text": "Hello from a user!",
  "ts": "1609459200.000100",
  "event_ts": "1609459200.000100",
  "channel_type": "channel"
}
```

The `type` field in the outer envelope is always `event_callback` for real events. The inner `event.type` identifies the specific event (e.g. `message`).

---

## Extracting Sender and Text

| Data | JSON path | Example value |
|---|---|---|
| Sender user ID | `event.user` | `"U987XYZ321"` |
| Message text | `event.text` | `"Hello from a user!"` |
| Channel ID | `event.channel` | `"C123ABC456"` |
| Message timestamp | `event.ts` | `"1609459200.000100"` |
| Workspace ID | `team_id` | `"T123ABC456"` |

> Note: `event.user` is not present on all event types — only on events triggered by a user action. Bot-posted messages will have a `bot_id` field instead of `user`.

---

## Signature Verification

All incoming requests from Slack must be verified using HMAC-SHA256 before processing.

### Headers used

| Header | Description |
|---|---|
| `X-Slack-Signature` | Slack's computed signature, formatted as `v0=<hex_digest>` |
| `X-Slack-Request-Timestamp` | Unix epoch timestamp when Slack sent the request |

### Verification steps

1. **Replay-attack check** — reject the request if `X-Slack-Request-Timestamp` is more than 5 minutes (300 seconds) away from your current server time.

2. **Construct the basestring** by joining the following three parts with `:`:
   ```
   v0:{X-Slack-Request-Timestamp}:{raw_request_body}
   ```
   Example:
   ```
   v0:1531420618:token=xyzz0WbapA4vBCDEFasx0q6G&team_id=T123ABC456&...
   ```
   Use the **raw, unparsed request body** — not a re-serialised version.

3. **Compute HMAC-SHA256** of the basestring using your app's **Signing Secret** as the key.

4. **Format the computed signature** as `v0=<lowercase_hex_digest>`.

5. **Compare** the computed signature against the `X-Slack-Signature` header using a **constant-time equality function** (not a standard string comparison) to prevent timing attacks.

6. **Accept** the request only if the signatures match.

### Pseudocode

```python
import hmac, hashlib, time

def verify_slack_request(signing_secret, timestamp, body, slack_signature):
    # Step 1: Replay-attack guard
    if abs(time.time() - int(timestamp)) > 300:
        return False  # stale request

    # Step 2 & 3: Build basestring and compute HMAC-SHA256
    basestring = f"v0:{timestamp}:{body}"
    computed = "v0=" + hmac.new(
        signing_secret.encode(),
        basestring.encode(),
        hashlib.sha256
    ).hexdigest()

    # Step 5: Constant-time comparison
    return hmac.compare_digest(computed, slack_signature)
```

The **Signing Secret** is found in your app's settings under **Basic Information > App Credentials > Signing Secret**.

---

## URL Verification Challenge

When you first configure (or update) your Events API endpoint, Slack sends a one-time verification request:

```json
{
  "token": "<verification_token>",
  "challenge": "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P",
  "type": "url_verification"
}
```

Your endpoint must:

1. Respond with HTTP `200 OK` **within 3 seconds**.
2. Return the `challenge` value verbatim — either as plain text or as JSON:
   ```json
   { "challenge": "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P" }
   ```

---

## Setup Steps

1. **Create a Slack app** at https://api.slack.com/apps if you have not already.
2. In the app settings, go to **Event Subscriptions** and toggle it **on**.
3. Choose a delivery method:
   - **HTTP endpoint** — Slack POSTs events to a public HTTPS URL you host.
   - **Socket Mode** — events are delivered over a persistent WebSocket (no public URL needed; useful for development).
4. Enter your **Request URL** (HTTP mode). Slack will immediately send the URL verification challenge; your endpoint must respond correctly to save the setting.
5. Under **Subscribe to Bot Events** (or Workspace Events), add the relevant event type:
   - `message.channels` — public channel messages
   - `message.groups` — private channel messages
   - `message.im` — direct messages
   - `message.mpim` — group direct messages
6. **Save** changes and **reinstall** the app to the workspace so the new scopes take effect.
7. Ensure the corresponding OAuth scope is granted:
   - `channels:history` for `message.channels`
   - `groups:history` for `message.groups`
   - `im:history` for `message.im`
   - `mpim:history` for `message.mpim`

---

## Event Delivery Requirements

- Respond with HTTP `200` within **3 seconds** of receiving any event.
- If processing takes longer, acknowledge immediately (return `200`) and queue the work asynchronously.
- Slack retries failed deliveries **3 times** with exponential backoff.
- If your success rate drops below **5% in a 60-minute window**, Slack will automatically disable your event subscriptions and notify you via email.
