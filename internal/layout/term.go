package layout

import (
	"os"
	"strconv"

	"golang.org/x/term"
)

const (
	DefaultTerminalWidth = 100
	MinTerminalWidth     = 30
)

// DetectTerminalWidth determines the terminal width by querying stdout,
// inspecting the COLUMNS environment variable, or falling back to a sensible default.
func DetectTerminalWidth() int {
	// 1. Check COLUMNS environment variable
	if colStr := os.Getenv("COLUMNS"); colStr != "" {
		if col, err := strconv.Atoi(colStr); err == nil && col >= MinTerminalWidth {
			return col
		}
	}

	// 2. Query stdout terminal size
	if term.IsTerminal(int(os.Stdout.Fd())) {
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w >= MinTerminalWidth {
			return w
		}
	}

	// 3. Query stdin terminal size
	if term.IsTerminal(int(os.Stdin.Fd())) {
		if w, _, err := term.GetSize(int(os.Stdin.Fd())); err == nil && w >= MinTerminalWidth {
			return w
		}
	}

	return DefaultTerminalWidth
}
