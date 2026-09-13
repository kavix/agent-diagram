package renderer

import (
	"fmt"
	"strings"

	"github.com/kavix/agent-diagram/internal/ast"
	"github.com/kavix/agent-diagram/internal/layout"
)

// RenderSequence renders an AST SequenceDiagram according to RenderOptions.
func RenderSequence(diag *ast.SequenceDiagram, opts RenderOptions) (string, error) {
	lay := layout.ComputeSequenceLayout(diag, opts.Width, opts.Mode)

	switch lay.Mode {
	case layout.ModeFull:
		return renderSequenceFull(diag, lay, opts), nil
	case layout.ModeNarrow:
		return renderSequenceNarrow(diag, lay, opts), nil
	default:
		return renderSequenceCompact(diag, lay, opts), nil
	}
}

func renderSequenceFull(diag *ast.SequenceDiagram, lay *layout.SequenceLayout, opts RenderOptions) string {
	box := UnicodeBox
	if opts.ASCIIOnly {
		box = ASCIIBox
	}
	useColor := !opts.NoColor

	var b strings.Builder

	// Title if present
	if diag.Title() != "" {
		b.WriteString(Style(diag.Title(), ColorBold+ColorCyan, useColor))
		b.WriteString("\n\n")
	}

	totalWidth := lay.TotalWidth
	if totalWidth < lay.TargetWidth {
		totalWidth = lay.TargetWidth
	}

	// 1. Participant Top Boxes
	// Line 1: ┌──────────┐ ┌──────────┐
	// Line 2: │   STS    │ │ Workload │
	// Line 3: └────┬─────┘ └────┬─────┘
	topLine := make([]rune, totalWidth+5)
	midLine := make([]rune, totalWidth+5)
	botLine := make([]rune, totalWidth+5)
	for i := range topLine {
		topLine[i] = ' '
		midLine[i] = ' '
		botLine[i] = ' '
	}

	for _, col := range lay.Columns {
		bw := col.BoxWidth
		sx := col.StartX
		ex := sx + bw - 1

		// Box top
		topLine[sx] = []rune(box.TopLeft)[0]
		for x := sx + 1; x < ex; x++ {
			topLine[x] = []rune(box.Horizontal)[0]
		}
		topLine[ex] = []rune(box.TopRight)[0]

		// Box mid
		midLine[sx] = []rune(box.Vertical)[0]
		midLine[ex] = []rune(box.Vertical)[0]
		// Center label inside box
		innerW := bw - 2
		paddedLabel := layout.PadCenter(col.Participant.Label, innerW)
		labelRunes := []rune(paddedLabel)
		for i, r := range labelRunes {
			if sx+1+i < ex {
				midLine[sx+1+i] = r
			}
		}

		// Box bot with connector to lifeline
		botLine[sx] = []rune(box.BottomLeft)[0]
		for x := sx + 1; x < ex; x++ {
			if x == col.LifelineX {
				botLine[x] = []rune(box.TeeDown)[0]
			} else {
				botLine[x] = []rune(box.Horizontal)[0]
			}
		}
		botLine[ex] = []rune(box.BottomRight)[0]
	}

	b.WriteString(Style(strings.TrimRight(string(topLine), " "), ColorCyan, useColor) + "\n")
	b.WriteString(Style(strings.TrimRight(string(midLine), " "), ColorBold, useColor) + "\n")
	b.WriteString(Style(strings.TrimRight(string(botLine), " "), ColorCyan, useColor) + "\n")

	// Helper to create blank lifeline row
	makeLifelineRow := func() []rune {
		row := make([]rune, totalWidth+5)
		for i := range row {
			row[i] = ' '
		}
		for _, col := range lay.Columns {
			row[col.LifelineX] = []rune(box.Vertical)[0]
		}
		return row
	}

	// 2. Events & Lifelines
	for eventIdx, ev := range diag.Events {
		fromCol := lay.ColByPID[ev.From]
		toCol := lay.ColByPID[ev.To]
		if fromCol == nil || toCol == nil {
			continue
		}

		// Lifeline spacing row before arrow
		b.WriteString(Style(strings.TrimRight(string(makeLifelineRow()), " "), ColorGray, useColor) + "\n")

		// Message label row (placed above the arrow)
		msgRow := makeLifelineRow()
		msgText := ev.Message
		if ev.Number > 0 {
			msgText = fmt.Sprintf("%d. %s", ev.Number, ev.Message)
		} else if diag.AutoNumber {
			msgText = fmt.Sprintf("%d. %s", eventIdx+1, ev.Message)
		}

		minX := fromCol.LifelineX
		maxX := toCol.LifelineX
		if minX > maxX {
			minX, maxX = toCol.LifelineX, fromCol.LifelineX
		}
		availableSpan := maxX - minX - 2
		if availableSpan < 4 {
			availableSpan = 4
		}
		truncatedMsg := layout.Truncate(msgText, availableSpan)

		// Place message centered or aligned in span
		msgStart := minX + 2
		msgRunes := []rune(truncatedMsg)
		for i, r := range msgRunes {
			if msgStart+i < maxX {
				msgRow[msgStart+i] = r
			}
		}
		b.WriteString(Style(strings.TrimRight(string(msgRow), " "), ColorYellow, useColor) + "\n")

		// Arrow row
		arrowRow := makeLifelineRow()
		horizChar := []rune(box.Horizontal)[0]
		if ev.Arrow == ast.ArrowDotted || ev.Arrow == ast.ArrowDottedArrow || ev.Arrow == ast.ArrowDottedCross {
			horizChar = []rune(box.DottedHoriz)[0]
		}

		if fromCol.LifelineX < toCol.LifelineX {
			// Left to Right: ├───────►
			arrowRow[fromCol.LifelineX] = []rune(box.TeeRight)[0]
			for x := fromCol.LifelineX + 1; x < toCol.LifelineX; x++ {
				// Check if another lifeline is crossed in between
				if arrowRow[x] == []rune(box.Vertical)[0] {
					arrowRow[x] = []rune(box.Cross)[0]
				} else {
					arrowRow[x] = horizChar
				}
			}
			arrowRow[toCol.LifelineX] = []rune(box.ArrowRight)[0]
		} else if fromCol.LifelineX > toCol.LifelineX {
			// Right to Left: ◄───────┤
			arrowRow[fromCol.LifelineX] = []rune(box.TeeLeft)[0]
			for x := toCol.LifelineX + 1; x < fromCol.LifelineX; x++ {
				if arrowRow[x] == []rune(box.Vertical)[0] {
					arrowRow[x] = []rune(box.Cross)[0]
				} else {
					arrowRow[x] = horizChar
				}
			}
			arrowRow[toCol.LifelineX] = []rune(box.ArrowLeft)[0]
		} else {
			// Self message (from == to): ├──┐ then └──►
			arrowRow[fromCol.LifelineX] = []rune(box.TeeRight)[0]
			if fromCol.LifelineX+3 < len(arrowRow) {
				arrowRow[fromCol.LifelineX+1] = horizChar
				arrowRow[fromCol.LifelineX+2] = horizChar
				arrowRow[fromCol.LifelineX+3] = []rune(box.ArrowRight)[0]
			}
		}

		b.WriteString(Style(strings.TrimRight(string(arrowRow), " "), ColorGreen, useColor) + "\n")
	}

	// 3. Notes (if any)
	for _, note := range diag.Notes {
		b.WriteString(Style(strings.TrimRight(string(makeLifelineRow()), " "), ColorGray, useColor) + "\n")
		noteRow := makeLifelineRow()
		noteStr := fmt.Sprintf(" Note: %s ", note.Text)
		noteRunes := []rune(noteStr)
		startX := 2
		for i, r := range noteRunes {
			if startX+i < len(noteRow) {
				noteRow[startX+i] = r
			}
		}
		b.WriteString(Style(strings.TrimRight(string(noteRow), " "), ColorMagenta, useColor) + "\n")
	}

	// Lifeline spacing row before bottom boxes
	b.WriteString(Style(strings.TrimRight(string(makeLifelineRow()), " "), ColorGray, useColor) + "\n")

	// 4. Participant Bottom Boxes
	botTopLine := make([]rune, totalWidth+5)
	botMidLine := make([]rune, totalWidth+5)
	botEndLine := make([]rune, totalWidth+5)
	for i := range botTopLine {
		botTopLine[i] = ' '
		botMidLine[i] = ' '
		botEndLine[i] = ' '
	}

	for _, col := range lay.Columns {
		bw := col.BoxWidth
		sx := col.StartX
		ex := sx + bw - 1

		// Box top with connector to lifeline
		botTopLine[sx] = []rune(box.TopLeft)[0]
		for x := sx + 1; x < ex; x++ {
			if x == col.LifelineX {
				botTopLine[x] = []rune(box.TeeUp)[0]
			} else {
				botTopLine[x] = []rune(box.Horizontal)[0]
			}
		}
		botTopLine[ex] = []rune(box.TopRight)[0]

		// Box mid
		botMidLine[sx] = []rune(box.Vertical)[0]
		botMidLine[ex] = []rune(box.Vertical)[0]
		innerW := bw - 2
		paddedLabel := layout.PadCenter(col.Participant.Label, innerW)
		labelRunes := []rune(paddedLabel)
		for i, r := range labelRunes {
			if sx+1+i < ex {
				botMidLine[sx+1+i] = r
			}
		}

		// Box bottom
		botEndLine[sx] = []rune(box.BottomLeft)[0]
		for x := sx + 1; x < ex; x++ {
			botEndLine[x] = []rune(box.Horizontal)[0]
		}
		botEndLine[ex] = []rune(box.BottomRight)[0]
	}

	b.WriteString(Style(strings.TrimRight(string(botTopLine), " "), ColorCyan, useColor) + "\n")
	b.WriteString(Style(strings.TrimRight(string(botMidLine), " "), ColorBold, useColor) + "\n")
	b.WriteString(Style(strings.TrimRight(string(botEndLine), " "), ColorCyan, useColor) + "\n")

	return b.String()
}

