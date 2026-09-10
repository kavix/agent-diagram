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
	ID    string
	Label string
	Order int
}

// SequenceEvent represents a message exchange or interaction between participants.
type SequenceEvent struct {
	Number  int
	From    string
	To      string
	Message string
	Arrow   ArrowType
}

// SequenceNote represents an explanatory note attached to one or more participants.
type SequenceNote struct {
	Participants []string
	Position     NotePosition
	Text         string
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
