package mcp_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kavix/agent-diagram/internal/mcp"
)

func TestMCPServer_EndToEnd(t *testing.T) {
	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"render_diagram","arguments":{"source":"sequenceDiagram\nparticipant A\nparticipant B\nA->>B: Hello\nB-->>A: Hi","width":80,"mode":"compact"}}}`,
	}

	inBuf := bytes.NewBufferString(strings.Join(requests, "\n") + "\n")
	outBuf := &bytes.Buffer{}

	server := mcp.NewServer(inBuf, outBuf)
	if err := server.Serve(); err != nil {
		t.Fatalf("server error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 responses, got %d: %s", len(lines), outBuf.String())
	}

	// 1. Check initialize response
	var initResp mcp.JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[0]), &initResp); err != nil {
		t.Fatalf("failed unmarshaling init response: %v", err)
	}
	if initResp.Error != nil {
		t.Fatalf("unexpected init error: %+v", initResp.Error)
	}

	// 2. Check tools/list response
	var listResp mcp.JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[1]), &listResp); err != nil {
		t.Fatalf("failed unmarshaling tools/list response: %v", err)
	}
	if !strings.Contains(lines[1], "render_diagram") {
		t.Fatalf("expected render_diagram in tools/list: %s", lines[1])
	}

	// 3. Check tools/call response
	var callResp mcp.JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[2]), &callResp); err != nil {
		t.Fatalf("failed unmarshaling tools/call response: %v", err)
	}
	if !strings.Contains(lines[2], "Hello") || !strings.Contains(lines[2], "Hi") {
		t.Fatalf("expected rendered diagram output in tool call result: %s", lines[2])
	}
}
