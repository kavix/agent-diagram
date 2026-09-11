package mermaid_test

import (
	"testing"

	"github.com/kavix/agent-diagram/internal/ast"
	"github.com/kavix/agent-diagram/internal/parser/mermaid"
)

func TestParseSequence_Basic(t *testing.T) {
	input := `
sequenceDiagram
    autonumber
    participant STS as StatefulSet
    participant WL as Workload
    participant Rec as Reconciler
    participant Client as Kubernetes API

    Rec->>WL: Reconcile scale to zero
    Rec->>Client: releaseScaleDownReservation()
    Rec->>Client: Update PodSets[0].Count = 5
    Rec-->>Client: clearOnHold()
    Note over Rec,Client: Reconciliation completed
`

	diag, err := mermaid.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error parsing sequence: %v", err)
	}

	seq, ok := diag.(*ast.SequenceDiagram)
	if !ok {
		t.Fatalf("expected *ast.SequenceDiagram, got %T", diag)
	}

	if !seq.AutoNumber {
		t.Errorf("expected AutoNumber to be true")
	}

	if len(seq.Participants) != 4 {
		t.Fatalf("expected 4 participants, got %d", len(seq.Participants))
	}

	if seq.Participants[0].ID != "STS" || seq.Participants[0].Label != "StatefulSet" {
		t.Errorf("unexpected participant 0: %+v", seq.Participants[0])
	}
	if seq.Participants[3].ID != "Client" || seq.Participants[3].Label != "Kubernetes API" {
		t.Errorf("unexpected participant 3: %+v", seq.Participants[3])
	}

	if len(seq.Events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(seq.Events))
	}

	if seq.Events[0].Number != 1 || seq.Events[0].Message != "Reconcile scale to zero" {
		t.Errorf("unexpected event 0: %+v", seq.Events[0])
	}

	if seq.Events[3].Arrow != ast.ArrowDottedArrow {
		t.Errorf("expected dotted arrow for event 3, got %s", seq.Events[3].Arrow)
	}

	if len(seq.Notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(seq.Notes))
	}
	if seq.Notes[0].Text != "Reconciliation completed" {
		t.Errorf("unexpected note text: %s", seq.Notes[0].Text)
	}
}
