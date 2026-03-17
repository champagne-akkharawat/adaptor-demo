# TODO

## Context

This project is building a **multi-platform messaging adaptor** for the Aura Wellness platform. The goal is to allow Aura to send and receive messages across multiple consumer and enterprise messaging platforms (LINE, Facebook Messenger, Instagram, Discord, Microsoft Teams, Slack, Twitter/X, WhatsApp) through a single unified interface.

The adaptor normalises each platform's proprietary webhook payloads and REST APIs into a common canonical schema, so the Aura application layer never needs to know which platform a message came from or is going to.

The research phase documents how each platform handles each message type (receiving via webhook, sending via API) so we can design the canonical schema and write the adaptors accurately. The schema design phase produces and selects the unified message format that all adaptors will translate to and from.

## Research

### Platform API Research
- [x] LINE
  - [x] text_message_read
  - [x] text_message_write
  - [ ] deep_research
    - [x] webhook.md (setup, auth, all webhook event payload types)
    - [x] message_api.md (setup, auth, all send endpoints and message object types)
    - [x] Validate all finding against official docs via web fetch
    - [ ] Confirm all ⚠️ Spot-check
- [x] Facebook Messenger
  - [x] text_message_read
  - [x] text_message_write
  - [ ] deep_research
    - [x] webhook_reference.md (setup, auth, all webhook event payload types)
    - [x] message_api.md (setup, auth, all send endpoints and message object types)
    - [ ] Validate all findings against official docs via web fetch
    - [ ] Confirm all ⚠️ Spot-check items
- [x] Instagram
  - [x] text_message_read
  - [x] text_message_write
    - [x] webhook.md (setup, auth, all webhook event payload types)
    - [ ] message_api.md (setup, auth, all send endpoints and message object types)
- [x] Discord
  - [x] text_message_read
  - [x] text_message_write
- [x] Microsoft Teams
  - [x] text_message_read
  - [x] text_message_write
- [x] Slack
  - [x] text_message_read
  - [x] text_message_write
- [x] Twitter / X
  - [x] text_message_read
  - [x] text_message_write
- [ ] WhatsApp _(skipped for now)_
  - [ ] text_message_read
  - [ ] text_message_write

### Schema Design
- [x] Unified schema proposals (5 proposals documented in research_results/proposals/unified_schema.md)
- [ ] Select a schema proposal to adopt
