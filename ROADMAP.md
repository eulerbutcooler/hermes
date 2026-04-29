# Hermes Roadmap

Two major features to evolve Hermes from a linear-step automation platform into an intelligent, conversational workflow engine.

---

## Feature 1 — DAG Execution Engine

> Replace the current linear `order_index` chain with a full directed acyclic graph (DAG) so that actions can fan-out, fan-in, branch conditionally, and run in parallel.

### Current State (What We Have)

- Workflows ("relays") are a flat, ordered list of `relay_actions` keyed by `(relay_id, order_index)`.
- The worker pool (`worker_pool.go`) iterates `actions` sequentially: step 0 → step 1 → step 2 → …
- Template engine references use `prev`, `steps[i]` — both assume linear ordering.
- The DB schema enforces a `UNIQUE(relay_id, order_index)` constraint.

### Target State

```
         ┌──► B ──┐
Trigger ─► A      ├──► D ──► E
         └──► C ──┘
```

- Each action (node) declares which other nodes it depends on.
- Nodes with no unmet dependencies run in parallel.
- Fan-in nodes wait for **all** parents to complete before starting.
- Conditional edges allow branching (`if`, `switch`).
- Cycle detection rejects invalid graphs at save time.

---

### Phase 1 — Data Model & Schema (Week 1-2)

#### 1.1 New Migration: `000007_dag_edges.up.sql`

```sql
-- Add a stable node_id to relay_actions (replaces order_index as the identity)
ALTER TABLE relay_actions
  ADD COLUMN IF NOT EXISTS node_id TEXT;

-- Backfill existing rows: use 'node_<order_index>' so nothing breaks
UPDATE relay_actions
  SET node_id = 'node_' || order_index
  WHERE node_id IS NULL;

ALTER TABLE relay_actions
  ALTER COLUMN node_id SET NOT NULL;

-- Each node_id must be unique within a relay
ALTER TABLE relay_actions
  ADD CONSTRAINT uq_relay_actions_node_id UNIQUE (relay_id, node_id);

-- Edge table: defines "node_id depends on parent_node_id"
CREATE TABLE IF NOT EXISTS relay_edges (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relay_id     UUID NOT NULL REFERENCES relays(id) ON DELETE CASCADE,
    parent_node_id TEXT NOT NULL,
    child_node_id  TEXT NOT NULL,
    condition    JSONB,  -- optional: {"expr": "steps['fetch'].output.status == 200"}
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(relay_id, parent_node_id, child_node_id),
    FOREIGN KEY (relay_id, parent_node_id) REFERENCES relay_actions(relay_id, node_id) ON DELETE CASCADE,
    FOREIGN KEY (relay_id, child_node_id)  REFERENCES relay_actions(relay_id, node_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_relay_edges_relay ON relay_edges(relay_id);
CREATE INDEX IF NOT EXISTS idx_relay_edges_child  ON relay_edges(relay_id, child_node_id);
```

#### 1.2 Update Go Models (`models.go`)

```go
type CreateRelayActionInput struct {
    NodeID     string         `json:"node_id"`               // NEW: stable identifier e.g. "fetch_data"
    ActionType string         `json:"action_type"`
    Config     map[string]any `json:"config"`
    OrderIndex int            `json:"order_index,omitempty"`  // DEPRECATED: kept for backwards compat
}

type RelayEdge struct {
    ParentNodeID string         `json:"parent_node_id"`
    ChildNodeID  string         `json:"child_node_id"`
    Condition    map[string]any `json:"condition,omitempty"`
}

type CreateRelayRequest struct {
    Name          string                   `json:"name"`
    UserID        string                   `json:"user_id"`
    Description   string                   `json:"description"`
    TriggerType   TriggerType              `json:"trigger_type,omitempty"`
    TriggerConfig map[string]any           `json:"trigger_config,omitempty"`
    Actions       []CreateRelayActionInput `json:"actions"`
    Edges         []RelayEdge              `json:"edges"`              // NEW
}

type RelayWithActions struct {
    Relay
    Actions []RelayAction `json:"actions"`
    Edges   []RelayEdge   `json:"edges"`   // NEW
}
```

#### 1.3 Backwards Compatibility Strategy

- If `edges` is empty/nil in a `CreateRelayRequest`, auto-generate a linear chain from `order_index` (node_0 → node_1 → node_2 → …). Old clients keep working.
- If `edges` is provided, `order_index` is ignored during execution.

