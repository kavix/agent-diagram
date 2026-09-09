package mermaid

import (
	"fmt"
	"strings"

	"github.com/kavix/agent-diagram/internal/ast"
)

// Parse inspects the input source, strips any markdown code fences, detects
// the diagram type (sequenceDiagram, flowchart, graph), and delegates to the
// appropriate specialized parser.
func Parse(source string) (ast.Diagram, error) {
	cleaned := cleanSource(source)
	if cleaned == "" {
		return nil, fmt.Errorf("empty diagram source")
	}

	lines := strings.Split(cleaned, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "sequencediagram") {
			return ParseSequence(lines[i:])
		}
		if strings.HasPrefix(lower, "flowchart") || strings.HasPrefix(lower, "graph") {
			return ParseFlowchart(lines[i:])
		}
		if strings.HasPrefix(lower, "statediagram") {
			return ParseState(lines[i:])
		}

		return nil, fmt.Errorf("unsupported diagram header on line %d: %q (expected sequenceDiagram, flowchart, graph, or stateDiagram)", i+1, trimmed)
	}

	return nil, fmt.Errorf("no valid diagram header found in source")
}

// cleanSource removes markdown code blocks, normalizes newlines, and trims margins.
func cleanSource(source string) string {
	s := strings.TrimSpace(source)

	// Strip ```mermaid ... ``` or ``` ... ```
	if strings.HasPrefix(s, "```") {
		firstNL := strings.Index(s, "\n")
		if firstNL != -1 {
			s = s[firstNL+1:]
		}
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
	}

	return strings.ReplaceAll(strings.TrimSpace(s), "\r\n", "\n")
}
