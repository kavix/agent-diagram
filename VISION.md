# The Vision: The Next Software Engineering Tool in the AI Era

> **"agent-diagram is the interface adapter between the AI agent, the software engineer, formal diagramming, and human hands typing and directing tools."**

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
| [#8](https://github.com/kavix/agent-diagram/issues/8) | **Interactive Terminal TUI Mode with Bubble Tea** (Pan, Zoom, Event Step-Through) | Intermediate |
| [#9](https://github.com/kavix/agent-diagram/issues/9) | **High-Resolution Terminal Graphic Protocols** (Kitty, iTerm2 Inline, Sixel) | Advanced |
| [#10](https://github.com/kavix/agent-diagram/issues/10) | **Git Graph (`gitGraph`) Visualization** for Branching & Worktrees | Beginner / Good First Issue |
| [#11](https://github.com/kavix/agent-diagram/issues/11) | **Entity-Relationship and Class Diagrams** (`erDiagram` / `classDiagram`) | Intermediate |
| [#12](https://github.com/kavix/agent-diagram/issues/12) | **Code-Aware Node Navigation** (Jump from Diagram Node to `file:line`) | Advanced |
| [#13](https://github.com/kavix/agent-diagram/issues/13) | **Packaging: Homebrew Formula & GoReleaser** Automated Distribution | Beginner / Good First Issue |

---

## 4. How to Get Involved

1. Check out [`CONTRIBUTING.md`](CONTRIBUTING.md) for local environment setup.
2. Pick any issue labeled `good first issue` or `help wanted` on [GitHub Issues](https://github.com/kavix/agent-diagram/issues).
3. Open a pull request!

---

## 5. The Agentic Era: Why Standard UML Falls Short

Traditional software architecture is **deterministic**: a service makes an API call and either succeeds or fails. Agentic AI architectures are **probabilistic and autonomous**—they feature non-deterministic looping, dynamic reasoning, and stateful memory that grows with every iteration.

Standard sequence diagrams fail here because they assume a linear, predictable timeline with bounded latency. `agent-diagram` is evolving to address this with **AgentUML**: a formal extension of Mermaid notation for the agentic era.

### Core Agentic Coordination Patterns

```
  ReAct Loop          Orchestrator-Worker     Hierarchical            Generator-Critic
  (Risk: LOW)         (Risk: HIGH ⚠)          (Risk: MEDIUM)          (Risk: LOW)

  User                User                    User                    User
   │                   │                       │                       │
   ▼                   ▼                       ▼                       ▼
 Agent              Guardrail               Manager               Generator
   │                   │                    ╱     ╲                    │
  loop             Orchestrator        Supervisor  Supervisor        Critic
  Reason              ╱ │ ╲            ╱    ╲      ╱    ╲             │
  ──►Tool          W1  W2  W3       Leaf  Leaf  Leaf  Leaf        loop until
  ◄──Obs            └──┬──┘          └───┬──┘    └───┬──┘          pass rubric
  until done       Synthesise        Compress    Compress
```

### Context-Window Overflow Risk by Pattern

| Pattern | Risk | Root Cause |
| :--- | :--- | :--- |
| **Orchestrator-Worker** | 🔴 HIGH | Orchestrator accumulates **all** worker outputs in one context window. With N workers × avg response size, overflow is linear and predictable. |
| **Hierarchical** | 🟡 MEDIUM | Context is partitioned per level, but manager nodes aggregate children's summaries — lossy compression can discard critical details. |
| **ReAct / Generator-Critic** | 🟢 LOW | Single context window with natural loop bounds; observation history can be truncated after N steps. |

**For current projects**: The **Orchestrator-Worker pattern** carries the highest practical risk. As you scale the number of specialised workers (web search, code analysis, SQL query, etc.), the orchestrator's context grows linearly. The mitigation is to externalise worker results to a `<<memory_semantic>>` vector store and have the orchestrator retrieve only the top-k relevant chunks rather than accumulating full outputs.

### AgentUML Notation (in `agent-diagram`)

`agent-diagram` implements AgentUML stereotypes in its internal AST:

- `<<orchestrator>>` — routing node, does not execute business logic
- `<<agent>>` — autonomous reasoning loop with tool access
- `<<llm>>` — stateless model inference endpoint
- `<<tool>>` / `<<action>>` — deterministic external API
- `<<memory_working>>` — short-term context window (token-bounded)
- `<<memory_semantic>>` — long-term vector store (RAG)
- `<<guardrail>>` — policy enforcement gate

See [`docs/AGENT_UML_SPEC.md`](docs/AGENT_UML_SPEC.md) for the full specification and [`examples/`](examples/) for runnable diagrams of all four patterns.