---

### Phase 2 — DAG Validation Library (Week 2)

#### 2.1 New Package: `packages/hermes-common/pkg/dag/`

Build a pure-Go DAG library that is used by both `hermes-core` (API validation) and `hermes-worker` (execution planning).

```
packages/hermes-common/pkg/dag/
├── dag.go          // Graph construction, topological sort, cycle detection
├── dag_test.go     // Unit tests
├── scheduler.go    // Produces execution "waves" (levels of parallelizable nodes)
└── scheduler_test.go
```

**Core API**:
```go
package dag

type Node struct {
    ID string
}

type Edge struct {
    From      string
    To        string
    Condition map[string]any  // nil = unconditional
}

type Graph struct { ... }

func New(nodes []Node, edges []Edge) (*Graph, error)  // returns error on cycle
func (g *Graph) TopologicalOrder() []string
func (g *Graph) Waves() [][]string   // groups of nodes that can run in parallel
func (g *Graph) RootNodes() []string // nodes with no incoming edges (entry points)
func (g *Graph) Parents(nodeID string) []string
```

**Cycle Detection**: Use Kahn's algorithm (BFS-based topological sort). If not all nodes are visited, there is a cycle → reject.

#### 2.2 Validation at API Layer

In `handlers.go` → `CreateRelay` and `UpdateRelayActions`:
1. Build a `dag.Graph` from the submitted nodes + edges.
2. If `dag.New()` returns an error (cycle), respond `400 VALIDATION_ERROR: "workflow contains a cycle"`.
3. Verify all `node_id` values referenced in edges actually exist in the actions list.
4. Verify there is at least one root node (entry point).

---

### Phase 3 — DAG Execution Engine (Week 3-4)

#### 3.1 Refactor `worker_pool.go` → `process()`

Replace the linear `for _, act := range actions` loop with a wave-based executor:

```
              CURRENT                           TARGET
    ┌────────────────────────┐      ┌─────────────────────────────────┐
    │ for each action (seq)  │      │ waves := dag.Waves()            │
    │   resolve secrets      │      │ for each wave:                  │
    │   resolve templates    │      │   run all nodes in parallel     │
    │   execute              │      │   wait for all to complete      │
    │   collect output       │      │   evaluate conditional edges    │
    │ end                    │      │   merge outputs into shared map │
    └────────────────────────┘      │ end                             │
                                    └─────────────────────────────────┘
```

**Key design decisions**:

| Concern | Decision |
|---|---|
| **Parallelism within a wave** | `errgroup.Group` with a per-wave context. If one node fails, cancel siblings (configurable: fail-fast vs. continue). |
| **Output storage** | Replace `[]StepOutput` (ordered list) with `map[string]StepOutput` keyed by `node_id`. Thread-safe via `sync.Map` or a mutex. |
| **Conditional edges** | After a node completes, evaluate its outbound edge conditions. Unevaluated children are skipped. Use a simple expression evaluator (e.g. `expr-lang/expr`). |
| **Template engine update** | `steps[i]` syntax is deprecated. Add `steps['node_id']` syntax. `prev` becomes ambiguous in DAG → deprecate or resolve to the single parent (error if multiple). |

#### 3.2 Updated Execution Step Schema

```sql
ALTER TABLE execution_steps
  ADD COLUMN IF NOT EXISTS node_id TEXT;
```

The `node_id` lets you reconstruct the execution graph shape in the UI.

#### 3.3 Template Engine Updates (`templateengine.go`)

Add support for node-id based references alongside the existing index-based ones:

```
{{ steps['fetch_data'].output.body.items[0].name }}   ← NEW (preferred)
{{ steps[0].output.body }}                             ← EXISTING (still works for linear)
{{ prev.output }}                                      ← DEPRECATED (ambiguous in DAG)
```

Internal change: `steps` parameter changes from `[]StepOutput` to a `map[string]StepOutput`, but the `Resolve()` function signature accepts both for backwards compat.

---

### Phase 4 — API & Store Updates (Week 4-5)

#### 4.1 Store Layer Changes (`relay_store.go`)

- `CreateRelay`: Insert edges into `relay_edges` table within the same transaction.
- `GetRelay`: JOIN and return edges alongside actions.
- `UpdateRelayActions`: Accept new `edges` field, delete-and-replace edges in the same transaction.
- New query: `GetRelayGraph(ctx, relayID) → ([]RelayAction, []RelayEdge)` used by the worker.

