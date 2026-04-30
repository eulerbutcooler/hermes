# Hermes — Project Breakdown

---

## 1. What Is Hermes?

Hermes is a **self-hosted workflow automation platform** — think Zapier or Make.com, but one you own and run yourself.

The core concept is a **Relay**: a named workflow that has one trigger and one or more action nodes wired together in a graph. When the trigger fires, Hermes executes all the action nodes in the correct order, resolves secret values on the fly, substitutes template expressions in configs, persists a full execution audit trail, and ACKs or NAKs the original message depending on the outcome.

**What a user can do:**
- Create relays with a webhook, cron schedule, or manual button as the trigger.
- Wire actions together — send to Discord, Slack, an HTTP endpoint, or email — with the output of one step available as input to the next via `{{steps['nodeId'].output.field}}` templates.
- Store encrypted secrets and reference them in action configs so plaintext credentials never touch the workflow definition.
- View the full execution history per relay, including per-step input/output and timing.
- Connect OAuth providers (Google, Microsoft) so email actions can send on behalf of the user's account.

---

## 2. High-Level Architecture

```
┌────────────────────────────────────────────────────────────┐
│                        Browser / Client                    │
│                     (Next.js frontend)                     │
└──────────────┬──────────────────────────┬──────────────────┘
               │  REST API                │  Webhook POST
               ▼                          ▼
┌──────────────────────┐     ┌──────────────────────────┐
│    hermes-core        │     │      hermes-hooks         │
│  (Auth, Relay CRUD,  │     │  (Webhook ingestion,      │
│   Secret mgmt,       │     │   NATS publisher)         │
│   Manual triggers)   │     │                           │
└──────────┬───────────┘     └────────────┬─────────────┘
           │  publishes                   │  publishes
           │  events.>                    │  events.>
           └──────────────┬───────────────┘
                          ▼
               ┌─────────────────────┐
               │   NATS JetStream    │
               │  (Message broker)   │
               └──────────┬──────────┘
                          │ consumes
                          ▼
               ┌─────────────────────┐
               │   hermes-worker     │
               │  (Execution engine) │
               │  - WorkerPool       │
               │  - CronScheduler    │
               │  - DAG executor     │
               └──────────┬──────────┘
                          │ reads/writes
                          ▼
               ┌─────────────────────┐
               │     PostgreSQL      │
               └─────────────────────┘
```

There is also a **hermes-telegram** service that provides a Telegram bot interface for interacting with the platform from Telegram.

---

## 3. Services

### 3.1 hermes-core
The main REST API. Handles:
- **Auth** — registration, login, JWT issuance, OAuth callback (Google, Microsoft).
- **Relay CRUD** — create, read, update, delete relays. Every create/update runs DAG validation before writing to the DB so invalid graphs are rejected early.
- **Secrets** — create, list, delete. Values are encrypted with AES-GCM at rest.
- **Connections** — OAuth token storage for email integrations.
- **Manual trigger** — publishes an event to NATS JetStream for relays with `trigger_type = manual`.
- **Execution history** — serves execution records and per-step details back to the frontend.

Router: **chi** with JWT middleware on protected routes and CORS configured for the frontend origin.

### 3.2 hermes-hooks
A thin, high-throughput webhook ingestion service. Receives inbound HTTP POSTs at `/hooks/:relayID`, reads up to 1MB of body, extracts or generates an `event_id` (from header `X-Event-ID`, query param, or UUID), then publishes a structured `ExecutionEvent` to NATS JetStream on `events.{relayID}`.

It deliberately does **no DB access** — it just translates HTTP to a queue message as fast as possible.

### 3.3 hermes-worker
The execution engine. Runs three concurrent subsystems:

| Subsystem | Role |
|---|---|
| `NATS Consumer` | Subscribes to `events.>` with a durable consumer, bridges messages into the internal job channel |
| `CronScheduler` | Polls the DB every 30 seconds for cron relays whose `next_run_at <= NOW()`, enqueues jobs directly |
| `WorkerPool` | N goroutines (default 10) that drain the job channel and execute the relay's action graph |

### 3.4 hermes-telegram
A Telegram bot integration that lets users interact with their relays from Telegram.

### 3.5 hermes-common
A shared Go module imported by all services. Contains:
- `dag` — graph construction, cycle detection, wave-parallel scheduling
- `templateengine` — `{{payload.x}}` / `{{steps['id'].output.x}}` resolver
- `cronutil` — cron expression parser and next-run calculator
- `encryptor` — AES-GCM encrypt/decrypt
- `actions` — canonical registry of action types and their config validation schemas
- `oauth` — Google and Microsoft OAuth provider implementations
- `logger` — structured `slog`-based logger factory

---

## 4. Technologies Used

