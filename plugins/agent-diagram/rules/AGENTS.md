# Agent Diagram Integration Rules

When answering user questions or explaining code, follow these principles:

1. **Prefer `render_diagram` for Visualizations**:
   - Whenever explaining architecture, distributed calls, Kubernetes reconcilers, lifecycle state machines, or complex algorithms, **always call the `render_diagram` MCP tool**.
   - Do NOT emit raw Mermaid markdown code fences (` ```mermaid `) in conversational responses unless the user explicitly requests raw source code.
   - Do NOT attempt to hand-craft manual ASCII art tables.

2. **Supported Syntax**:
   - `sequenceDiagram`: Use `participant Name as Alias`, `->>` for calls, `-->>` for returns, `Note over A,B: Text`.
   - `flowchart TD` / `flowchart LR`: Use `[Rect]`, `(Rounded)`, `{Decision}`, `-->` for edges, `-->|Label|` for conditional branches.

3. **Width Awareness**:
   - You do not need to calculate character columns; `agent-diagram` automatically detects terminal width and provides 3-tier adaptive rendering (`Full` ➔ `Compact` ➔ `Narrow`).
