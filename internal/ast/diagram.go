package ast

// DiagramType represents the kind of diagram being visualized.
type DiagramType string

const (
	TypeSequence  DiagramType = "sequence"
	TypeFlowchart DiagramType = "flowchart"
	TypeState     DiagramType = "state"
	TypeAgent     DiagramType = "agent" // AgentUML — agentic workflow diagrams
)

// Diagram is the base interface that all diagram ASTs implement.
type Diagram interface {
	Type() DiagramType
	Title() string
}
