package renderer

import (
	"fmt"
	"strings"

	"github.com/kavix/agent-diagram/internal/ast"
	"github.com/kavix/agent-diagram/internal/layout"
)

// RenderFlowchart renders an AST FlowchartDiagram according to RenderOptions.
func RenderFlowchart(diag *ast.FlowchartDiagram, opts RenderOptions) (string, error) {
	lay := layout.ComputeFlowchartLayout(diag, opts.Width, opts.Mode)

	if lay.Mode == layout.ModeCompact || lay.Mode == layout.ModeNarrow {
		return renderFlowchartCompact(diag, lay, opts), nil
	}
	return renderFlowchartFull(diag, lay, opts), nil
}

func renderFlowchartFull(diag *ast.FlowchartDiagram, lay *layout.FlowchartLayout, opts RenderOptions) string {
	box := opts.BoxStyle()
	useColor := !opts.NoColor

	var b strings.Builder
	title := "Flowchart"
	if diag.Title() != "" {
		title = diag.Title()
	}
	b.WriteString(Style(fmt.Sprintf("%s (%s)", title, diag.Direction), ColorBold+ColorCyan, useColor) + "\n\n")

	// Pre-index outgoing edges per node
	outgoing := make(map[string][]*ast.FlowEdge)
	for _, edge := range diag.Edges {
		outgoing[edge.From] = append(outgoing[edge.From], edge)
	}

	for layerIdx, layerNodes := range lay.Layers {
		// Render boxes for this layer side-by-side
		var boxLines [3][]rune
		totalWidth := lay.TargetWidth
		if totalWidth < 80 {
			totalWidth = 80
		}
		for i := 0; i < 3; i++ {
			boxLines[i] = make([]rune, totalWidth)
			for j := range boxLines[i] {
				boxLines[i][j] = ' '
			}
		}

		currentX := 2
		const gap = 4
		nodeCenterX := make(map[string]int)

		for _, nl := range layerNodes {
			w := nl.Width
			sx := currentX
			ex := sx + w - 1

			// Shape borders
			tl, tr, bl, br := box.TopLeft, box.TopRight, box.BottomLeft, box.BottomRight
			if nl.Node.Shape == ast.ShapeRounded && !opts.ASCIIOnly {
				tl, tr, bl, br = "╭", "╮", "╰", "╯"
			} else if nl.Node.Shape == ast.ShapeDiamond {
				tl, tr, bl, br = "◇", "◇", "◇", "◇"
			}

			// Top line
			boxLines[0][sx] = []rune(tl)[0]
			for x := sx + 1; x < ex; x++ {
				boxLines[0][x] = []rune(box.Horizontal)[0]
			}
			boxLines[0][ex] = []rune(tr)[0]

			// Mid line with text
			boxLines[1][sx] = []rune(box.Vertical)[0]
			boxLines[1][ex] = []rune(box.Vertical)[0]
			innerW := w - 2
			padded := layout.PadCenter(nl.Node.Text, innerW)
			pRunes := []rune(padded)
			for i, r := range pRunes {
				if sx+1+i < ex {
					boxLines[1][sx+1+i] = r
				}
			}

			// Bot line with TeeDown if has outgoing edges
			hasOutgoing := len(outgoing[nl.Node.ID]) > 0
			cx := sx + (w / 2)
			nodeCenterX[nl.Node.ID] = cx

			boxLines[2][sx] = []rune(bl)[0]
			for x := sx + 1; x < ex; x++ {
				if x == cx && hasOutgoing {
					boxLines[2][x] = []rune(box.TeeDown)[0]
				} else {
					boxLines[2][x] = []rune(box.Horizontal)[0]
				}
			}
			boxLines[2][ex] = []rune(br)[0]

			currentX += w + gap
		}

		// Print the 3 lines of this layer's boxes
		b.WriteString(Style(strings.TrimRight(string(boxLines[0]), " "), ColorCyan, useColor) + "\n")
		b.WriteString(Style(strings.TrimRight(string(boxLines[1]), " "), ColorBold, useColor) + "\n")
		b.WriteString(Style(strings.TrimRight(string(boxLines[2]), " "), ColorCyan, useColor) + "\n")

		// If there is a next layer, render connector arrows down
		if layerIdx < len(lay.Layers)-1 {
			var connLine1 = make([]rune, totalWidth)
			var labelLine = make([]rune, totalWidth)
			var connLine2 = make([]rune, totalWidth)
			for i := range connLine1 {
				connLine1[i] = ' '
				labelLine[i] = ' '
				connLine2[i] = ' '
			}

			for _, nl := range layerNodes {
				cx, hasCx := nodeCenterX[nl.Node.ID]
				if !hasCx {
					continue
				}
				edges := outgoing[nl.Node.ID]
				if len(edges) == 0 {
					continue
				}

				connLine1[cx] = []rune(box.Vertical)[0]
				connLine2[cx] = []rune(box.ArrowDown)[0]

				// If edge has text, print it next to connector
				if len(edges) == 1 && edges[0].Text != "" {
					lblRunes := []rune(fmt.Sprintf("(%s)", edges[0].Text))
					for i, r := range lblRunes {
						if cx+2+i < len(labelLine) {
							labelLine[cx+2+i] = r
						}
					}
				}
			}

			b.WriteString(Style(strings.TrimRight(string(connLine1), " "), ColorGray, useColor) + "\n")
			if strings.TrimSpace(string(labelLine)) != "" {
				b.WriteString(Style(strings.TrimRight(string(labelLine), " "), ColorYellow, useColor) + "\n")
			}
			b.WriteString(Style(strings.TrimRight(string(connLine2), " "), ColorGreen, useColor) + "\n")
		}
	}

	return b.String()
}

