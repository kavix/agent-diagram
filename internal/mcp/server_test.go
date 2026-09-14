package mcp_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kavix/agent-diagram/internal/mcp"
)

// helper: send a batch of newline-separated JSON-RPC requests and return response lines.
func runRequests(t *testing.T, requests []string) []string {
	t.Helper()
	inBuf := bytes.NewBufferString(strings.Join(requests, "\n") + "\n")
	outBuf := &bytes.Buffer{}
	server := mcp.NewServer(inBuf, outBuf)
	if err := server.Serve(); err != nil {
		t.Fatalf("server error: %v", err)
	}
	return strings.Split(strings.TrimSpace(outBuf.String()), "\n")
}

func TestMCPServer_EndToEnd(t *testing.T) {
	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		// Use response=full to keep the existing assertion that "Hello" appears
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"render_diagram","arguments":{"source":"sequenceDiagram\nparticipant A\nparticipant B\nA->>B: Hello\nB-->>A: Hi","width":80,"mode":"compact","response":"full"}}}`,
	}

	lines := runRequests(t, requests)
	if len(lines) != 3 {
		t.Fatalf("expected 3 responses, got %d: %s", len(lines), strings.Join(lines, "\n"))
	}

	// 1. initialize
	var initResp mcp.JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[0]), &initResp); err != nil {
		t.Fatalf("failed unmarshaling init response: %v", err)
	}
	if initResp.Error != nil {
		t.Fatalf("unexpected init error: %+v", initResp.Error)
	}

	// 2. tools/list — must contain both render_diagram and validate_diagram
	if !strings.Contains(lines[1], "render_diagram") {
		t.Fatalf("expected render_diagram in tools/list: %s", lines[1])
	}
	if !strings.Contains(lines[1], "validate_diagram") {
		t.Fatalf("expected validate_diagram in tools/list: %s", lines[1])
	}

	// 3. tools/call render_diagram (full response mode) — diagram content present
	var callResp mcp.JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[2]), &callResp); err != nil {
		t.Fatalf("failed unmarshaling tools/call response: %v\nraw: %s", err, lines[2])
	}
	if !strings.Contains(lines[2], "Hello") || !strings.Contains(lines[2], "Hi") {
		t.Fatalf("expected rendered diagram in full-response call: %s", lines[2])
	}
}

// TestMCPServer_SummaryMode verifies that summary mode returns a short confirmation
// rather than the full rendered diagram, keeping LLM response tokens minimal.
func TestMCPServer_SummaryMode(t *testing.T) {
	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"render_diagram","arguments":{"source":"sequenceDiagram\nA->>B: Hello\nB-->>A: Hi","width":80,"response":"summary"}}}`,
	}

	lines := runRequests(t, requests)

	resp := lines[1]
	// Summary must contain the confirmation tick, not raw diagram art
	if !strings.Contains(resp, "✓") {
		t.Errorf("summary mode should return confirmation tick: %s", resp)
	}
	// Summary must say "Shown in terminal" — confirms the diagram went to terminal, not LLM context
	if !strings.Contains(resp, "Shown in terminal") {
		t.Errorf("summary mode must confirm diagram shown in terminal: %s", resp)
	}
	// Summary must NOT contain multi-line box art (many vertical pipes means full render leaked)
	lineCount := strings.Count(resp, "\\n")
	if lineCount > 5 {
		t.Errorf("summary mode returned too many lines (%d) — full diagram leaked into LLM response: %s", lineCount, resp)
	}
}

// TestMCPServer_ValidateTool verifies the validate_diagram tool returns structured
// error hints for invalid Mermaid source, enabling single-shot LLM self-correction.
func TestMCPServer_ValidateTool_InvalidSource(t *testing.T) {
	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"validate_diagram","arguments":{"source":"sequenceDiagram\nA B: No arrow here"}}}`,
	}

	lines := runRequests(t, requests)
	resp := lines[1]

	// Must contain a fix hint for the LLM
	if !strings.Contains(resp, "Fix") && !strings.Contains(resp, "fix") {
		t.Errorf("validate tool should return fix hints: %s", resp)
	}
}

// TestMCPServer_ValidateTool_ValidSource verifies that valid Mermaid passes validation.
func TestMCPServer_ValidateTool_ValidSource(t *testing.T) {
	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"validate_diagram","arguments":{"source":"sequenceDiagram\nA->>B: Hello"}}}`,
	}

	lines := runRequests(t, requests)
	resp := lines[1]

	if !strings.Contains(resp, "✓") && !strings.Contains(resp, "Valid") {
		t.Errorf("valid source should pass validation: %s", resp)
	}
}

// TestMCPServer_ValidationBeforeRender verifies that the render pipeline rejects
// invalid source before doing AST work, returning a structured error with fix hints.
func TestMCPServer_ValidationBeforeRender_RejectsInvalidSource(t *testing.T) {
	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"render_diagram","arguments":{"source":"not a valid diagram at all"}}}`,
	}

	lines := runRequests(t, requests)
	resp := lines[1]

	// Must be an error response, not a successful render
	if !strings.Contains(resp, "isError") && !strings.Contains(resp, "error") {
		t.Errorf("invalid source should return error: %s", resp)
	}
}

// TestMCPServer_ANSIStripping verifies that LLM responses never contain ANSI codes
// even when the renderer emits them (ANSI adds tokens and breaks LLM tokenization).
func TestMCPServer_ANSIStripping(t *testing.T) {
	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"render_diagram","arguments":{"source":"sequenceDiagram\nA->>B: test","response":"full"}}}`,
	}

	lines := runRequests(t, requests)
	resp := lines[1]

	if strings.Contains(resp, "\x1b[") {
		t.Errorf("LLM response must not contain ANSI escape codes: %s", resp)
	}
}

// TestTokenBudget_StripANSI verifies the ANSI stripping function.
func TestTokenBudget_StripANSI(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"\x1b[32mGreen\x1b[0m", "Green"},
		{"\x1b[1m\x1b[36mBold Cyan\x1b[0m", "Bold Cyan"},
		{"no escapes", "no escapes"},
		{"", ""},
	}
	for _, tc := range cases {
		got := mcp.StripANSI(tc.input)
		if got != tc.want {
			t.Errorf("StripANSI(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