| Layer | Technology | Why |
|---|---|---|
| Backend language | **Go** | Low overhead, strong concurrency primitives, fast builds |
| Frontend | **Next.js + TypeScript** | Full-stack React framework with type safety |
| HTTP router | **chi** | Lightweight, idiomatic Go router with middleware support |
| Message broker | **NATS JetStream** | Persistent, at-least-once delivery, durable consumers, built-in back-pressure |
| Database | **PostgreSQL** | JSONB for flexible config storage, strong referential integrity for graph edges |
| DB driver | **pgx / pgxpool** | High-performance native Go PostgreSQL driver with connection pooling |
| DB migrations | **golang-migrate** (SQL files) | Plain SQL migrations, easy to inspect and roll back |
| Auth | **JWT (golang-jwt)** | Stateless tokens, 168h expiry, HS256 signing |
| Password hashing | **bcrypt** | Standard adaptive hashing |
| Secret encryption | **AES-GCM** | Authenticated encryption for secrets stored in PostgreSQL |
| OAuth | **Google + Microsoft** | OAuth2 PKCE flow for email connection integrations |
| Containerisation | **Docker + Docker Compose** | Dev and prod parity across all services |
| Linting/Formatting | **Biome** (frontend) | Fast JS/TS formatter and linter |
| Monorepo | **Go workspaces** (`go.work`) | Lets multiple Go modules share local package paths |

---

## 5. Database Schema (Key Tables)

```
users
  id, username, email, password_hash, created_at, updated_at

relays
  id, user_id, name, description, webhook_path
  trigger_type (webhook | cron | manual)
  trigger_config (JSONB)
  next_run_at, last_run_at
  is_active, created_at, updated_at

relay_actions                         ← DAG nodes
  id, relay_id, node_id (TEXT, unique per relay)
  action_type, config (JSONB)
  created_at, updated_at

relay_edges                           ← DAG edges
  id, relay_id
  parent_node_id, child_node_id
  condition (JSONB, optional)
  created_at
  [FK: (relay_id, parent_node_id) → relay_actions(relay_id, node_id)]
  [FK: (relay_id, child_node_id)  → relay_actions(relay_id, node_id)]

executions
  id, relay_id, event_id, status
  trigger_payload (JSONB)
  error_message, started_at, finished_at

execution_steps
  id, execution_id, node_id
  action_type, status
  input (JSONB), output (JSONB)
  error_message, started_at, finished_at

secrets
  id, user_id, name, value (encrypted AES-GCM)

connections
  id, user_id, provider
  access_token, refresh_token (both encrypted)
  account_email, token_expiry

processed_events                      ← deduplication table
  relay_id, event_id  (unique together)
```

---

## 6. DAG Workflow System

### 6.1 Why DAG?

Earlier, workflows were a flat ordered list of actions (`order_index`). This meant every step ran sequentially even when steps were logically independent. DAG (Directed Acyclic Graph) support allows:

- **Non-linear execution** — independent branches run concurrently.
- **Fan-out** — one node's output feeding into multiple downstream nodes simultaneously.
- **Fan-in** — a node that waits for all parents to complete before it runs.
- **Conditional routing** — edges carry an optional `condition` field for future use.

### 6.2 Data Model

A relay's workflow is stored as two DB tables:
- `relay_actions` — each row is a **node** identified by a user-defined `node_id` string (unique within the relay).
- `relay_edges` — each row is a **directed edge** from `parent_node_id` → `child_node_id`. Foreign keys enforce that both endpoints reference real nodes in the same relay.

### 6.3 In-Memory Graph (`hermes-common/pkg/dag`)

At relay creation and update time (and at execution time) the raw nodes and edges are loaded and used to construct a `dag.Graph`:

```
Graph {
    Nodes:    map[nodeID] → Node
    OutEdges: map[nodeID] → []Edge   // outgoing edges (who does this node feed into?)
    InEdges:  map[nodeID] → []Edge   // incoming edges (what does this node depend on?)
}
```

`OutEdges` is used to walk downstream during scheduling.  
`InEdges` is used to compute in-degree (number of unresolved dependencies) for each node.

### 6.4 Cycle Detection — Kahn's Algorithm

Hermes prevents cycles using **Kahn's topological sort algorithm**, run inside `dag.New()` every time a graph is constructed:

**Step-by-step:**

1. Compute the **in-degree** of every node — how many incoming edges does it have?
2. Push all nodes with **in-degree = 0** (no dependencies, i.e. root nodes) into a ready queue.
3. While the queue is not empty:
   - Dequeue a node, increment `visitedCount`.
   - For each outgoing edge from that node, decrement the in-degree of the target node.
   - If any target's in-degree drops to **0**, it has no remaining unresolved dependencies → enqueue it.
