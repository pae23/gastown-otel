# Spec: Gastown Waterfall v2 — Chrome DevTools-style Agent Orchestration View

> **Authoritative reference**: `/Users/pa/dev/third-party/gastown/docs/waterfall-spec.md`
> This document extends the reference with Go/frontend implementation constraints for `gastown-trace`.

---

## Context

Gastown is a multi-agent orchestration system running Claude Code agents in tmux sessions. Agents have roles (mayor, deacon, witness, refinery, polecat, dog, boot, crew) and are organised into **rigs** (e.g. `fai`, `mol`, `gt-wyvern`). They communicate via **beads** (work items managed by `bd`), **mails**, **slings** (bead dispatches), and **prompts** sent into tmux panes.

All telemetry lives in VictoriaLogs (structured OTLP logs) queried via LogsQL. The Go tool `gastown-trace` queries VictoriaLogs and renders HTML views. The current `/waterfall` page is a prototype — this spec replaces it entirely.

### Primary key: `run.id`

Each agent spawn generates a unique UUID — the **`run.id`** (`GT_RUN`) — propagated into the tmux environment and into `OTEL_RESOURCE_ATTRIBUTES` for all `bd` sub-processes. It is the universal correlation key across all events in a run. **All correlation logic must prefer `run.id` over legacy `_stream` fields.**

---

## Goal

Build a **Chrome DevTools Network Waterfall**-style view at `/waterfall` showing the complete timeline of a Gastown instance: every agent session, every inter-agent exchange, every API call — laid out horizontally on a shared time axis, with interactive filtering, drill-down, and communication-flow visualisation.

Think: Azure DevOps pipeline view × Chrome Network tab — for an AI agent swarm.

---

## Data Sources (VictoriaLogs events)

All data comes from `vlQuery()` calls. Available event types:

| Event | Key fields | What it represents |
|-------|-----------|-------------------|
| `agent.instantiate` | `run.id`, `instance`, `town_root`, `agent_type`, `role`, `agent_name`, `session_id`, `rig` | **Root of each run** — emitted once per spawn |
| `session.start` | `run.id`, `session_id`, `role`, `status` | Agent session started in tmux |
| `session.stop` | `run.id`, `session_id`, `role`, `status` | Agent session stopped |
| `prime` | `run.id`, `role`, `hook_mode`, `formula`, `status` | Startup context injection (rendered TOML formula) |
| `bd.call` | `run.id`, `subcommand`, `args`, `stdout`, `stderr`, `duration_ms`, `status` | bd CLI operation |
| `claude_code.api_request` | `session.id`, `model`, `input_tokens`, `output_tokens`, `cache_read_tokens`, `cost_usd`, `duration_ms` | LLM API call *(source: claude-code OTEL instrumentation, independent of gastown)* |
| `claude_code.tool_result` | `session.id`, `tool_name`, `tool_parameters`, `duration_ms`, `success` | Tool execution *(source: same)* |
| `agent.event` | `run.id`, `session`, `native_session_id`, `agent_type`, `event_type`, `role` *(LLM role: `"assistant"` / `"user"`, ≠ Gastown role)*, `content` | Agent conversation turn (text/tool_use/tool_result/thinking) |
| `agent.state_change` | `run.id`, `agent_id`, `new_state`, `hook_bead` | Agent state transition; `hook_bead` is the bead ID being processed (empty string if none) |
| `prompt.send` | `run.id`, `session`, `keys_len`, `debounce_ms`, `status` | Prompt injected into agent via tmux *(full text `keys` is P1)* |
| `pane.output` | `run.id`, `session`, `content` | Raw tmux pane output *(opt-in: `GT_LOG_PANE_OUTPUT=true`)* |
| `sling` | `run.id`, `bead`, `target`, `status` | Bead dispatched from one agent to another |
| `mail` | `run.id`, `operation`, `msg.id`, `msg.from`, `msg.to`, `msg.subject`, `msg.body`, `msg.thread_id`, `msg.priority`, `msg.type`, `status` | Inter-agent mail operation |
| `nudge` | `run.id`, `target`, `status` | Agent nudged (re-triggered) |
| `polecat.spawn` | `run.id`, `name`, `status` | Polecat sub-agent spawned |
| `polecat.remove` | `run.id`, `name`, `status` | Polecat removed |
| `done` | `run.id`, `exit_type` (COMPLETED/ESCALATED/DEFERRED), `status` | Agent completed its work item |
| `formula.instantiate` | `run.id`, `formula_name`, `bead_id`, `status` | Work template instantiated |
| `convoy.create` | `run.id`, `bead_id`, `status` | Auto-convoy (batch) created |
| `daemon.restart` | `run.id`, `agent_type` | Daemon restarted |
| `mol.cook` | `run.id`, `formula_name`, `status` | Formula compiled to a proto (prerequisite for wisp creation) |
| `mol.wisp` | `run.id`, `formula_name`, `wisp_root_id`, `bead_id`, `status` | Proto instantiated as a live wisp; `bead_id` empty for standalone formula slinging |
| `mol.squash` | `run.id`, `mol_id`, `done_steps`, `total_steps`, `digest_created`, `status` | Molecule execution completed and collapsed to a digest |
| `mol.burn` | `run.id`, `mol_id`, `children_closed`, `status` | Molecule destroyed without creating a digest |
| `bead.create` | `run.id`, `bead_id`, `parent_id`, `mol_source` | Child bead created during molecule instantiation |

