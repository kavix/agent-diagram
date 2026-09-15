# Integrating `agent-diagram` into AI Coding Agents

This guide explains how to connect `agent-diagram` to AI coding assistants including **Google Antigravity CLI (`agy`)**, **Claude Code**, **Cursor**, **Windsurf**, and custom LLM agent pipelines via the **Model Context Protocol (MCP)** or CLI subprocess.

---

## 1. Antigravity CLI (`agy`)

Antigravity CLI discovers MCP servers from user configuration and presents them as native callable tools to Gemini agents.

### Step 1: Add to `mcp_config.json`

Add the server to your machine-wide configuration at `~/.gemini/config/mcp_config.json` (or within a workspace's `plugins/<name>/mcp_config.json`):

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

> **Tip:** If `agent-diagram` is not on your system `$PATH`, specify the absolute binary path (e.g. `/usr/local/bin/agent-diagram` or `/Users/<username>/go/bin/agent-diagram`).

### Step 2: Add Agent Rule (`GEMINI.md` or `AGENTS.md`)

Add the following rule to your repository root or `~/.gemini/config/rules/` to guide the agent on when to use `agent-diagram`:

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

## 2. Claude Code CLI

Claude Code has first-class support for MCP servers using the `claude mcp` CLI.

### Step 1: Register the MCP Server

```bash
claude mcp add agent-diagram -- agent-diagram mcp
```

### Step 2: Verify Registration

```bash
claude mcp list
```
You should see:
```text
agent-diagram: agent-diagram mcp (running)
```

### Step 3: Add Steering Instruction (`CLAUDE.md`)

In your repository's `CLAUDE.md`, add:

```markdown
## Diagram Rendering
When explaining multi-component flows, code execution paths, or architecture, prefer calling the `render_diagram` MCP tool with Mermaid syntax. Do not output raw Mermaid code fences or wide ASCII tables.
```

---

## 3. Cursor & Windsurf

In Cursor or Windsurf, configure MCP servers through the editor settings or config file.

In `~/.cursor/mcp.json` (or project `.cursor/mcp.json`):

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

## 4. Custom Python Agents (LangChain, LlamaIndex, OpenAI SDK)

If you are building custom AI agents in Python, you can invoke `agent-diagram` either via standard subprocess or MCP client:

### Option A: Subprocess Execution

```python
import subprocess
import shutil

def render_terminal_diagram(mermaid_code: str, width: int = 100) -> str:
    """Invokes agent-diagram CLI to convert Mermaid into terminal Unicode."""
    binary = shutil.which("agent-diagram")
    if not binary:
        return f"```mermaid\n{mermaid_code}\n```"
    
    proc = subprocess.run(
        [binary, "render", "--width", str(width), "--no-color"],
        input=mermaid_code,
        text=True,
        capture_output=True
    )
    if proc.returncode == 0:
        return proc.stdout
    return f"Render error: {proc.stderr}"
```

### Option B: Python Function Tool for Function Calling

```python
from pydantic import BaseModel, Field

class RenderDiagramInput(BaseModel):
    source: str = Field(..., description="Mermaid diagram definition (sequenceDiagram or flowchart TD/LR)")
    mode: str = Field("auto", description="Layout mode: 'auto', 'full', 'compact', or 'narrow'")

def render_diagram_tool(input: RenderDiagramInput) -> str:
    return render_terminal_diagram(input.source)
```

---

## 5. MCP Protocol Specification

The `agent-diagram` MCP server implements JSON-RPC 2.0 over standard I/O:

- **Method**: `tools/list`
  - Returns `render_diagram` tool descriptor.
- **Method**: `tools/call`
  - Arguments:
    - `source` (string, required): Mermaid diagram text.
    - `width` (integer, optional): Desired column width.
    - `mode` (string, optional): `"auto"`, `"full"`, `"compact"`, `"narrow"`.
    - `ascii` (boolean, optional): Set `true` to disable UTF-8 box-drawing characters.
  - Return:
    - Text content block containing the rendered diagram.
