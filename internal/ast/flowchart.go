package ast

// FlowDirection indicates the flow layout orientation.
type FlowDirection string

const (
	DirectionTD FlowDirection = "TD" // Top to Bottom
	DirectionTB FlowDirection = "TB" // Top to Bottom
	DirectionLR FlowDirection = "LR" // Left to Right
	DirectionRL FlowDirection = "RL" // Right to Left
	DirectionBT FlowDirection = "BT" // Bottom to Top
)

// NodeShape defines the visual boundary of a flowchart node.
type NodeShape string

const (
	ShapeRect     NodeShape = "rect"     // [text]
	ShapeRounded  NodeShape = "rounded"  // (text)
	ShapeStadium  NodeShape = "stadium"  // ([text])
	ShapeCylinder NodeShape = "cylinder" // [(text)]
	ShapeCircle   NodeShape = "circle"   // ((text))
	ShapeDiamond  NodeShape = "diamond"  // {text}
)

// EdgeStyle defines line characteristics.
type EdgeStyle string

const (
	EdgeSolid  EdgeStyle = "solid"
	EdgeDotted EdgeStyle = "dotted"
	EdgeThick  EdgeStyle = "thick"
)

// FlowNode represents a vertex or step in the flowchart.
type FlowNode struct {
	ID    string
	Text  string
	Shape NodeShape
	Order int
}

// FlowEdge represents a directed or undirected connection between two nodes.
type FlowEdge struct {
	From  string
	To    string
	Text  string
	Style EdgeStyle
	Arrow bool // true for directed ->, false for undirected --
}

// FlowSubgraph models a grouping of nodes.
type FlowSubgraph struct {
	ID      string
	Title   string
	NodeIDs []string
}

// FlowchartDiagram models a parsed Mermaid flowchart or graph.
type FlowchartDiagram struct {
	DiagramTitle string
	Direction    FlowDirection
	Nodes        map[string]*FlowNode
	NodeOrder    []string
	Edges        []*FlowEdge
	Subgraphs    []*FlowSubgraph
}

func (d *FlowchartDiagram) Type() DiagramType {
	return TypeFlowchart
}

func (d *FlowchartDiagram) Title() string {
	return d.DiagramTitle
}

// GetNode safely retrieves a node by its ID.
func (d *FlowchartDiagram) GetNode(id string) *FlowNode {
	if d.Nodes == nil {
		return nil
	}
	return d.Nodes[id]
}
