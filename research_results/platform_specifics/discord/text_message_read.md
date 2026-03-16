# Discord — Receive an Incoming Text Message

## Two Delivery Methods (Mutually Exclusive)

Discord provides exactly two ways to receive messages. You must choose one:

| Method           | Mechanism                          | Best For                          |
|------------------|------------------------------------|-----------------------------------|
| **Gateway**      | Persistent WebSocket connection    | Bots that need full event streams |
| **HTTP Webhook** | Discord POSTs to your endpoint     | Serverless / interaction-only apps|

> "These two methods are mutually exclusive; you can only receive Interactions one of the two ways."
> — Discord docs

Note: the HTTP Webhook method (Interactions Endpoint URL) is primarily designed for slash-command interactions, not arbitrary message events. For receiving plain `MESSAGE_CREATE` events the Gateway is the standard approach.

---

## Method 1 — Gateway (MESSAGE_CREATE Event)

### Event Payload Structure

When any message is created Discord dispatches a `MESSAGE_CREATE` Gateway event (opcode 0):

```json
{
  "op": 0,
  "s": 42,
  "t": "MESSAGE_CREATE",
  "d": {
    "id": "1234567890123456789",
    "channel_id": "9876543210987654321",
    "guild_id": "1111111111111111111",
    "author": {
      "id": "2222222222222222222",
      "username": "someuser",
      "discriminator": "0",
      "avatar": "abc123"
    },
    "member": {
      "nick": "Server Nickname",
      "roles": ["..."]
    },
    "content": "Hello, bot!",
    "timestamp": "2026-03-16T12:00:00.000000+00:00",
    "edited_timestamp": null,
    "mentions": [],
    "embeds": []
  }
}
```

### Extracting Sender and Text

| Value Needed     | JSON Path               | Notes                                               |
|------------------|-------------------------|-----------------------------------------------------|
| Message text     | `d.content`             | Requires `MESSAGE_CONTENT` privileged intent.       |
| Sender user ID   | `d.author.id`           | Always present.                                     |
| Sender username  | `d.author.username`     | Always present.                                     |
| Guild member info| `d.member`              | Present only for guild messages, not DMs.           |
| Channel          | `d.channel_id`          | Use to reply via Create Message endpoint.           |
| Guild            | `d.guild_id`            | Absent for DMs and ephemeral messages.              |

### Gateway Intents Required

Intents are a bitwise integer sent in the Identify payload. For plain text messages:

| Intent            | Bit       | Value  | Purpose                                   |
|-------------------|-----------|--------|-------------------------------------------|
| `GUILD_MESSAGES`  | `1 << 9`  | 512    | MESSAGE_CREATE events in guild channels.  |
| `DIRECT_MESSAGES` | `1 << 12` | 4096   | MESSAGE_CREATE events in DMs.             |
| `MESSAGE_CONTENT` | `1 << 15` | 32768  | **Privileged.** Required to read `content`.|

Without `MESSAGE_CONTENT`, the `content`, `embeds`, `attachments`, and `components` fields are empty strings/arrays for all messages except:
- Messages the bot itself sent
- DMs sent directly to the bot
- Messages where the bot is mentioned

### Setup Steps — Gateway

1. **Enable intents in Developer Portal**: Go to your Application → Bot → Privileged Gateway Intents → enable "Message Content Intent".
2. **Fetch the Gateway URL**:
   ```http
   GET https://discord.com/api/v10/gateway/bot
   Authorization: Bot <BOT_TOKEN>
   ```
   Response: `{ "url": "wss://gateway.discord.gg" }`. Cache and reuse this URL.
3. **Open WebSocket connection**:
   ```
   wss://gateway.discord.gg/?v=10&encoding=json
   ```
4. **Receive Hello** (`op: 10`) — contains `heartbeat_interval` in milliseconds.
5. **Begin heartbeating** — send `op: 1` every `heartbeat_interval` ms.
6. **Send Identify** (`op: 2`):
   ```json
   {
     "op": 2,
     "d": {
       "token": "<BOT_TOKEN>",
       "intents": 33280,
       "properties": {
         "os": "linux",
         "browser": "my_bot",
         "device": "my_bot"
       }
     }
   }
   ```
   Intent value `33280` = `GUILD_MESSAGES (512)` + `DIRECT_MESSAGES (4096)` + `MESSAGE_CONTENT (32768)`.
7. **Receive Ready** (`t: "READY"`) — connection confirmed.
8. **Listen for `MESSAGE_CREATE`** — filter on `t === "MESSAGE_CREATE"` and read `d.content`.

### Authentication / Security

- The bot authenticates to Discord (not the other way around) via the `token` field in the Identify payload.
- The WebSocket connection itself is over TLS (`wss://`).
- No inbound signature verification is needed for Gateway connections; Discord is the server.

---

## Method 2 — HTTP Webhook (Interactions Endpoint)

This method is designed for **slash command interactions** rather than passive message reading, but is documented here for completeness.

### Interaction Object Structure (Received via POST)

Discord POSTs to your configured Interactions Endpoint URL:

```json
{
  "id": "interaction_snowflake",
  "type": 2,
  "data": {
    "name": "mycommand",
    "options": [
      { "name": "message", "value": "Hello from slash command" }
    ]
  },
  "user": { "id": "2222222222222222222", "username": "someuser" },
  "member": { "user": { "id": "..." } },
  "guild_id": "1111111111111111111",
  "channel_id": "9876543210987654321",
  "token": "interaction_token"
}
```

### Extracting Sender and Content (Interactions)

| Value Needed   | JSON Path                     | Context                    |
|----------------|-------------------------------|----------------------------|
| Sender (DM)    | `user.id`                     | Direct message context.    |
| Sender (guild) | `member.user.id`              | Guild context.             |
| Command text   | `data.options[n].value`       | Slash command parameters.  |
| Component data | `data.custom_id`, `data.values`| Button/select interactions.|
| Modal input    | `data.components[n].value`    | Modal submissions.         |

### Security Verification (Webhook Signature)

Webhook delivery requires HMAC-SHA256 request signature validation:

- Discord signs each request with your application's **Public Key**.
- Validate the `X-Signature-Ed25519` and `X-Signature-Timestamp` headers against the raw request body using Ed25519.
- Reject any request that fails verification with HTTP `401`.
- Discord sends a `PING` interaction (`type: 1`) on setup; respond with `{ "type": 1 }` to confirm the endpoint.

### Response Requirement

You must respond within **3 seconds**. The response callback endpoint is:

```
POST /interactions/{interaction.id}/{interaction.token}/callback
```

Interaction tokens remain valid for **15 minutes** for followup messages.

### Setup Steps — HTTP Webhook (Interactions)

1. Host a public HTTPS endpoint.
2. Implement Ed25519 signature verification on every incoming request.
3. Handle the initial `PING` (`type: 1`) with `{ "type": 1 }`.
4. In Developer Portal → your Application → General Information → set **Interactions Endpoint URL** to your endpoint.
5. Discord will verify the endpoint before saving.

---

## Key Notes

- For a general-purpose message-reading bot, **use the Gateway** (`MESSAGE_CREATE`). The Interactions Endpoint is for slash commands/components only.
- `MESSAGE_CONTENT` is a **privileged intent** — it must be explicitly enabled in the Developer Portal. Apps with 100+ servers additionally require approval during Discord's verification process.
- All IDs are snowflakes returned as strings. Never parse them as integers.
- Heartbeat maintenance is critical on the Gateway; failing to heartbeat causes a disconnect.
- On disconnect, use the `resume_gateway_url` and session ID from the Ready event to resume without replaying missed events from scratch.
