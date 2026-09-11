package mermaid

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kavix/agent-diagram/internal/ast"
)

var (
	headerFlowchartRegex = regexp.MustCompile(`^(?:flowchart|graph)\s+([A-Za-z]{2})\b`)
	subgraphStartRegex   = regexp.MustCompile(`^subgraph\s+([a-zA-Z0-9_\-]+)(?:\s+\[?([^\]]+)\]?)?$`)
)

// ParseFlowchart parses lines containing a flowchart or graph diagram.
func ParseFlowchart(lines []string) (*ast.FlowchartDiagram, error) {
	diag := &ast.FlowchartDiagram{
		Direction: ast.DirectionTD,
		Nodes:     make(map[string]*ast.FlowNode),
		NodeOrder: make([]string, 0),
		Edges:     make([]*ast.FlowEdge, 0),
		Subgraphs: make([]*ast.FlowSubgraph, 0),
	}

	var currentSubgraph *ast.FlowSubgraph

	for lineIdx, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "%%") {
			continue
		}

		// Header check
		if m := headerFlowchartRegex.FindStringSubmatch(line); len(m) > 1 {
			diag.Direction = ast.FlowDirection(strings.ToUpper(m[1]))
			continue
		}

		// Subgraph start
		if m := subgraphStartRegex.FindStringSubmatch(line); len(m) > 1 {
			sgID := m[1]
			title := sgID
			if len(m) > 2 && m[2] != "" {
				title = strings.TrimSpace(m[2])
			}
			currentSubgraph = &ast.FlowSubgraph{
				ID:      sgID,
				Title:   title,
				NodeIDs: make([]string, 0),
			}
			diag.Subgraphs = append(diag.Subgraphs, currentSubgraph)
			continue
		}

		// Subgraph end
		if strings.EqualFold(line, "end") {
			currentSubgraph = nil
			continue
		}

		// Try parsing edges or nodes
		if err := parseFlowchartStatement(line, diag, currentSubgraph); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineIdx+1, err)
		}
	}

	return diag, nil
}

func ensureNode(diag *ast.FlowchartDiagram, id, text string, shape ast.NodeShape) *ast.FlowNode {
	if n, exists := diag.Nodes[id]; exists {
		if text != "" && (n.Text == id || n.Text == "") {
			n.Text = text
		}
		if shape != ast.ShapeRect && n.Shape == ast.ShapeRect {
			n.Shape = shape
		}
		return n
	}

	if text == "" {
		text = id
	}
	if shape == "" {
		shape = ast.ShapeRect
	}

	n := &ast.FlowNode{
		ID:    id,
		Text:  text,
		Shape: shape,
		Order: len(diag.NodeOrder),
	}
	diag.Nodes[id] = n
	diag.NodeOrder = append(diag.NodeOrder, id)
	return n
}

// parseNodeToken extracts id, label, and shape from a token like `A["My Label"]` or `B(Process)`.
func parseNodeToken(token string) (id string, text string, shape ast.NodeShape, matched bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", "", "", false
	}

	// Order of shapes: stadium `([])`, cylinder `[()]`, diamond `{}`, rounded `()`, rect `[]`, circle `(())`
	// Stadium: id([text])
	if idx := strings.Index(token, "(["); idx != -1 && strings.HasSuffix(token, "])") {
		id = strings.TrimSpace(token[:idx])
		text = token[idx+2 : len(token)-2]
		return id, cleanNodeText(text), ast.ShapeStadium, true
	}
	// Cylinder: id[(text)]
	if idx := strings.Index(token, "[("); idx != -1 && strings.HasSuffix(token, ")]") {
		id = strings.TrimSpace(token[:idx])
		text = token[idx+2 : len(token)-2]
		return id, cleanNodeText(text), ast.ShapeCylinder, true
	}
	// Circle: id((text))
	if idx := strings.Index(token, "(("); idx != -1 && strings.HasSuffix(token, "))") {
		id = strings.TrimSpace(token[:idx])
		text = token[idx+2 : len(token)-2]
		return id, cleanNodeText(text), ast.ShapeCircle, true
	}
	// Diamond: id{text}
	if idx := strings.Index(token, "{"); idx != -1 && strings.HasSuffix(token, "}") {
		id = strings.TrimSpace(token[:idx])
		text = token[idx+1 : len(token)-1]
		return id, cleanNodeText(text), ast.ShapeDiamond, true
	}
	// Rounded: id(text)
	if idx := strings.Index(token, "("); idx != -1 && strings.HasSuffix(token, ")") {
		id = strings.TrimSpace(token[:idx])
		text = token[idx+1 : len(token)-1]
		return id, cleanNodeText(text), ast.ShapeRounded, true
	}
	// Rect: id[text]
	if idx := strings.Index(token, "["); idx != -1 && strings.HasSuffix(token, "]") {
		id = strings.TrimSpace(token[:idx])
		text = token[idx+1 : len(token)-1]
		return id, cleanNodeText(text), ast.ShapeRect, true
	}

	// Plain ID without brackets
	return token, token, ast.ShapeRect, true
}

