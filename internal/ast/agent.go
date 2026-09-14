package ast

// AgentStereotype classifies an agentic component per the AgentUML notation.
// These map to the UML Profile stereotype extension mechanism (<<stereotype>>).
type AgentStereotype string

const (
	// StereotypeOrchestrator is a routing node that determines control flow
	// but does not execute business logic directly. It breaks down tasks and
	// delegates to worker agents.
	StereotypeOrchestrator AgentStereotype = "orchestrator"

	// StereotypeAgent is an autonomous reasoning loop capable of tool calling,
	// self-reflection, and iterative goal pursuit.
	StereotypeAgent AgentStereotype = "agent"

	// StereotypeLLM is the raw foundational model inference endpoint
	// (e.g., Gemini, GPT-4, Claude). It is stateless and probabilistic.
	StereotypeLLM AgentStereotype = "llm"

	// StereotypeTool represents a strict, deterministic external function or
	// API executed by an agent (e.g., web_search, read_file, shell_exec).
	StereotypeTool AgentStereotype = "tool"

	// StereotypeAction is an alias for Tool, emphasizing side-effectful operations
	// such as sending emails, writing files, or making API calls.
	StereotypeAction AgentStereotype = "action"

	// StereotypeMemoryWorking represents the short-term context window or
	// in-flight scratchpad of an agent. Subject to token limits.
	StereotypeMemoryWorking AgentStereotype = "memory_working"

	// StereotypeMemorySemantic represents a long-term vector database used
	// for RAG (Retrieval-Augmented Generation) semantic lookups.
	StereotypeMemorySemantic AgentStereotype = "memory_semantic"

	// StereotypeGuardrail is a deterministic policy enforcement boundary.
	// Examples: content filters, topic routers, rate limiters, PII redactors.
	StereotypeGuardrail AgentStereotype = "guardrail"
)

// AgentFlowType classifies the coordination pattern of an agentic workflow.
type AgentFlowType string

const (
	// FlowReAct is the Single-Agent Loop (Reason + Act). The baseline agentic
	// pattern: observe → reason → act → observe, until the goal is met.
	FlowReAct AgentFlowType = "react"

	// FlowOrchestratorWorker is a central routing LLM that breaks down a
	// complex task and dynamically delegates to specialised worker agents,
	// then synthesises their outputs. Highest risk of context-window overflow
	// when the orchestrator accumulates all worker responses before replying.
	FlowOrchestratorWorker AgentFlowType = "orchestrator_worker"

	// FlowHierarchical is a tree-structured delegation pattern: top-level
	// manager → supervisors → leaf execution agents. Necessary for highly
	// ambiguous tasks to prevent a single context window from overflowing.
	FlowHierarchical AgentFlowType = "hierarchical"

	// FlowGeneratorCritic is the Reflection pattern: one agent generates an
	// output, and a separate critic agent evaluates it against a rubric,
	// forcing a revision loop if the quality threshold is not met.
	FlowGeneratorCritic AgentFlowType = "generator_critic"
)

// AgentArrowType classifies message arrows specific to agentic diagrams.
// Extends the standard SequenceDiagram ArrowType with probabilistic notation.
type AgentArrowType string

const (
	// AgentArrowDeterministic is a solid arrow for guaranteed, synchronous calls
	// (e.g., tool execution, guardrail enforcement).
	AgentArrowDeterministic AgentArrowType = "->"

	// AgentArrowProbabilistic is a dashed arrow annotated with a confidence
	// threshold or condition (e.g., "-. [if confidence < 0.8] .->").
	AgentArrowProbabilistic AgentArrowType = "-.>"

	// AgentArrowAsync is an open arrow representing a fire-and-forget message
	// between agents where the caller does not block on a response.
	AgentArrowAsync AgentArrowType = "-->>"

	// AgentArrowReflection is a self-loop arrow representing an internal
	// critic or reflection pass within the same agent.
	AgentArrowReflection AgentArrowType = "self"
)

// AgentNode represents a component in an agentic workflow diagram.
type AgentNode struct {
	ID         string
	Label      string
	Stereotype AgentStereotype
	// ContextBudget is the approximate token budget for this node's working memory.
	// 0 means unspecified / not modelled.
	ContextBudget int
}

// AgentMessage represents an interaction between two agents or components.
type AgentMessage struct {
	Number     int
	From       string
	To         string
	Message    string
	Arrow      AgentArrowType
	// Condition holds probabilistic branch conditions, e.g. "confidence < 0.8".
	// Empty for deterministic calls.
	Condition  string
	// IsLoop marks this message as part of an internal retry or reflection loop.
	IsLoop     bool
}

// AgentDiagram is the AST for an AgentUML agentic workflow diagram.
// It extends Mermaid sequenceDiagram notation with stereotype annotations,
// probabilistic branches, and context-window budget tracking.
type AgentDiagram struct {
	DiagramTitle string
	FlowType     AgentFlowType
	Nodes        []*AgentNode
	Messages     []*AgentMessage
	// GuardrailBoundaries maps guardrail node IDs to the set of node IDs
	// they protect (downstream of the guardrail).
	GuardrailBoundaries map[string][]string
}

func (d *AgentDiagram) Type() DiagramType {
	return TypeAgent
}

func (d *AgentDiagram) Title() string {
	return d.DiagramTitle
}

// FindNode returns the AgentNode with the given ID, or nil if not found.
func (d *AgentDiagram) FindNode(id string) *AgentNode {
	for _, n := range d.Nodes {
		if n.ID == id {
			return n
		}
	}
	return nil
}

// ContextOverflowRisk returns an estimated risk level ("low", "medium", "high")
// based on the coordination pattern and the number of agents accumulating context.
//
// Analysis derived from the AgentUML specification:
//   - OrchestratorWorker: HIGH — the orchestrator accumulates all worker responses
//     in a single context window before synthesising a final reply. With N workers,
//     the prompt grows linearly and can easily exceed 128k tokens.
//   - Hierarchical: MEDIUM — context is partitioned across levels, but a manager
//     node still aggregates summaries from all its children.
//   - ReAct / GeneratorCritic: LOW — a single agent loops with bounded observations.
func (d *AgentDiagram) ContextOverflowRisk() string {
	switch d.FlowType {
	case FlowOrchestratorWorker:
		return "high"
	case FlowHierarchical:
		return "medium"
	case FlowReAct, FlowGeneratorCritic:
		return "low"
	default:
		return "unknown"
	}
}
