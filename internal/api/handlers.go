package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/relentlessworks/basekit/internal/model"
)

// Server holds the HTTP server state.
type Server struct {
	addr string
}

// NewServer creates a new API server.
func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	mux := s.routes()
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) routes() *http.ServeMux {
	mux := &http.ServeMux{}

	mux.HandleFunc("/help", s.handleHelp)
	mux.HandleFunc("/.well-known/agent.md", s.handleHelp)
	mux.HandleFunc("/encode", s.handleEncode)
	mux.HandleFunc("/decode", s.handleDecode)
	mux.HandleFunc("/convert", s.handleConvert)
	mux.HandleFunc("/detect", s.handleDetect)
	mux.HandleFunc("/list", s.handleList)
	mux.HandleFunc("/mcp", s.handleMCP)
	mux.HandleFunc("/", s.handleRoot)

	return mux
}

// wantsJSON checks if the client wants JSON output.
func wantsJSON(r *http.Request) bool {
	if r.URL.Query().Get("format") == "json" {
		return true
	}
	return strings.Contains(r.Header.Get("Accept"), "application/json")
}

// writeError writes an instructive error response.
func writeError(w http.ResponseWriter, status int, msg, hint string, r *http.Request) {
	w.WriteHeader(status)
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"error": msg,
			"hint":  hint,
		})
		return
	}
	fmt.Fprintf(w, "error: %s | hint: %s\n", msg, hint)
}

// writeRecord writes a single key=value record.
func writeRecord(w http.ResponseWriter, r *http.Request, fields map[string]string) {
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fields)
		return
	}
	var parts []string
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	fmt.Fprintln(w, strings.Join(parts, " "))
}

// writeRecords writes multiple records.
func writeRecords(w http.ResponseWriter, r *http.Request, records []map[string]string) {
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)
		return
	}
	for _, rec := range records {
		var parts []string
		for k, v := range rec {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
		fmt.Fprintln(w, strings.Join(parts, " "))
	}
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "unknown endpoint", "GET /help for available endpoints", r)
		return
	}
	writeRecord(w, r, map[string]string{
		"service": "basekit",
		"version": "0.1.0",
		"status":  "ok",
		"help":    "GET /help",
	})
}

func (s *Server) handleHelp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, helpText)
}

func (s *Server) handleEncode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "use POST with encoding and data fields", r)
		return
	}

	enc := r.URL.Query().Get("encoding")
	if enc == "" {
		enc = r.FormValue("encoding")
	}

	if enc == "" {
		writeError(w, http.StatusBadRequest, "missing encoding parameter", "specify encoding: base64, base64url, base32, base32hex, base58, base85, hex, url", r)
		return
	}

	var data []byte
	var err error

	// Check if data is in query param or body
	input := r.URL.Query().Get("data")
	if input == "" {
		input = r.FormValue("data")
	}

	if input != "" {
		data = []byte(input)
	} else {
		// Read raw body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read request body", "send data as form field 'data' or raw body", r)
			return
		}
		data = body
	}

	if len(data) == 0 {
		writeError(w, http.StatusBadRequest, "empty input data", "provide data to encode via 'data' field or request body", r)
		return
	}

	result, err := model.Encode(data, model.Encoding(enc))
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("encoding failed: %v", err), "check encoding name and input data", r)
		return
	}

	writeRecord(w, r, map[string]string{
		"encoding": enc,
		"input":    string(data),
		"output":   result,
	})
}

func (s *Server) handleDecode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "use POST with encoding and data fields", r)
		return
	}

	enc := r.URL.Query().Get("encoding")
	if enc == "" {
		enc = r.FormValue("encoding")
	}

	if enc == "" {
		writeError(w, http.StatusBadRequest, "missing encoding parameter", "specify encoding: base64, base64url, base32, base32hex, base58, base85, hex, url", r)
		return
	}

	input := r.URL.Query().Get("data")
	if input == "" {
		input = r.FormValue("data")
	}

	if input == "" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read request body", "send encoded data as form field 'data' or raw body", r)
			return
		}
		input = strings.TrimSpace(string(body))
	}

	if input == "" {
		writeError(w, http.StatusBadRequest, "empty input data", "provide encoded data to decode via 'data' field or request body", r)
		return
	}

	result, err := model.Decode(input, model.Encoding(enc))
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("decode failed: %v", err), "check encoding name and input data format", r)
		return
	}

	writeRecord(w, r, map[string]string{
		"encoding": enc,
		"input":    input,
		"output":   string(result),
	})
}