> **Note — `mail`**: use `RecordMailMessage` for operations where the message is available (send, read); use `RecordMail` for content-less operations (list, archive-by-id). A targeted query `mail AND operation:send` is needed alongside the generic `mail` query because `bd.list` / `bd.mol` operations vastly outnumber `send` events and can push them past the per-query limit.

> **Note — `agent.event.role`**: this field is the **LLM role** (`"assistant"` or `"user"`), not the Gastown role (mayor/witness/…). The Gastown role is on `agent.instantiate.role` and propagated via `gt.role` in `_stream` fields.

> **Note — `session.start`**: v1 of this spec listed `gt.topic`, `gt.prompt`, `gt.agent` on this event. These fields are not in the reference; they come from an older `_stream` fields version. Ignore them for correlation logic — prefer `agent.instantiate`.

### Resource attributes on all events

Two systems coexist — prefer **direct attributes** (new model) over **`_stream` fields** (legacy):

**Direct attributes (new model, authoritative):**
- `run.id` — run UUID (primary key)
- `instance` — `hostname:basename(town_root)` (e.g. `laptop:gt`)
- `role` — Gastown role (mayor, witness, polecat, …)
- `rig` — rig name (empty = town-level)
- `session_id` — tmux pane name

**`_stream` fields (legacy, useful for older events):**
- `gt.role`, `gt.rig`, `gt.session`, `gt.actor`, `gt.agent`, `gt.town`

---

## Layout

### Two levels of view

**Level 1: Instance view** (`/waterfall`) — Swim lanes of all active/recent runs, grouped by rig, on a shared time axis.

**Level 2: Run detail view** (`/waterfall?run=<uuid>` or detail panel on click) — Hierarchical timeline of an individual run from `agent.instantiate` to `session.stop`.

### Global structure

