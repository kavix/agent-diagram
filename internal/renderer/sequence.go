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
	box := opts.BoxStyle()
	useColor := !opts.NoColor

	var b strings.Builder

	if diag.Title() != "" {
		b.WriteString(Style(diag.Title(), ColorBold+ColorCyan, useColor))
		b.WriteString("\n\n")
	}

	totalWidth := lay.TotalWidth
	if totalWidth < lay.TargetWidth {
		totalWidth = lay.TargetWidth
	}

	// ── 1. Participant Top Boxes ─────────────────────────────────────────────
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

		topLine[sx] = []rune(box.TopLeft)[0]
		for x := sx + 1; x < ex; x++ {
			topLine[x] = []rune(box.Horizontal)[0]
		}
		topLine[ex] = []rune(box.TopRight)[0]

		midLine[sx] = []rune(box.Vertical)[0]
		midLine[ex] = []rune(box.Vertical)[0]
		innerW := bw - 2
		paddedLabel := layout.PadCenter(col.Participant.Label, innerW)
		for i, r := range []rune(paddedLabel) {
			if sx+1+i < ex {
				midLine[sx+1+i] = r
			}
		}

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

	// Helper: blank lifeline row with vertical bars at each lifeline X
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

	writeLifelineRow := func() {
		b.WriteString(Style(strings.TrimRight(string(makeLifelineRow()), " "), ColorGray, useColor) + "\n")
	}

	// ── 2. Events ────────────────────────────────────────────────────────────
	for eventIdx, ev := range diag.Events {
		fromCol := lay.ColByPID[ev.From]
		toCol := lay.ColByPID[ev.To]
		if fromCol == nil || toCol == nil {
			continue
		}

		writeLifelineRow()

		// Message label row
		msgRow := makeLifelineRow()
		msgText := ev.Message
		if ev.Number > 0 {
			msgText = fmt.Sprintf("%d. %s", ev.Number, ev.Message)
		} else if diag.AutoNumber {
			msgText = fmt.Sprintf("%d. %s", eventIdx+1, ev.Message)
		}

		isSelf := fromCol.LifelineX == toCol.LifelineX
		minX := fromCol.LifelineX
		maxX := toCol.LifelineX
		if minX > maxX {
			minX, maxX = maxX, minX
		}
		horizChar := []rune(box.Horizontal)[0]
		if ev.Arrow == ast.ArrowDotted || ev.Arrow == ast.ArrowDottedArrow || ev.Arrow == ast.ArrowDottedCross {
			horizChar = []rune(box.DottedHoriz)[0]
		}

		if isSelf {
			// ── Self-Loop: 3-row prominent box with ↺ icon & arrow returning to lifeline ──
			selfEndX := fromCol.LifelineX + 6
			if selfEndX >= totalWidth {
				selfEndX = totalWidth - 1
			}

			loopIcon := "↺"
			if opts.ASCIIOnly {
				loopIcon = "@"
			}

			// Row 1 (top arm): ├───┐ (or ╠═══╗ in TOON)
			topArm := makeLifelineRow()
			topArm[fromCol.LifelineX] = []rune(box.TeeRight)[0]
			for x := fromCol.LifelineX + 1; x < selfEndX; x++ {
				topArm[x] = horizChar
			}
			if selfEndX < len(topArm) {
				topArm[selfEndX] = []rune(box.TopRight)[0]
			}
			b.WriteString(Style(strings.TrimRight(string(topArm), " "), ColorGreen, useColor) + "\n")

			// Row 2 (loop body with icon and message): │ ↺ │  1. message
			midRow := makeLifelineRow()
			midRow[fromCol.LifelineX] = []rune(box.Vertical)[0]
			midRow[selfEndX] = []rune(box.Vertical)[0]
			// Center loop icon inside the loop
			iconX := fromCol.LifelineX + 1 + (selfEndX-fromCol.LifelineX-1)/2
			if iconX < selfEndX {
				midRow[iconX] = []rune(loopIcon)[0]
			}
			// Place message label to the right of the loop box
			msgStart := selfEndX + 2
			for i, r := range []rune(msgText) {
				if msgStart+i < len(midRow) {
					midRow[msgStart+i] = r
				}
			}
			b.WriteString(Style(strings.TrimRight(string(midRow), " "), ColorYellow, useColor) + "\n")

			// Row 3 (return arm with arrow returning to lifeline): ◄───┘ (or ◀═══╝ in TOON)
			botArm := makeLifelineRow()
			botArm[fromCol.LifelineX] = []rune(box.ArrowLeft)[0]
			for x := fromCol.LifelineX + 1; x < selfEndX; x++ {
				botArm[x] = horizChar
			}
			if selfEndX < len(botArm) {
				botArm[selfEndX] = []rune(box.BottomRight)[0]
			}
			b.WriteString(Style(strings.TrimRight(string(botArm), " "), ColorGreen, useColor) + "\n")

		} else {
			// Normal message between two different lifelines
			availableSpan := maxX - minX - 2
			if availableSpan < 4 {
				availableSpan = 4
			}

			truncatedMsg := layout.Truncate(msgText, availableSpan)
			msgStart := minX + 2
			for i, r := range []rune(truncatedMsg) {
				if msgStart+i < len(msgRow) {
					msgRow[msgStart+i] = r
				}
			}
			b.WriteString(Style(strings.TrimRight(string(msgRow), " "), ColorYellow, useColor) + "\n")

			// Arrow row
			arrowRow := makeLifelineRow()
			if fromCol.LifelineX < toCol.LifelineX {
				// Left → Right: ├─────►
				arrowRow[fromCol.LifelineX] = []rune(box.TeeRight)[0]
				for x := fromCol.LifelineX + 1; x < toCol.LifelineX; x++ {
					if arrowRow[x] == []rune(box.Vertical)[0] {
						arrowRow[x] = []rune(box.Cross)[0]
					} else {
						arrowRow[x] = horizChar
					}
				}
				arrowRow[toCol.LifelineX] = []rune(box.ArrowRight)[0]
				b.WriteString(Style(strings.TrimRight(string(arrowRow), " "), ColorGreen, useColor) + "\n")

			} else {
				// Right → Left: ◄─────┤
				arrowRow[fromCol.LifelineX] = []rune(box.TeeLeft)[0]
				for x := toCol.LifelineX + 1; x < fromCol.LifelineX; x++ {
					if arrowRow[x] == []rune(box.Vertical)[0] {
						arrowRow[x] = []rune(box.Cross)[0]
					} else {
						arrowRow[x] = horizChar
					}
				}
				arrowRow[toCol.LifelineX] = []rune(box.ArrowLeft)[0]
				b.WriteString(Style(strings.TrimRight(string(arrowRow), " "), ColorGreen, useColor) + "\n")
			}
		}
	}

	// ── 3. Notes (rendered after all events, in declaration order) ───────────
	for _, note := range diag.Notes {
		renderNoteBox(&b, note, lay, box, makeLifelineRow, totalWidth, useColor)
	}

	writeLifelineRow()

	// ── 4. Participant Bottom Boxes ──────────────────────────────────────────
	btTop := make([]rune, totalWidth+5)
	btMid := make([]rune, totalWidth+5)
	btBot := make([]rune, totalWidth+5)
	for i := range btTop {
		btTop[i] = ' '
		btMid[i] = ' '
		btBot[i] = ' '
	}
	for _, col := range lay.Columns {
		bw := col.BoxWidth
		sx := col.StartX
		ex := sx + bw - 1

		btTop[sx] = []rune(box.TopLeft)[0]
		for x := sx + 1; x < ex; x++ {
			if x == col.LifelineX {
				btTop[x] = []rune(box.TeeUp)[0]
			} else {
				btTop[x] = []rune(box.Horizontal)[0]
			}
		}
		btTop[ex] = []rune(box.TopRight)[0]

		btMid[sx] = []rune(box.Vertical)[0]
		btMid[ex] = []rune(box.Vertical)[0]
		innerW := bw - 2
		paddedLabel := layout.PadCenter(col.Participant.Label, innerW)
		for i, r := range []rune(paddedLabel) {
			if sx+1+i < ex {
				btMid[sx+1+i] = r
			}
		}

		btBot[sx] = []rune(box.BottomLeft)[0]
		for x := sx + 1; x < ex; x++ {
			btBot[x] = []rune(box.Horizontal)[0]
		}
		btBot[ex] = []rune(box.BottomRight)[0]
	}
	b.WriteString(Style(strings.TrimRight(string(btTop), " "), ColorCyan, useColor) + "\n")
	b.WriteString(Style(strings.TrimRight(string(btMid), " "), ColorBold, useColor) + "\n")
	b.WriteString(Style(strings.TrimRight(string(btBot), " "), ColorCyan, useColor) + "\n")

	return b.String()
}

