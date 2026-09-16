package ast

// ArrowType specifies the visual styling of a sequence message arrow.
type ArrowType string

const (
	ArrowSolid       ArrowType = "->"   // Solid line without arrowhead
	ArrowSolidArrow  ArrowType = "->>"  // Solid line with arrowhead
	ArrowDotted      ArrowType = "-->"  // Dotted line without arrowhead
	ArrowDottedArrow ArrowType = "-->>" // Dotted line with arrowhead (response)
	ArrowCross       ArrowType = "-x"   // Solid line with cross at end
	ArrowDottedCross ArrowType = "--x"  // Dotted line with cross at end
)

// NotePosition indicates where a note is placed relative to participants.
type NotePosition string

const (
	NoteOver    NotePosition = "over"
	NoteLeftOf  NotePosition = "left of"
	NoteRightOf NotePosition = "right of"
)

// Participant represents an actor or system component in the sequence.
type Participant struct {
	ID     string    `json:"id"`
	Label  string    `json:"label"`
	Order  int       `json:"order"`
	Source SourceRef `json:"source,omitempty"`
}

func (p *Participant) GetSource() SourceRef {
	if p == nil {
		return SourceRef{}
	}
	return p.Source
}

// SequenceEvent represents a message exchange or interaction between participants.
type SequenceEvent struct {
	Number  int       `json:"number,omitempty"`
	From    string    `json:"from"`
	To      string    `json:"to"`
	Message string    `json:"message"`
	Arrow   ArrowType `json:"arrow"`
	Source  SourceRef `json:"source,omitempty"`
}

func (e *SequenceEvent) GetSource() SourceRef {
	if e == nil {
		return SourceRef{}
	}
	return e.Source
}

// SequenceNote represents an explanatory note attached to one or more participants.
type SequenceNote struct {
	Participants []string     `json:"participants"`
	Position     NotePosition `json:"position"`
	Text         string       `json:"text"`
	Source       SourceRef    `json:"source,omitempty"`
}

func (n *SequenceNote) GetSource() SourceRef {
	if n == nil {
		return SourceRef{}
	}
	return n.Source
}

// SequenceDiagram models a parsed Mermaid sequence diagram.
type SequenceDiagram struct {
	DiagramTitle string
	AutoNumber   bool
	Participants []*Participant
	Events       []*SequenceEvent
	Notes        []*SequenceNote
}

func (d *SequenceDiagram) Type() DiagramType {
	return TypeSequence
}

func (d *SequenceDiagram) Title() string {
	return d.DiagramTitle
}

// FindParticipant returns a participant by ID, or nil if not found.
func (d *SequenceDiagram) FindParticipant(id string) *Participant {
	for _, p := range d.Participants {
		if p.ID == id {
			return p
		}
	}
	return nil
}
