package mermaid

import (
	"fmt"
	"strings"
)

// ValidationError captures a parse error with a structured hint so the LLM
// can self-correct in a single follow-up without a full round-trip parse attempt.
type ValidationError struct {
	Line    int
	Code    string // short machine-readable error code
	Message string // human-readable detail
	Hint    string // one-sentence fix instruction for the LLM
}

func (e *ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("line %d [%s]: %s — hint: %s", e.Line, e.Code, e.Message, e.Hint)
	}
	return fmt.Sprintf("[%s]: %s — hint: %s", e.Code, e.Message, e.Hint)
}

// ValidationResult summarises a fast pre-parse validation.
type ValidationResult struct {
	Valid       bool
	DiagramType string // "sequence", "flowchart", "state", ""
	Errors      []*ValidationError
	// Stats for token-budget decisions
	LineCount         int
	ParticipantCount  int
	MessageCount      int
	EstimatedRenderKB float32 // rough estimate of rendered output size in KB
}

// Validate performs a fast O(n) pre-validation pass over raw Mermaid source.
// It does NOT build a full AST — it only checks structural validity.
// This is called before the full parser to catch common LLM generation errors
// cheaply, returning machine-readable hints for single-shot self-correction.
func Validate(source string) *ValidationResult {
	cleaned := cleanSource(source)
	result := &ValidationResult{}

	if cleaned == "" {
		result.Errors = append(result.Errors, &ValidationError{
			Code:    "EMPTY_SOURCE",
			Message: "diagram source is empty",
			Hint:    "provide a non-empty Mermaid diagram starting with sequenceDiagram, flowchart TD/LR, or stateDiagram-v2",
		})
		return result
	}

	lines := strings.Split(cleaned, "\n")
	result.LineCount = len(lines)

	// Find header
	headerFound := false
	for lineNum, rawLine := range lines {
		trimmed := strings.TrimSpace(rawLine)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "sequencediagram") {
			result.DiagramType = "sequence"
			headerFound = true
			result.Errors = append(result.Errors, validateSequenceLines(lines[lineNum:])...)
		} else if strings.HasPrefix(lower, "flowchart") || strings.HasPrefix(lower, "graph") {
			result.DiagramType = "flowchart"
			headerFound = true
			result.Errors = append(result.Errors, validateFlowchartLines(lines[lineNum:])...)
		} else if strings.HasPrefix(lower, "statediagram") {
			result.DiagramType = "state"
			headerFound = true
			// State diagrams have flexible syntax; basic check only
		} else {
			result.Errors = append(result.Errors, &ValidationError{
				Line:    lineNum + 1,
				Code:    "UNKNOWN_HEADER",
				Message: fmt.Sprintf("unrecognized diagram header: %q", trimmed),
				Hint:    "first non-comment line must be one of: sequenceDiagram, flowchart TD, flowchart LR, graph TD, stateDiagram-v2",
			})
		}
		break
	}

	if !headerFound && len(result.Errors) == 0 {
		result.Errors = append(result.Errors, &ValidationError{
			Code:    "NO_HEADER",
			Message: "no diagram header found",
			Hint:    "start with sequenceDiagram, flowchart TD, or stateDiagram-v2",
		})
	}

	result.Valid = len(result.Errors) == 0

	// Estimate render output size for token budget decisions
	// Full mode: ~80 chars/line × (events × 3 rows) + header overhead
	// Compact mode: ~50 chars/line × events
	result.EstimatedRenderKB = float32(result.LineCount*80) / 1024.0

	return result
}

// validateSequenceLines checks sequence diagram statements for common LLM errors.
func validateSequenceLines(lines []string) []*ValidationError {
	var errs []*ValidationError
	participantCount := 0
	messageCount := 0

	for i, rawLine := range lines {
		lineNum := i + 1
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "%%") || strings.EqualFold(line, "sequenceDiagram") || strings.EqualFold(line, "autonumber") {
			continue
		}

		lower := strings.ToLower(line)

		// Participant / actor
		if strings.HasPrefix(lower, "participant") || strings.HasPrefix(lower, "actor") {
			participantCount++
			// Validate: participant X as Label — check no colons in ID
			if strings.Count(line, ":") > 1 {
				errs = append(errs, &ValidationError{
					Line:    lineNum,
					Code:    "PARTICIPANT_SYNTAX",
					Message: fmt.Sprintf("participant declaration has extra colons: %q", line),
					Hint:    "participant syntax is: participant ID as Label (no colons in participant lines)",
				})
			}
			continue
		}

		// Loop / par / alt / else / end / note — structural keywords
		if matchesKeyword(lower, "loop", "par", "alt", "else", "end", "note", "rect", "critical", "break", "opt") {
			continue
		}

		// Title
		if strings.HasPrefix(lower, "title") {
			continue
		}

		// Message arrows — the most common LLM error source
		if strings.Contains(line, ":") {
			messageCount++
			if !containsArrow(line) {
				errs = append(errs, &ValidationError{
					Line:    lineNum,
					Code:    "MISSING_ARROW",
					Message: fmt.Sprintf("line looks like a message but has no arrow: %q", line),
					Hint:    "sequence messages need an arrow: use A->>B: message or A-->B: message",
				})
			}
		} else if containsArrow(line) {
			errs = append(errs, &ValidationError{
				Line:    lineNum,
				Code:    "MISSING_COLON",
				Message: fmt.Sprintf("arrow found but no message label (missing colon): %q", line),
				Hint:    "add a colon and message after the arrow: A->>B: your message here",
			})
		}
	}

	// Warn if too many participants (likely to overflow terminal width)
	if participantCount > 8 {
		errs = append(errs, &ValidationError{
			Code:    "TOO_MANY_PARTICIPANTS",
			Message: fmt.Sprintf("%d participants declared — will overflow most terminal widths", participantCount),
			Hint:    "keep sequence diagrams to ≤6 participants; split complex flows into multiple diagrams",
		})
	}
	_ = messageCount
	return errs
}

// validateFlowchartLines checks flowchart statements for common LLM generation errors.
func validateFlowchartLines(lines []string) []*ValidationError {
	var errs []*ValidationError

	for i, rawLine := range lines {
		lineNum := i + 1
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "%%") {
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "flowchart") || strings.HasPrefix(lower, "graph") ||
			strings.HasPrefix(lower, "subgraph") || strings.HasPrefix(lower, "end") ||
			strings.HasPrefix(lower, "style") || strings.HasPrefix(lower, "classDef") {
			continue
		}

		// Common error: using -> instead of --> in flowcharts
		if strings.Contains(line, "->") && !strings.Contains(line, "-->") && !strings.Contains(line, "->>") {
			errs = append(errs, &ValidationError{
				Line:    lineNum,
				Code:    "WRONG_ARROW_FLOWCHART",
				Message: fmt.Sprintf("flowchart uses -> but should use --> for directed edges: %q", line),
				Hint:    "flowchart edges use --> not -> (two dashes); use ==> for thick edges, -.-> for dotted",
			})
		}
	}
	return errs
}

func containsArrow(s string) bool {
	return strings.Contains(s, "->>") || strings.Contains(s, "-->") ||
		strings.Contains(s, "->") || strings.Contains(s, "--x") ||
		strings.Contains(s, "-x")
}

func matchesKeyword(lower string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.HasPrefix(lower, kw) {
			return true
		}
	}
	return false
}
