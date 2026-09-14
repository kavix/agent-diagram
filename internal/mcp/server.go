package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

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
	JSONRPC string      `json:"jsonrpc"`
	ID      any         `json:"id"`
	Result  any         `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
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
					"description": "Render a Mermaid sequence diagram or flowchart as a terminal-native, width-adaptive Unicode visualization. " +
						"Use this tool when explaining code execution, architecture, controller/reconciler flows, distributed systems, or Kubernetes lifecycle. " +
						"Returns cleanly formatted terminal art that fits the user's terminal width.",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"source": map[string]any{
								"type":        "string",
								"description": "Mermaid diagram definition (sequenceDiagram or flowchart TD/LR).",
							},
							"width": map[string]any{
								"type":        "integer",
								"description": "Target terminal width in columns. Optional; defaults to auto-detected terminal width.",
							},
							"mode": map[string]any{
								"type":        "string",
								"enum":        []string{"auto", "full", "compact", "narrow"},
								"description": "Rendering mode. Defaults to 'auto'.",
							},
							"ascii": map[string]any{
								"type":        "boolean",
								"description": "If true, renders using pure ASCII characters instead of Unicode box-drawing.",
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

	if callParams.Name != "render_diagram" {
		s.sendError(req.ID, -32601, fmt.Sprintf("Unknown tool: %s", callParams.Name))
		return
	}

	var args ToolCallArgs
	if err := json.Unmarshal(callParams.Arguments, &args); err != nil {
		s.sendError(req.ID, -32602, fmt.Sprintf("Invalid tool arguments: %v", err))
		return
	}

	diag, err := mermaid.Parse(args.Source)
	if err != nil {
		s.sendResult(req.ID, map[string]any{
			"isError": true,
			"content": []map[string]any{
				{
					"type": "text",
					"text": fmt.Sprintf("Error parsing diagram: %v", err),
				},
			},
		})
		return
	}

	renderWidth := args.Width
	if renderWidth <= 0 {
		renderWidth = layout.DetectTerminalWidth()
	}

	rendered, err := renderer.Render(diag, renderer.RenderOptions{
		Width:     renderWidth,
		Mode:      layout.RenderMode(args.Mode),
		NoColor:   true, // Agent tool calls prefer clean text without escape artifacts
		ASCIIOnly: args.ASCII,
	})
	if err != nil {
		s.sendResult(req.ID, map[string]any{
			"isError": true,
			"content": []map[string]any{
				{
					"type": "text",
					"text": fmt.Sprintf("Error rendering diagram: %v", err),
				},
			},
		})
		return
	}

	s.sendResult(req.ID, map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": rendered,
			},
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