func renderSequenceCompact(diag *ast.SequenceDiagram, lay *layout.SequenceLayout, opts RenderOptions) string {
	box := UnicodeBox
	if opts.ASCIIOnly {
		box = ASCIIBox
	}
	useColor := !opts.NoColor

	var b strings.Builder
	title := "Sequence Diagram (Compact)"
	if diag.Title() != "" {
		title = fmt.Sprintf("%s (Compact)", diag.Title())
	}
	b.WriteString(Style(title, ColorBold+ColorCyan, useColor) + "\n")
	b.WriteString(strings.Repeat(box.Horizontal, 60) + "\n")

	nameFor := func(id string) string {
		p := diag.FindParticipant(id)
		if p != nil && p.Label != "" {
			return p.Label
		}
		return id
	}

	for idx, ev := range diag.Events {
		num := ev.Number
		if num <= 0 {
			num = idx + 1
		}

		arrowSymbol := box.Horizontal + box.Horizontal + box.ArrowRight
		if ev.Arrow == ast.ArrowDotted || ev.Arrow == ast.ArrowDottedArrow {
			arrowSymbol = box.DottedHoriz + box.DottedHoriz + box.ArrowRight
		}

		fromName := Style(nameFor(ev.From), ColorBold, useColor)
		toName := Style(nameFor(ev.To), ColorBold, useColor)
		arrowFormatted := Style(arrowSymbol, ColorGreen, useColor)

		b.WriteString(fmt.Sprintf(" %d. %s %s %s\n", num, fromName, arrowFormatted, toName))
		b.WriteString(fmt.Sprintf("    %s\n", Style(ev.Message, ColorYellow, useColor)))
	}

	for _, note := range diag.Notes {
		b.WriteString(fmt.Sprintf("\n %s Note [%s]: %s\n", Style("◈", ColorMagenta, useColor), strings.Join(note.Participants, ", "), note.Text))
	}

	return b.String()
}

func renderSequenceNarrow(diag *ast.SequenceDiagram, lay *layout.SequenceLayout, opts RenderOptions) string {
	box := UnicodeBox
	if opts.ASCIIOnly {
		box = ASCIIBox
	}
	useColor := !opts.NoColor

	var b strings.Builder
	b.WriteString(Style("Sequence Diagram", ColorBold, useColor) + "\n")
	b.WriteString(strings.Repeat(box.Horizontal, 35) + "\n")

	for idx, ev := range diag.Events {
		num := ev.Number
		if num <= 0 {
			num = idx + 1
		}
		arrowSymbol := box.Horizontal + box.ArrowRight
		b.WriteString(fmt.Sprintf("[%d] %s %s %s\n", num, ev.From, arrowSymbol, ev.To))
		b.WriteString(fmt.Sprintf("    %s\n", ev.Message))
	}

	return b.String()
}