func cleanNodeText(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"") && len(text) >= 2 {
		text = text[1 : len(text)-1]
	}
	return text
}

// Regex to find edge operators (most specific patterns must come first)
var edgeOpRegex = regexp.MustCompile(`(-->\|[^|\n]+\||---\|[^|\n]+\||--\s*[^-\n>]+\s*-->|-\.\s*[^.\n>]+\s*\.->|-->|==>|-\.->|---)`)

func parseFlowchartStatement(line string, diag *ast.FlowchartDiagram, sg *ast.FlowSubgraph) error {
	matches := edgeOpRegex.FindAllStringIndex(line, -1)
	if len(matches) == 0 {
		// Single standalone node declaration like A[Label]
		id, text, shape, ok := parseNodeToken(line)
		if !ok || id == "" {
			return fmt.Errorf("invalid flowchart token: %q", line)
		}
		ensureNode(diag, id, text, shape)
		if sg != nil {
			sg.NodeIDs = append(sg.NodeIDs, id)
		}
		return nil
	}

	// Chain of edges: e.g., A --> B --> C
	var tokens []string
	var edgeOps []string

	lastEnd := 0
	for _, m := range matches {
		nodePart := strings.TrimSpace(line[lastEnd:m[0]])
		edgePart := strings.TrimSpace(line[m[0]:m[1]])
		tokens = append(tokens, nodePart)
		edgeOps = append(edgeOps, edgePart)
		lastEnd = m[1]
	}
	tokens = append(tokens, strings.TrimSpace(line[lastEnd:]))

	for i := 0; i < len(edgeOps); i++ {
		fromToken := tokens[i]
		toToken := tokens[i+1]
		op := edgeOps[i]

		fromID, fromText, fromShape, _ := parseNodeToken(fromToken)
		toID, toText, toShape, _ := parseNodeToken(toToken)

		if fromID != "" {
			ensureNode(diag, fromID, fromText, fromShape)
			if sg != nil {
				sg.NodeIDs = append(sg.NodeIDs, fromID)
			}
		}
		if toID != "" {
			ensureNode(diag, toID, toText, toShape)
			if sg != nil {
				sg.NodeIDs = append(sg.NodeIDs, toID)
			}
		}

		edgeStyle := ast.EdgeSolid
		arrow := true
		var edgeText string

		if strings.HasPrefix(op, "-->|") && strings.HasSuffix(op, "|") {
			edgeText = strings.TrimSuffix(strings.TrimPrefix(op, "-->|"), "|")
		} else if strings.HasPrefix(op, "---|") && strings.HasSuffix(op, "|") {
			edgeText = strings.TrimSuffix(strings.TrimPrefix(op, "---|"), "|")
			arrow = false
		} else if strings.HasPrefix(op, "--") && strings.HasSuffix(op, "-->") && len(op) > 5 {
			edgeText = strings.TrimSuffix(strings.TrimPrefix(op, "--"), "-->")
		} else if strings.HasPrefix(op, "-.") && strings.HasSuffix(op, ".->") && len(op) > 5 {
			edgeStyle = ast.EdgeDotted
			edgeText = strings.TrimSuffix(strings.TrimPrefix(op, "-."), ".->")
		} else if op == "==>" {
			edgeStyle = ast.EdgeThick
		} else if op == "-.->" {
			edgeStyle = ast.EdgeDotted
		} else if op == "---" {
			arrow = false
		}

		diag.Edges = append(diag.Edges, astFlowedgeWrapper(fromID, toID, strings.TrimSpace(edgeText), edgeStyle, arrow))
	}

	return nil
}

func astFlowedgeWrapper(from, to, text string, style ast.EdgeStyle, arrow bool) *ast.FlowEdge {
	return &ast.FlowEdge{
		From:  from,
		To:    to,
		Text:  text,
		Style: style,
		Arrow: arrow,
	}
}
