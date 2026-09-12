package layout

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// StringWidth returns the visual cell width of a string in a terminal.
func StringWidth(s string) int {
	return runewidth.StringWidth(s)
}

// Truncate ensures a string occupies at most maxWidth terminal cells,
// appending ellipsis "…" if truncated.
func Truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if StringWidth(s) <= maxWidth {
		return s
	}
	if maxWidth == 1 {
		return "…"
	}

	var sb strings.Builder
	currWidth := 0
	target := maxWidth - 1 // Leave room for ellipsis

	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if currWidth+rw > target {
			break
		}
		sb.WriteRune(r)
		currWidth += rw
	}
	sb.WriteString("…")
	return sb.String()
}

// PadCenter centers string s within width cells with spaces.
func PadCenter(s string, width int) string {
	sw := StringWidth(s)
	if sw >= width {
		return s
	}
	totalPad := width - sw
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad
	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

// PadRight pads string s with trailing spaces up to width cells.
func PadRight(s string, width int) string {
	sw := StringWidth(s)
	if sw >= width {
		return s
	}
	return s + strings.Repeat(" ", width-sw)
}
