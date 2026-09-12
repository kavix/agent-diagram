package layout

import (
	"github.com/kavix/agent-diagram/internal/ast"
)

// ParticipantCol contains positioning data for a participant in full sequence layout.
type ParticipantCol struct {
	Participant *ast.Participant
	BoxWidth    int // Width of box: label + padding + borders
	StartX      int // X start of header box
	LifelineX   int // X coordinate of vertical lifeline
}

// SequenceLayout holds computed layout geometry for a sequence diagram.
type SequenceLayout struct {
	Mode         RenderMode
	Columns      []*ParticipantCol
	ColByPID     map[string]*ParticipantCol
	TotalWidth   int
	TargetWidth  int
	Diagram      *ast.SequenceDiagram
	MessageWidth int // Available width for message annotations between lifelines
}

// ComputeSequenceLayout calculates geometry and selects the optimal rendering mode.
func ComputeSequenceLayout(diag *ast.SequenceDiagram, targetWidth int, requestedMode RenderMode) *SequenceLayout {
	if targetWidth <= 0 {
		targetWidth = DetectTerminalWidth()
	}

	numParts := len(diag.Participants)
	if numParts == 0 {
		return &SequenceLayout{
			Mode:        ModeCompact,
			TargetWidth: targetWidth,
			Diagram:     diag,
		}
	}

	// Calculate minimal box widths
	boxWidths := make([]int, numParts)
	minTotalWidth := 0
	const minGap = 4

	for i, p := range diag.Participants {
		labelLen := StringWidth(p.Label)
		// Box format: [ Label ] => labelLen + 4 cells (space + border + space)
		bw := labelLen + 4
		if bw < 8 {
			bw = 8
		}
		boxWidths[i] = bw
		minTotalWidth += bw
	}
	minTotalWidth += (numParts - 1) * minGap

	// Decide mode
	mode := requestedMode
	if mode == "" || mode == ModeAuto {
		if targetWidth < 50 {
			mode = ModeNarrow
		} else if minTotalWidth > targetWidth {
			mode = ModeCompact
		} else {
			mode = ModeFull
		}
	}

	layout := &SequenceLayout{
		Mode:        mode,
		TargetWidth: targetWidth,
		Diagram:     diag,
		Columns:     make([]*ParticipantCol, numParts),
		ColByPID:    make(map[string]*ParticipantCol),
	}

	if mode != ModeFull {
		return layout
	}

	// Distribute extra spacing evenly across gaps
	extraSpace := targetWidth - minTotalWidth
	if extraSpace < 0 {
		extraSpace = 0
	}
	extraPerGap := 0
	if numParts > 1 {
		extraPerGap = extraSpace / (numParts - 1)
		if extraPerGap > 12 {
			extraPerGap = 12 // Cap max gap to prevent overly sparse diagrams
		}
	}

	currentX := 1 // 1-cell left margin
	for i, p := range diag.Participants {
		bw := boxWidths[i]
		col := &ParticipantCol{
			Participant: p,
			BoxWidth:    bw,
			StartX:      currentX,
			LifelineX:   currentX + (bw / 2),
		}
		layout.Columns[i] = col
		layout.ColByPID[p.ID] = col

		currentX += bw + minGap + extraPerGap
	}

	layout.TotalWidth = currentX
	return layout
}