```
┌─────────────────────────────────────────────────────────────────────┐
│ nav: [Dashboard] [Flow] [Waterfall*] [Sessions] [Beads] ...        │
│      window: [1h] [24h] [7d] [30d] [custom range]                  │
├─────────────────────────────────────────────────────────────────────┤
│ INSTANCE: laptop:gt   town: /Users/pa/gt                            │
│ FILTERS BAR                                                         │
│ [Rig ▼] [Role ▼] [Agent ▼] [Event types ▼] [Search ___________]   │
├─────────────────────────────────────────────────────────────────────┤
│ SUMMARY CARDS                                                       │
│ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐            │
│ │ 12   │ │ 3    │ │ 847  │ │ 42   │ │$1.23 │ │ 2m30s│            │
│ │ Runs │ │ Rigs │ │Events│ │Beads │ │ Cost │ │ Span │            │
│ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘            │
├──────────────┬──────────────────────────────────────────────────────┤
│ SWIM LANES   │  TIME AXIS ──────────────────────────────────►       │
│              │  0s    30s    1m     1m30   2m     2m30    3m        │
│──────────────┼──────────────────────────────────────────────────────│
│              │                                                       │
│ ── fai ──    │  (rig header, collapsible)                           │
│              │                                                       │
│ fai/mayor    │  ████████████████████████████████████████████████──  │
│   API calls  │    ▪▪  ▪▪▪  ▪▪    ▪▪▪▪▪   ▪▪  ▪▪▪  ▪▪             │
│   tools      │     ◆  ◆◆    ◆      ◆◆◆    ◆    ◆                  │
│              │        ╔══▶ sling:bead-42 ══════════▶                │
│ fai/deacon   │         ░░░░░░░░░░░░░░░░░░░░░░░░────                │
│   API calls  │           ▪▪  ▪▪▪   ▪▪▪  ▪▪▪                       │
│              │              ╔══▶ mail → fai/witness ══▶             │
│ fai/witness  │               ░░░░░░░░░░░░░░░░░░░──                 │
│   API calls  │                 ▪▪  ▪▪  ▪▪  ▪▪                      │
│              │                                                       │
│ ── mol ──    │  (rig header, collapsible)                           │
│              │                                                       │
│ mol/witness  │       ██████████████████████████████████████████──── │
│ mol/polecat  │              ░░░░░░░░░░░░░░──                        │
│   ↑jana      │              ░░░░░░░░░──                             │
│              │                                                       │
├──────────────┴──────────────────────────────────────────────────────┤
│ DETAIL PANEL (click any element)                                    │
│ ┌───────────────────────────────────────────────────────────────┐   │
│ │ Run: 6ba7b810…  Role: witness  Rig: fai  Agent: witness      │   │
│ │ Started: 14:32:05  Duration: 1m42s  Cost: $0.3241            │   │
│ │                                                               │   │
│ │ [14:32:01] ● instantiate   claudecode/fai-witness             │   │
│ │ [14:32:05] ─ session.start                                   │   │
│ │ [14:32:06]   prime         polecat formula (2 KB)            │   │
│ │ [14:32:06] ▶ prompt.send   "You have bead gt-abc…"          │   │
│ │ [14:32:08] ◀ thinking      847 chars                         │   │
│ │ [14:32:10] ◀ text          "I'll review the assigned bead…"  │   │
│ │ [14:32:11] 🔧 tool_use     bd list --assignee=fai/witness    │   │
│ │ [14:32:11]   bd.call       list (38ms) ✓                     │   │
│ │ [14:32:11] ↩ tool_result   [{id:"bead-42"…}]                │   │
│ │ [14:32:15] 🔧 tool_use     Bash "git diff HEAD~1"            │   │
│ │ [14:32:18] ↩ tool_result   (320 lines)                       │   │
│ │ [14:32:25] ◀ text          "The changes look correct…"       │   │
│ │ [14:32:26] 🔧 tool_use     bd update bead-42 --status=done   │   │
│ │ [14:32:26] ■ done          COMPLETED                         │   │
│ │ [14:32:26] ─ session.stop                                    │   │
│ └───────────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────────┤
│ COMMUNICATION MAP (collapsible section)                             │
│                                                                     │
│   mayor ──sling──▶ deacon ──mail──▶ witness                        │
│     │                                    │                          │
│     └──────────── mail ◀─────────────────┘                         │
│                                                                     │
│   mayor ──spawn──▶ polecat/jana                                     │
│     │               │                                               │
│     └── nudge ──────┘                                               │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Swim lanes — detail

Each **agent run** (`agent.instantiate`) gets one horizontal swim lane. Lanes are grouped by **rig**, with collapsible rig headers. Inside each lane:

1. **Session bar**: solid colour bar spanning the full width (colour = role) from `session.start` to `session.stop` (or now if still running). Pulsing animation if in progress.

2. **API ticks**: small vertical marks on the session bar for each `claude_code.api_request`. Colour intensity = cost. Hover: model, tokens, cost, duration.

3. **Tool markers**: diamond markers below the bar for each `claude_code.tool_result`. Colour = success (green) / failure (red). Hover: tool name, command, duration.

4. **Inter-agent arrows**: horizontal arrows between lanes showing communications:
   - **Sling** (bead dispatch): solid arrow, labelled with the bead ID
   - **Mail** (send/deliver): wavy arrow, labelled with `msg.subject` or `msg.from→msg.to`
   - **Nudge**: dashed arrow
   - **Polecat spawn**: thick arrow to the child lane
   - **Done/escalate**: return arrow to the parent

   > **Note**: v1 defined an `assign` type (derived from `bd update --assignee`). This is not a native event — it is a heuristic. Show it as `bd.call` with `subcommand=update` and args containing `--assignee`, not as a first-class communication type.

5. **Lifecycle bead overlay**: optional coloured segments on session bars showing which bead is currently being worked on (derived from `bd.call` create/update correlation).

### Time axis

- Shared horizontal time axis at the top, auto-scaled to the window
- Tick marks at sensible intervals (every 10s, 30s, 1m, 5m, etc.)
- Vertical grid lines (subtle) for alignment
- Zoom: scroll wheel on the timeline area
- Pan: click-drag on the timeline area
- Current-time marker (if live/recent view): vertical red line

### Filters

| Filter | Type | Source |
|--------|------|--------|
| Rig | multi-select dropdown | `rig` attribute on `agent.instantiate` |
| Role | multi-select dropdown | `role`: mayor, deacon, witness, refinery, polecat, dog, boot, crew |
| Agent | multi-select dropdown | `agent_name` or `session_id` |
| Event types | checkbox group | Runs, API calls, Tool calls, BD calls, Slings, Mails, Nudges, Spawns |
| Search | text input | Full-text search on content, bead IDs, tool names |

Filters are URL query-param driven (`?rig=fai&role=witness&types=api,tool`) for link sharing.

### Detail panel — right panel (Chrome DevTools Network style)

Clicking any row in the waterfall opens a **right side panel** displayed alongside the waterfall (vertical split layout, ~40% of width), exactly like the Chrome DevTools Network detail panel. The waterfall resizes to make room — it does not disappear.

```
┌──────────────────────────┬──────────────────────────────────────┐
│  WATERFALL (60%)         │  DETAIL PANEL (40%)                  │
│                          │                                       │
│  fai/mayor  ████████─    │  ┌─ fai-witness / polecat ──────[✕]─┐│
│  fai/deacon  ░░░░░──     │  │  run: 6ba7b810…  dur: 4m32s      ││
│► fai/witness ░░░────     │  │  rig: wyvern  cost: $0.0341       ││
│  mol/witness ████──      │  ├──────────────────────────────────┤│
│                          │  │ [Overview][Prompt][Conversation]  ││
│                          │  │ [BD Calls][Mails][Timeline]       ││
│                          │  ├──────────────────────────────────┤│
│                          │  │                                   ││
│                          │  │  (active tab content)             ││
│                          │  │                                   ││
│                          │  └──────────────────────────────────┘│
└──────────────────────────┴──────────────────────────────────────┘
```

#### Panel tabs (based on element clicked)

**Click on a run lane** → run panel with 6 tabs:

| Tab | Content |
|-----|---------|
| **Overview** | Metadata: `run.id`, `role`, `rig`, `agent_name`, `agent_type`, `session_id`, `instance`, `started_at`, `ended_at`, duration, total cost, event count |
| **Prompt** | Full text of `prompt.send` received by the agent (`keys` attribute if available, otherwise `keys_len` + missing notice). Monospace, dark background, scrollable. Render Markdown if the prompt contains it. |
| **Conversation** | All `agent.event` records for the run, shown as chat bubbles: `thinking` (lavender, italic), `assistant/text` (dark green, right-aligned), `user/text` (light green, left-aligned), `tool_use` (amber, code block), `tool_result` (blue, code block). Full content, no truncation. Scrollable. |
| **BD Calls** | Table of all `bd.call` records: `time`, `subcommand`, `args`, `duration_ms`, `status`. If `GT_LOG_BD_OUTPUT=true`, show `stdout` in a collapsible `<details>`. |
| **Mails** | Table of all `mail` events: `operation`, `msg.from`, `msg.to`, `msg.subject`, `msg.priority`. Full body (`msg.body`) in a collapsible `<details>`. |
| **Timeline** | Mini-waterfall of this run only: same horizontal view as the global waterfall but zoomed to this run, with nested sub-events (see §Nesting). |

**Click on an API tick** → API panel with 2 tabs:

| Tab | Content |
|-----|---------|
| **Headers** | Model, `session.id`, timestamps, duration |
| **Tokens** | Table: input / output / cache_read tokens, cost USD. Proportional visual bar. |

**Click on a tool marker** → Tool panel with 2 tabs:

| Tab | Content |
|-----|---------|
| **Summary** | Tool name, duration, success/failure, `session.id` |
| **Parameters** | `tool_parameters` JSON formatted with syntax highlighting (JSON.stringify indent 2). |

**Click on a communication arrow** → Comm panel with 2 tabs:

| Tab | Content |
|-----|---------|
| **Info** | Type (`sling`/`mail`/`nudge`/`spawn`/`done`), source, target, timestamp, bead ID if applicable |
| **Bead** | For `sling`: full bead lifecycle from `/bead/{id}` (state transition table). For `mail`: full `msg.body`. |

#### Panel behaviour

- **Open**: slide-in from the right, 150ms animation
- **Close**: `✕` button top-right, or `Escape` key
- **Resize**: drag on the panel's left edge (width between 25% and 70%)
- **Tab persistence**: active tab memorised per type (run/api/tool/comm) within the session
- **Run navigation**: `↑` / `↓` keys to move to the previous/next run without closing the panel
- **External link**: "Open in full view" button → `/session/{session_id}` or `/bead/{id}`

### Communication map

Collapsible section below the waterfall showing a **node-link diagram** of all inter-agent communication within the window:

- Nodes = agents (coloured by role)
- Edges = communication events (slings, mails, spawns, nudges, dones)
- Edge thickness = frequency
- Edge label = count + last bead ID or mail subject
- Hover on a node: highlight all its communication edges
- Click on a node: filter the waterfall to that agent

### Colour coding

| Event | Colour |
|-------|--------|
| `agent.instantiate` | purple |
| `session.start` / `session.stop` | grey |
| `prime` / `prime.context` | blue |
| `prompt.send` | cyan |
| `agent.event` thinking | lavender |
| `agent.event` text assistant | dark green |
| `agent.event` tool_use | orange |
| `agent.event` tool_result | light orange |
| `agent.event` user | green |
| `bd.call` | red |
| `mail` | yellow |
| `sling` / `nudge` | pink |
| `mol.cook` / `mol.wisp` | teal |
| `mol.squash` / `mol.burn` | indigo |
| `bead.create` | sky blue |
| `done` COMPLETED | bright green |
| `done` ESCALATED / DEFERRED | bright orange |
| `status: "error"` | bright red border |

### Nesting rules (run detail view)

OTel logs do not carry a native parent span ID; the hierarchy is reconstructed by:
1. Grouping on `run.id`
2. Chronological ordering by `_time`
3. The following rules:

```
agent.instantiate                    ← absolute root (1 per run)
  ├─ session.start                   ← tmux lifecycle
  ├─ prime                           ← context injection
  ├─ prompt.send                     ← daemon → agent
  │
  ├─ agent.event [user/text]         ← received text message
  ├─ agent.event [user/tool_result]  ← received tool result
  │
  ├─ agent.event [assistant/thinking]
  ├─ agent.event [assistant/text]
  ├─ agent.event [assistant/tool_use]  ← tool call
  │    ↳ bd.call                         if tool = bd (time window)
  │    ↳ mail                            if tool = mail
  │    ↳ sling                           if tool = gt sling
  │    ↳ nudge                           if tool = gt nudge
  │
  ├─ mol.cook                        ← formula compilation
  │    ↳ mol.wisp                       wisp instantiation
  │         ↳ bead.create               child beads created
  │
  ├─ mol.squash / mol.burn           ← molecule lifecycle end
  ├─ agent.state_change              ← state transition
  ├─ done                            ← work item completed
  └─ session.stop                    ← lifecycle end
