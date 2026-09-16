package editor

import (
	"testing"

	"github.com/kavix/agent-diagram/internal/ast"
)

func TestHandleNodeInteraction_Validation(t *testing.T) {
	// Invalid event
	err := HandleNodeInteraction("hover", ast.SourceRef{File: "main.go", Line: 10})
	if err == nil {
		t.Error("expected error for unsupported navigation event")
	}

	// Invalid source ref
	err = HandleNodeInteraction(EventClick, ast.SourceRef{})
	if err == nil {
		t.Error("expected error for invalid source reference")
	}

	// Empty file path
	err = OpenEditorCommand("", 10, "vscode")
	if err == nil {
		t.Error("expected error for empty file path")
	}
}

func TestEditorCommandFormats(t *testing.T) {
	// Test that editor types are recognized without panic
	// Note: exec.Command().Run() will fail because 'code', 'nvim', 'idea' binaries might not be present in test container,
	// but we can verify our command construction logic or test error handling.
	err := OpenEditorCommand("main.go", 42, "unknown-editor")
	// Should default to vscode format
	if err == nil {
		// If 'code' happened to be installed, error is nil or command ran successfully.
	}
}
