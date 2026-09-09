package renderer

import (
	"fmt"
	"strings"

	"github.com/kavix/agent-diagram/internal/ast"
)

// RenderState renders an AST StateDiagram into formatted terminal visualization.
func RenderState(diag *ast.StateDiagram, opts RenderOptions) (string, error) {
	box := UnicodeBox
	if opts.ASCIIOnly {
		box = ASCIIBox
	}
	useColor := !opts.NoColor

	var b strings.Builder
	title := "State Machine"
	if diag.Title() != "" {
		title = diag.Title()
	}
	b.WriteString(Style(title, ColorBold+ColorCyan, useColor) + "\n")
	b.WriteString(strings.Repeat(box.Horizontal, 60) + "\n\n")

	formatNode := func(id string, isEnd bool) string {
		if id == "[*]" {
			if isEnd {
				return Style("(◉)", ColorMagenta, useColor)
			}
			return Style("(●)", ColorMagenta, useColor)
		}
		s := diag.States[id]
		label := id
		if s != nil && s.Label != "" {
			label = s.Label
		}
		return Style(fmt.Sprintf("[%s]", label), ColorBold, useColor)
	}

	for _, t := range diag.Transitions {
		arrow := box.Horizontal + box.Horizontal + box.ArrowRight
		if opts.ASCIIOnly {
			arrow = "-->"
		}

		fromFormatted := formatNode(t.From, false)
		toFormatted := formatNode(t.To, true)

		if t.Trigger != "" {
			triggerFormatted := fmt.Sprintf(" %s(%s)%s%s ", Style(box.Horizontal+box.Horizontal, ColorGreen, useColor), Style(t.Trigger, ColorYellow, useColor), Style(box.Horizontal+box.Horizontal, ColorGreen, useColor), Style(box.ArrowRight, ColorGreen, useColor))
			if opts.ASCIIOnly {
				triggerFormatted = fmt.Sprintf(" --(%s)--> ", t.Trigger)
			}
			b.WriteString(fmt.Sprintf("  %s%s%s\n", fromFormatted, triggerFormatted, toFormatted))
		} else {
			arrowFormatted := Style(arrow, ColorGreen, useColor)
			b.WriteString(fmt.Sprintf("  %s %s %s\n", fromFormatted, arrowFormatted, toFormatted))
		}
	}

	// State descriptions if present
	hasDescs := false
	for _, s := range diag.States {
		if s.Description != "" {
			if !hasDescs {
				b.WriteString("\n" + Style("State Descriptions:", ColorBold, useColor) + "\n")
				hasDescs = true
			}
			b.WriteString(fmt.Sprintf("  • %s: %s\n", Style(s.Label, ColorCyan, useColor), s.Description))
		}
	}

	return b.String(), nil
}
