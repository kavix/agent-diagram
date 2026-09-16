package ast

// StateType indicates special state categories.
type StateType string

const (
	StateNormal  StateType = "normal"
	StateStart   StateType = "start" // [*] initial
	StateEnd     StateType = "end"   // [*] terminal
	StateChoice  StateType = "choice"
)

// StateNode represents a state machine state.
type StateNode struct {
	ID          string    `json:"id"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	Type        StateType `json:"type"`
	Order       int       `json:"order"`
	Source      SourceRef `json:"source,omitempty"`
}

func (s *StateNode) GetSource() SourceRef {
	if s == nil {
		return SourceRef{}
	}
	return s.Source
}

// StateTransition represents a directed transition between two states.
type StateTransition struct {
	From    string
	To      string
	Trigger string
}

// StateDiagram models a parsed Mermaid state diagram (stateDiagram / stateDiagram-v2).
type StateDiagram struct {
	DiagramTitle string
	States       map[string]*StateNode
	StateOrder   []string
	Transitions  []*StateTransition
}

func (d *StateDiagram) Type() DiagramType {
	return "state"
}

func (d *StateDiagram) Title() string {
	return d.DiagramTitle
}

func (d *StateDiagram) GetState(id string) *StateNode {
	if d.States == nil {
		return nil
	}
	return d.States[id]
}
