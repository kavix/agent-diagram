package mcp

import (
	"fmt"
	"strings"
)

// ResponseMode controls how much of the rendered diagram is returned to the LLM.
// The LLM almost never needs to read the full rendered art — the user sees it
// directly on screen. Returning the full art wastes input tokens on the next turn.
type ResponseMode string

const (
	// ResponseModeFull returns the complete rendered diagram. Use when the LLM
	// explicitly needs to reason about diagram content (e.g. code intelligence).
	ResponseModeFull ResponseMode = "full"

	// ResponseModeSummary returns only a one-line confirmation + stats.
	// This is the default and lowest-cost mode: ~10–15 tokens instead of 200–2000.
	ResponseModeSummary ResponseMode = "summary"

	// ResponseModeTruncated returns up to MaxResponseLines lines, then a
	// truncation notice. Good for medium diagrams where the LLM may want
	// to verify a few key participants.
	ResponseModeTruncated ResponseMode = "truncated"
)

// DefaultMaxResponseLines is the cap for ResponseModeTruncated.
const DefaultMaxResponseLines = 20

// BudgetedResponse builds the text returned to the LLM from a rendered diagram.
// It applies the ResponseMode to keep the token cost predictable.
//
// Token cost comparison (approximate, gpt-4o pricing):
//   - ResponseModeFull:      200–2000 tokens  ($0.001–$0.010 per render)
//   - ResponseModeTruncated: 50–100 tokens    ($0.0003 per render)
//   - ResponseModeSummary:   10–15 tokens     ($0.00008 per render) ← default
func BudgetedResponse(rendered string, mode ResponseMode, diagramType string, eventCount int) string {
	switch mode {
	case ResponseModeFull:
		return rendered

	case ResponseModeTruncated:
		lines := strings.Split(rendered, "\n")
		if len(lines) <= DefaultMaxResponseLines {
			return rendered
		}
		head := strings.Join(lines[:DefaultMaxResponseLines], "\n")
		return fmt.Sprintf("%s\n[... %d more lines — diagram rendered to terminal ...]",
			head, len(lines)-DefaultMaxResponseLines)

	default: // ResponseModeSummary
		lines := strings.Split(strings.TrimSpace(rendered), "\n")
		// Extract the first meaningful content line (title or first box row)
		firstContent := ""
		for _, l := range lines {
			t := strings.TrimSpace(l)
			if t != "" {
				firstContent = t
				break
			}
		}
		return fmt.Sprintf("✓ Rendered %s diagram (%d events, %d lines). Shown in terminal.\n  Preview: %s",
			diagramType, eventCount, len(lines), truncateStr(firstContent, 60))
	}
}

// StripANSI removes ANSI escape codes from a string.
// MCP responses sent to LLMs must never contain ANSI — they add token cost
// and confuse the model's token boundary detection.
func StripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Skip until 'm' (end of ANSI sequence)
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++ // consume 'm'
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// TrimTrailingWhitespaceLines removes lines that are purely whitespace from the
// end of rendered output. Empty trailing lines add tokens with zero value.
func TrimTrailingWhitespaceLines(s string) string {
	lines := strings.Split(s, "\n")
	end := len(lines)
	for end > 0 && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return strings.Join(lines[:end], "\n")
}

func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}
