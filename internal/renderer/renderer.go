package renderer

import (
	"fmt"

	"github.com/kavix/agent-diagram/internal/ast"
	"github.com/kavix/agent-diagram/internal/layout"
)

// RenderOptions configures diagram rendering.
type RenderOptions struct {
	Width     int
	Mode      layout.RenderMode
	NoColor   bool
	ASCIIOnly bool
}

// Render accepts an AST diagram and options and returns a formatted terminal diagram string.
func Render(diag ast.Diagram, opts RenderOptions) (string, error) {
	if diag == nil {
		return "", fmt.Errorf("nil diagram provided to renderer")
	}

	if opts.Width <= 0 {
		opts.Width = layout.DetectTerminalWidth()
	}

	switch d := diag.(type) {
	case *ast.SequenceDiagram:
		return RenderSequence(d, opts)
	case *ast.FlowchartDiagram:
		return RenderFlowchart(d, opts)
	default:
		return "", fmt.Errorf("unsupported diagram AST type: %T", diag)
	}
}