#### 4.2 Worker Store (`hermes-worker/internal/store/`)

- `GetRelayActions` → also returns edges so the worker can build the DAG.
- Or add a new `GetRelayEdges(ctx, relayID) → []RelayEdge` query.

#### 4.3 API Response Changes

All relay responses now include an `edges` field. Old clients that don't read it are unaffected.

---

### Phase 5 — Frontend Updates (Week 5-6)

#### 5.1 Visual DAG Builder (`web/`)

Replace the current linear action list UI with a node-graph editor:

- **Library**: Use `@xyflow/react` (React Flow) — the standard for visual node editors.
- Each action type becomes a custom node component with typed handles (inputs/outputs).
- Edges are draggable connections between nodes.
- Validate on save (call a validation endpoint or do client-side cycle detection).
- Export the graph as `{actions: [...], edges: [...]}` JSON.

#### 5.2 Execution Visualization

- Show completed executions as a DAG with color-coded nodes (green = success, red = failed, grey = skipped).
- Clicking a node shows its input/output/error details.

---

### Phase 6 — Testing & Hardening (Week 6-7)

- Unit tests for `dag` package: cycle detection, topological sort, wave generation, edge cases (single node, diamond, disconnected subgraphs).
- Integration tests: create a DAG relay via API, trigger it, verify parallel execution, verify conditional skipping.
- Migration test: existing linear relays still execute correctly after migration (backwards compat).
- Load test: DAG with 50+ nodes, verify memory/goroutine usage is bounded.

---

## Feature 2 — Telegram Bot (Natural Language → Workflow)

> A Telegram bot where you text something like *"when I get a webhook, fetch my GitHub issues and send a summary to Discord"* and it creates a real Hermes workflow behind the scenes.

### Architecture

```
┌──────────────┐     ┌──────────────────┐     ┌────────────────┐     ┌─────────────┐
│  Telegram    │     │  hermes-telegram  │     │  LLM Provider  │     │ hermes-core │
│  (user msg)  │────►│  (Go service)     │────►│  (OpenAI /     │     │ (REST API)  │
│              │◄────│                   │◄────│   Gemini, etc) │     │             │
│              │     │  - bot handler    │     └────────────────┘     │             │
│              │     │  - auth linking   │──────────────────────────►│  POST /relays│
│              │     │  - session mgmt   │◄─────────────────────────│             │
└──────────────┘     └──────────────────┘                            └─────────────┘
```

### Phase 1 — New Service Scaffolding (Week 1)

#### 1.1 Create `services/hermes-telegram/`

```
services/hermes-telegram/
├── cmd/
│   └── bot/
│       └── main.go           // Entry point
├── internal/
│   ├── bot/
│   │   ├── bot.go            // Telegram bot setup, command routing
│   │   ├── handlers.go       // Message + command handlers
│   │   └── session.go        // Per-user conversation state
│   ├── ai/
│   │   ├── client.go         // LLM API client (OpenAI/Gemini)
│   │   ├── prompts.go        // System prompts, few-shot examples
│   │   └── parser.go         // Parse LLM JSON output → CreateRelayRequest
│   ├── hermes/
│   │   └── client.go         // HTTP client for hermes-core API
│   └── config/
│       └── config.go         // Env-based configuration
├── go.mod
├── go.sum
├── Dockerfile
└── .env
```

#### 1.2 Add to Go Workspace

```go
// go.work
use (
    ./packages/hermes-common
    ./services/hermes-core
    ./services/hermes-hooks
    ./services/hermes-worker
    ./services/hermes-telegram   // NEW
)
```

#### 1.3 Dependencies

- `github.com/go-telegram-bot-api/telegram-bot-api/v5` — Telegram Bot API client.
- `github.com/sashabaranov/go-openai` (or equivalent for Gemini) — LLM client.
- `hermes-common` — reuse `actions.Types()` for the LLM prompt.

#### 1.4 Environment Variables

```env
# hermes-telegram .env
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
LLM_PROVIDER=openai          # "openai" | "gemini" | "anthropic"
LLM_API_KEY=sk-...
LLM_MODEL=gpt-4o-mini        # cost-effective for structured output
HERMES_CORE_URL=http://localhost:3000
```

---

