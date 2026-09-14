# Agent Diagram (`agent-diagram`)

[![CI](https://github.com/kavix/agent-diagram/actions/workflows/ci.yml/badge.svg)](https://github.com/kavix/agent-diagram/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kavix/agent-diagram)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![MCP Compatible](https://img.shields.io/badge/MCP-Protocol%202024--11--05-brightgreen.svg)](https://modelcontextprotocol.io)

**Terminal-native, width-adaptive diagram visualization engine for AI coding agents (`agy`, Claude Code, Cursor) and developer CLIs.**

AI coding agents love generating Mermaid diagrams to explain code paths, reconciliation loops, and distributed interactions. But terminal UIs choke on them:

```text
Diagram exceeds terminal width (141 > 80 cols)
Displayed as raw code block.
```

Or worse, models emit broken ASCII art with misaligned columns.

`agent-diagram` parses Mermaid into an internal Abstract Syntax Tree (AST), inspects current terminal geometry, and renders pixel-perfect Unicode terminal diagrams that dynamically adapt to the available width.

---

## Visual Demonstration

### Full Lifeline Mode (Terminal Width ≥ 100 cols)

```text
 ┌─────────────┐             ┌──────────┐             ┌────────────┐             ┌────────────────┐
 │ StatefulSet │             │ Workload │             │ Reconciler │             │ Kubernetes API │
 └──────┬──────┘             └─────┬────┘             └──────┬─────┘             └────────┬───────┘
        │                          │                         │                            │
        │                          │ 1. Reconcile scale to 0 │                            │
        │                          ◄─────────────────────────┤                            │
        │                          │                         │ 2. releaseReservation()    │
        │                          │                         ├────────────────────────────►
        │                          │                         │ 3. Update PodSets Count = 5│
        │                          │                         ├────────────────────────────►
        │                          │                         │ 4. clearOnHold()           │
        │                          │                         ├┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄►
```

### Compact Mode (Terminal Width 50–99 cols)

When your terminal is split into panes or resized, the same diagram automatically adapts:

```text
Sequence Diagram (Compact)
────────────────────────────────────────────────────────────
 1. Reconciler ──► Workload
    Reconcile scale to zero
 2. Reconciler ──► Kubernetes API
    releaseScaleDownReservation()
 3. Reconciler ──► Kubernetes API
    Update PodSets[0].Count = 5
 4. Reconciler ┄┄► Kubernetes API
    clearOnHold()

 ◈ Note [Rec, Client]: Resize completed successfully
```

### Flowchart / DAG Mode (`flowchart TD`)

```text
  ┌─────────────────┐
  │ Start Reconcile │
  └────────┬────────┘
           │
           ▼
  ◇──────────────────◇
  │ Workload OnHold? │
  ◇─────────┬────────◇
            │
            ▼
  ┌───────────────────┐    ◇─────────────────────◇
  │ Clear OnHold Flag │    │ Scale Down Pending? │
  └─────────┬─────────┘    ◇──────────┬──────────◇
            │                         │
            ▼                         ▼
  ┌───────────────────┐    ┌───────────────────┐
  │ Update API Server │    │ Sync PodSet Count │
  └───────────────────┘    └───────────────────┘
```

---

## Supported Diagram Types

- **Sequence Diagrams (`sequenceDiagram`)**:
  - `participant` and `actor` with custom aliases (`participant STS as StatefulSet`)
  - Auto-registration of undeclared participants
  - Solid arrows `->>`, `->`, dotted return arrows `-->>`, `-->`
  - Cross terminations `-x`, `--x`
  - `autonumber` and event numbering
  - `Note over`, `Note left of`, `Note right of`
- **State Diagrams (`stateDiagram-v2`, `stateDiagram`)**:
  - Initial `[*]` and terminal `[*]` transitions
  - State aliases (`state "Name" as ID`)
  - Transition event triggers (`State1 --> State2 : Trigger`)
  - State descriptions and annotations
- **Flowcharts & DAGs (`flowchart TD`, `flowchart LR`, `graph TD`, `graph LR`)**:
  - Rectangular `[text]`, rounded `(text)`, diamond `{condition}`, stadium `([text])`, cylinder `[(db)]`
  - Directed edges `-->`, thick edges `==>`, dotted edges `-.->`, undirected `---`
  - Inline edge annotations (`-->|Yes|` or `-- Yes -->`)
  - Topological layer calculation & multi-layer vertical connector routing

---

## AgentUML: Diagrams for AI Agentic Workflows

AI agents are not microservices. They loop non-deterministically, grow their context window with every observation, and can delegate to hierarchies of sub-agents. `agent-diagram` supports **AgentUML** — a stereotype-annotated extension of Mermaid sequence syntax that makes these dynamics visible.

### Supported AgentUML Stereotypes

Append `<<stereotype>>` to any participant name:

| Stereotype | Role |
| :--- | :--- |
| `<<orchestrator>>` | Central routing LLM; decomposes & delegates tasks |
| `<<agent>>` | Autonomous ReAct loop with tool-calling |
| `<<llm>>` | Stateless model inference endpoint |
| `<<tool>>` / `<<action>>` | Deterministic external API |
| `<<memory_working>>` | Short-term context window (token-bounded) |
| `<<memory_semantic>>` | Long-term RAG vector store |
| `<<guardrail>>` | Policy enforcement gate |

### Example: Orchestrator-Worker Pattern ⚠ High Context-Window Risk

```bash
agent-diagram render examples/agent_orchestrator_worker.mmd
```

```text
Sequence Diagram — Orchestrator-Worker (AgentUML)
────────────────────────────────────────────────────────────────────────────────
  ┌──────┐    ┌──────────┐    ┌─────────┐    ┌──────┐    ┌──────┐    ┌──────┐
  │ User │    │Guardrail │    │Orchestr.│    │  W1  │    │  W2  │    │  W3  │
  └──┬───┘    └────┬─────┘    └────┬────┘    └──┬───┘    └──┬───┘    └──┬───┘
     │              │               │             │            │            │
     │─────────────►│ 1. Send prompt│             │            │            │
     │              │──────────────►│ 2. Approve  │            │            │
     │              │               │────────────►│ 7a. Task   │            │
     │              │               │─────────────┼───────────►│ 7b. Task   │
     │              │               │◄────────────│ 8a. Result⚠│            │
     │              │               │◄────────────┼────────────│ 8b. Result⚠│
     │              │               │─────────────────────────►│ 7c. Task   │
  ◈ Note [Orchestr.,Mem]: ⚠ HIGH OVERFLOW RISK: context grows with each worker
     │◄─────────────┼───────────────│ 12. Response│            │            │
```

### Context-Window Overflow Risk by Pattern

| Pattern | Risk | Example |
| :--- | :--- | :--- |
| Orchestrator-Worker | 🔴 **HIGH** | `examples/agent_orchestrator_worker.mmd` |
| Hierarchical | 🟡 **MEDIUM** | `examples/agent_hierarchical.mmd` |
| ReAct / Generator-Critic | 🟢 **LOW** | `examples/agent_react_loop.mmd`, `examples/agent_generator_critic.mmd` |

See [`docs/AGENT_UML_SPEC.md`](docs/AGENT_UML_SPEC.md) for the full AgentUML specification, including probabilistic branch notation, guardrail boundaries, and production anti-patterns.

---

## Installation

### From Source

```bash
go install github.com/kavix/agent-diagram/cmd/agent-diagram@latest
```

Or clone and build locally:

```bash
git clone https://github.com/kavix/agent-diagram.git
cd agent-diagram
make build
# Binary created at ./agent-diagram
```

---

## CLI Usage

### Render a File

```bash
agent-diagram render examples/kueue_resize.mmd
```

### Stdin / Pipe Support

```bash
cat diagram.mmd | agent-diagram render
```

```bash
echo "flowchart TD\n  A[Start] --> B[End]" | agent-diagram render
```

### Force Layout Mode or Width

```bash
# Force 80 columns rendering
agent-diagram render diagram.mmd --width 80

# Force Compact or Full mode explicitly
agent-diagram render diagram.mmd --mode compact
agent-diagram render diagram.mmd --mode full

# Pure ASCII fallback without Unicode box characters
agent-diagram render diagram.mmd --ascii --no-color
```

---

## AI Agent Integration (MCP & Beyond)

`agent-diagram` implements a standard [Model Context Protocol (MCP)](https://modelcontextprotocol.io) stdio server over JSON-RPC 2.0 and supports OpenAI Function Calling, shell pipes, and native plugin bundles.

Detailed instructions for every platform can be found in [docs/AGENT_INTEGRATION.md](docs/AGENT_INTEGRATION.md).

### 1. Google Antigravity CLI (`agy`)
Run the 1-command installer to install the plugin bundle and configure MCP:
```bash
agent-diagram install-plugin
```

### 2. Anthropic Claude Code CLI
```bash
claude mcp add agent-diagram -- agent-diagram mcp
```

### 3. OpenAI / Codex & GitHub Copilot
Use the native JSON tool definition in Python / TypeScript SDK:
```python
tools = [{
    "type": "function",
    "function": {
        "name": "render_diagram",
        "description": "Render a Mermaid diagram into terminal Unicode",
        "parameters": {"type": "object", "properties": {"source": {"type": "string"}}, "required": ["source"]}
    }
}]
```

### 4. OpenCode CLI
In `~/.config/opencode/mcp.json`:
```json
{
  "mcpServers": {
    "agent-diagram": {
      "command": "agent-diagram",
      "args": ["mcp"]
    }
  }
}
```

### 5. Cursor & Windsurf
In `~/.cursor/mcp.json`:
```json
{
  "mcpServers": {
    "agent-diagram": {
      "command": "agent-diagram",
      "args": ["mcp"]
    }
  }
}
```

### 6. Aider & Shell Agents
Stream directly via standard input:
```bash
echo "flowchart TD\n A --> B" | agent-diagram render
```

---

## Architecture

```text
                  Input (Mermaid / JSON)
                           │
                           ▼
                  ┌─────────────────┐
                  │   Parser        │
                  │ (Lexer & Regex) │
                  └────────┬────────┘
                           │
                           ▼
                  ┌─────────────────┐
                  │   Unified AST   │
                  │ (Sequence/Flow) │
                  └────────┬────────┘
                           │
            ┌──────────────┴──────────────┐
            ▼                             ▼
   Terminal Geometry               Layout Engine
  (Columns / TTY probe)          (Topological/Grid)
            │                             │
            └──────────────┬──────────────┘
                           ▼
                  ┌─────────────────┐
                  │ Terminal        │
                  │ Renderer        │
                  │ (Full/Compact)  │
                  └────────┬────────┘
                           │
             ┌─────────────┴─────────────┐
             ▼                           ▼
        CLI Output                  MCP Server
      (stdout / pipe)             (stdio JSON-RPC)
```

---

## Roadmap & The Vision

Our mission is to establish `agent-diagram` as the definitive terminal-native visualization standard in the AI era. Read our strategic manifesto in [VISION.md](VISION.md).

We have active roadmap issues ready for OSS contributors:

| Issue | Title | Status / Area |
| :--- | :--- | :--- |
| [#8](https://github.com/kavix/agent-diagram/issues/8) | **Interactive TUI Mode with Bubble Tea** (Pan, Zoom, Event Step-Through) | `enhancement`, `ui` |
| [#9](https://github.com/kavix/agent-diagram/issues/9) | **High-Res Terminal Graphics Protocols** (Kitty, iTerm2 Inline, Sixel) | `graphics`, `protocol` |
| [#10](https://github.com/kavix/agent-diagram/issues/10) | **Git Graph (`gitGraph`) Visualization** for Branching & Worktrees | `good first issue` |
| [#11](https://github.com/kavix/agent-diagram/issues/11) | **Entity-Relationship & Class Diagrams** (`erDiagram`, `classDiagram`) | `diagram-type` |
| [#12](https://github.com/kavix/agent-diagram/issues/12) | **Code Intelligence Linkage** (Jump from Diagram Node to `file:line`) | `core-architecture` |
| [#13](https://github.com/kavix/agent-diagram/issues/13) | **Packaging & Distribution** (Homebrew Formula & GoReleaser) | `good first issue` |
| [#14](https://github.com/kavix/agent-diagram/issues/14) | **AgentUML Native Stereotype Parsing** (`<<agent>>`, `<<guardrail>>`, probabilistic arrows) | `agentic`, `enhancement` |

---

## Contributing

We welcome contributions from the open-source community! Check out [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development environment setup
- Architectural deep dive
- **Step-by-step guide to adding new diagram types**
- Pull Request guidelines

---

## License

MIT © [Kavindu Sachinthe](LICENSE)