```

Any event without an inferable parent → shown flat.

---

## Implementation notes

### Existing code to reuse

- `data.go`: `loadSessions()`, `loadBeadLifecycles()`, `loadAPIRequests()`, `loadToolCalls()`, `loadBDCalls()`, `loadFlowEvents()`, `loadPaneOutput()`, `correlateClaudeSessions()` — usable typed structs
- `vl.go`: `vlQuery()` for VictoriaLogs queries, `extractStreamField()` for parsing `_stream` attributes
- `main.go`: existing handler pattern, template helpers (`roleColor`, `fmtTime`, `fmtDur`, `fmtCost`, etc.)
- `waterfall.go`: `rigFromSession()`, `loadWaterfallData()` — partially reusable, deep refactoring needed

### New data to load

1. **Runs**: `vlQuery(cfg.LogsURL, "agent.instantiate", limit, since, end)` — fields: `run.id`, `instance`, `town_root`, `agent_type`, `role`, `agent_name`, `session_id`, `rig`
2. **Slings**: `vlQuery(cfg.LogsURL, "sling", limit, since, end)` — fields: `run.id`, `bead`, `target`, `status`
3. **Mails (send)**: `vlQuery(cfg.LogsURL, "mail AND operation:send", limit, since, end)` — fields: `run.id`, `operation`, `msg.from`, `msg.to`, `msg.subject`, `msg.body`, `msg.thread_id`, `msg.priority`, `msg.type`, `status`
4. **Mails (all)**: `vlQuery(cfg.LogsURL, "mail", limit, since, end)` — general mail operations; run after the send query and dedup by event ID
5. **Nudges**: `vlQuery(cfg.LogsURL, "nudge", limit, since, end)` — fields: `run.id`, `target`, `status`
6. **Spawns**: `vlQuery(cfg.LogsURL, "polecat.spawn", limit, since, end)` — fields: `run.id`, `name`, `status`
7. **Dones**: `vlQuery(cfg.LogsURL, "done", limit, since, end)` — fields: `run.id`, `exit_type`, `status`
8. **Prime**: `vlQuery(cfg.LogsURL, "prime", limit, since, end)` — fields: `run.id`, `role`, `formula`, `hook_mode`, `status`
9. **Mol lifecycle**: `vlQuery(cfg.LogsURL, "mol.cook", ...)`, `vlQuery(cfg.LogsURL, "mol.wisp", ...)`, `vlQuery(cfg.LogsURL, "mol.squash", ...)`, `vlQuery(cfg.LogsURL, "mol.burn", ...)` — fields per event (see Data Sources table)
10. **Bead creation**: `vlQuery(cfg.LogsURL, "bead.create", limit, since, end)` — fields: `run.id`, `bead_id`, `parent_id`, `mol_source`

> **Suggestion**: query `agent.instantiate` first to get all `run.id`s in the window, then query all events with `run.id:<uuid1> OR run.id:<uuid2> OR …` to avoid N+1 queries. See `waterfall-spec.md §4.1`.

### Data pipeline

```
loadWaterfallV2Data(cfg, since, filters) →
  1. Load agent.instantiate  → run list → group by rig
  2. Load session.start/stop → run durations
  3. Load prime              → startup context per run
  4. Load API requests       → assign to runs via correlateClaudeSessions() + run.id
  5. Load tool calls         → assign to runs via session.id
  6. Load agent events       → assign to runs via native_session_id + run.id
  7. Load BD calls           → extract slings, assigns, creates
  8. Load slings/mails       → build communication edges (source run → target)
  9. Load spawns/dones       → build lifecycle edges
  10. Load mol.* / bead.create → attach to runs by run.id
  11. Compute time axis      → min(started_at) to max(ended_at or now)
  12. Apply filters          → rig, role, agent, event type
  13. Serialize to JSON      → send to frontend for rendering
