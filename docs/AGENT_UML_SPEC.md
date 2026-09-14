# AgentUML Specification: Extending UML for the Agentic Era

> **Status**: Draft v0.1 — for community review  
> **Authors**: agent-diagram project contributors  
> **Based on**: OMG UML 2.5.1 Profile Diagram mechanism (formal/2017-12-05)

---

## 1. Motivation

Standard UML sequence diagrams assume a **deterministic, linear timeline**: a caller makes a request, and the callee either returns a response or throws an error. This model breaks down completely for agentic AI systems, which are:

- **Probabilistic** — the same input can produce different outputs across runs.
- **Non-deterministically looping** — an agent may iterate 0, 1, or N times before completing.
- **Context-bounded** — every interaction grows a finite context window; exceeding it causes catastrophic degradation.
- **Autonomously delegating** — agents spawn sub-agents, tools, and critics without human orchestration.

AgentUML formalises a minimal set of extensions to Mermaid sequence and flowchart syntax to make these properties **visible and diagnosable**.

---

## 2. Stereotype Reference

Stereotypes extend UML's Profile Diagram mechanism. In Mermaid, they are represented as `<<stereotype>>` suffix annotations in participant names.

| Stereotype | Symbol | Description |
| :--- | :--- | :--- |
| `<<orchestrator>>` | 🔀 | Routing node that decomposes tasks and delegates to workers. Does **not** execute business logic. |
| `<<agent>>` | 🤖 | Autonomous reasoning loop capable of tool calling, reflection, and iterative goal pursuit. |
| `<<llm>>` | ⚡ | Raw foundational model inference endpoint (stateless, probabilistic). |
| `<<tool>>` | 🔧 | Deterministic external function or API with bounded latency and defined schema. |
| `<<action>>` | 🎯 | Side-effectful operation (file write, email send, API mutation). |
| `<<memory_working>>` | 📋 | Short-term context window / scratchpad. Subject to token limits. |
| `<<memory_semantic>>` | 🗄️ | Long-term vector database for RAG retrieval (persistent, indexed). |
| `<<guardrail>>` | 🛡️ | Deterministic policy enforcement boundary (topic filter, PII redactor, safety classifier). |

---

## 3. Arrow Notation

AgentUML extends standard Mermaid arrow syntax with probabilistic and asynchronous variants:

| Arrow | Mermaid Syntax | Meaning |
| :--- | :--- | :--- |
| Deterministic call | `->>` | Synchronous, guaranteed execution (tool call, guardrail check) |
| Async fire-and-forget | `-->>` | Non-blocking message; caller does not wait for response |
| Probabilistic branch | `-. [condition] .->` | Branch conditioned on confidence threshold or model output |
| Self-reflection loop | `A->>A:` | Internal critic or re-reasoning pass within the same agent |
| Guardrail gate | `--x` | Message blocked by policy enforcement (cross marks rejection) |

### Probabilistic Branch Convention

When an agent branches non-deterministically, annotate the arrow with the condition:

```
Agent -.-> Tool : [if confidence >= 0.8] execute_tool
Agent -.-> LLM  : [if confidence < 0.8] re-reason
```

---

## 4. Coordination Patterns & Context-Window Risk

### 4.1. Risk Matrix

| Pattern | Context-Window Risk | Mitigation Strategy |
| :--- | :--- | :--- |
| **ReAct (Single-Agent Loop)** | 🟢 LOW | Bounded observation history; truncate after N steps |
| **Generator-Critic (Reflection)** | 🟢 LOW | Fixed revision cap (max 3 iterations recommended) |
| **Hierarchical Task Decomposition** | 🟡 MEDIUM | Context partitioned per level; compress summaries before bubbling up |
| **Orchestrator-Worker** | 🔴 HIGH | Orchestrator accumulates **all** worker outputs in one window; use streaming reduction or parallel-safe summarisation |

### 4.2. Orchestrator-Worker: Highest Risk Analysis

The Orchestrator-Worker pattern presents the **highest risk of context-window overflow** for three reasons:

1. **Linear accumulation**: With N workers, the orchestrator context grows as `O(N × avg_worker_output_tokens)`. For a 128k-token window with 10 workers averaging 10k tokens each, overflow is guaranteed.
2. **Serialised synthesis**: The orchestrator cannot synthesise a final answer until all worker results arrive (or time out), creating a "collect all" anti-pattern that blocks parallelism.
3. **No natural truncation boundary**: Unlike the ReAct loop (which can truncate old observations), the orchestrator cannot discard any worker result without losing task-critical information.

**Recommended Mitigation in Diagrams:**
- Always model a `<<memory_semantic>>` store for worker outputs to externalise results.
- Add an explicit `summarise()` step after each worker returns before appending to the orchestrator's working memory.
- Mark context growth visually with `Note over` annotations showing token budget.

### 4.3. Hierarchical: Medium Risk Analysis

Hierarchical decomposition **partitions** context across levels but introduces a **summary compression** step where managers must reduce their children's outputs before bubbling up. The risk is:

- **Lossy compression**: Critical details may be lost when a supervisor compresses 5 leaf-agent outputs into a single summary.
- **Deep trees amplify latency**: A 3-level hierarchy multiplies the worst-case latency (each level waits for all children).

**Recommended Mitigation:** Use structured JSON schemas for inter-level communication instead of free-text summaries to preserve information fidelity.

---

## 5. Guardrail Boundary Notation

Guardrails must be modelled as **physical gates** — not soft suggestions. In AgentUML:

1. A guardrail node (`<<guardrail>>`) is placed **between** the user/external input and the first agent.
2. The guardrail appears **again** between the agent and any sensitive resource (databases, file systems, external APIs).
3. Rejected messages use the cross-termination arrow (`--x`).

```
User ->> InputFilter : 1. Send prompt
InputFilter --x User : [BLOCKED] Policy violation detected
InputFilter -->> Orchestrator : [PASS] Clean request forwarded
...
Orchestrator ->> DBGuardrail : Read query
DBGuardrail -->> DB : [ENFORCED READ-ONLY] SELECT only
DBGuardrail --x Orchestrator : [BLOCKED] Mutation attempt rejected
```

---

## 6. Context Window Budget Notation

To prevent "diagram drift" — where the diagram fails to model the actual token pressure — use `Note over` blocks to annotate context budgets at critical accumulation points:

```
Note over WorkingMemory: Token budget: ~45k / 128k used
Note over Orchestrator,WorkingMemory: ⚠ Approaching limit — summarise before next worker
```

In `agent-diagram`, the `<<memory_working>>` stereotype automatically triggers a token-budget annotation in the rendered output when a `ContextBudget` field is set on the node.

---

## 7. D2 Syntax Reference (Production Use)

For teams using D2 as their diagram-as-code tool, the following class map implements AgentUML stereotypes natively:

```d2
direction: right

classes: {
  orchestrator: { shape: rectangle; style.fill: "#e3f2fd"; style.stroke: "#1565c0" }
  agent:        { shape: cylinder;  style.fill: "#e8f5e9"; style.stroke: "#2e7d32" }
  llm:          { shape: oval;      style.fill: "#fce4ec"; style.stroke: "#c62828" }
  tool:         { shape: step;      style.fill: "#fff8e1"; style.stroke: "#f57f17" }
  memory:       { shape: document;  style.fill: "#f3e5f5"; style.stroke: "#6a1b9a" }
  guardrail:    { shape: hexagon;   style.fill: "#ffebee"; style.stroke: "#b71c1c" }
}
```

---

## 8. Production Anti-Patterns

### ❌ Anti-Pattern 1: Modelling Agents as Microservices

```
# WRONG — treats agent as synchronous request/response service
User ->> Agent : Call
Agent ->> User : Response (latency: ~50ms)
```

**Why it's wrong**: Multi-agent workflows can take **minutes**, fail non-deterministically, and loop internally. Synchronous sequence diagrams imply guaranteed bounded latency.

**Fix**: Model with `loop` blocks and async arrows. Show internal iteration explicitly.

---

### ❌ Anti-Pattern 2: Ignoring Working Memory

```
# WRONG — loop appears to have no side effects on the agent's state
loop
  Agent ->> Tool : execute()
  Tool -->> Agent : result
end
```

**Why it's wrong**: Each iteration **appends to the context window**. By iteration 10, the prompt may be 10× larger than iteration 1. Failing to model this hides the overflow risk.

**Fix**: Always show the `<<memory_working>>` node receiving appended observations.

---

### ❌ Anti-Pattern 3: Diagram Drift

Diagrams drawn in GUI tools (Visio, Lucidchart) become stale within sprints. Agentic architectures evolve rapidly.

**Fix**: Store all diagrams as `.mmd` files alongside application code in version control. Use `agent-diagram` in CI to validate syntax and render previews on every PR.

---

## 9. Mapping to the C4 Model

AgentUML complements the C4 Model's zoom levels:

| C4 Level | AgentUML Application |
| :--- | :--- |
| **Context (L1)** | Show the `<<orchestrator>>` as the system boundary actor interacting with external users and third-party APIs. |
| **Container (L2)** | Show individual `<<agent>>` containers, their `<<memory_semantic>>` stores, and `<<guardrail>>` boundaries. |
| **Component (L3)** | Decompose each `<<agent>>` into its internal ReAct loop: LLM inference → tool selection → observation → context append. |
| **Code (L4)** | Map agent reasoning steps to specific function calls using Code Intelligence jump-to-source (Roadmap Issue #12). |

---

## 10. Roadmap

| Feature | Status |
| :--- | :--- |
| Native Mermaid `<<stereotype>>` parsing for `agent`, `orchestrator`, `guardrail` nodes | Planned |
| Token-budget progress bar in `<<memory_working>>` node rendering | Planned |
| Probabilistic branch arrow `-. [condition] .->` rendering with dashed style | Planned |
| Context overflow risk badge in diagram header | Planned |
| D2 syntax support via AgentUML class map import | Research |