4. After the queue is empty, compare `visitedCount` to `len(Nodes)`.
   - If they are **equal** → every node was reachable in topological order → **no cycle**.
   - If `visitedCount < len(Nodes)` → some nodes were never reachable because they are part of a cycle → **cycle detected → return error**.

This runs at two enforcement points:
- **API layer** (`validateDAG` in `hermes-core/api/handlers.go`) — rejects invalid graphs at create/update with HTTP 400 before anything touches the DB.
- **Runtime layer** (`dag.New()` in the worker) — double-checks when loading the graph for execution.

### 6.5 Wave-Parallel Scheduling (`dag.Waves()`)

Once the graph is validated, `Waves()` produces a **2D slice of node IDs** grouped by execution layer:

```
Wave 0: [A, B]        ← root nodes, no dependencies, run in parallel
Wave 1: [C, D]        ← depend only on wave 0, run in parallel after wave 0 finishes
Wave 2: [E]           ← depends on C and D, runs last
```

The algorithm is identical to Kahn's but instead of just counting visited nodes it groups them by layer:
1. Collect all zero-in-degree nodes into the first wave.
2. For each node in the current wave, decrement the in-degree of its children.
3. Collect all newly zero-in-degree children into the next wave.
4. Repeat until no nodes remain.

---

## 7. Worker Pool — How Workers Are Spawned and Execute Jobs

### 7.1 Startup

When the worker service starts (`cmd/main.go`):

```
NewWorkerPool(maxWorkers=10, db, registry, logger)
    └── creates a buffered Job channel (capacity 100)

pool.Start(ctx)
    └── spawns 10 goroutines, each running worker(id)

NewCronScheduler(db, pool.JobQueue, logger)
    └── starts a 30s ticker goroutine

queue.NewConsumer(natsURL, pool.JobQueue, logger)
    └── subscribes to NATS JetStream "events.>" with durable consumer
```

All three producers (NATS consumer, cron scheduler, manual trigger via hermes-core) write `Job` structs into the same shared buffered channel.

### 7.2 Job struct

```go
type Job struct {
    RelayID string        // which relay to execute
    EventID string        // deduplication key
    Payload []byte        // JSON trigger payload
    MsgAck  func(bool)    // callback: true = ACK, false = NAK the NATS message
}
```

For cron and manual jobs, `MsgAck` is a no-op. For NATS-originated jobs it calls `msg.Ack()` or `msg.Nak()` so JetStream knows whether to redeliver.

### 7.3 Worker Loop

Each of the 10 worker goroutines runs a `select` loop:

```
for {
    select {
    case <-ctx.Done():
        exit (context cancelled during shutdown)

    case job, ok := <-JobQueue:
        if !ok → exit (channel closed)
        process(ctx, job)
        job.MsgAck(err == nil)
    }
}
```

Workers exit cleanly on context cancellation without requiring the channel to be closed, which avoids send-on-closed panics.

### 7.4 process() — Execution Flow

For each job, `process()` performs the following steps in order:

```
1. Deduplication
   └── INSERT INTO processed_events (relay_id, event_id) ON CONFLICT DO NOTHING
   └── If 0 rows inserted → duplicate, skip silently

2. Create Execution record
   └── INSERT INTO executions (status='running', trigger_payload, started_at)
   └── Deferred: UPDATE executions SET status, finished_at on function exit

3. Fetch relay owner (user_id) for secret resolution

4. Load the relay graph
   └── GetRelayGraph() → []RelayAction (nodes) + []RelayEdge (edges)

5. Build DAG
   └── dag.New(nodes, edges) → validates + detects cycles

6. Get execution waves
   └── graph.Waves() → [][]nodeID

7. For each wave (outer loop — sequential between waves):
   For each nodeID in wave (inner loop — parallelisable):
       a. Resolve secrets
          └── _ref suffix fields → look up encrypted secret → decrypt → inject plain value
       b. Resolve templates
          └── templateengine.Resolve(config, payload, completedOutputs)
          └── Replaces {{payload.x}} and {{steps['nodeId'].output.x}}
       c. Create ExecutionStep record (status='running')
       d. Look up executor in Registry by action_type
       e. executor.Execute(ctx, resolvedConfig, payload, prevOutputs)
       f. Update ExecutionStep (status='success'/'failed', output)
       g. Store output in completedOutputs map keyed by nodeID

8. Deferred CompleteExecution fires → marks execution success or failed
```

### 7.5 Shutdown Order (Correct Producer-then-Consumer sequence)

```
1. consumer.Stop()       ← drains NATS subscription, no new messages enter channel
2. cronScheduler.Stop()  ← closes done channel (sync.Once), waits for goroutine exit
3. pool.Shutdown()       ← cancels context, waits for all 10 workers to finish current jobs
```

This order guarantees no new jobs are enqueued after shutdown begins, and all in-flight jobs run to completion.

---

## 8. Template Engine