### Phase 2 — Telegram Bot Core (Week 1-2)

#### 2.1 Bot Setup & Commands (`bot.go`)

```
/start        → Welcome message + auth instructions
/login        → Link Telegram user to Hermes account (email + password or token)
/new          → Start creating a workflow (enters conversational mode)
/list         → List your existing relays
/trigger <id> → Manually trigger a relay
/help         → Show available commands
```

#### 2.2 User Authentication & Linking (`session.go`)

Two options (implement the simpler one first):

**Option A — Token-based linking (recommended for v1)**:
1. User runs `/login` in Telegram.
2. Bot says: "Send me your Hermes API token. You can generate one at [dashboard URL]."
3. User sends the JWT token.
4. Bot validates the token by calling `GET /api/v1/relays` with it.
5. If valid, store `(telegram_user_id, hermes_jwt)` mapping in Postgres.

**Option B — OAuth deep-link (future)**:
1. Bot sends a login URL that redirects through Hermes OAuth.
2. Callback writes the link to the DB.

#### 2.3 Session / Conversation State (`session.go`)

Each Telegram user can be in one of these states:

```go
type SessionState int

const (
    StateIdle       SessionState = iota
    StateAwaitLogin              // waiting for JWT/credentials
    StateDescribing              // user is describing a workflow in natural language
    StateConfirming              // bot has generated a relay, waiting for user confirmation
    StateEditing                 // user is refining the generated relay
)

type Session struct {
    State        SessionState
    HermesToken  string
    DraftRelay   *models.CreateRelayRequest  // the relay being built
    Conversation []Message                   // LLM conversation history
}
```

Store sessions in-memory with a `sync.Map` (simple) or Redis (production).

---

### Phase 3 — LLM Integration (Week 2-3)

#### 3.1 System Prompt Design (`prompts.go`)

This is the most critical piece. The system prompt must:
1. Describe what Hermes is and what action types are available.
2. Provide the exact JSON schema for `CreateRelayRequest`.
3. Include few-shot examples mapping natural language → JSON.
4. Instruct the LLM to ask clarifying questions if the user's request is ambiguous.

```go
func BuildSystemPrompt() string {
    actionTypes := actions.Types() // ["debug_log", "discord_send", "email_send", ...]
    return fmt.Sprintf(`You are Hermes Assistant, an AI that helps users create automation workflows.

Available action types: %s

When the user describes a workflow, respond with ONLY a JSON object matching this schema:
{
  "ready": true/false,
  "questions": ["...", "..."],     // if ready=false, ask these
  "relay": {                       // if ready=true, the workflow
    "name": "...",
    "description": "...",
    "trigger_type": "webhook" | "manual" | "cron",
    "trigger_config": {},
    "actions": [
      {
        "node_id": "step_name",
        "action_type": "...",
        "config": { ... },
        "order_index": 0
      }
    ],
    "edges": [
      {"parent_node_id": "a", "child_node_id": "b"}
    ]
  }
}

Rules:
- If the user's request is clear enough, set "ready": true and provide the full relay JSON.
- If you need more information (e.g. which Discord webhook URL, what email to send to), set "ready": false and list your questions.
- Use descriptive node_ids like "fetch_github", "send_discord", not generic names.
- For secrets like webhook URLs or API keys, use the "_ref" suffix pattern (e.g. "webhook_url_ref": "my_discord_webhook") and tell the user they need to save the secret in Hermes first.
- Build DAG edges based on data dependencies. Steps that don't depend on each other should be parallel.

Action type configs:
- debug_log: {"message": "..."}
- discord_send: {"webhook_url_ref": "secret_name", "message": "..."}
- slack_send: {"webhook_url_ref": "secret_name", "message": "..."}
- http_request: {"url": "...", "method": "GET|POST|...", "headers": {}, "body": "..."}
- email_send: {"connection_id": "...", "to": "...", "subject": "...", "body": "..."}
`, strings.Join(actionTypes, ", "))
}
```

#### 3.2 LLM Client (`client.go`)

```go
type LLMClient interface {
    Chat(ctx context.Context, messages []Message) (string, error)
}

type Message struct {
    Role    string `json:"role"`    // "system", "user", "assistant"
    Content string `json:"content"`
}
```

Implement for OpenAI first. Use `response_format: { type: "json_object" }` to force JSON output.

#### 3.3 Response Parser (`parser.go`)

Parse the LLM's JSON response:

