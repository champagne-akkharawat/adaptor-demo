# Research Plan: Instagram Message API (Graph API)

**Target file:** `research_results/platform_specifics/instagram/deep_research/message_api.md`

---

## Format Reference Files

- [facebook/deep_research/message_api.md](../../facebook/deep_research/message_api.md) — primary format model (section structure, field tables, JSON style, error table, cURL section)
- [instagram/deep_research/webhook.md](webhook.md) — same platform doc, confirms IGSID usage, document header format, and existing setup steps that sections 1–2 must not contradict
- [line/deep_research/message_api.md](../../line/deep_research/message_api.md) — confirms ToC anchor link convention and numbered sub-step pattern for setup sections

---

## Execution Steps

### Step 1 — Foundation (run in parallel)

Establish endpoint, permissions, and token flow before writing any sections.

| # | Type | Resource | Purpose |
|---|---|---|---|
| 1.1 | Fetch | `https://developers.facebook.com/docs/messenger-platform/instagram/` | Top-level IG Messaging overview, confirm supported features |
| 1.2 | Fetch | `https://developers.facebook.com/docs/messenger-platform/instagram/features/send-messages` | Endpoint overview, all supported message types for Instagram |
| 1.3 | Fetch | `https://developers.facebook.com/docs/instagram-platform/instagram-api-with-instagram-login/get-access-tokens-and-permissions` | Instagram-specific token acquisition flow and permission list |
| 1.4 | Fetch | `https://developers.facebook.com/docs/development/create-an-app/` | App creation steps (Business type requirement) |
| 1.5 | Search | `Instagram Messaging API instagram_manage_messages permission setup 2024 site:developers.facebook.com` | Confirm exact permission names, app type, linking steps |
| 1.6 | Search | `Meta System User Token never-expiring Instagram Messaging API setup site:developers.facebook.com` | System User Token flow for IG |
| 1.7 | Search | `Instagram Messaging API rate limits standard high volume tier 2024` | Rate limit tiers (standard vs. high-volume) |

> **Blocker:** If endpoint version differs from `v21.0`, update all subsequent sections before writing.

---

### Step 2 — Auth and App Review (run in parallel, after Step 1)

| # | Type | Resource | Purpose |
|---|---|---|---|
| 2.1 | Fetch | `https://developers.facebook.com/docs/permissions/reference/instagram_manage_messages` | Exact permission description, App Review requirements, advanced access gate |
| 2.2 | Fetch | `https://developers.facebook.com/docs/permissions/reference/instagram_basic` | Confirms this permission is needed alongside `instagram_manage_messages` |
| 2.3 | Fetch | `https://developers.facebook.com/docs/facebook-login/guides/access-tokens/get-long-lived` | Short → long-lived → never-expiring token exchange; confirm applies to Instagram |
| 2.4 | Search | `App Review instagram_manage_messages advanced access requirements business verification` | Which permissions require App Review, business verification gate |
| 2.5 | Search | `"Page Access Token" Instagram Messaging API long-lived never-expiring instagram_manage_messages` | Confirm token flow is identical to Facebook or IG-specific |
| 2.6 | Search | `Instagram Graph API "Authorization: Bearer" vs access_token query parameter difference` | Confirm both auth methods work; preference for server-side CS integrations |

---

### Step 3 — Send API Payloads (run all in parallel)

One search + one doc fetch per message type. All are independent of each other.

| Section | Fetch URL | Search Query |
|---|---|---|
| Endpoint overview | `https://developers.facebook.com/docs/messenger-platform/reference/send-api` | — |
| Text | `…/instagram/features/send-messages#text` | `Instagram Messaging API send text message POST me/messages recipient IGSID 2024` |
| Attachments | `…/instagram/features/send-messages#attachments` | `Instagram Messaging API send image video audio file attachment by URL site:developers.facebook.com` |
| Generic template | `…/send-messages/templates/generic` | `Instagram Messaging API generic template image title subtitle buttons site:developers.facebook.com` |
| Generic template (IG-specific) | `…/instagram/features/generic-template` | — |
| Product template | `…/instagram/features/product-template` | `Instagram Messaging API product template catalog site:developers.facebook.com` |
| Quick replies | `…/send-messages/quick-replies` | `Instagram Messaging API quick replies text user_phone_number user_email site:developers.facebook.com` |
| Sender actions | `…/instagram/features/send-messages#sender-actions` | `Instagram Messaging API sender_action typing_on typing_off mark_seen site:developers.facebook.com` |
| Ice breakers | `…/instagram/features/ice-breakers` | `Instagram Messaging API ice breakers configuration setup site:developers.facebook.com` |
| Persistent menu | `…/instagram/features/persistent-menu` | `Instagram Messaging API persistent menu configuration setup site:developers.facebook.com` |
| Private replies | `…/instagram/features/private-replies` | `Instagram Messaging API private reply to comment story mention 2024 site:developers.facebook.com` |
| Handover protocol | `https://developers.facebook.com/docs/messenger-platform/handover-protocol` | `Instagram Messaging API handover protocol pass_thread_control take_thread_control 2024` |

---

### Step 4 — Gap-fill and Validation (sequential, after Step 3)

| # | Type | Resource | Purpose |
|---|---|---|---|
| 4.1 | Search | `Instagram Messaging API limitations compared to Facebook Messenger API 2024` | Uncover capability gaps to document as constraints |
| 4.2 | Search | `Instagram-Scoped User ID IGSID vs PSID difference how to obtain` | Ensure recipient ID section clearly explains the difference |
| 4.3 | Search | `Instagram Messaging API v21 v22 changelog 2024 breaking changes` | Confirm whether v21.0 is current or if a newer version supersedes it |
| 4.4 | Fetch | `https://developers.facebook.com/docs/graph-api/overview/rate-limiting` | Graph API rate limiting general reference, per-IGSID limits |