```

### VictoriaLogs queries

```
# All recent runs (instance view)
GET /select/logsql/query?query=_msg:agent.instantiate AND instance:laptop:gt AND _time:[now-1h,now]&limit=100

# All events for a run
GET /select/logsql/query?query=run.id:<uuid>&limit=10000

# Filter by rig
GET /select/logsql/query?query=_msg:agent.instantiate AND rig:fai

# Filter by role
GET /select/logsql/query?query=_msg:agent.instantiate AND role:polecat

# Mol lifecycle for a run
GET /select/logsql/query?query=(mol.cook OR mol.wisp OR mol.squash OR mol.burn OR bead.create) AND run.id:<uuid>

# Mail send events (targeted query to avoid limit cutoff)
GET /select/logsql/query?query=mail AND operation:send&limit=500
```

### Frontend rendering

The waterfall MUST be rendered client-side (JavaScript + Canvas or SVG) for interactivity (zoom, pan, hover, click). The Go handler serves:

1. An HTML page with the shell (nav, filters, summary cards, detail panel)
2. A `<script>` block with waterfall data as JSON: `const DATA = {{.JSONData}};`
3. JavaScript that renders the waterfall in a `<canvas>` or SVG container

Use Canvas for performance (hundreds of events). SVG is appropriate for the communication map.

### API endpoint

Add `GET /api/waterfall.json?window=24h&rig=fai&role=witness` returning the structured JSON data. This allows:
- The `/waterfall` page to fetch data dynamically (filter changes without full reload)
- A separate frontend to consume the same API

### JSON shape

```typescript
interface WaterfallEvent {
  id:        string;       // VictoriaLogs internal ID
  run_id:    string;       // GASTOWN run UUID (GT_RUN)
  body:      string;       // event name ("bd.call", "agent.event", "mail", "mol.cook", …)
  timestamp: string;       // RFC3339
  severity:  "info" | "error";
  attrs: {
    // Present on all events
    instance?:          string;
    town_root?:         string;
    session_id?:        string;
    rig?:               string;
    role?:              string;   // Gastown role on agent.instantiate/session.*
                                  // LLM role ("assistant"/"user") on agent.event
    agent_type?:        string;
    agent_name?:        string;
    status?:            string;
    // agent.event
    event_type?:        string;
    "agent.role"?:      string;  // "assistant" | "user" (LLM role, alias of role on agent.event)
    content?:           string;  // full content — no truncation
    native_session_id?: string;
    // agent.state_change
    agent_id?:          string;
    new_state?:         string;
    hook_bead?:         string;  // bead ID being processed; empty string if none
    // bd.call
    subcommand?:        string;
    args?:              string;
    duration_ms?:       number;
    stdout?:            string;
    stderr?:            string;
    // mail
    "msg.id"?:          string;
    "msg.from"?:        string;
    "msg.to"?:          string;
    "msg.subject"?:     string;
    "msg.body"?:        string;  // full body — no truncation
    "msg.thread_id"?:   string;
    "msg.priority"?:    string;
    "msg.type"?:        string;
    // prime
    formula?:           string;
    hook_mode?:         boolean;
    // sling
    bead?:              string;
    target?:            string;
    // done
    exit_type?:         string;
    // mol lifecycle
    formula_name?:      string;
    wisp_root_id?:      string;
    bead_id?:           string;
    mol_id?:            string;
    done_steps?:        number;
    total_steps?:       number;
    digest_created?:    boolean;
    children_closed?:   number;
    // bead.create
    parent_id?:         string;
    mol_source?:        string;
    [key: string]:      unknown;
  };
}

