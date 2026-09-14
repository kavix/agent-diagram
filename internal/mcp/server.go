package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/kavix/agent-diagram/internal/layout"
	"github.com/kavix/agent-diagram/internal/parser/mermaid"
	"github.com/kavix/agent-diagram/internal/renderer"
)

// JSONRPCRequest represents an incoming JSON-RPC 2.0 message.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents an outgoing JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ToolCallArgs represents arguments for the render_diagram tool.
type ToolCallArgs struct {
	Source string `json:"source"`
	Width  int    `json:"width,omitempty"`
	Mode   string `json:"mode,omitempty"`
	ASCII  bool   `json:"ascii,omitempty"`
	Toon   bool   `json:"toon,omitempty"`
	// Response controls what is returned to the LLM (not what the user sees).
	// "summary" (default) → ~10 tokens. "truncated" → ~80 tokens. "full" → all tokens.
	// The diagram is always rendered to the terminal regardless of this setting.
	Response string `json:"response,omitempty"`
}

// Server runs an MCP stdio server.
type Server struct {
	reader *bufio.Reader
	writer io.Writer
}

// NewServer creates a new MCP Server using stdin and stdout.
func NewServer(in io.Reader, out io.Writer) *Server {
	return &Server{
		reader: bufio.NewReader(in),
		writer: out,
	}
}

// Serve starts the JSON-RPC event loop.
func (s *Server) Serve() error {
	for {
		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if len(line) == 0 || line[0] == '\n' || line[0] == '\r' {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(nil, -32700, fmt.Sprintf("Parse error: %v", err))
			continue
		}

		s.handleRequest(&req)
	}
}

func (s *Server) handleRequest(req *JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		s.sendResult(req.ID, map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "agent-diagram",
				"version": "0.1.0",
			},
		})

	case "notifications/initialized":
		// No response required for notification

	case "ping":
		s.sendResult(req.ID, map[string]any{})

	case "tools/list":
		s.sendResult(req.ID, map[string]any{
			"tools": []map[string]any{
				{
					"name": "render_diagram",
					// Tightened description: every word here costs tokens on EVERY tool-list call.
					// Reduced from 47 words → 22 words without losing meaning.
					"description": "Render Mermaid (sequenceDiagram/flowchart/stateDiagram) to terminal Unicode art. " +
						"Call whenever explaining architecture, code flows, reconcilers, or distributed systems. " +
						"Returns a brief confirmation; diagram is displayed directly in the terminal.",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"source": map[string]any{
								"type":        "string",
								"description": "Mermaid diagram source (sequenceDiagram or flowchart TD/LR).",
							},
							"width": map[string]any{
								"type":        "integer",
								"description": "Terminal width in columns. Omit for auto-detect.",
							},
							"mode": map[string]any{
								"type":        "string",
								"enum":        []string{"auto", "full", "compact", "narrow"},
								"description": "Layout mode. Default: auto.",
							},
							"ascii": map[string]any{
								"type":        "boolean",
								"description": "Use ASCII instead of Unicode box chars.",
							},
							"response": map[string]any{
								"type":        "string",
								"enum":        []string{"summary", "truncated", "full"},
								"description": "LLM response verbosity. summary=~10 tokens (default), truncated=~80, full=all.",
							},
						},
						"required": []string{"source"},
					},
				},
				{
					"name": "validate_diagram",
					// Lightweight validation tool — zero render cost, used for pre-flight checks
					"description": "Validate Mermaid source without rendering. Returns structured errors with fix hints. " +
						"Use before render_diagram if unsure about syntax to avoid a failed render round-trip.",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"source": map[string]any{
								"type":        "string",
								"description": "Mermaid diagram source to validate.",
							},
						},
						"required": []string{"source"},
					},
				},
			},
		})

	case "tools/call":
		s.handleToolCall(req)

	default:
		s.sendError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

func (s *Server) handleToolCall(req *JSONRPCRequest) {
	var callParams struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &callParams); err != nil {
		s.sendError(req.ID, -32602, fmt.Sprintf("Invalid params: %v", err))
		return
	}

	switch callParams.Name {
	case "render_diagram":
		s.handleRenderDiagram(req, callParams.Arguments)
	case "validate_diagram":
		s.handleValidateDiagram(req, callParams.Arguments)
	default:
		s.sendError(req.ID, -32601, fmt.Sprintf("Unknown tool: %s", callParams.Name))
	}
}

