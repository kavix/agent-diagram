package editor

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/kavix/agent-diagram/internal/ast"
)

// EditorType represents supported code editors for node navigation.
type EditorType string

const (
	EditorVSCode   EditorType = "vscode"
	EditorNeovim   EditorType = "neovim"
	EditorGoLand   EditorType = "goland"
	EditorIDEA     EditorType = "idea"
)

// NavigationEvent represents an interaction type triggering node navigation (Click or Enter).
type NavigationEvent string

const (
	EventClick NavigationEvent = "click"
	EventEnter NavigationEvent = "enter"
)

// OpenEditorCommand builds and executes the shell command to open a source file at a specific line.
// Supports VS Code, Neovim, GoLand, and IDEA, with easy extensibility for other editors.
func OpenEditorCommand(file string, line int, editorType ...string) error {
	if file == "" {
		return fmt.Errorf("source file path cannot be empty")
	}
	if line < 1 {
		line = 1
	}

	ed := string(EditorVSCode)
	if len(editorType) > 0 && editorType[0] != "" {
		ed = strings.ToLower(editorType[0])
	}

	var cmd *exec.Cmd
	switch ed {
	case "nvim", "vim":
		// Neovim: nvim +<line> <file>
		cmd = exec.Command("nvim", fmt.Sprintf("+%d", line), file)
	case "goland", "idea":
		// GoLand / IDEA: idea --line <line> <file>
		cmd = exec.Command("idea", "--line", fmt.Sprintf("%d", line), file)
	case "code", "visual-studio-code":
		fallthrough
	default:
		// VS Code: code --goto <file>:<line>
		target := fmt.Sprintf("%s:%d", file, line)
		cmd = exec.Command("code", "--goto", target)
	}

	return cmd.Run()
}

// HandleNodeInteraction handles an interactive event ('click' or 'enter') on a diagram node or lifeline
// carrying a source reference, and executes the corresponding editor command.
func HandleNodeInteraction(event NavigationEvent, source ast.SourceRef, editorType ...string) error {
	if event != EventClick && event != EventEnter {
		return fmt.Errorf("unsupported navigation event: %s (expected 'click' or 'enter')", event)
	}
	if !source.IsValid() {
		return fmt.Errorf("invalid or missing source reference (file/line) on node")
	}
	return OpenEditorCommand(source.File, source.Line, editorType...)
}