```go
type LLMResponse struct {
    Ready     bool                      `json:"ready"`
    Questions []string                  `json:"questions"`
    Relay     *models.CreateRelayRequest `json:"relay"`
}
```

Validate the parsed relay:
- All action types exist in `actions.Registry`.
- Config fields pass `actions.ValidateConfig()`.
- DAG is valid (no cycles) via `dag.New()`.
- If validation fails, send the error back to the LLM for self-correction (1 retry).

---

### Phase 4 — Conversational Flow (Week 3-4)

#### 4.1 Message Handler Flow (`handlers.go`)

```
User sends a message
        │
        ▼
  ┌─ Is user authenticated? ─── No ──► "Please /login first"
  │
  Yes
  │
  ▼
  ┌─ Current state?
  │
  ├─ Idle ──────► Treat message as a new workflow description
  │               Append to conversation, call LLM
  │               ├─ LLM says ready=false ──► Send questions to user, state=Describing
  │               └─ LLM says ready=true  ──► Show summary, state=Confirming
  │
  ├─ Describing ──► Append answer to conversation, call LLM again
  │                 ├─ ready=false ──► Ask more questions
  │                 └─ ready=true  ──► Show summary, state=Confirming
  │
  ├─ Confirming ──► User says "yes"/"confirm" ──► POST to hermes-core API
  │                 User says "no"/"edit"     ──► state=Editing
  │                 User says "cancel"         ──► state=Idle
  │
  └─ Editing ────► Append edit instruction to conversation, call LLM
                    └─ ... (same as Describing)
```

#### 4.2 User-Facing Messages

When the LLM returns `ready=true`, format the relay as a readable summary:

```
🔗 *New Workflow: GitHub Issue Notifier*

Trigger: Webhook
Steps:
  1. 📡 fetch_github → HTTP Request (GET https://api.github.com/repos/.../issues)
  2. 📨 notify_discord → Discord Send (message: "New issues: {{steps['fetch_github'].output.body}}")

Flow: fetch_github → notify_discord

Reply "confirm" to create, "edit" to modify, or "cancel" to discard.
```

#### 4.3 Error Handling

- LLM returns invalid JSON → Retry once with a corrective prompt.
- LLM uses unknown action types → Tell user "I don't know how to do X yet. Available actions are: ..."
- Hermes API returns an error → Forward the error message to the user.

---

### Phase 5 — Hermes Core Client (`hermes/client.go`) (Week 4)

A typed HTTP client that calls the hermes-core REST API:

```go
type HermesClient struct {
    baseURL    string
    httpClient *http.Client
}

func (c *HermesClient) CreateRelay(ctx context.Context, token string, req models.CreateRelayRequest) (*models.RelayWithActions, error)
func (c *HermesClient) ListRelays(ctx context.Context, token string) ([]models.Relay, error)
func (c *HermesClient) TriggerRelay(ctx context.Context, token string, relayID string, payload map[string]any) error
func (c *HermesClient) GetExecutions(ctx context.Context, token string, relayID string) ([]models.Execution, error)
```

---

### Phase 6 — Additional Bot Features (Week 5)

#### 6.1 Execution Notifications

- When a relay execution completes (success or failure), send a Telegram notification to the relay owner.
- Implementation: The worker publishes execution events to a NATS subject. `hermes-telegram` subscribes and routes notifications to the right Telegram chat.

#### 6.2 Inline Relay Management

```
/list                     → Shows relays as inline buttons
/status <relay_name>      → Shows last 5 executions
/toggle <relay_name>      → Enable/disable a relay
/delete <relay_name>      → Delete with confirmation
/logs <relay_name>        → Show recent execution logs
```

#### 6.3 Quick Templates

Pre-built workflow templates the user can activate with one tap:

```
/templates
  📋 GitHub → Discord (notify on new webhooks)
  📋 Cron → Email (daily report)
  📋 Webhook → Slack + Email (fan-out notification)
```

---

### Phase 7 — Infrastructure & Deployment (Week 5-6)

#### 7.1 Docker Compose Update

```yaml
# docker-compose.yml
hermes-telegram:
  build: ./services/hermes-telegram
  environment:
    - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
    - LLM_API_KEY=${LLM_API_KEY}
    - HERMES_CORE_URL=http://hermes-core:3000
    - DATABASE_URL=postgres://user:password@postgres:5432/hermes?sslmode=disable
    - NATS_URL=nats://nats:4222
  depends_on:
    - postgres
    - nats
```

