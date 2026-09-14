package mermaid

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kavix/agent-diagram/internal/ast"
)

var (
	// participant Alice as Alice User / actor Bob
	participantRegex = regexp.MustCompile(`^(?:participant|actor)\s+([a-zA-Z0-9_]+)(?:\s+as\s+(.+))?$`)

	// A->>B: message or A-->>B: message
	messageRegex = regexp.MustCompile(`^([a-zA-Z0-9_]+)\s*(-->>|->>|-->|->|--x|-x)\s*([a-zA-Z0-9_]+)\s*:\s*(.*)$`)

	// Note over A,B: message
	noteRegex = regexp.MustCompile(`^Note\s+(over|left\s+of|right\s+of)\s+([a-zA-Z0-9_,\s]+)\s*:\s*(.*)$`)

	titleRegex = regexp.MustCompile(`^title\s+(.+)$`)
)

// ParseSequence parses a sequenceDiagram block into an ast.SequenceDiagram.
func ParseSequence(lines []string) (*ast.SequenceDiagram, error) {
	diag := &ast.SequenceDiagram{
		Participants: make([]*ast.Participant, 0),
		Events:       make([]*ast.SequenceEvent, 0),
		Notes:        make([]*ast.SequenceNote, 0),
	}

	participantMap := make(map[string]*ast.Participant)

	ensureParticipant := func(id string, label string) *ast.Participant {
		if p, exists := participantMap[id]; exists {
			if label != "" && p.Label == id {
				p.Label = label
			}
			return p
		}
		if label == "" {
			label = id
		}
		p := &ast.Participant{
			ID:    id,
			Label: label,
			Order: len(diag.Participants),
		}
		diag.Participants = append(diag.Participants, p)
		participantMap[id] = p
		return p
	}

	eventCounter := 1

	for lineIdx, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "%%") {
			continue
		}

		// Header
		if strings.EqualFold(line, "sequencediagram") {
			continue
		}

		// Title
		if m := titleRegex.FindStringSubmatch(line); len(m) > 1 {
			diag.DiagramTitle = strings.TrimSpace(m[1])
			continue
		}

		// Autonumber
		if strings.EqualFold(line, "autonumber") {
			diag.AutoNumber = true
			continue
		}

		// Participant
		if m := participantRegex.FindStringSubmatch(line); len(m) > 1 {
			id := m[1]
			label := id
			if len(m) > 2 && m[2] != "" {
				label = strings.TrimSpace(m[2])
			}
			ensureParticipant(id, label)
			continue
		}

		// Message
		if m := messageRegex.FindStringSubmatch(line); len(m) == 5 {
			from := m[1]
			arrowStr := m[2]
			to := m[3]
			msg := strings.TrimSpace(m[4])

			ensureParticipant(from, "")
			ensureParticipant(to, "")

			eventNum := 0
			if diag.AutoNumber {
				eventNum = eventCounter
				eventCounter++
			}

			event := &ast.SequenceEvent{
				Number:  eventNum,
				From:    from,
				To:      to,
				Message: msg,
				Arrow:   ast.ArrowType(arrowStr),
			}
			diag.Events = append(diag.Events, event)
			continue
		}

		// Note
		if m := noteRegex.FindStringSubmatch(line); len(m) == 4 {
			posStr := strings.ToLower(strings.TrimSpace(m[1]))
			partStr := m[2]
			text := strings.TrimSpace(m[3])

			var pos ast.NotePosition
			switch posStr {
			case "over":
				pos = ast.NoteOver
			case "left of":
				pos = ast.NoteLeftOf
			case "right of":
				pos = ast.NoteRightOf
			default:
				pos = ast.NoteOver
			}

			parts := strings.Split(partStr, ",")
			var noteParticipants []string
			for _, p := range parts {
				pClean := strings.TrimSpace(p)
				if pClean != "" {
					ensureParticipant(pClean, "")
					noteParticipants = append(noteParticipants, pClean)
				}
			}

			diag.Notes = append(diag.Notes, &ast.SequenceNote{
				Participants: noteParticipants,
				Position:     pos,
				Text:         text,
			})
			continue
		}

		// Structural grouping keywords — loop, par, alt, else, end, opt, critical, break, rect.
		// These are valid Mermaid syntax but grouping boxes are not yet rendered.
		// We skip them gracefully rather than aborting the parse.
		lower := strings.ToLower(line)
		if isStructuralKeyword(lower) {
			continue
		}

		// If unknown line in sequence diagram, return error
		return nil, fmt.Errorf("line %d: unrecognized sequence diagram statement: %q", lineIdx+1, line)
	}

	return diag, nil
}

// isStructuralKeyword returns true for Mermaid grouping/control-flow keywords
// that are syntactically valid but not yet rendered by agent-diagram.
func isStructuralKeyword(lower string) bool {
	keywords := []string{
		"loop", "par", "alt", "else", "end", "opt",
		"critical", "break", "rect", "and", "activate", "deactivate",
	}
	for _, kw := range keywords {
		if strings.HasPrefix(lower, kw+" ") || lower == kw {
			return true
		}
	}
	return false
}