---

### Step 5 — Write the Document

Write top-to-bottom in section order. Sections 1–2 first (setup/auth), then 3 (endpoint overview), then 4–12 (payload types), then 13–15 (errors, cURL, production).

---

## Document Structure

```
# Instagram Messaging API — Send Message (Graph API) Comprehensive Reference

> **Scope:** CS integration hub — outbound messages from an Instagram Professional Account.
> **Primary docs:** [list of source URLs]

## Table of Contents

1. Setup — Manual Steps
   1.1 Prerequisites
   1.2 Create a Meta App (Business type)
   1.3 Link Instagram Professional Account to a Facebook Page
   1.4 Enable Instagram Messaging in the App Dashboard
   1.5 Required Permissions
   1.6 App Review — instagram_manage_messages Advanced Access
   1.7 Business Verification Requirement
   1.8 Rate Limit Tiers

2. Authentication / Authorization
   2.1 Instagram-Scoped User ID (IGSID)
   2.2 Short-lived User Token
   2.3 Long-lived User Token
   2.4 Long-lived Page Access Token (never-expiring)
   2.5 Never-expiring System User Token
   2.6 access_token param vs. Authorization: Bearer header
   2.7 Business Verification gate

3. Send API — Endpoint Overview
   (endpoint, top-level fields table, messaging_type values, response envelope)

4. Text Messages

5. Attachments
   5.1 Image
   5.2 Video
   5.3 Audio
   5.4 File

6. Templates
   6.1 Generic Template
   6.2 Product Template (Catalog-Linked)

7. Quick Replies

8. Sender Actions

9. Profile Setup — Ice Breakers

10. Profile Setup — Persistent Menu

11. Handover Protocol
    11.1 Pass Thread Control
    11.2 Take Thread Control

12. Private Replies
    12.1 Reply to a Comment
    12.2 Reply to a Story Mention

13. Common Errors

14. Minimal cURL Example

15. Production Requirements
```

### Per-Section Content Notes

**Section 1 — Setup:** numbered step lists, one `> **Reference**: URL` blockquote per sub-section, UI paths where applicable.

**Section 2 — Auth:** each token type gets an HTTP example showing the exchange request; 2.1 explains IGSID vs. PSID clearly for CS hub implementers.

**Section 3 — Endpoint Overview:** top-level request body fields table (Field / Type / Required / Notes), `messaging_type` values with the 24-hour window rule noted, response envelope JSON.

**Sections 4–12 — Payloads:** each section follows this pattern:
1. Brief description
2. `> **Reference**: URL`
3. Full JSON example (with inline comments)
4. Field table (Field / Type / Required / Notes)
5. Constraints sub-section

**Sections 9–10 — Profile Setup:** framed as setup steps, not per-message sends. Include both the PUT (create/update) and DELETE (remove) endpoint examples.

**Section 12 — Private Replies:** document `comment_id` as the recipient field (not IGSID) for comment replies; clarify the recipient source from the inbound webhook event.

**Section 15 — Production Requirements:** checklist format (`- [ ]` items), not prose.

---

## Instagram-Specific Risks — Must Verify Before Writing

Do not assume parity with Facebook Messenger. Each of these must be resolved from a fetched source. If a feature is unsupported, document it explicitly with a "Not supported on Instagram" note and a reference URL — do not silently omit it.

| Risk Area | What to Verify | Impact if Wrong |
|---|---|---|
| Attachment upload API | Does `POST /me/message_attachments` work for IG, or is it URL-only? | Entire attachment section structure changes |
| Generic template elements | Max 1 element or carousel (multiple)? | JSON examples would be wrong |
| Product template structure | `product_elements[].id` vs. `elements[].id`? | Field table wrong |
| Quick reply `image_url` | Supported for IG quick replies? | Field table notes wrong |
| `request_thread_control` | Supported for IG handover? | Section 11 scope changes |
| Private reply window | 7 days for comments? 24h for story mentions? | Constraints section wrong |
| Ice breakers | Confirmed for IG or Facebook-only? | Section 9 may need "Not supported" note |
| Persistent menu | Confirmed for IG? Same `/me/messenger_profile` endpoint? | Section 10 structure |
| `MESSAGE_TAG` messaging_type | Works for IG send or Facebook-only? | Section 3 notes wrong |
| API version | Is `v21.0` still valid or superseded by `v22+`? | All endpoint URLs |

---

## Pre-Save Checklist

Before saving the final `message_api.md`, verify:

- [ ] All JSON examples use `IGSID` in comments (not `PSID`)
- [ ] All endpoint URLs use the version confirmed in Step 1 (not assumed `v21.0`)
- [ ] Every section has at least one `> **Reference**: URL` blockquote
- [ ] Field tables have exactly 4 columns: Field / Type / Required / Notes
- [ ] No section assumes a Facebook feature works on Instagram without a fetched source
- [ ] Section 12 explains `comment_id` vs. `IGSID` recipient clearly
- [ ] Sections 9 and 10 are framed as profile setup steps, not per-message sends
- [ ] Section 15 is a checklist, not prose
- [ ] Document opens with `> **Scope:**` and `> **Primary docs:**` block matching `webhook.md` header format