interface WaterfallRun {
  run_id:      string;
  instance:    string;
  town_root:   string;
  agent_type:  string;
  role:        string;
  agent_name:  string;
  session_id:  string;
  rig:         string;
  started_at:  string;
  ended_at?:   string;      // present if session.stop received
  duration_ms?: number;
  running:     boolean;
  cost?:       number;      // from claude_code.api_request
  events:      WaterfallEvent[];
}

interface WaterfallInstance {
  instance:   string;
  town_root:  string;
  window:     { start: string; end: string };
  summary: {
    runCount:      number;
    rigCount:      number;
    eventCount:    number;
    beadCount:     number;
    totalCost:     number;
    totalDuration: string;
  };
  rigs: Array<{
    name:      string;
    collapsed: boolean;
    runs:      WaterfallRun[];
  }>;
  communications: Array<{
    time:      string;
    type:      "sling" | "mail" | "nudge" | "spawn" | "done";
    from:      string;   // run_id or actor (rig/role)
    to:        string;
    beadID?:   string;
    label:     string;
    // mail only
    subject?:  string;
    body?:     string;
  }>;
  beads: Array<{
    id:        string;
    title:     string;
    type:      string;
    state:     string;
    createdBy: string;
    assignee:  string;
    createdAt: string;
    doneAt?:   string;
  }>;
}
```

> **Note**: v1 had `rigs > lanes > apiCalls/toolCalls/agentEvents` (artificial separation). The new shape normalises everything as `WaterfallEvent[]` inside each `WaterfallRun`, aligned with the TypeScript reference. `apiCalls` and `toolCalls` from `claude_code.*` remain separate in `events` with their specific `body`.

> **Note**: The `"assign"` type in `communications` (v1) is removed — it does not exist as a native event. Bead assignment via `bd update --assignee` is visible in `bd.call` events, not as inter-agent communication.

---

## Environment variables

| Variable | Set by | Role |
|----------|--------|------|
| `GT_RUN` | tmux session env + subprocess | run UUID, waterfall correlation key |
| `GT_OTEL_LOGS_URL` | daemon startup | VictoriaLogs OTLP endpoint |
| `GT_OTEL_METRICS_URL` | daemon startup | VictoriaMetrics OTLP endpoint |
| `GT_LOG_AGENT_OUTPUT` | operator | opt-in Claude JSONL streaming |
| `GT_LOG_BD_OUTPUT` | operator | opt-in bd stdout/stderr content |
| `GT_LOG_PANE_OUTPUT` | operator | opt-in raw tmux pane output |

`GT_RUN` is surfaced as `gt.run_id` in `OTEL_RESOURCE_ATTRIBUTES` for all `bd` sub-processes, correlating their telemetry to the parent run.

---

## Interactions

| Action | Result |
|--------|--------|
| Hover over session bar | Lightweight tooltip: run.id (8 chars), role, rig, duration, cost |
| **Click on a lane (run)** | **Right panel slides in: Overview / Prompt / Conversation / BD Calls / Mails / Timeline tabs** |
| Hover over API tick | Tooltip: model, tokens, cost, latency |
| **Click on API tick** | **Right panel: Headers / Tokens tabs** |
| Hover over tool marker | Tooltip: tool name, duration, success |
| **Click on tool marker** | **Right panel: Summary / Parameters (formatted JSON) tabs** |
| Hover over communication arrow | Highlight source + target lanes, comm label |
| **Click on communication arrow** | **Right panel: Info / Bead or Info / Mail (full body) tabs** |
| Scroll wheel on timeline | Zoom in/out centred on cursor |
| Click-drag on timeline | Pan left/right |
| Click rig header | Collapse/expand rig group |
| Click node in comm map | Filter waterfall to that agent |
| `Escape` key | Close right panel |
| `↑` / `↓` keys (panel open) | Previous / next run without closing panel |
| Drag panel left edge | Resize panel width (25%–70%) |
| "Open in full view" button | Navigate to `/session/{session_id}` or `/bead/{id}` |

---

## Non-goals (v1)

- Real-time streaming (SSE/WebSocket) — use `/live-view` for that
- Editable state (no bead updates from this view)
- Historical diff (comparing two time windows)
- Mobile layout

---

## Acceptance criteria

1. `/waterfall` renders a horizontal Canvas timeline with swim lanes grouped by rig
2. Each swim lane corresponds to a `run.id` from `agent.instantiate`
3. All active filters are reflected in URL query params and persist on reload
4. **Click on a lane opens the right panel (vertical split) with all 6 tabs**
5. **Prompt tab shows full `prompt.send` text (`keys`) in monospace, untruncated**
6. **Conversation tab shows all `agent.event` records as chat bubbles, full content**
7. **BD Calls tab lists all `bd.call` records with collapsible `stdout`**
8. **Mails tab lists all `mail` events with collapsible `msg.body`**
9. **Panel is resizable by dragging its left edge**
10. **`↑`/`↓` keys navigate between runs without closing the panel**
11. Inter-agent communication arrows render between the correct swim lanes
12. Zoom/pan works smoothly for up to 50 runs and 5000 events
13. `/api/waterfall.json` returns complete structured data
14. Communication map section renders a readable node-link diagram
15. Dark theme consistent with existing gastown-trace pages

---

## Implementation status (reference: waterfall-spec.md §7)

| Component | Status |
|-----------|--------|
| `run.id` generated at spawn (lifecycle, polecat, witness, refinery) | ✅ |
| `GT_RUN` propagated via tmux env + subprocess `agent-log` | ✅ |
| `GT_RUN` in `OTEL_RESOURCE_ATTRIBUTES` for bd | ✅ |
| `run.id` injected into every OTel event | ✅ |
| `agent.instantiate` with `instance`, `role`, `town_root` | ✅ |
| `RecordMailMessage` with full content | ✅ (call sites added in `mail/`) |
| `agent.event` content without truncation | ✅ |
| bd stdout/stderr without truncation | ✅ |
| `agent.state_change` with `hook_bead: string` (replaces `has_hook_bead: bool`) | ✅ |
| `mol.cook` / `mol.wisp` recorder functions + metric counters | ✅ |
| `mol.squash` / `mol.burn` recorder functions | ✅ |
| `bead.create` recorder function | ✅ |
| Full prompt text in `prompt.send` (`keys` attribute) | ⬜ P1 |
| `RecordMailMessage` called from `mail/router` + `delivery` | ⬜ P2 |
| Bead ID of work item in `agent.instantiate` | ⬜ P2 |
| Token usage from Claude JSONL | ⬜ P3 |
| **Right panel with tabs (Overview/Prompt/Conversation/BD/Mails/Timeline)** | ⬜ to implement |
| Frontend waterfall v2 (base) | ✅ implemented |
