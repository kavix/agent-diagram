# AI Agent Integration Guide: `agent-diagram`

Comprehensive instructions for integrating `agent-diagram` across the entire ecosystem of AI coding assistants, including **Google Antigravity CLI (`agy`)**, **Anthropic Claude Code**, **OpenAI / Codex**, **OpenCode**, **Cursor**, **Windsurf**, and **Aider / Terminal CLI agents**.

---

## Supported Agent Platforms

| Agent / CLI Surface | Integration Mechanism | Protocol / Interface |
| :--- | :--- | :--- |
| **Google Antigravity (`agy`)** | Native Plugin or Global MCP | Stdio MCP (`~/.gemini/config/mcp_config.json`) |
| **Claude Code CLI** | CLI MCP Registration | Stdio MCP (`claude mcp add`) |
| **OpenAI / Codex** | Function Calling & Shell Tool | OpenAI JSON Tool Schema / Subprocess |
| **OpenCode CLI** | Native MCP Config | Stdio MCP (`~/.config/opencode/mcp.json`) |
| **Cursor & Windsurf** | Editor MCP Config | Stdio MCP (`~/.cursor/mcp.json`) |
| **Aider & Terminal Shells** | Direct Stdin Pipe / Subprocess | UNIX Pipes (`cat flow.mmd \| agent-diagram render`) |

---

## 1. Google Antigravity CLI (`agy`)

Antigravity CLI provides two methods for integration:

### Method A: Automated 1-Command Plugin Install (Recommended)

From your `agent-diagram` workspace:

```bash
agent-diagram install-plugin
```

This automatically:
- Installs the plugin into `~/.gemini/config/plugins/agent-diagram/`
- Registers the MCP server in `~/.gemini/config/mcp_config.json`
- Injects behavioral rules into `rules/AGENTS.md` so the agent proactively invokes `render_diagram`

### Method B: Manual MCP Configuration

Add the server to `~/.gemini/config/mcp_config.json`:

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

Add this rule to your project's `AGENTS.md` or `GEMINI.md`:

```markdown
# Visualizing Interactions & Architectures

Whenever explaining:
- Controller / reconciler event loops (e.g. Kubernetes, Kueue)
- Distributed network calls, APIs, or microservice RPCs
- Multi-component sequence interactions or state transitions
- Complex algorithms and decision flowcharts

**Always invoke the `render_diagram` tool** using valid Mermaid syntax (`sequenceDiagram` or `flowchart TD/LR`) rather than writing manual ASCII art or dumping unrendered markdown code blocks. The tool adapts automatically to the user's terminal geometry.
```

---

## 2. Anthropic Claude Code CLI

Claude Code has first-class support for MCP servers.

### Step 1: Register the Tool
```bash
claude mcp add agent-diagram -- agent-diagram mcp
```

### Step 2: Verify Connection
```bash
claude mcp list
```
Expected output:
```text
agent-diagram: agent-diagram mcp (running)
```

### Step 3: Add Steering Instruction (`CLAUDE.md`)
Add this to your repository's `CLAUDE.md`:
```markdown
## Diagram Rendering
When explaining multi-component flows, code execution paths, or architecture, prefer calling the `render_diagram` MCP tool with Mermaid syntax. Do not output raw Mermaid code fences or wide ASCII tables.
```

---

## 3. OpenAI / Codex & GitHub Copilot

For agents built on OpenAI models (`gpt-4o`, `o1`, Codex, or custom fine-tunes), define `render_diagram` as an OpenAI Function Calling tool.

### OpenAI Function Calling Schema (Python SDK)

```python
import json
import subprocess
from openai import OpenAI

client = OpenAI()

# 1. Define tool schema for Codex / OpenAI
diagram_tool = {
    "type": "function",
    "function": {
        "name": "render_diagram",
        "description": (
            "Render a Mermaid sequence diagram or flowchart as a terminal-native, "
            "width-adaptive Unicode visualization. Use this tool when explaining code "
            "execution, architecture, reconcilers, or distributed systems."
        ),
        "parameters": {
            "type": "object",
            "properties": {
                "source": {
                    "type": "string",
                    "description": "Mermaid diagram syntax (e.g. sequenceDiagram or flowchart TD/LR)"
                },
                "width": {
                    "type": "integer",
                    "description": "Optional terminal width (auto-detected if omitted)"
                },
                "mode": {
                    "type": "string",
                    "enum": ["auto", "full", "compact", "narrow"],
                    "description": "Layout mode (default: auto)"
                }
            },
            "required": ["source"]
        }
    }
}

# 2. Local execution handler
def execute_render_diagram(source: str, width: int = 0, mode: str = "auto") -> str:
    cmd = ["agent-diagram", "render", "--mode", mode]
    if width > 0:
        cmd.extend(["--width", str(width)])
    
    proc = subprocess.run(cmd, input=source, text=True, capture_output=True)
    return proc.stdout if proc.returncode == 0 else proc.stderr

# 3. Agent execution loop
response = client.chat.completions.create(
    model="gpt-4o",
    messages=[
        {"role": "system", "content": "When explaining code architecture or flows, always call render_diagram."},
        {"role": "user", "content": "Explain the Kueue StatefulSet resize flow."}
    ],
    tools=[diagram_tool]
)

tool_call = response.choices[0].message.tool_calls[0]
arguments = json.loads(tool_call.function.arguments)
terminal_art = execute_render_diagram(arguments["source"])
print(terminal_art)
```

---

## 4. OpenCode CLI

OpenCode natively supports MCP servers.

In `~/.config/opencode/mcp.json` (or `~/.opencode/config.json`):

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

## 5. Cursor & Windsurf IDEs

In Cursor or Windsurf, configure the tool in `~/.cursor/mcp.json` or within workspace settings:

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

## 6. Aider & Shell-Based AI CLIs (Piping & Subprocess)

For agents like Aider, Mentat, or custom shell-based REPLs that execute shell commands directly:

### Standard Input Pipe
Instruct the agent to stream Mermaid code through `agent-diagram render`:

```bash
printf "sequenceDiagram\nparticipant A\nparticipant B\nA->>B: Ping\nB-->>A: Pong\n" | agent-diagram render
```

### Temporary File Rendering
```bash
agent-diagram render /tmp/flow.mmd --width 80
```

---

## 7. Universal Agent Steering Rules

Regardless of which agent platform you use, add this prompt instruction to ensure the model proactively calls `agent-diagram`:

```text
Whenever explaining multi-component architectures, controller/reconciler loops, 
network protocols, or execution flows:
1. Always formulate the flow as valid Mermaid (sequenceDiagram or flowchart TD/LR).
2. Call the `render_diagram` tool with the source string.
3. Do NOT output raw ```mermaid markdown fences or hand-drawn ASCII art.
```
