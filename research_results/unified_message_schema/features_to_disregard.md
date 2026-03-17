# Unified Message Schema — Features to Disregard

> **Date:** 2026-03-17
> **Usecase:** CS hub integration — reactive 1-on-1 customer service inbox with agent replies
>
> **Source documents:**
> - LINE webhook events: [platform_specifics/line/deep_research_2/webhook.md](../platform_specifics/line/deep_research_2/webhook.md)
> - LINE send API: [platform_specifics/line/deep_research_2/message_api.md](../platform_specifics/line/deep_research_2/message_api.md)
> - Instagram webhook events: [platform_specifics/instagram/deep_research/webhook.md](../platform_specifics/instagram/deep_research/webhook.md)
> - Instagram send API: [platform_specifics/instagram/deep_research/message_api.md](../platform_specifics/instagram/deep_research/message_api.md)
> - Facebook webhook events: [platform_specifics/facebook/deep_research/webhook_reference.md](../platform_specifics/facebook/deep_research/webhook_reference.md)
> - Facebook send API: [platform_specifics/facebook/deep_research/message_api.md](../platform_specifics/facebook/deep_research/message_api.md)

Features listed here are deliberately excluded from the unified schema because they do not serve the CS reply workflow. They may still be logged as raw events for auditing, but the hub does not need to model, route, or act on them.

---

## Table of Contents

- [LINE — Disregard](#line--disregard)
- [Instagram — Disregard](#instagram--disregard)
- [Facebook — Disregard](#facebook--disregard)
- [Cross-Platform — Treat as Status Metadata, Not Events](#cross-platform--treat-as-status-metadata-not-events)

---

## LINE — Disregard

| Feature | Reason |
|---|---|
| **`beacon` events** | Requires physical LINE Beacon BLE hardware; irrelevant to CS messaging |
| **`things` events** | IoT device link/unlink/scenario; not a messaging CS concern; possibly deprecated |
| **`videoPlayComplete` event** | Tracks playback of *bot-sent* videos by tracking ID; not relevant for CS reply flows |
| **`membership` events** | LINE paid membership plan management; niche, separate product surface |
| **`join` / `leave` events** | Bot added/removed from group chats; CS hub is 1-on-1 focused |
| **`memberJoined` / `memberLeft` events** | Group chat lifecycle management; CS is 1-on-1 |
| **`accountLink` events** | Linking user's external service account to LINE; not a CS inbox feature |
| **Broadcast send endpoint** | Mass outbound campaigns; CS is 1-on-1 reactive replies |
| **Multicast send endpoint** | Same — batch sends to multiple users; not a CS reply |
| **Narrowcast send endpoint** | Same — demographic/audience targeting; not a CS reply |
| **Imagemap message type (outbound)** | Complex image hotspot layout; no analog on other platforms; too specialized |
| **`textV2` group-mention substitutions** | Group-only variant; plain text covers CS needs; adds complexity for zero gain |
| **Beacon reply token** | Follows from disregarding beacon events |

---

## Instagram — Disregard

| Feature | Reason |
|---|---|
| **`messaging_optins` / Recurring Notification tokens** | Marketing opt-in / subscription messaging; not CS reactive replies |
| **`messaging_referrals` (standalone event)** | Traffic attribution for ad/link tracking; log for analytics if needed but not a CS conversation feature |
| **Standby channel events** | Handover Protocol is deprecated on Instagram; Conversation Routing is static config, not API-driven |
| **Handover Protocol (`pass_thread_control` / `take_thread_control`)** | Deprecated on Instagram; migration is irreversible; omit from send layer |
| **Product Template (catalog-linked)** | Requires Commerce Manager catalog linkage; specialized e-commerce feature, not universal CS |
| **Ephemeral / view-once media** | No content delivered (`type: "ephemeral"`); log as a system event but no CS action possible |
| **Story mention attachment** | Inbound-only, URL expires, Meta policy prohibits storing media; surface as informational note only |
| **Ice Breakers setup endpoint** | One-time profile configuration, not per-conversation CS data |
| **Persistent Menu setup endpoint** | One-time profile configuration, not per-conversation CS data |
| **`messaging_seen` on bot-sent messages** | Read status on outbound; useful for UI tick marks but not a CS conversation event |

---

## Facebook — Disregard

| Feature | Reason |
|---|---|
| **`messaging_payments`** | Beta payments processing; out of CS scope |
| **`messaging_game_plays`** | Instant Games beta; not a CS platform feature |
| **`messaging_account_linking`** | Account linking flow for external identity; niche |
| **`messaging_feedback`** | Post-interaction NPS / Customer Feedback template; separate product feature |
| **`messaging_policy_enforcement`** | Platform enforcement webhook; handle as a system alert/log, not a CS conversation event |
| **`messaging_optins` / Recurring Notifications** | Marketing opt-in, not CS |
| **`messaging_referrals` (standalone event)** | Traffic attribution, not a CS conversation feature |
| **Receipt Template** | Very specific e-commerce order confirmation layout; too specialized to mandate in a unified CS reply |
| **Standby channel** | Handover Protocol state management; only relevant if the hub is itself a Secondary Receiver — treat as edge case, not a first-class feature |
| **`message_echoes` (outbound echo webhook)** | Echo of bot-sent messages; redundant if the hub tracks its own sends |
| **Persona API** | Custom sender name/avatar per message; potentially useful but adds schema complexity — defer to v2 |
| **`send_cart`** | Niche/beta field; not CS-relevant |
| **`messenger_template_status_update`** | Utility Message template review status; not CS inbox data |
| **`response_feedback`** | Feedback button clicks; separate analytics surface |

---

## Cross-Platform — Treat as Status Metadata, Not Events

These features exist on multiple platforms but should **not** be modelled as first-class conversation events. Instead, surface them as status updates on the originating message.

| Feature | Recommended treatment |
|---|---|
| **Read receipts** (`message_reads` FB / `messaging_seen` IG / no equivalent on LINE) | Update `read_at` on the stored outbound message; do not create a new conversation item |
| **Delivery receipts** (`message_deliveries` FB) | Update `delivered_at` on the stored outbound message; same as above |
| **Reaction events** (FB `message_reactions`, IG `reaction`) | Attach as metadata to the reacted-to message; not a replyable event in most CS workflows |
| **`unsend` events** (LINE `unsend`, IG/FB `is_deleted: true`) | Must delete/redact local copy per platform policy; surface as a tombstone marker on the message, not a new event |
