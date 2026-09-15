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
- **Flowcharts & DAGs (`flowchart TD`, `flowchart LR`, `graph TD`, `graph LR`)**:
  - Rectangular `[text]`, rounded `(text)`, diamond `{condition}`, stadium `([text])`, cylinder `[(db)]`
  - Directed edges `-->`, thick edges `==>`, dotted edges `-.->`, undirected `---`
  - Inline edge annotations (`-->|Yes|` or `-- Yes -->`)
  - Topological layer calculation & multi-layer vertical connector routing

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

## AI Agent Integration (MCP)

`agent-diagram` implements a standard [Model Context Protocol (MCP)](https://modelcontextprotocol.io) stdio server over JSON-RPC 2.0.

### 1. Antigravity CLI (`agy`)

Add `agent-diagram` to your global or project configuration:

`~/.gemini/config/mcp_config.json`:
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

Now whenever you ask `agy`:
> *"Explain how the StatefulSet resize reconciliation works in Kueue"*

`agy` automatically calls `render_diagram` and prints the width-adapted visualization directly into your terminal session!

### 2. Claude Code

Register with Claude Code CLI:

```bash
claude mcp add agent-diagram -- agent-diagram mcp
```

Verify it's connected:

```bash
claude mcp list
```

### 3. Cursor / Windsurf

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

## Contributing

We welcome contributions from the open-source community! Check out [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development environment setup
- Architectural deep dive
- **Step-by-step guide to adding new diagram types** (`stateDiagram-v2`, `gitGraph`, `erDiagram`)
- Pull Request guidelines

---

## License

MIT © [Kavindu Sachinthe](LICENSE)