// renderNoteBox renders a SequenceNote as a properly-aligned bordered box.
// It eliminates background lifeline bleeding, supports multi-line notes (<br/>, \n),
// and guarantees pixel-perfect visual width alignment even with wide emojis.
func renderNoteBox(
	b *strings.Builder,
	note *ast.SequenceNote,
	lay *layout.SequenceLayout,
	box BoxChars,
	makeLifelineRow func() []rune,
	totalWidth int,
	useColor bool,
) {
	const minNoteWidth = 24

	// Find leftmost StartX and rightmost EndX of involved participants
	leftX := totalWidth
	rightX := 0
	for _, pid := range note.Participants {
		col := lay.ColByPID[pid]
		if col == nil {
			continue
		}
		if col.StartX < leftX {
			leftX = col.StartX
		}
		end := col.StartX + col.BoxWidth - 1
		if end > rightX {
			rightX = end
		}
	}

	// Fallback if no participants matched
	if leftX > rightX || rightX == 0 {
		leftX = 2
		rightX = totalWidth - 4
	}
	if rightX-leftX < minNoteWidth {
		rightX = leftX + minNoteWidth
	}
	if rightX >= totalWidth+3 {
		rightX = totalWidth + 2
	}

	innerWidth := rightX - leftX - 1 // space between the two border chars

	// Helper to build background lifelines for columns left of leftX
	buildPrefix := func() string {
		row := make([]rune, leftX)
		for i := range row {
			row[i] = ' '
		}
		for _, col := range lay.Columns {
			if col.LifelineX < leftX {
				row[col.LifelineX] = []rune(box.Vertical)[0]
			}
		}
		return string(row)
	}

	// Helper to build background lifelines for columns right of rightX
	buildSuffix := func() string {
		start := rightX + 1
		if start >= totalWidth+4 {
			return ""
		}
		length := (totalWidth + 4) - start
		row := make([]rune, length)
		for i := range row {
			row[i] = ' '
		}
		for _, col := range lay.Columns {
			if col.LifelineX > rightX && col.LifelineX-start < length {
				row[col.LifelineX-start] = []rune(box.Vertical)[0]
			}
		}
		return strings.TrimRight(string(row), " ")
	}

	prefix := buildPrefix()
	suffix := buildSuffix()

	// Parse note text into lines (supports Mermaid <br/> and \n)
	rawText := strings.ReplaceAll(note.Text, "<br/>", "\n")
	rawText = strings.ReplaceAll(rawText, "<br>", "\n")
	textLines := strings.Split(rawText, "\n")

	// Spacer row above note
	spacer := makeLifelineRow()
	b.WriteString(Style(strings.TrimRight(string(spacer), " "), ColorGray, useColor) + "\n")

	// Top border: ╭──────────────────╮ (or ┏━━━━━━━━━━┓ in TOON)
	b.WriteString(Style(prefix, ColorGray, useColor))
	b.WriteString(Style(box.NoteTopLeft+strings.Repeat(box.NoteHorizontal, innerWidth)+box.NoteTopRight, ColorMagenta, useColor))
	if suffix != "" {
		b.WriteString(Style(suffix, ColorGray, useColor))
	}
	b.WriteString("\n")

	// Text rows: │   padded text    │ (or ┃   text   ┃ in TOON)
	for _, rawLine := range textLines {
		line := strings.TrimSpace(rawLine)
		if layout.StringWidth(line) > innerWidth-2 {
			line = layout.Truncate(line, innerWidth-2)
		}
		padded := layout.PadCenter(line, innerWidth)

		b.WriteString(Style(prefix, ColorGray, useColor))
		b.WriteString(Style(box.NoteVertical, ColorMagenta, useColor))
		b.WriteString(Style(padded, ColorBold+ColorMagenta, useColor))
		b.WriteString(Style(box.NoteVertical, ColorMagenta, useColor))
		if suffix != "" {
			b.WriteString(Style(suffix, ColorGray, useColor))
		}
		b.WriteString("\n")
	}

	// Bottom border: ╰──────────────────╯ (or ┗━━━━━━━━━━┛ in TOON)
	b.WriteString(Style(prefix, ColorGray, useColor))
	b.WriteString(Style(box.NoteBottomLeft+strings.Repeat(box.NoteHorizontal, innerWidth)+box.NoteBottomRight, ColorMagenta, useColor))
	if suffix != "" {
		b.WriteString(Style(suffix, ColorGray, useColor))
	}
	b.WriteString("\n")
}