func (s *Server) handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "use POST with from, to, and data fields", r)
		return
	}

	from := r.URL.Query().Get("from")
	if from == "" {
		from = r.FormValue("from")
	}
	to := r.URL.Query().Get("to")
	if to == "" {
		to = r.FormValue("to")
	}

	if from == "" {
		writeError(w, http.StatusBadRequest, "missing 'from' parameter", "specify source encoding: base64, base64url, base32, base32hex, base58, base85, hex, url", r)
		return
	}
	if to == "" {
		writeError(w, http.StatusBadRequest, "missing 'to' parameter", "specify target encoding: base64, base64url, base32, base32hex, base58, base85, hex, url", r)
		return
	}

	input := r.URL.Query().Get("data")
	if input == "" {
		input = r.FormValue("data")
	}

	if input == "" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read request body", "send data as form field 'data' or raw body", r)
			return
		}
		input = strings.TrimSpace(string(body))
	}

	if input == "" {
		writeError(w, http.StatusBadRequest, "empty input data", "provide data to convert via 'data' field or request body", r)
		return
	}

	result, err := model.Convert(input, model.Encoding(from), model.Encoding(to))
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("convert failed: %v", err), "check 'from' and 'to' encoding names and input data format", r)
		return
	}

	writeRecord(w, r, map[string]string{
		"from":   from,
		"to":     to,
		"input":  input,
		"output": result,
	})
}

func (s *Server) handleDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "use POST with data field or GET ?data=...", r)
		return
	}

	input := r.URL.Query().Get("data")
	if input == "" {
		input = r.FormValue("data")
	}

	if input == "" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read request body", "send data as form field 'data' or raw body", r)
			return
		}
		input = strings.TrimSpace(string(body))
	}

	if input == "" {
		writeError(w, http.StatusBadRequest, "empty input data", "provide data to detect via 'data' field or request body", r)
		return
	}

	candidates := model.Detect(input)
	if len(candidates) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "no matching encoding detected", "the input does not match any known encoding pattern", r)
		return
	}

	var records []map[string]string
	for _, c := range candidates {
		records = append(records, map[string]string{
			"encoding": string(c),
		})
	}
	writeRecords(w, r, records)
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	encodings := model.AllEncodings()
	var records []map[string]string
	for _, e := range encodings {
		records = append(records, map[string]string{
			"encoding": string(e),
		})
	}
	writeRecords(w, r, records)
}

const helpText = `basekit — Agentic-first base encoding & decoding service
=========================================================

Encode and decode data using Base64, Base64URL, Base32, Base32Hex,
Base58 (Bitcoin), Base85 (ASCII85), Hex, and URL encoding.

ENDPOINTS
---------

GET  /                    Service info
GET  /help                This help text
GET  /.well-known/agent.md  Same as /help
GET  /list                List all supported encodings

POST /encode              Encode data
  Params: encoding=<enc>  data=<text>
  Body:   raw text (if no data param)
  Encodings: base64, base64url, base32, base32hex, base58, base85, hex, url

POST /decode              Decode data
  Params: encoding=<enc>  data=<encoded>
  Body:   encoded text (if no data param)

POST /convert             Convert between encodings
  Params: from=<enc>  to=<enc>  data=<encoded>

POST /detect              Detect possible encodings
  Params: data=<text>
  Body:   text (if no data param)

POST /mcp                 MCP JSON-RPC 2.0 endpoint

OUTPUT FORMAT
-------------
Plain text by default: one labeled line per record.
  encoding=base64 input=hello output=aGVsbG8=
JSON on demand: Accept: application/json or ?format=json

ERRORS
------
4xx responses include a hint:
  error: <message> | hint: <what to do next>

EXAMPLES
--------
  curl -X POST 'http://localhost:7700/encode?encoding=base64&data=hello'
  -> encoding=base64 input=hello output=aGVsbG8=

  curl -X POST 'http://localhost:7700/decode?encoding=base64&data=aGVsbG8='
  -> encoding=base64 input=aGVsbG8= output=hello

  curl -X POST 'http://localhost:7700/convert?from=hex&to=base64&data=68656c6c6f'
  -> from=hex to=base64 input=68656c6c6f output=aGVsbG8=

  curl -X POST 'http://localhost:7700/detect?data=aGVsbG8='
  -> encoding=base64

CONFIGURATION
-------------
  -addr string    Listen address (default ":7700")
  BASEKIT_ADDR    Env var for listen address
`