func renderFlowchartCompact(diag *ast.FlowchartDiagram, lay *layout.FlowchartLayout, opts RenderOptions) string {
	box := opts.BoxStyle()
	useColor := !opts.NoColor

	var b strings.Builder
	title := "Flowchart"
	if diag.Title() != "" {
		title = diag.Title()
	}
	b.WriteString(Style(fmt.Sprintf("%s (%s - Compact)", title, diag.Direction), ColorBold+ColorCyan, useColor) + "\n")
	b.WriteString(strings.Repeat(box.Horizontal, 60) + "\n")

	// Pre-index edges
	outgoing := make(map[string][]*ast.FlowEdge)
	for _, edge := range diag.Edges {
		outgoing[edge.From] = append(outgoing[edge.From], edge)
	}

	for _, nodeID := range diag.NodeOrder {
		node := diag.Nodes[nodeID]
		edges := outgoing[nodeID]

		nodeLabel := fmt.Sprintf("[%s] %s", node.ID, node.Text)
		if node.ID == node.Text {
			nodeLabel = fmt.Sprintf("[%s]", node.ID)
		}
		b.WriteString(Style(nodeLabel, ColorBold, useColor) + "\n")

		for i, edge := range edges {
			isLast := i == len(edges)-1
			branch := box.TeeRight + box.Horizontal + box.Horizontal + box.ArrowRight
			if isLast {
				branch = box.BottomLeft + box.Horizontal + box.Horizontal + box.ArrowRight
			}

			targetNode := diag.Nodes[edge.To]
			targetName := edge.To
			if targetNode != nil && targetNode.Text != "" {
				targetName = fmt.Sprintf("[%s] %s", targetNode.ID, targetNode.Text)
			}

			edgeDesc := ""
			if edge.Text != "" {
				edgeDesc = fmt.Sprintf(" (%s) %s", Style(edge.Text, ColorYellow, useColor), Style(box.Horizontal+box.ArrowRight, ColorGreen, useColor))
			}

			b.WriteString(fmt.Sprintf("  %s%s %s\n", Style(branch, ColorGreen, useColor), edgeDesc, targetName))
		}

		if len(edges) == 0 {
			b.WriteString(fmt.Sprintf("  %s (Terminal / End)\n", Style(box.BottomLeft+box.Horizontal, ColorGray, useColor)))
		}
		b.WriteString("\n")
	}

	return b.String()
}
