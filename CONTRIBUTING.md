# Contributing to `agent-diagram`

Thank you for your interest in contributing to `agent-diagram`! This document guides you through our architecture, development workflow, and how to add new diagram types and features.

---

## 1. Project Philosophy

1. **Terminal-Native First**: The primary output target is the terminal. Never assume infinite screen real-estate or modern graphical web engines.
2. **Width-Adaptive**: Terminal width is not an error constraint—it is an input to our layout algorithms. If width is narrow, the layout degrades gracefully (`Full` ➔ `Compact` ➔ `Narrow`).
3. **Decoupled AST**: Input formats (Mermaid, JSON, etc.) parse into a clean internal Go AST. Rendering backends (Unicode, ASCII, Sixel) consume only this AST.
4. **Zero Heavyweight Dependencies**: Single binary with fast startup (< 5ms) and no CGO dependencies so AI coding agents can launch it instantly via MCP.

---

## 2. Directory Structure

```
agent-diagram/
├── cmd/
│   └── agent-diagram/        # CLI & entrypoint (flag routing, stdio streaming)
├── internal/
│   ├── ast/                  # Core AST data models (Diagram interface, Sequence, Flowchart)
│   ├── parser/
│   │   └── mermaid/          # Mermaid grammar parsers
│   ├── layout/               # Terminal geometry, width measuring, topological layers
│   ├── renderer/             # Unicode box-drawing, ANSI colors, Full & Compact renderers
│   └── mcp/                  # Model Context Protocol stdio JSON-RPC 2.0 server
├── examples/                 # Real-world Mermaid sample files (.mmd)
├── .github/                  # CI workflows, PR templates, and Issue templates
├── Makefile                  # Build and test shortcuts
└── go.mod
```

---

## 3. Getting Started

### Prerequisites

- Go `1.22+` (or latest stable)
- `git`
- (Optional) `golangci-lint`

### Build & Run Tests

```bash
# Clone your fork
git clone https://github.com/<your-username>/agent-diagram.git
cd agent-diagram

# Run all unit and integration tests
make test

# Build local executable
make build

# Run against sample diagrams
./agent-diagram render examples/kueue_resize.mmd --width 120
./agent-diagram render examples/kueue_resize.mmd --width 70
./agent-diagram render examples/kueue_reconcile_flow.mmd --mode compact
```

---

## 4. How to Add a New Diagram Type (Step-by-Step)

Want to add support for **State Diagrams (`stateDiagram-v2`)**, **Git Graphs (`gitGraph`)**, or **Class/ER Diagrams (`erDiagram`)**? Follow this 5-step checklist:

### Step 1: Define the AST (`internal/ast/`)
Create a new file `internal/ast/<diagram_type>.go`:
```go
package ast

type StateDiagram struct {
    DiagramTitle string
    States       map[string]*StateNode
    Transitions  []*StateTransition
}

func (d *StateDiagram) Type() DiagramType {
    return "state"
}

func (d *StateDiagram) Title() string {
    return d.DiagramTitle
}
```

### Step 2: Implement the Parser (`internal/parser/mermaid/`)
Add a new parser file `internal/parser/mermaid/<diagram_type>.go`:
- Register the diagram header in `internal/parser/mermaid/parser.go` (e.g. `strings.HasPrefix(lower, "statediagram")`).
- Tokenize and populate your AST struct.
- Add unit tests in `<diagram_type>_test.go`.

### Step 3: Compute the Layout (`internal/layout/`)
Add layout logic in `internal/layout/<diagram_type>.go`:
- Inspect available width (`targetWidth`).
- Calculate minimal bounding boxes.
- Choose between `ModeFull` and `ModeCompact`.

### Step 4: Implement Terminal Renderer (`internal/renderer/`)
Add rendering logic in `internal/renderer/<diagram_type>.go`:
- Use `UnicodeBox` (or `ASCIIBox` if `opts.ASCIIOnly` is true).
- Use `Style(...)` for ANSI color support (can be toggled via `opts.NoColor`).
- Implement `renderStateFull(...)` and `renderStateCompact(...)`.
- Hook your renderer into `internal/renderer/renderer.go:Render()`.

### Step 5: Add Unit Tests & Examples
- Add sample `.mmd` file in `examples/`.
- Add test cases in `internal/renderer/renderer_test.go`.
- Run `make test` to ensure 100% green tests!

---

## 5. Pull Request Guidelines

1. **Branch Naming**: Use descriptive branch names like `feature/state-diagram`, `fix/flowchart-edge-regex`, `docs/contributing-update`.
2. **Atomic Commits**: Keep commits logical and concise.
3. **Tests Required**: Every new parser feature, shape, or layout mode must have accompanying unit tests.
4. **Formatting**: Always format your code with standard `go fmt ./...`.
5. **No Breaking Changes to MCP Schema**: Ensure the `render_diagram` MCP tool remains backward-compatible for all AI coding agents.

---

## 6. Code of Conduct

Please adhere to the standards described in [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). Be respectful and welcoming to all contributors!
