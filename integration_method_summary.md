# Messaging Platform API Integration Summary

| Platform | Method | Send Endpoint | Auth Type | Public URL Required | Sig. Verification | Documentation |
|---|---|---|---|---|---|---|
| Facebook Messenger | Webhook | `graph.facebook.com/v{v}/me/messages` | Page Access Token | Yes | Yes (HMAC-SHA256) | [developers.facebook.com/docs/messenger-platform](https://developers.facebook.com/docs/messenger-platform) |
| Instagram (Meta) | Webhook | `graph.facebook.com/v{v}/me/messages` | Page Access Token | Yes | Yes (HMAC-SHA256) | [developers.facebook.com/docs/messenger-platform/instagram](https://developers.facebook.com/docs/messenger-platform/instagram) |
| WhatsApp (Meta) | Webhook | `graph.facebook.com/v{v}/{phone-id}/messages` | System User Token | Yes | Yes (HMAC-SHA256) | [developers.facebook.com/docs/whatsapp/cloud-api](https://developers.facebook.com/docs/whatsapp/cloud-api/) |
| LINE | Webhook | `api.line.me/v2/bot/message/reply` | Channel Access Token | Yes | Yes (HMAC-SHA256) | [developers.line.biz/en/docs/messaging-api](https://developers.line.biz/en/docs/messaging-api/) |
| Telegram | Webhook or Long-poll | `api.telegram.org/bot{token}/sendMessage` | Bot Token | Optional | No | [core.telegram.org/bots/api](https://core.telegram.org/bots/api) |
| Slack | Webhook / Events API / Socket Mode | `slack.com/api/chat.postMessage` | OAuth 2.0 / Bot Token | Optional | Yes (HMAC-SHA256) | [docs.slack.dev/messaging](https://docs.slack.dev/messaging/) |
| Discord | Webhook or Gateway WebSocket | `discord.com/api/channels/{id}/messages` | Bot Token | Optional | Yes (Ed25519) | [docs.discord.com/developers/reference](https://docs.discord.com/developers/reference) |
| Twitter/X | Webhook or REST | `api.twitter.com/2/dm_conversations/...` | OAuth 1.0a / 2.0 | Yes (for webhooks) | Yes (CRC token) | [developer.x.com/en/docs/x-api/direct-messages](https://developer.x.com/en/docs/x-api/direct-messages/manage/introduction) |
| WeChat | Webhook | `api.weixin.qq.com/cgi-bin/message/...` | Access Token (AppID + Secret) | Yes | Yes (SHA1) | [developers.weixin.qq.com/doc/offiaccount/en](https://developers.weixin.qq.com/doc/offiaccount/en/Getting_Started/Overview.html) |
| KakaoTalk | Webhook | `kapi.kakao.com/v2/api/talk/memo/...` | OAuth 2.0 | Yes | No | [developers.kakao.com/docs/latest/en/kakaotalk-message/rest-api](https://developers.kakao.com/docs/latest/en/kakaotalk-message/rest-api) |
| Viber | Webhook | `chatapi.viber.com/pa/send_message` | Auth Token | Yes | Yes (HMAC-SHA256) | [developers.viber.com/docs/api/rest-bot-api](https://developers.viber.com/docs/api/rest-bot-api/) |
| Microsoft Teams | Webhook / Bot Framework | `smba.trafficmanager.net/...` | OAuth 2.0 (Azure AD) | Yes | Yes (HMAC-SHA256) | [learn.microsoft.com/microsoftteams/platform/bots/overview](https://learn.microsoft.com/en-us/microsoftteams/platform/bots/overview) |
| Google Chat | Webhook / Events API | `chat.googleapis.com/v1/spaces/{id}/messages` | OAuth 2.0 / Service Account | Yes | No | [developers.google.com/workspace/chat](https://developers.google.com/workspace/chat) |
| Zalo | Webhook | `openapi.zalo.me/v3.0/oa/message/...` | OAuth 2.0 / Access Token | Yes | Yes (HMAC-SHA256) | [developers.zalo.me/docs/api/official-account-api-230](https://developers.zalo.me/docs/api/official-account-api-230) |
| Twilio SMS/MMS | Webhook | `api.twilio.com/2010-04-01/Accounts/{id}/Messages` | API Key + Secret (Basic Auth) | Yes | Yes (HMAC-SHA256) | [twilio.com/docs/messaging/api](https://www.twilio.com/docs/messaging/api) |
| RCS (Google) | Webhook | `rcsbusinessmessaging.googleapis.com/...` | OAuth 2.0 / Service Account | Yes | Yes | [developers.google.com/business-communications/rcs-business-messaging](https://developers.google.com/business-communications/rcs-business-messaging/reference/rest) |
| Skype / Bot Framework | Webhook | `smba.trafficmanager.net/...` | OAuth 2.0 (Azure AD) | Yes | Yes | [learn.microsoft.com/azure/bot-service](https://learn.microsoft.com/en-us/azure/bot-service/) |
| Apple Messages for Business | Webhook (via MSP) | Via approved MSP only | OAuth 2.0 | Yes (via MSP) | Yes | [register.apple.com/resources/messages/msp-rest-api](https://register.apple.com/resources/messages/msp-rest-api/) |
| Kik | Webhook | `api.kik.com/v1/message` | API Key (Basic Auth) | Yes | No | [dev.kik.com](https://dev.kik.com/) |
| Matrix/Element | Client-Server API | `matrix.org/_matrix/client/v3/rooms/{id}/send/...` | Access Token | Optional | No | [spec.matrix.org/latest/client-server-api](https://spec.matrix.org/latest/client-server-api/) |

## Key Observations

- **Meta family** (Messenger, Instagram, WhatsApp) share almost identical integration patterns.
- **Asian platforms** (WeChat, KakaoTalk, Zalo, LINE) tend to require stricter business verification.
- **Enterprise platforms** (Teams, Slack, Google Chat) use OAuth 2.0 / service accounts.
- **Apple Messages** requires working through an approved Message Service Provider (MSP) — no direct API access.
- **Telegram** is the most developer-friendly: no app review, supports polling so no public URL needed.

## Common Architecture Pattern

```
User → Platform → Webhook POST → Your Server
Your Server → Platform REST API → User
```

1. **Inbound:** Platform pushes events to your webhook endpoint
2. **Outbound:** You call the platform's REST API to send messages
3. **Auth:** Token in `Authorization: Bearer <token>` header
4. **Verification:** Signature validation on incoming webhooks (HMAC-SHA256 typically)

> Last verified: March 2026
