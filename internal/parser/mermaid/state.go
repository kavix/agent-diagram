package mermaid

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kavix/agent-diagram/internal/ast"
)

var (
	// State1 --> State2 : Trigger or [*] --> State1
	stateTransitionRegex = regexp.MustCompile(`^(\[\*\]|[a-zA-Z0-9_\-\.]+)\s*-->\s*(\[\*\]|[a-zA-Z0-9_\-\.]+)(?:\s*:\s*(.*))?$`)

	// state "Label" as StateID
	stateAliasRegex = regexp.MustCompile(`^state\s+"([^"]+)"\s+as\s+([a-zA-Z0-9_\-\.]+)$`)

	// StateID : Description
	stateDescRegex = regexp.MustCompile(`^([a-zA-Z0-9_\-\.]+)\s*:\s*(.+)$`)
)

// ParseState parses lines containing a stateDiagram or stateDiagram-v2 block.
func ParseState(lines []string) (*ast.StateDiagram, error) {
	diag := &ast.StateDiagram{
		States:      make(map[string]*ast.StateNode),
		StateOrder:  make([]string, 0),
		Transitions: make([]*ast.StateTransition, 0),
	}

	stateMap := make(map[string]*ast.StateNode)

	ensureState := func(id, label string, sType ast.StateType) *ast.StateNode {
		if id == "[*]" {
			// Virtual start or end
			return &ast.StateNode{
				ID:    id,
				Label: "[*]",
				Type:  sType,
			}
		}
		if s, exists := stateMap[id]; exists {
			if label != "" && s.Label == id {
				s.Label = label
			}
			return s
		}
		if label == "" {
			label = id
		}
		if sType == "" {
			sType = ast.StateNormal
		}
		s := &ast.StateNode{
			ID:    id,
			Label: label,
			Type:  sType,
			Order: len(diag.StateOrder),
		}
		stateMap[id] = s
		diag.StateOrder = append(diag.StateOrder, id)
		return s
	}

	for lineIdx, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "%%") {
			continue
		}

		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "statediagram") {
			continue
		}

		// Check title
		if m := titleRegex.FindStringSubmatch(line); len(m) > 1 {
			diag.DiagramTitle = strings.TrimSpace(m[1])
			continue
		}

		// State alias: state "Label" as ID
		if m := stateAliasRegex.FindStringSubmatch(line); len(m) == 3 {
			label := m[1]
			id := m[2]
			ensureState(id, label, ast.StateNormal)
			continue
		}

		// Transition: A --> B : Trigger
		if m := stateTransitionRegex.FindStringSubmatch(line); len(m) >= 3 {
			from := m[1]
			to := m[2]
			trigger := ""
			if len(m) >= 4 {
				trigger = strings.TrimSpace(m[3])
			}

			fromType := ast.StateNormal
			if from == "[*]" {
				fromType = ast.StateStart
			}
			toType := ast.StateNormal
			if to == "[*]" {
				toType = ast.StateEnd
			}

			ensureState(from, "", fromType)
			ensureState(to, "", toType)

			diag.Transitions = append(diag.Transitions, &ast.StateTransition{
				From:    from,
				To:      to,
				Trigger: trigger,
			})
			continue
		}

		// State description: StateID : Desc
		if m := stateDescRegex.FindStringSubmatch(line); len(m) == 3 {
			id := m[1]
			desc := strings.TrimSpace(m[2])
			s := ensureState(id, "", ast.StateNormal)
			s.Description = desc
			continue
		}

		return nil, fmt.Errorf("line %d: unrecognized state diagram statement: %q", lineIdx+1, line)
	}

	diag.States = stateMap
	return diag, nil
}