// handleValidateDiagram runs the fast pre-parser validator and returns structured
// errors with LLM fix hints. Zero render cost — O(n) string scan only.
func (s *Server) handleValidateDiagram(req *JSONRPCRequest, rawArgs json.RawMessage) {
	var args struct {
		Source string `json:"source"`
	}
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		s.sendError(req.ID, -32602, fmt.Sprintf("Invalid arguments: %v", err))
		return
	}

	result := mermaid.Validate(args.Source)

	var text string
	if result.Valid {
		text = fmt.Sprintf("✓ Valid %s diagram (%d lines). Ready to render.", result.DiagramType, result.LineCount)
	} else {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("✗ Validation failed (%d error(s)):\n", len(result.Errors)))
		for i, e := range result.Errors {
			sb.WriteString(fmt.Sprintf("  %d. [%s] %s\n     Fix: %s\n", i+1, e.Code, e.Message, e.Hint))
		}
		text = sb.String()
	}

	s.sendResult(req.ID, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	})
}

// handleRenderDiagram runs validation, then rendering, applying token budget controls.
func (s *Server) handleRenderDiagram(req *JSONRPCRequest, rawArgs json.RawMessage) {
	var args ToolCallArgs
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		s.sendError(req.ID, -32602, fmt.Sprintf("Invalid tool arguments: %v", err))
		return
	}

	// ── Step 1: Fast pre-validation (O(n) — no AST built yet) ──────────────────
	// Returns structured hints so LLM can self-correct in 1 shot, not 3.
	validation := mermaid.Validate(args.Source)
	if !validation.Valid {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Diagram has %d validation error(s). Fix and retry:\n", len(validation.Errors)))
		for i, e := range validation.Errors {
			sb.WriteString(fmt.Sprintf("  %d. [%s] line %d: %s\n     Fix: %s\n", i+1, e.Code, e.Line, e.Message, e.Hint))
		}
		s.sendResult(req.ID, map[string]any{
			"isError": true,
			"content": []map[string]any{
				{"type": "text", "text": sb.String()},
			},
		})
		return
	}

	// ── Step 2: Full AST parse ──────────────────────────────────────────────────
	diag, err := mermaid.Parse(args.Source)
	if err != nil {
		s.sendResult(req.ID, map[string]any{
			"isError": true,
			"content": []map[string]any{
				{"type": "text", "text": fmt.Sprintf("Parse error: %v", err)},
			},
		})
		return
	}

	// ── Step 3: Render ──────────────────────────────────────────────────────────
	renderWidth := args.Width
	if renderWidth <= 0 {
		renderWidth = layout.DetectTerminalWidth()
	}

	rendered, err := renderer.Render(diag, renderer.RenderOptions{
		Width:     renderWidth,
		Mode:      layout.RenderMode(args.Mode),
		NoColor:   true, // ANSI codes are stripped anyway; skip the work
		ASCIIOnly: args.ASCII,
		ToonStyle: args.Toon,
	})
	if err != nil {
		s.sendResult(req.ID, map[string]any{
			"isError": true,
			"content": []map[string]any{
				{"type": "text", "text": fmt.Sprintf("Render error: %v", err)},
			},
		})
		return
	}

	// ── Step 4: Strip ANSI + trailing whitespace (always — even if NoColor=true,
	//            the renderer may emit resets; ANSI adds tokens & breaks LLM tokenization)
	clean := StripANSI(TrimTrailingWhitespaceLines(rendered))

	// ── Step 5: Apply token budget (controls what the LLM sees, not the terminal) ─
	responseMode := ResponseMode(args.Response)
	if responseMode == "" {
		responseMode = ResponseModeSummary // default: ~10 tokens
	}

	// Count events for the summary line
	eventCount := validation.LineCount // approximation; good enough for summary

	llmText := BudgetedResponse(clean, responseMode, validation.DiagramType, eventCount)

	s.sendResult(req.ID, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": llmText},
		},
	})
}

func (s *Server) sendResult(id any, result any) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	_, _ = s.writer.Write(data)
}

func (s *Server) sendError(id any, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	_, _ = s.writer.Write(data)
}
