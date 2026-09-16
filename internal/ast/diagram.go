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

// SourceRef represents a reference to a source code file and line.
type SourceRef struct {
	File string `json:"file,omitempty"`
	Line int    `json:"line,omitempty"`
}

// IsValid checks if the SourceRef contains valid file and line values.
func (s SourceRef) IsValid() bool {
	return s.File != "" || s.Line > 0
}

// SourceProvider is an optional interface implemented by AST elements that carry source code references.
type SourceProvider interface {
	GetSource() SourceRef
}