#### 7.2 Makefile Addition

```makefile
dev-telegram: ## Run hermes-telegram bot
    @echo "$(YELLOW)Starting hermes-telegram...$(NC)"
    @cd services/hermes-telegram && go run cmd/bot/main.go
```

#### 7.3 Database: Telegram User Links

```sql
-- 000008_telegram_links.up.sql
CREATE TABLE IF NOT EXISTS telegram_links (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    telegram_user_id BIGINT NOT NULL UNIQUE,
    telegram_username TEXT,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_telegram_links_user ON telegram_links(user_id);
CREATE INDEX IF NOT EXISTS idx_telegram_links_tg   ON telegram_links(telegram_user_id);
```

---

## Dependency Order & Timeline

```mermaid
gantt
    title Hermes Roadmap
    dateFormat  YYYY-MM-DD
    axisFormat  %b %d

    section DAG Engine
    Data Model & Migration       :dag1, 2026-05-01, 7d
    DAG Validation Library       :dag2, after dag1, 5d
    DAG Execution Engine         :dag3, after dag2, 10d
    API & Store Updates          :dag4, after dag2, 7d
    Frontend DAG Builder         :dag5, after dag4, 10d
    Testing & Hardening          :dag6, after dag3, 7d

    section Telegram Bot
    Service Scaffolding          :tg1, 2026-05-01, 5d
    Bot Core & Auth              :tg2, after tg1, 7d
    LLM Integration              :tg3, after tg2, 7d
    Conversational Flow          :tg4, after tg3, 7d
    Hermes Client & Features     :tg5, after tg4, 5d
    Infra & Deployment           :tg6, after tg5, 5d
```

> **Both features can be developed in parallel** since they touch different parts of the codebase. The only intersection point is that the Telegram bot should generate DAG-format relays — so the DAG data model (Phase 1) should land before the LLM prompt design (Telegram Phase 3).

---

## Files Changed / Created Summary

### DAG Engine

| Action | Path |
|--------|------|
| NEW | `services/hermes-core/db/migrations/000007_dag_edges.up.sql` |
| NEW | `services/hermes-core/db/migrations/000007_dag_edges.down.sql` |
| NEW | `services/hermes-core/db/migrations/000008_exec_step_node_id.up.sql` |
| NEW | `packages/hermes-common/pkg/dag/dag.go` |
| NEW | `packages/hermes-common/pkg/dag/dag_test.go` |
| NEW | `packages/hermes-common/pkg/dag/scheduler.go` |
| NEW | `packages/hermes-common/pkg/dag/scheduler_test.go` |
| MODIFY | `services/hermes-core/internal/models/models.go` |
| MODIFY | `services/hermes-core/internal/store/relay_store.go` |
| MODIFY | `services/hermes-core/internal/api/handlers.go` |
| MODIFY | `services/hermes-core/internal/api/interfaces.go` |
| MODIFY | `services/hermes-worker/internal/engine/worker_pool.go` |
| MODIFY | `services/hermes-worker/internal/engine/executor.go` |
| MODIFY | `services/hermes-worker/internal/store/` |
| MODIFY | `packages/hermes-common/pkg/templateengine/templateengine.go` |

### Telegram Bot

| Action | Path |
|--------|------|
| NEW | `services/hermes-telegram/` (entire service) |
| NEW | `services/hermes-core/db/migrations/000008_telegram_links.up.sql` |
| MODIFY | `go.work` |
| MODIFY | `docker-compose.yml` |
| MODIFY | `Makefile` |
| MODIFY | `.env.example` |

---

## Open Questions

1. **LLM Provider**: OpenAI (gpt-4o-mini) vs Google Gemini vs Anthropic Haiku? gpt-4o-mini is cheapest for structured JSON output. Gemini has a generous free tier.
2. **Conditional edges**: Should we implement a full expression language (like `expr`) or start with simple comparisons (`status == 200`, `output.count > 0`)?
3. **Fan-in failure policy**: When a fan-in node has multiple parents and one fails, should the fan-in node be skipped, or should it run with partial inputs?
4. **Telegram session storage**: In-memory `sync.Map` (loses state on restart) vs Postgres table vs Redis?
5. **Rate limiting**: Should the LLM calls be rate-limited per user to control costs?
