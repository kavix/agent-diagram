package layout

import (
	"github.com/kavix/agent-diagram/internal/ast"
)

// FlowNodeLayout contains layout coordinates for a single flowchart node.
type FlowNodeLayout struct {
	Node   *ast.FlowNode
	Layer  int
	Index  int // index within layer
	Width  int
	Height int
	X      int
	Y      int
}

// FlowchartLayout holds computed geometry for a flowchart diagram.
type FlowchartLayout struct {
	Mode        RenderMode
	Direction   ast.FlowDirection
	TargetWidth int
	TotalWidth  int
	TotalHeight int
	Layers      [][]*FlowNodeLayout
	NodeLayouts map[string]*FlowNodeLayout
	Diagram     *ast.FlowchartDiagram
}

// ComputeFlowchartLayout determines node layers and coordinates.
func ComputeFlowchartLayout(diag *ast.FlowchartDiagram, targetWidth int, requestedMode RenderMode) *FlowchartLayout {
	if targetWidth <= 0 {
		targetWidth = DetectTerminalWidth()
	}

	mode := requestedMode
	if mode == "" || mode == ModeAuto {
		if targetWidth < 50 {
			mode = ModeCompact
		} else {
			mode = ModeFull
		}
	}

	layout := &FlowchartLayout{
		Mode:        mode,
		Direction:   diag.Direction,
		TargetWidth: targetWidth,
		NodeLayouts: make(map[string]*FlowNodeLayout),
		Diagram:     diag,
	}

	if len(diag.NodeOrder) == 0 {
		return layout
	}

	// Compute node in-degrees and adjacency
	inDegree := make(map[string]int)
	adj := make(map[string][]string)
	for _, id := range diag.NodeOrder {
		inDegree[id] = 0
		adj[id] = make([]string, 0)
	}
	for _, edge := range diag.Edges {
		if _, ok := inDegree[edge.To]; ok {
			inDegree[edge.To]++
		}
		if _, ok := adj[edge.From]; ok {
			adj[edge.From] = append(adj[edge.From], edge.To)
		}
	}

	// Compute layers via topological BFS
	nodeLayer := make(map[string]int)
	queue := make([]string, 0)
	for _, id := range diag.NodeOrder {
		if inDegree[id] == 0 {
			queue = append(queue, id)
			nodeLayer[id] = 0
		}
	}

	// In case of cycles or isolated graphs, ensure all nodes get placed
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		currLayer := nodeLayer[curr]

		for _, next := range adj[curr] {
			if nodeLayer[next] < currLayer+1 {
				nodeLayer[next] = currLayer + 1
				queue = append(queue, next)
			}
		}
	}

	// Any unassigned nodes (e.g. part of cycles)
	for _, id := range diag.NodeOrder {
		if _, ok := nodeLayer[id]; !ok {
			nodeLayer[id] = 0
		}
	}

	// Group by layer
	maxLayer := 0
	for _, layer := range nodeLayer {
		if layer > maxLayer {
			maxLayer = layer
		}
	}

	layers := make([][]*FlowNodeLayout, maxLayer+1)
	for _, id := range diag.NodeOrder {
		node := diag.Nodes[id]
		l := nodeLayer[id]
		textLen := StringWidth(node.Text)
		nodeWidth := textLen + 4 // [ text ]
		if nodeWidth < 8 {
			nodeWidth = 8
		}

		nl := &FlowNodeLayout{
			Node:   node,
			Layer:  l,
			Index:  len(layers[l]),
			Width:  nodeWidth,
			Height: 3, // 3 lines for box
		}
		layers[l] = append(layers[l], nl)
		layout.NodeLayouts[id] = nl
	}

	layout.Layers = layers

	// Check if full layout fits
	if mode == ModeFull {
		// Calculate width needed for the widest layer in TD mode
		maxWidthNeeded := 0
		for _, layerNodes := range layers {
			layerW := 0
			for _, nl := range layerNodes {
				layerW += nl.Width + 4
			}
			if layerW > maxWidthNeeded {
				maxWidthNeeded = layerW
			}
		}
		if maxWidthNeeded > targetWidth && targetWidth < 70 {
			layout.Mode = ModeCompact
		}
	}

	return layout
}