The template engine allows action configs to reference dynamic data using `{{expr}}` syntax.

| Expression | Resolves to |
|---|---|
| `{{payload}}` | Full raw JSON trigger payload |
| `{{payload.user.name}}` | Drilled field from the trigger payload |
| `{{steps['nodeId'].output}}` | Full output JSON from a completed node |
| `{{steps['nodeId'].output.status}}` | Specific field from a completed node's output |

Resolution happens per-node just before execution, using outputs collected from all previously completed nodes. Non-string config values (integers, booleans) pass through untouched.

---

## 9. Security Model

| Concern | Mechanism |
|---|---|
| Authentication | JWT (HS256, 168h expiry) issued on login/register/OAuth callback |
| Password storage | bcrypt with default cost |
| Secret values | AES-GCM encrypted before writing to DB, decrypted only at execution time |
| Log safety | `redactConfig()` strips secret-ref fields and known-sensitive keys (`api_key`, `token`, `password`, `webhook_url`, `secret`) from execution step inputs before persisting |
| Transport | CORS restricted to configured frontend origin |
| Payload size | Webhook bodies capped at 1MB in hermes-hooks |
| Duplicate events | `processed_events` table with unique constraint prevents replay attacks |

---

## 10. Integration Plugins

Each plugin implements the `ActionExecutor` interface:

```go
type ActionExecutor interface {
    Execute(ctx context.Context, config map[string]any, payload []byte, prevOutputs map[string]StepOutput) (json.RawMessage, error)
}
```

Registered plugins:

| Name | What it does |
|---|---|
| `debug_log` | Logs config and payload, useful for testing |
| `discord_send` | Posts a message to a Discord webhook URL |
| `slack_send` | Posts a message to a Slack incoming webhook |
| `http_request` | Makes an outbound HTTP request (GET/POST/PUT/PATCH/DELETE) |
| `email_send` | Sends an email via a connected OAuth provider (Google/Microsoft) |
| `condition` | Evaluates a condition (for conditional routing, in development) |

New plugins are registered in `main.go` with a single line: `reg.Register("name", impl.New())`. No other code changes needed.

---

## 11. Request Lifecycle (End-to-End Example)

**Scenario: a webhook fires on a relay that sends a Slack message then calls an HTTP API.**

```
1. External service POSTs to https://hermes/hooks/{relayID}
2. hermes-hooks reads body, generates event_id, publishes to NATS:
       events.{relayID} → { event_id, relay_id, payload, received_at }
3. NATS JetStream persists the message, delivers to WORKER_CONSUMER subscription
4. hermes-worker NATS Consumer receives it, wraps it in a Job, pushes to channel
5. A free worker picks up the Job:
   a. Deduplication check → new event, proceed
   b. Creates execution record in DB
   c. Loads relay graph (2 nodes: slack_send → http_request, 1 edge)
   d. dag.New() validates, no cycle
   e. Waves: [[slack_send], [http_request]]
   f. Wave 0: resolves secrets, resolves templates, executes slack_send
              → Slack returns 200, output: {"ok": true}
              → step recorded in DB
   g. Wave 1: resolves http_request config, {{steps['slack_send'].output.ok}} = true
              → HTTP call made, response captured
              → step recorded in DB
   h. Deferred: execution marked success
6. Worker ACKs the NATS message
```

---

## 12. Monorepo Layout

```
hermes/
├── go.work                          ← Go workspace linking all modules
├── docker-compose.yml               ← Spins up all services + Postgres + NATS
├── web/                             ← Next.js frontend (TypeScript)
├── packages/
│   └── hermes-common/               ← Shared Go library
│       └── pkg/
│           ├── dag/                 ← DAG graph, cycle detection, wave scheduling
│           ├── templateengine/      ← {{expr}} resolver
│           ├── cronutil/            ← Cron next-run calculator
│           ├── encryptor/           ← AES-GCM encrypt/decrypt
│           ├── actions/             ← Action type registry + config validation
│           ├── oauth/               ← Google + Microsoft OAuth providers
│           └── logger/              ← Structured slog factory
└── services/
    ├── hermes-core/                 ← REST API (auth, relays, secrets, connections)
    │   ├── internal/api/            ← Handlers, middleware, router
    │   ├── internal/store/          ← DB queries
    │   ├── internal/models/         ← Request/response types
    │   └── db/migrations/           ← SQL migration files
    ├── hermes-hooks/                ← Webhook ingestion → NATS publisher
    ├── hermes-worker/               ← Execution engine
    │   ├── internal/engine/         ← WorkerPool, CronScheduler, Registry, DAG executor
    │   ├── internal/queue/          ← NATS consumer
    │   ├── internal/store/          ← Worker-side DB queries
    │   └── internal/integrations/   ← Plugin implementations
    └── hermes-telegram/             ← Telegram bot interface
```
