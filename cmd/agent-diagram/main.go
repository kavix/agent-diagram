package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kavix/agent-diagram/internal/layout"
	"github.com/kavix/agent-diagram/internal/mcp"
	"github.com/kavix/agent-diagram/internal/parser/mermaid"
	"github.com/kavix/agent-diagram/internal/renderer"
	"golang.org/x/term"
)

const Version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcmd := os.Args[1]
	switch subcmd {
	case "render":
		runRender(os.Args[2:])
	case "mcp":
		runMCP()
	case "version", "--version", "-v":
		fmt.Printf("agent-diagram v%s\n", Version)
	case "help", "--help", "-h":
		printUsage()
	default:
		// Check if user passed a file directly as first argument (e.g., `agent-diagram diagram.mmd`)
		if _, err := os.Stat(subcmd); err == nil {
			runRender(os.Args[1:])
			return
		}
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", subcmd)
		printUsage()
		os.Exit(1)
	}
}

func runRender(args []string) {
	renderFlags := flag.NewFlagSet("render", flag.ExitOnError)
	widthFlag := renderFlags.Int("width", 0, "Target width in columns (default: auto-detected from terminal)")
	modeFlag := renderFlags.String("mode", "auto", "Layout mode: auto, full, compact, narrow")
	noColorFlag := renderFlags.Bool("no-color", false, "Disable ANSI color output")
	asciiFlag := renderFlags.Bool("ascii", false, "Use ASCII box drawing instead of Unicode")

	// Separate flags from positional arguments so flag order doesn't matter
	var flagArgs []string
	var posArgs []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flagArgs = append(flagArgs, arg)
			// Check if flag takes a value in next argument
			if (arg == "--width" || arg == "-width" || arg == "--mode" || arg == "-mode") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flagArgs = append(flagArgs, args[i])
			}
		} else {
			posArgs = append(posArgs, arg)
		}
	}

	_ = renderFlags.Parse(flagArgs)
	positional := posArgs

	var sourceBytes []byte
	var err error

	if len(positional) > 0 && positional[0] != "-" {
		filePath := positional[0]
		sourceBytes, err = os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
			os.Exit(1)
		}
	} else {
		// Read from stdin
		sourceBytes, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
	}

	source := string(sourceBytes)
	if len(source) == 0 {
		fmt.Fprintln(os.Stderr, "Error: empty diagram input")
		os.Exit(1)
	}

	diag, err := mermaid.Parse(source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse Error: %v\n", err)
		os.Exit(1)
	}

	targetWidth := *widthFlag
	if targetWidth <= 0 {
		targetWidth = layout.DetectTerminalWidth()
	}

	// Auto-disable color if output is not a terminal and --no-color was not specified
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))
	noColor := *noColorFlag || !isTTY

	out, err := renderer.Render(diag, renderer.RenderOptions{
		Width:     targetWidth,
		Mode:      layout.RenderMode(*modeFlag),
		NoColor:   noColor,
		ASCIIOnly: *asciiFlag,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Render Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(out)
}

func runMCP() {
	server := mcp.NewServer(os.Stdin, os.Stdout)
	if err := server.Serve(); err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`agent-diagram v%s - Terminal-native visualization engine for AI coding agents

USAGE:
  agent-diagram render [file] [flags]     Render a Mermaid diagram
  agent-diagram mcp                       Start Model Context Protocol (MCP) server
  agent-diagram version                   Show version information

RENDER FLAGS:
  --width int       Target column width (default: auto-detected)
  --mode string     Layout mode: auto, full, compact, narrow (default: auto)
  --no-color        Disable ANSI syntax colors
  --ascii           Use pure ASCII characters (+, -, |) instead of Unicode

EXAMPLES:
  # Render a sequence diagram file
  agent-diagram render diagram.mmd

  # Pipe Mermaid text directly into agent-diagram
  echo "flowchart TD\n  A --> B" | agent-diagram render

  # Preview with constrained width (e.g. 80 columns)
  agent-diagram render diagram.mmd --width 80

  # Run as MCP server for agy CLI or Claude Code
  agent-diagram mcp
`, Version)
}