func renderSequenceCompact(diag *ast.SequenceDiagram, lay *layout.SequenceLayout, opts RenderOptions) string {
	box := opts.BoxStyle()
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
		arrowFmt := Style(arrowSymbol, ColorGreen, useColor)

		if ev.From == ev.To {
			b.WriteString(fmt.Sprintf(" %d. %s ↺ (self)\n", num, fromName))
			b.WriteString(fmt.Sprintf("    %s\n", Style(ev.Message, ColorYellow, useColor)))
		} else {
			b.WriteString(fmt.Sprintf(" %d. %s %s %s\n", num, fromName, arrowFmt, toName))
			b.WriteString(fmt.Sprintf("    %s\n", Style(ev.Message, ColorYellow, useColor)))
		}
	}

	for _, note := range diag.Notes {
		participantStr := strings.Join(note.Participants, ", ")
		b.WriteString(fmt.Sprintf("\n %s Note [%s]:\n    %s\n",
			Style("◈", ColorMagenta, useColor),
			participantStr,
			Style(note.Text, ColorMagenta, useColor),
		))
	}

	return b.String()
}

func renderSequenceNarrow(diag *ast.SequenceDiagram, lay *layout.SequenceLayout, opts RenderOptions) string {
	box := opts.BoxStyle()
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
		if ev.From == ev.To {
			b.WriteString(fmt.Sprintf("[%d] %s ↺\n", num, ev.From))
		} else {
			b.WriteString(fmt.Sprintf("[%d] %s %s %s\n", num, ev.From, arrowSymbol, ev.To))
		}
		b.WriteString(fmt.Sprintf("    %s\n", ev.Message))
	}

	return b.String()
}
