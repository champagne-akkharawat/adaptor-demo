# System Architecture

## Table of Contents

- [Container Overview](#container-overview)
- [Database Schema](#database-schema)
- [Redis Key Space](#redis-key-space)
- [REST API Summary](#rest-api-summary)
- [WebSocket Event Summary](#websocket-event-summary)
- Sequence Diagrams
  - [Sequence: Agent Login](#sequence-agent-login)
  - [Sequence: Incoming Customer Message](#sequence-incoming-customer-message)
  - [Sequence: Agent Takes Ownership](#sequence-agent-takes-ownership)
  - [Sequence: Agent Opens a Conversation](#sequence-agent-opens-a-conversation)
  - [Sequence: WS Hub — view:open / view:close Handling](#sequence-ws-hub--viewopen--viewclose-handling)
  - [Sequence: WS Hub — Incoming message.new](#sequence-ws-hub--incoming-messagenew)
- [Redis Channel Partitioning — Design Comparison](#redis-channel-partitioning--design-comparison)

---

## Container Overview

```mermaid
graph TB
    subgraph External["External (upstream)"]
        EXT["Channel Platforms (LINE / IG / FB)"]
        UPSTR["Adaptor"]
        EXT -->|raw channel events| UPSTR
    end

    subgraph Infra["Infrastructure (Docker Compose)"]
        PG[("PostgreSQL 16")]
        RDB[("Redis 7")]
        SQS_IN["SQS aura-incoming"]
        SQS_OUT["SQS aura-outgoing"]
    end

    subgraph Backend["Backend (Go — :8080)"]
        API["HTTP Server"]
        WORKER["SQS Worker"]
        HUB["WebSocket Hub"]
    end

    subgraph Frontend["Frontend (Next.js — :3000)"]
        BROWSER["Browser — React + Zustand"]
    end

    UPSTR -->|unified SQS message| SQS_IN
    WORKER -->|polls| SQS_IN
    WORKER -->|writes messages & conversations| PG
    WORKER -->|publishes events| RDB

    API -->|reads/writes| PG
    API -->|session R/W| RDB
    API -->|publishes events| RDB
    API -->|enqueues outgoing replies| SQS_OUT

    HUB -->|subscribes events:global| RDB
    HUB -->|subscribes events:conversation:*| RDB

    BROWSER -->|REST + Bearer token| API
    BROWSER <-->|WebSocket wss://.../ws?token=| HUB
```

---

## Database Schema

```mermaid
erDiagram
    agents {
        uuid agent_id PK
        text email
        text password_hash
        text full_name
        text short_name
        timestamptz created_at
    }

    channels {
        serial channel_id PK
        text name
        text type
    }

    customers {
        uuid customer_id PK
        text full_name
        text short_name
        text phone
        text email
        text location
        timestamptz created_at
    }

    conversations {
        uuid conversation_id PK
        text status
        int channel_id FK
        uuid customer_id FK
        uuid owner_agent_id FK
        timestamptz created_at
        timestamptz updated_at
    }

    conversation_histories {
        uuid message_id PK
        uuid conversation_id FK
        text sender_type
        uuid sender_agent_id FK
        text sender_name
        text sender_short_name
        text content
        timestamptz created_at
    }

    conversations }o--|| channels : "via channel_id"
    conversations }o--|| customers : "via customer_id"
    conversations }o--o| agents : "owner_agent_id (nullable)"
    conversation_histories }o--|| conversations : "via conversation_id"
    conversation_histories }o--o| agents : "sender_agent_id (nullable)"
```

---

## Redis Key Space

| Key pattern | Type | TTL | Purpose |
|---|---|---|---|
| `session:{token}` | Hash (`agentId`, `fullName`, `shortName`) | 24 h | Auth session store |
| `events:global` | Pub/Sub channel | — | `conversation.new`, `conversation.updated` — broadcast to all agents |
| `events:conversation:{id}` | Pub/Sub channel | — | `message.new` — delivered only to Hub instances with an active viewer |

---

## REST API Summary

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/auth/login` | — | Authenticate agent; returns token + agent identity |
| `POST` | `/auth/logout` | Bearer | Invalidate session token |
| `GET` | `/conversations` | Bearer | List conversations (filter: status, channel, search, owned_by) |
| `GET` | `/conversations/{id}` | Bearer | Get single conversation metadata |
| `PUT` | `/conversations/{id}/ownership` | Bearer | Take or release ownership |
| `GET` | `/conversations/{id}/messages` | Bearer | Get message thread |
| `POST` | `/conversations/{id}/messages` | Bearer | Send a reply (enqueues to SQS outgoing) |
| `GET` | `/customers/{id}` | Bearer | Get customer profile |
| `GET` | `/channels` | Bearer | List connected channels |
| `GET` | `/ws` | token query param | Upgrade to WebSocket |
| `GET` | `/health` | — | Health check |

---

## WebSocket Event Summary

### Server → Client

All events share the envelope `{ "event": "string", "data": {} }`.

#### `conversation.new`
Trigger: new conversation created via ingest. Delivered to all connected agents.
```json
{
  "event": "conversation.new",
  "data": {
    "conversationId": "uuid",
    "status": "waiting | ongoing",
    "channel": "LINE | IG | FB",
    "customer": { "customerId": "uuid", "fullName": "string", "shortName": "string" },
    "owner": { "agentId": "uuid", "fullName": "string", "shortName": "string" },
    "lastMessage": { "preview": "string", "timestamp": "ISO 8601" }
  }
}
```
> `owner` is `null` when `status` is `"waiting"`.

#### `conversation.updated`
Trigger: ownership change or status change. Delivered to all connected agents.
```json
{
  "event": "conversation.updated",
  "data": {
    "conversationId": "uuid",
    "changes": {
      "status": "waiting | ongoing",
      "owner": { "agentId": "uuid", "fullName": "string", "shortName": "string" },
      "lastMessage": { "preview": "string", "timestamp": "ISO 8601" }
    }
  }
}
```
> Only changed fields are included in `changes`. `owner` may be `null` to indicate ownership was released.

#### `message.new`
Trigger: new message in a conversation. Delivered only to agents with that conversation open.
```json
{
  "event": "message.new",
  "data": {
    "conversationId": "uuid",
    "message": {
      "messageId": "uuid",
      "sender": { "type": "agent | customer", "fullName": "string", "shortName": "string" },
      "content": "string",
      "timestamp": "ISO 8601"
    }
  }
}
```

### Client → Server

| Event | Payload | Effect |
|---|---|---|
| `view:open` | `{conversationId}` | Hub subscribes to `events:conversation:{id}` if no other local viewer |
| `view:close` | `{conversationId}` | Hub unsubscribes from `events:conversation:{id}` if no remaining local viewers |

---

## Sequence: Agent Login

```mermaid
sequenceDiagram
    participant A as Agent Browser
    participant LC as localStorage
    participant API as REST API
    participant PG as PostgreSQL
    participant RD as Redis
    participant HUB as WebSocket Hub

    A->>API: POST /auth/login {email, password}
    API->>PG: SELECT agent WHERE email = ?
    PG-->>API: agent row (password_hash)
    API->>API: bcrypt.CompareHashAndPassword
    API->>RD: SET session:{uuid-token} {agentId, fullName, shortName} TTL 24h
    API-->>A: 200 {token, agent}

    A->>LC: localStorage.setItem("auth_token", token)
    A->>LC: localStorage.setItem("auth_agent", agent)
    A->>A: AuthContext state updated (token, agent)
    A->>A: router.push("/inbox")

    Note over A: InboxLayout mounts

    A->>API: GET /ws?token= (HTTP Upgrade)
    API->>RD: GET session:{token}
    RD-->>API: {agentId, fullName, shortName}
    API->>HUB: Register new Connection (agentId)
    API-->>A: 101 Switching Protocols
    Note over A,HUB: WebSocket open — agent receives conversation.new and conversation.updated from events:global

    par InboxColumn fetch
        A->>API: GET /conversations
        API->>PG: SELECT conversations JOIN customers, channels, agents
        PG-->>API: conversations[]
        API-->>A: 200 conversations[]
        A->>A: store.setConversations() — renders Waiting + All Chat lists
    and MyOngoingColumn fetch
        A->>API: GET /conversations?status=ongoing&owner=me
        API->>PG: SELECT conversations WHERE owner_agent_id = ? AND status = 'ongoing'
        PG-->>API: conversations[]
        API-->>A: 200 conversations[]
        A->>A: renders My Ongoing list
    end
```

---

## Sequence: Incoming Customer Message

```mermaid
sequenceDiagram
    participant CH as Channel Platform
    participant UP as Adaptor
    participant SQS as SQS aura-incoming
    participant WK as SQS Worker
    participant IS as IngestService
    participant PG as PostgreSQL
    participant RD as Redis
    participant HUB as WS Hub
    participant FE as Frontend (all agents)

    CH->>UP: channel-specific event
    UP->>SQS: publish {customerId, channelId, content}
    WK->>SQS: poll (long-poll)
    SQS-->>WK: message
    WK->>IS: ingestService.receive(payload)
    IS->>PG: findCustomer
    IS->>PG: findOrCreateConversation

    alt New conversation
        PG-->>IS: {conversation, isNew: true}
        IS->>PG: insertMessage
        IS->>RD: PUBLISH events:global {conversation.new}
        RD-->>HUB: events:global message
        HUB->>FE: WS push conversation.new
        FE->>FE: store.conversationPrepend()
    else Existing conversation
        PG-->>IS: {conversation, isNew: false}
        IS->>PG: insertMessage
        IS->>RD: PUBLISH events:conversation:{id} {message.new}
        RD-->>HUB: events:conversation:{id} message
        HUB->>FE: WS push message.new (only to agents viewing that conversation)
        FE->>FE: store.messageAppend()
    end

    WK->>SQS: delete message
```

---

## Sequence: Agent Takes Ownership

```mermaid
sequenceDiagram
    participant A as Agent Browser
    participant API as REST API
    participant PG as PostgreSQL
    participant RD as Redis
    participant HUB as WS Hub
    participant ALL as All Connected Agents

    A->>API: PUT /conversations/{id}/ownership {action:"take"}
    API->>PG: UPDATE conversations SET owner_agent_id=?, status='ongoing'
    API->>RD: PUBLISH events:global {conversation.updated, {status, owner}}
    API-->>A: 200 {conversationId, owner}

    RD-->>HUB: events:global message
    HUB->>ALL: WS push conversation.updated
    ALL->>ALL: store.conversationMerge() → re-render all inbox lists
```

---

## Sequence: Agent Opens a Conversation

```mermaid
sequenceDiagram
    participant A as Agent Browser
    participant API as REST API
    participant WS as WebSocket (same server)
    participant HUB as WS Hub
    participant RD as Redis

    A->>API: GET /conversations/{id}/messages
    API-->>A: messages[]
    A->>WS: send {event:"view:open", data:{conversationId}}
    WS->>HUB: handleViewOpen(conn, conversationId)

    alt First local viewer for this conversation
        HUB->>RD: SUBSCRIBE events:conversation:{id}
    end

    Note over HUB,RD: Hub now routes message.new events<br/>for this conversation to this agent
```

---

## Sequence: WS Hub — view:open / view:close Handling

```mermaid
sequenceDiagram
    participant C as Client
    participant HUB as Hub
    participant RD as Redis

    C->>HUB: view:open {conversationId}
    HUB->>HUB: convSubs[conversationId]++
    alt first local viewer (count == 1)
        HUB->>RD: SUBSCRIBE events:conversation:{conversationId}
    end

    C->>HUB: view:close {conversationId}
    HUB->>HUB: convSubs[conversationId]--
    alt no remaining local viewers (count == 0)
        HUB->>RD: UNSUBSCRIBE events:conversation:{conversationId}
    end
```

---

## Sequence: WS Hub — Incoming message.new

```mermaid
sequenceDiagram
    participant RD as Redis
    participant HUB as Hub (subscribeConversation goroutine)
    participant CA as Connection A (viewing)
    participant CB as Connection B (viewing)
    participant CC as Connection C (not viewing)

    RD->>HUB: message on events:conversation:{conversationId}
    HUB->>HUB: broadcastConversation(conversationId, payload)
    loop each registered connection
        alt activeConversationId == conversationId
            HUB->>CA: conn.send <- payload
            HUB->>CB: conn.send <- payload
        else activeConversationId != conversationId
            HUB-->>CC: skip
        end
    end
    CA->>CA: WritePump writes to WebSocket
    CB->>CB: WritePump writes to WebSocket
```

---

## Redis Channel Partitioning — Design Comparison

Full analysis: [`requirements/current/system_designs/redis_channel_partition.md`](requirements/current/system_designs/redis_channel_partition.md)

### Context

With the introduction of `business_units` (an agent belongs to multiple BUs; a message belongs to one BU; an agent can view all and only messages from their BUs), the choice of Redis Pub/Sub channel granularity becomes a meaningful architectural decision.

Two schemes are compared:

- **A — Conversation-scoped:** `events:conversation:{conversation_id}` (current implementation)
- **B — BU-scoped:** `events:bu:{bu_id}`

### Summary

| Dimension | A — Conversation-scoped | B — BU-scoped |
|---|---|---|
| Subscription lifecycle | Dynamic (view:open / view:close) | Static (connect / disconnect) |
| Active Redis channels | Up to N viewed conversations | Up to N BUs (small, fixed) |
| SUBSCRIBE/UNSUBSCRIBE churn | Every conversation open/close | Only on agent connect/disconnect |
| Delivery precision | Only to active viewer | All agents in BU |
| Access control alignment | Mismatched (BU rule ≠ conversation unit) | Exact match |
| Security isolation | Implicit; needs extra validation | Structurally enforced at subscribe |
| Backend complexity | Higher (convSubs, dynamic lifecycle) | Lower (buSubs, static per session) |
| Frontend complexity | Lower (only receives open convo) | Higher (must handle non-active convos) |
| Real-time inbox features | Not supported without extra work | Natively supported |
| Horizontal scaling efficiency | More surgical | Slightly more broadcast overhead |
| Hot-channel risk | Low (per conversation) | Possible for large, busy BUs |

### Recommendation

**Switch to `events:bu:{bu_id}`.**

The BU access control rule ("an agent always sees all messages from their BU") changes the fundamental requirement from "deliver only to the active viewer" to "deliver to all authorized agents." Scheme A was designed for the former; it actively fights the latter.

Scheme B eliminates the dynamic subscription lifecycle, aligns the Redis channel boundary with the access control boundary (making isolation correct by construction), and enables real-time inbox features (unread badges, live previews) that Scheme A cannot provide without additional infrastructure.

The tradeoff is modest: the frontend must handle `message.new` for conversations not currently open, and busy BUs produce slightly more cross-instance delivery overhead. Neither is a blocking concern at typical customer-service scale.

#### Channel layout after migration

| Channel | Subscribed by | Events |
|---|---|---|
| `events:global` | All hub instances (always-on) | `conversation.new`, `conversation.updated` |
| `events:bu:{bu_id}` | Hub instances with ≥1 agent in that BU | `message.new` |
