# Discord — Send a Text Message

## Endpoint

```
POST /channels/{channel.id}/messages
```

Full URL (API v10):

```
https://discord.com/api/v10/channels/{channel.id}/messages
```

## Authentication

Pass a Bot token in the `Authorization` header:

```
Authorization: Bot <BOT_TOKEN>
```

OAuth2 bearer tokens are also accepted:

```
Authorization: Bearer <BEARER_TOKEN>
```

All requests must also include a `User-Agent` header in the form `DiscordBot ($url, $versionNumber)`. Omitting it risks Cloudflare blocking.

## Request Headers

| Header         | Value                          | Required |
|----------------|--------------------------------|----------|
| Authorization  | `Bot <BOT_TOKEN>`              | Yes      |
| Content-Type   | `application/json`             | Yes      |
| User-Agent     | `DiscordBot (<url>, <version>)`| Yes      |

## Request Body

At least one of `content`, `embeds`, `sticker_ids`, `components`, `files`, or `poll` must be present.

For a plain text message, `content` is the only field needed:

| Field              | Type    | Required | Notes                                         |
|--------------------|---------|----------|-----------------------------------------------|
| `content`          | string  | No*      | Message text. Max 2000 characters.            |
| `tts`              | boolean | No       | Send as text-to-speech. Default `false`.      |
| `embeds`           | array   | No       | Up to 10 rich embeds (6000 char combined).    |
| `allowed_mentions` | object  | No       | Controls which mentions trigger notifications.|
| `message_reference`| object  | No       | For replies or message forwards.              |
| `components`       | array   | No       | Interactive UI elements (buttons, etc.).      |
| `nonce`            | string  | No       | Idempotency string, max 25 characters.        |

*At least one payload field is required; `content` alone is sufficient for text.

Maximum request size: **25 MiB**.

## Minimal Example

```http
POST /channels/1234567890123456789/messages HTTP/1.1
Host: discord.com
Authorization: Bot MTk4NjIyNDgzNDcxOTI1MjQ4.Cl2FMQ.ZnCjm1XVW7vRze4b7Cq4se7kKWs
Content-Type: application/json
User-Agent: DiscordBot (https://example.com, 1.0.0)

{
  "content": "Hello from my bot!"
}
```

## Successful Response

HTTP `200 OK` with the created [Message object](https://docs.discord.com/developers/resources/message#message-object) as JSON, including the assigned `id`, `channel_id`, `author`, `content`, and `timestamp`.

## Key Notes

- **Channel ID**: The bot must have permission to send messages in the target channel. The channel ID is a snowflake returned as a string to avoid integer overflow.
- **Content limit**: `content` is capped at 2000 characters. Use `embeds` for longer structured content.
- **Rate limiting**: Discord rate-limits per route per bot. Respect `Retry-After` headers (RFC 6585). Repeated violations lead to API key revocation.
- **API version**: Always specify a version in the path (e.g., `/v10/`). v10 is current; v9 is available but older versions are deprecated or discontinued.
- **Snowflake IDs**: All IDs (channel, message, user) are 64-bit integers returned as strings. Parse them as strings, not integers.
- **No DM channel ID?** First create a DM channel via `POST /users/@me/channels` with the target user's ID, then use the returned channel `id`.
