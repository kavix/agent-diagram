# The Vision: The Next Software Engineering Tool in the AI Era

> *"The terminal is not a legacy text interface—it is the universal canvas where autonomous AI agents and human engineers collaborate in real time."*

---

## 1. The Problem in Modern Software Engineering

Software systems have grown enormously complex: distributed microservices, Kubernetes reconciliation controllers (e.g. Kueue, Gateway API), event buses, and asynchronous workflows.

When AI coding agents (such as Google Antigravity, Claude Code, Codex, or Cursor) reason about this code, they generate textual explanations and Mermaid specifications. But when presented to human developers:
1. **Terminals cannot display raw Mermaid markdown**.
2. **Standard ASCII tools break or overflow screen boundaries**.
3. **Switching to browser tabs disrupts the terminal developer flow**.

`agent-diagram` exists to eliminate this friction by becoming the **universal visualization engine for the terminal in the AI era**.

---

## 2. The Four Pillars of `agent-diagram`

```
┌─────────────────────────────────────────────────────────────┐
│                    Universal UML Grammar                     │
│  Sequence Diagrams  •  Flowcharts  •  State Machines  •  Git │
└──────────────────────────────┬──────────────────────────────┘
                               │
               ┌───────────────┴───────────────┐
               ▼                               ▼
     Adaptive Geometry                 Code Intelligence
   • Full 2D Spatial Grid            • Node ➔ Source Line Mapping
   • 1D Compact Timeline             • Click to Jump (VS Code/Nvim)
   • 0D Narrow Fallbacks             • Execution Trace Debugging
               │                               │
               └───────────────┬───────────────┘
                               ▼
                    Autonomous Agent Loop
            AI Agent writes ➔ agent-diagram renders
            Agent verifies ➔ Real-time user feedback
```

### Pillar 1: Universal UML & System Models
Support every conceptual model an AI coding agent reasons about:
- **Sequence Diagrams**: Tracing RPCs, HTTP APIs, and reconciler handoffs.
- **Flowcharts & DAGs**: Visualizing decision trees, PR CI pipelines, and controller loops.
- **State Machines (`stateDiagram-v2`)**: Modeling Pod lifecycles, connection states, and worker tasks.
- **Git Graphs (`gitGraph`)**: Visualizing branches, worktrees, and rebase trajectories.
- **Class & ER Diagrams**: Mapping struct hierarchies, interfaces, and database schemas.

### Pillar 2: Dynamic Terminal Geometry Adaptation
The terminal width is an active input to our layout engine:
- If width is plentiful ($\ge 100$ cols), draw 2D Unicode grids.
- If width is constrained (laptop screen, split pane, 50–99 cols), adapt automatically to compact numbered timelines.
- Never crash, never truncate without semantics, and never emit "exceeds terminal width" warnings.

### Pillar 3: Code Intelligence & Interactive Navigation
Diagrams must not be passive text—they are interactive maps into the codebase:
- Clicking or pressing `Enter` on a node jumps directly to the source file and line in your editor (`code --goto file.go:245` or `nvim +245 file.go`).
- Real-time step-through execution of distributed traces.

### Pillar 4: Autonomous Agent Integration (MCP)
Built from day one with the **Model Context Protocol (MCP)**:
- Antigravity CLI (`agy`), Claude Code, Codex, OpenCode, and Cursor can natively discover and call `render_diagram`.
- Sub-5ms startup with zero CGO dependencies ensures synchronous turn speed.

---

## 3. Open Issues for OSS Contributors

We have opened dedicated issues on GitHub for community contributors:

| Issue | Title | Difficulty |
| :--- | :--- | :--- |
| [#1](https://github.com/kavix/agent-diagram/issues/1) | **Interactive Terminal TUI Mode with Bubble Tea** (Pan, Zoom, Event Step-Through) | Intermediate |
| [#2](https://github.com/kavix/agent-diagram/issues/2) | **High-Resolution Terminal Graphic Protocols** (Kitty, iTerm2 Inline, Sixel) | Advanced |
| [#3](https://github.com/kavix/agent-diagram/issues/3) | **Git Graph (`gitGraph`) Visualization** for Branching & Worktrees | Beginner / Good First Issue |
| [#4](https://github.com/kavix/agent-diagram/issues/4) | **Entity-Relationship and Class Diagrams** (`erDiagram` / `classDiagram`) | Intermediate |
| [#5](https://github.com/kavix/agent-diagram/issues/5) | **Code-Aware Node Navigation** (Jump from Diagram Node to `file:line`) | Advanced |
| [#6](https://github.com/kavix/agent-diagram/issues/6) | **Packaging: Homebrew Formula & GoReleaser** Automated Distribution | Beginner / Good First Issue |

---

## 4. How to Get Involved

1. Check out [`CONTRIBUTING.md`](CONTRIBUTING.md) for local environment setup.
2. Pick any issue labeled `good first issue` or `help wanted` on [GitHub Issues](https://github.com/kavix/agent-diagram/issues).
3. Open a pull request!
