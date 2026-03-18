# Features to Consider — Schema Design Notes

> These features were deferred from the initial inbound/outbound schema prototypes.
> Each section covers: what the feature is, platform support, quirks, and a proposed queue schema.

---

## Table of Contents

- [Quick Reply](#quick-reply)
- [Postback](#postback)
- [Template](#template)

---

## Quick Reply

Quick reply chips are ephemeral action buttons that appear above the input box after a message is sent. When the user taps one, the chip disappears and the response is sent back as an inbound event.

**Direction:** Outbound (send) + Inbound (receive tap response)

### Platform Support

| Platform  | Outbound (send chips) | Inbound (receive tap) | Max chips | Chip types |
|-----------|----------------------|-----------------------|-----------|------------|
| LINE      | Yes                  | As postback or text   | 13        | `message`, `postback`, `datetimepicker`, `camera`, `cameraRoll`, `location` |
| Facebook  | Yes                  | As `messages` event with `quick_reply.payload` | 13 | `text`, `user_phone_number`, `user_email` |
| Instagram | Yes                  | As `messaging_postbacks` event                 | 13 | `text` only |

### Quirks

- **LINE does not preserve payload on `message`-type quick replies.** If the action type is `message`, the user's tap sends a regular text event with the label as the text — the `payload` string is lost. To preserve payload on LINE, use a `postback` action type on each chip. The inbound adaptor will then receive it as a `postback` event, not a `quick_reply_response`.
- **Facebook quick reply tap arrives as a `messages` event**, not a `messaging_postbacks` event. It has a `quick_reply.payload` field alongside the visible `text`.
- **Instagram quick reply tap arrives as a `messaging_postbacks` event**, unlike Facebook.
- Quick reply chips cannot be sent without an accompanying message (text, image, etc.) on any platform.

### Outbound Queue Schema

Quick reply chips are attached to any outbound message, not a standalone `message_type`. The `message_type` field reflects the accompanying message (e.g., `text`). The presence of a populated `quick_reply` block signals the adaptor to attach chips.

```json
"quick_reply": {
    "items": [
        {
            "type": "message",
            "label": "Order Status",
            "payload": "order_status"
        },
        {
            "type": "message",
            "label": "Track Shipment",
            "payload": "track_shipment"
        }
    ]
}
```

> **Note on `label` vs `payload`:** `label` is the human-readable text shown on the button. `payload` is the machine-readable string the consumer receives when the user taps. They should be distinct. The LINE caveat above means `payload` is not recoverable for `message`-type chips — use `postback`-type chips on LINE to preserve it.

### Inbound Queue Schema (`message_type: "quick_reply_response"`)

Used when a user taps a quick reply chip on Facebook. On LINE, the tap arrives as either `postback` (if postback action) or `text` (if message action) — there is no `quick_reply_response` event on LINE.

```json
{
    "direction": "inbound",
    "provider": "facebook|instagram",
    "message_type": "quick_reply_response",

    "quick_reply_response": {
        "label": "Order Status",
        "payload": "order_status",

        "meta_data": {
            "provider": "facebook|instagram",
            "content (FB)": {
                "text": "Order Status",
                "quick_reply": {
                    "payload": "order_status"
                }
            },
            "content (IG)": {
                "payload": "order_status"
            }
        }
    }
}
```

---

## Postback

A postback event is triggered when a user taps a button that has a postback action — either in a template (rich card), a quick reply chip (on LINE and Instagram), or the Get Started button. The button sends a developer-defined payload string back to the bot.

**Direction:** Inbound only — postbacks are received, never sent.

### Platform Support

| Platform  | Trigger sources | Payload field | Label available |
|-----------|----------------|---------------|----------------|
| LINE      | Template buttons, quick reply chips (postback action type), rich menu items | `postback.data` | No — only `data` |
| Facebook  | Template buttons, Get Started button | `postback.payload` | Yes — `postback.title` (button label) |
| Instagram | Template buttons, Get Started button | `postback.payload` | Yes — `postback.title` (button label) |

### Quirks

- **LINE `postback.data` can carry `params`** for date/time picker quick reply actions (e.g., `{ "date": "2026-04-01" }`). This is empty for regular postback buttons.
- **LINE postback events include a `replyToken`** — store it in `message_id.meta_data` alongside the event ID, same as message events.
- **Facebook postback label (`title`)** is the visible button text at time of tap — useful for display in the CS hub, but should not be used as logic input (user can see custom labels).

### Inbound Queue Schema (`message_type: "postback"`)

```json
{
    "direction": "inbound",
    "provider": "line|facebook|instagram",
    "message_type": "postback",

    "postback": {
        "label": "Buy Now",
        "payload": "action=buy&itemid=123",

        "meta_data": {
            "provider": "line|facebook|instagram",
            "content (FB+IG)": {
                "title": "Buy Now",
                "payload": "action=buy&itemid=123"
            },
            "content (Line)": {
                "data": "action=buy&itemid=123",
                "params": {}
            }
        }
    }
}
```

> `params` is only populated for LINE date/time picker actions — otherwise an empty object.
> `label` maps from `postback.title` on FB/IG and is unavailable on LINE (leave `null`).

---

## Template

Templates are structured rich messages sent outbound — cards with images, titles, text, and action buttons. They are the primary way to present product cards, menus, and confirmations.

**Direction:** Outbound only — users cannot send templates to the bot.

### Platform Support

| Template type     | LINE               | Facebook           | Instagram          |
|-------------------|--------------------|--------------------|--------------------|
| Buttons           | Yes (up to 4 btns) | Yes (up to 3 btns) | No                 |
| Confirm           | Yes (2 btns)       | No                 | No                 |
| Generic / Carousel| Yes (up to 10 cards, 3 btns each) | Yes (up to 10 cards, 3 btns each) | Yes (up to 10 cards, `web_url` + `postback` only) |
| Image Carousel    | Yes (image + 1 action per card) | No | No |
| Media Template    | No                 | Yes (image/video + 1 btn) | No |
| Flex Message      | Yes (custom layout) | No                | No                 |

### Button Types per Platform

| Button type | LINE               | Facebook           | Instagram          |
|-------------|--------------------|--------------------|---------------------|
| URL         | `uri` action       | `web_url`          | `web_url`           |
| Postback    | `postback` action  | `postback`         | `postback`          |
| Phone call  | `tel` action       | `phone_number`     | Not supported       |
| Share       | Not standard       | `element_share`    | Not supported       |

### Quirks

- **LINE Flex Message** is a completely different system from templates — it uses a custom JSON component tree (boxes, hero images, buttons, carousels). It has no cross-platform equivalent and should be treated as a separate `raw_payload`-style feature outside the normalized schema.
- **Instagram restricts button types** in Generic Template to `web_url` and `postback` only — no call or share buttons.
- **Facebook Media Template** supports a single image or video as the hero with up to 1 button — useful for product spotlight messages.
- **LINE Confirm template** is 2-button yes/no only — maps well to a `buttons` type card with 2 buttons but no image.
- Carousel cards on all platforms require a minimum of 1 button per card on LINE; Facebook/Instagram allow cards with no buttons.

### Outbound Queue Schema

A single normalized `template` block covers Buttons, Generic/Carousel, and Confirm across platforms. The adaptor maps `cards` count + button count to the appropriate platform template type.

```json
"template": {
    "type": "carousel|buttons|confirm",

    "cards": [
        {
            "image_url": "https://cdn.example.com/product.jpg",
            "title": "Classic White T-Shirt",
            "subtitle": "100% cotton. Available in S, M, L, XL.",
            "default_url": "https://shop.example.com/products/white-tshirt",

            "buttons": [
                {
                    "type": "url",
                    "label": "View Details",
                    "value": "https://shop.example.com/products/white-tshirt"
                },
                {
                    "type": "postback",
                    "label": "Add to Cart",
                    "value": "action=add_to_cart&product_id=wt001"
                },
                {
                    "type": "call",
                    "label": "Call Us",
                    "value": "+66812345678"
                }
            ]
        }
    ],

    "meta_data": null
}
```

**Field notes:**

| Field | Description |
|-------|-------------|
| `type` | `carousel` = multiple cards; `buttons` = single card with buttons + optional image; `confirm` = 2-button yes/no (LINE only) |
| `cards` | Array of 1–10 cards. Single card = Buttons template on FB; multiple = Generic/Carousel |
| `image_url` | Optional. Omit for text-only cards. |
| `title` | Required on LINE carousel. Optional on FB/IG. |
| `subtitle` | Maps to LINE `text`, FB/IG `subtitle`. |
| `default_url` | Maps to LINE's card-level `defaultAction` URI and FB's card-level `default_action`. Optional. |
| `button.type` | `url`, `postback`, `call`. Adaptor skips unsupported types per platform (e.g., drops `call` on Instagram). |
| `button.value` | URL for `url` type, payload string for `postback`, phone number for `call`. |

> **Adaptor responsibility:** The outbound adaptor must validate card count and button count against platform limits, and silently drop or truncate unsupported button types for the target platform.

> **LINE Flex Message:** Not covered by this normalized schema. If an agent needs to send a Flex Message, use `raw_payload` with the full Flex JSON. This keeps the normalized schema clean and avoids encoding LINE-specific layout concepts into the shared contract.
