package mermaid_test

import (
	"testing"

	"github.com/kavix/agent-diagram/internal/ast"
	"github.com/kavix/agent-diagram/internal/parser/mermaid"
)

func TestParseStateDiagram_Basic(t *testing.T) {
	input := `
stateDiagram-v2
    [*] --> Idle
    Idle --> Processing : JobQueued
    Processing --> Done : Success
    Processing --> Error : Fail
    Done --> [*]
    Error --> [*]

    Idle : Waiting for work
    Processing : Actively executing tasks
`

	diag, err := mermaid.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error parsing state diagram: %v", err)
	}

	stateDiag, ok := diag.(*ast.StateDiagram)
	if !ok {
		t.Fatalf("expected *ast.StateDiagram, got %T", diag)
	}

	if len(stateDiag.Transitions) != 6 {
		t.Fatalf("expected 6 transitions, got %d", len(stateDiag.Transitions))
	}

	if stateDiag.Transitions[1].Trigger != "JobQueued" {
		t.Errorf("expected trigger JobQueued, got %q", stateDiag.Transitions[1].Trigger)
	}

	idleState := stateDiag.States["Idle"]
	if idleState == nil || idleState.Description != "Waiting for work" {
		t.Errorf("unexpected Idle state metadata: %+v", idleState)
	}
}
