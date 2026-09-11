package mermaid_test

import (
	"testing"

	"github.com/kavix/agent-diagram/internal/ast"
	"github.com/kavix/agent-diagram/internal/parser/mermaid"
)

func TestParseFlowchart_Basic(t *testing.T) {
	input := `
flowchart TD
    A[Start Reconcile] --> B{Workload on hold?}
    B -->|Yes| C[Clear on hold flag]
    B -->|No| D[Check scale down]
    C --> E[Sync Status]
    D --> E
`

	diag, err := mermaid.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error parsing flowchart: %v", err)
	}

	flow, ok := diag.(*ast.FlowchartDiagram)
	if !ok {
		t.Fatalf("expected *ast.FlowchartDiagram, got %T", diag)
	}

	if flow.Direction != ast.DirectionTD {
		t.Errorf("expected direction TD, got %s", flow.Direction)
	}

	if len(flow.Nodes) != 5 {
		t.Fatalf("expected 5 nodes, got %d", len(flow.Nodes))
	}

	if flow.Nodes["B"].Shape != ast.ShapeDiamond {
		t.Errorf("expected diamond shape for node B, got %s", flow.Nodes["B"].Shape)
	}

	if len(flow.Edges) != 5 {
		t.Fatalf("expected 5 edges, got %d", len(flow.Edges))
	}

	if flow.Edges[1].Text != "Yes" {
		t.Errorf("expected edge text 'Yes', got %q", flow.Edges[1].Text)
	}
}
