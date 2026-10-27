package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/relentlessworks/basekit/internal/model"
)

// MCPRequest represents a JSON-RPC 2.0 request.
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response.
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError represents a JSON-RPC error.
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCPTool represents a tool definition.
type MCPTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCPToolCallParams represents tool call parameters.
type MCPToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "use POST for MCP JSON-RPC", r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeMCPError(w, nil, -32700, "parse error")
		return
	}

	var req MCPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeMCPError(w, nil, -32700, "parse error")
		return
	}

	switch req.Method {
	case "initialize":
		writeMCPResult(w, req.ID, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "basekit",
				"version": "0.1.0",
			},
		})

	case "tools/list":
		writeMCPResult(w, req.ID, map[string]interface{}{
			"tools": mcpTools(),
		})

	case "tools/call":
		var params MCPToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			writeMCPError(w, req.ID, -32602, "invalid params")
			return
		}
		result, err := s.handleMCPToolCall(params)
		if err != nil {
			writeMCPError(w, req.ID, -32603, err.Error())
			return
		}
		writeMCPResult(w, req.ID, map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": result,
				},
			},
		})

	default:
		writeMCPError(w, req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

func writeMCPResult(w http.ResponseWriter, id json.RawMessage, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func writeMCPError(w http.ResponseWriter, id json.RawMessage, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: msg,
		},
	})
}

func mcpTools() []MCPTool {
	stringType := "string"
	return []MCPTool{
		{
			Name:        "encode",
			Description: "Encode data using the specified encoding (base64, base64url, base32, base32hex, base58, base85, hex, url)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"encoding": map[string]interface{}{"type": stringType, "description": "Encoding to use"},
					"data":      map[string]interface{}{"type": stringType, "description": "Data to encode"},
				},
				"required": []string{"encoding", "data"},
			},
		},
		{
			Name:        "decode",
			Description: "Decode data using the specified encoding",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"encoding": map[string]interface{}{"type": stringType, "description": "Encoding to use"},
					"data":      map[string]interface{}{"type": stringType, "description": "Encoded data to decode"},
				},
				"required": []string{"encoding", "data"},
			},
		},
		{
			Name:        "convert",
			Description: "Convert data from one encoding to another",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"from": map[string]interface{}{"type": stringType, "description": "Source encoding"},
					"to":   map[string]interface{}{"type": stringType, "description": "Target encoding"},
					"data": map[string]interface{}{"type": stringType, "description": "Data to convert"},
				},
				"required": []string{"from", "to", "data"},
			},
		},
		{
			Name:        "detect",
			Description: "Detect possible encodings of a string",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"data": map[string]interface{}{"type": stringType, "description": "Data to analyze"},
				},
				"required": []string{"data"},
			},
		},
		{
			Name:        "list_encodings",
			Description: "List all supported encodings",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func (s *Server) handleMCPToolCall(params MCPToolCallParams) (string, error) {
	switch params.Name {
	case "encode":
		enc, _ := params.Arguments["encoding"].(string)
		data, _ := params.Arguments["data"].(string)
		if enc == "" || data == "" {
			return "", fmt.Errorf("encoding and data are required")
		}
		result, err := model.Encode([]byte(data), model.Encoding(enc))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("encoding=%s input=%s output=%s", enc, data, result), nil

	case "decode":
		enc, _ := params.Arguments["encoding"].(string)
		data, _ := params.Arguments["data"].(string)
		if enc == "" || data == "" {
			return "", fmt.Errorf("encoding and data are required")
		}
		result, err := model.Decode(data, model.Encoding(enc))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("encoding=%s input=%s output=%s", enc, data, string(result)), nil

	case "convert":
		from, _ := params.Arguments["from"].(string)
		to, _ := params.Arguments["to"].(string)
		data, _ := params.Arguments["data"].(string)
		if from == "" || to == "" || data == "" {
			return "", fmt.Errorf("from, to, and data are required")
		}
		result, err := model.Convert(data, model.Encoding(from), model.Encoding(to))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("from=%s to=%s input=%s output=%s", from, to, data, result), nil

	case "detect":
		data, _ := params.Arguments["data"].(string)
		if data == "" {
			return "", fmt.Errorf("data is required")
		}
		candidates := model.Detect(data)
		if len(candidates) == 0 {
			return "no matching encoding detected", nil
		}
		var parts []string
		for _, c := range candidates {
			parts = append(parts, string(c))
		}
		return fmt.Sprintf("detected encodings: %s", strings.Join(parts, ", ")), nil

	case "list_encodings":
		var parts []string
		for _, e := range model.AllEncodings() {
			parts = append(parts, string(e))
		}
		return fmt.Sprintf("supported encodings: %s", strings.Join(parts, ", ")), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", params.Name)
	}
}
