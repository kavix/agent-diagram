package renderer_test

import (
	"strings"
	"testing"

	"github.com/kavix/agent-diagram/internal/layout"
	"github.com/kavix/agent-diagram/internal/parser/mermaid"
	"github.com/kavix/agent-diagram/internal/renderer"
)

func TestRenderSequence_Modes(t *testing.T) {
	source := `
sequenceDiagram
    autonumber
    participant Rec as Reconciler
    participant WL as Workload
    participant Client as K8s API

    Rec->>WL: Reconcile scale to zero
    Rec->>Client: releaseScaleDownReservation()
    Client-->>Rec: Ok
`

	diag, err := mermaid.Parse(source)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// 1. Full Mode
	outFull, err := renderer.Render(diag, renderer.RenderOptions{
		Width:   120,
		Mode:    layout.ModeFull,
		NoColor: true,
	})
	if err != nil {
		t.Fatalf("render full error: %v", err)
	}
	if !strings.Contains(outFull, "Reconciler") || !strings.Contains(outFull, "Workload") {
		t.Errorf("full render missing participants:\n%s", outFull)
	}
	if !strings.Contains(outFull, "►") && !strings.Contains(outFull, ">") {
		t.Errorf("full render missing arrows:\n%s", outFull)
	}

	// 2. Compact Mode
	outCompact, err := renderer.Render(diag, renderer.RenderOptions{
		Width:   80,
		Mode:    layout.ModeCompact,
		NoColor: true,
	})
	if err != nil {
		t.Fatalf("render compact error: %v", err)
	}
	if !strings.Contains(outCompact, "Sequence Diagram (Compact)") {
		t.Errorf("compact render missing header:\n%s", outCompact)
	}
	if !strings.Contains(outCompact, "1. Reconciler ──► Workload") {
		t.Errorf("compact render missing formatted step 1:\n%s", outCompact)
	}

	// 3. Narrow Mode
	outNarrow, err := renderer.Render(diag, renderer.RenderOptions{
		Width:   40,
		Mode:    layout.ModeNarrow,
		NoColor: true,
	})
	if err != nil {
		t.Fatalf("render narrow error: %v", err)
	}
	if !strings.Contains(outNarrow, "[1] Rec ─► WL") {
		t.Errorf("narrow render missing compact step:\n%s", outNarrow)
	}
}

func TestRenderFlowchart_Modes(t *testing.T) {
	source := `
flowchart TD
    A[Start] --> B{Valid?}
    B -->|Yes| C[Process]
    B -->|No| D[Reject]
`

	diag, err := mermaid.Parse(source)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Full Mode
	outFull, err := renderer.Render(diag, renderer.RenderOptions{
		Width:   100,
		Mode:    layout.ModeFull,
		NoColor: true,
	})
	if err != nil {
		t.Fatalf("render full error: %v", err)
	}
	if !strings.Contains(outFull, "Start") || !strings.Contains(outFull, "Valid?") {
		t.Errorf("full flowchart missing nodes:\n%s", outFull)
	}

	// Compact Mode
	outCompact, err := renderer.Render(diag, renderer.RenderOptions{
		Width:   60,
		Mode:    layout.ModeCompact,
		NoColor: true,
	})
	if err != nil {
		t.Fatalf("render compact error: %v", err)
	}
	if !strings.Contains(outCompact, "[A] Start") || !strings.Contains(outCompact, "[B] Valid?") {
		t.Errorf("compact flowchart missing node entries:\n%s", outCompact)
	}
}
