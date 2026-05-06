package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"llama-server-loader/pkg/servercmd"
)

// Server represents the embedded HTTP server with JSON-RPC 2.0 API.
type Server struct {
	addr      string
	server    *http.Server
	logger    *log.Logger
	models    []ModelInfo
	status    string
	shutdown  chan struct{}
	cmd       *exec.Cmd        // Process for llama-server
	cmdCtx    context.Context  // Context for command cancellation
	cmdCancel context.CancelFunc
}

// ModelInfo contains information about a model available for serving.
type ModelInfo struct {
	Name       string                 `json:"name"`
	Path       string                 `json:"path"`
	IsMMProj   bool                   `json:"is_mmproj"`
	Size       int64                  `json:"size"`
	MMProjPath string                 `json:"mmproj_path,omitempty"`
	Flags      map[string]interface{} `json:"flags,omitempty"`
}

// RPCRequest represents a JSON-RPC 2.0 request.
type RPCRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
	ID     *int            `json:"id"`
}

// RPCResponse represents a JSON-RPC 2.0 response.
type RPCResponse struct {
	ID     int                    `json:"id"`
	Result interface{}            `json:"result"`
	Error  interface{}            `json:"error,omitempty"`
}

// RPCErrorResponse represents an error response.
type RPCErrorResponse struct {
	ID     int      `json:"id"`
	Code   int      `json:"code"`
	Message string   `json:"message"`
}

// GetModelsResult is the result of a getModels RPC call.
type GetModelsResult struct {
	Models []ModelInfo `json:"models"`
	Count  int         `json:"count"`
}

// StartServerParams contains parameters for starting the server.
type StartServerParams struct {
	ModelPath   string    `json:"model_path"`
	MMProjPath  string    `json:"mmproj_path,omitempty"`
	Threads     int       `json:"threads"`
	Temperature float64   `json:"temperature"`
	Port        int       `json:"port"`
}

// StatusResult contains the current server status.
type StatusResult struct {
	Status string `json:"status"`
	Addr   string `json:"addr,omitempty"`
}

// NewServer creates a new WebUI server instance.
func NewServer(addr string) *Server {
	return &Server{
		addr:     addr,
		logger:   log.New(log.Writer(), "webui: ", 0),
		status:   "idle",
		shutdown: make(chan struct{}),
	}
}

// SetModels sets the available models for the server.
func (s *Server) SetModels(models []ModelInfo) {
	s.models = models
}

// Start starts the HTTP server with JSON-RPC 2.0 API.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Handle JSON-RPC endpoint
	mux.HandleFunc("/rpc", s.handleRPC)

	// Serve static files
	mux.HandleFunc("/", s.handleStatic)

	s.server = &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			s.logger.Printf("server shutdown error: %v", err)
		}
	}()

	s.status = "running"
	s.logger.Printf("WebUI server starting on %s", s.addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.status = "shutting_down"
	// Also shutdown llama-server if running
	s.shutdownLlamaServer()

	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// shutdownLlamaServer gracefully stops the llama-server process.
func (s *Server) shutdownLlamaServer() {
	if s.cmdCancel != nil {
		s.cmdCancel()
	}
	if s.cmd != nil && s.cmd.Process != nil {
		s.logger.Printf("Sending interrupt to llama-server process")
		s.cmd.Process.Signal(os.Interrupt)
		// Wait with timeout
		done := make(chan error, 1)
		go func() {
			done <- s.cmd.Wait()
		}()
		select {
		case <-done:
			s.logger.Printf("llama-server stopped")
		case <-time.After(5 * time.Second):
			s.logger.Printf("Force killing llama-server")
			s.cmd.Process.Kill()
		}
	}
}

// handleRPC processes JSON-RPC 2.0 requests.
func (s *Server) handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSONRPCError(w, RPCErrorResponse{
			Code:   -32601,
			Message: "Method not found",
		})
		return
	}

	var req RPCRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		s.writeJSONRPCError(w, RPCErrorResponse{
			Code:   -32700,
			Message: "Parse error",
		})
		return
	}

	var resp interface{}
	switch req.Method {
	case "getModels":
		resp = s.handleGetModels()
	case "startServer":
		resp = s.handleStartServer(req.Params)
	case "getStatus":
		resp = s.handleGetStatus()
	case "shutdown":
		resp = s.handleShutdown()
	default:
		s.writeJSONRPCError(w, RPCErrorResponse{
			Code:   -32601,
			Message: fmt.Sprintf("Unknown method: %s", req.Method),
		})
		return
	}

	id := 0
	if req.ID != nil {
		id = *req.ID
	}

	s.writeJSONRPCResponse(w, RPCResponse{
		ID:     id,
		Result: resp,
	})
}

// handleGetModels returns the list of available models.
func (s *Server) handleGetModels() interface{} {
	return GetModelsResult{
		Models: s.models,
		Count:  len(s.models),
	}
}

// handleStartServer starts the llama-server with given parameters using pkg/servercmd.
func (s *Server) handleStartServer(params json.RawMessage) interface{} {
	if len(params) == 0 {
		return map[string]string{"message": "No parameters provided"}
	}

	var p StartServerParams
	if err := json.Unmarshal(params, &p); err != nil {
		return map[string]string{
			"error": fmt.Sprintf("Invalid params: %v", err),
		}
	}

	s.status = "starting"
	s.logger.Printf("Starting server for model: %s", p.ModelPath)

	// Build config for servercmd
	cfg := &servercmd.Config{
		Version: "1.0",
		Models: []servercmd.ModelConfig{
			{
				Name:       filepath.Base(strings.TrimSuffix(p.ModelPath, ".gguf")),
				ModelPath:  p.ModelPath,
				MMProjPath: p.MMProjPath,
				MMProjOn:   p.MMProjPath != "",
				Size:       0,
				Flags: map[string]any{
					"threads":     p.Threads,
					"temperature": p.Temperature,
					"port":        p.Port,
				},
			},
		},
	}

	// Build command using servercmd.BuildCommand
	cmd, err := servercmd.BuildCommand(cfg)
	if err != nil {
		s.status = "error"
		return map[string]string{
			"error": fmt.Sprintf("Failed to build command: %v", err),
		}
	}

	// Start the command with context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	s.cmdCtx = ctx
	s.cmdCancel = cancel

	if err := cmd.Start(); err != nil {
		cancel()
		s.status = "error"
		return map[string]string{
			"error": fmt.Sprintf("Failed to start server: %v", err),
		}
	}

	// Store command for shutdown
	s.cmd = cmd

	// Update status
	if p.Port > 0 {
		s.status = fmt.Sprintf("running on port %d", p.Port)
	} else {
		s.status = "running"
	}

	go func() {
		cmd.Wait()
		if s.status == "running" || strings.HasPrefix(s.status, "running") {
			s.logger.Printf("llama-server process exited")
		}
	}()

	return map[string]string{
		"message": "Server starting...",
		"status":  s.status,
	}
}

// handleGetStatus returns the current server status.
func (s *Server) handleGetStatus() interface{} {
	return StatusResult{
		Status: s.status,
		Addr:   s.addr,
	}
}

// handleShutdown initiates graceful shutdown of llama-server.
func (s *Server) handleShutdown() interface{} {
	s.shutdownLlamaServer()
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.status = "shutting_down"
	}()
	return map[string]string{
		"message": "Shutdown initiated",
	}
}

// handleStatic serves static files.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/index.html" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
		return
	}

	// Try to serve static files from embedded FS
	path := r.URL.Path
	if path != "" && path[0] == '/' {
		path = path[1:]
	}
	// Prepend "static/" for CSS, JS and other asset files
	if !strings.HasPrefix(path, "static/") && path != "index.html" {
		path = "static/" + path
	}

	data, err := staticFS.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch ext := path[len(path)-3:]; ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	default:
		// Try to detect MIME type from extension
	}

	w.Write(data)
}

// writeJSONRPCResponse writes a successful JSON-RPC response.
func (s *Server) writeJSONRPCResponse(w http.ResponseWriter, resp RPCResponse) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Printf("error writing response: %v", err)
	}
}

// writeJSONRPCError writes a JSON-RPC error response.
func (s *Server) writeJSONRPCError(w http.ResponseWriter, errResp RPCErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		s.logger.Printf("error writing error response: %v", err)
	}
}

// formatSize converts bytes to a human-readable string.
func formatSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	var exp int
	for bytes >= 1024 && exp < len(units)-1 {
		bytes /= 1024
		exp++
	}
	return fmt.Sprintf("%.2f %s", float64(bytes), units[exp])
}

// FormatModelName extracts a model name from its path.
func FormatModelName(path string) string {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	// Remove common suffixes like -v1, q4_k_m etc. for cleaner names
	name = strings.Split(name, "-")[0]
	return strings.ToLower(name)
}

// BuildFlagsString builds a flags string from a map of model configuration.
func BuildFlagsString(flags map[string]interface{}) string {
	var parts []string
	for k, v := range flags {
		switch val := v.(type) {
		case string:
			parts = append(parts, fmt.Sprintf("--%s=%s", k, val))
		case float64:
			parts = append(parts, fmt.Sprintf("--%s=%.0f", k, val))
		case bool:
			if val {
				parts = append(parts, fmt.Sprintf("--%s", k))
			}
		default:
			s := fmt.Sprintf("%v", val)
			num, err := strconv.ParseFloat(s, 64)
			if err == nil {
				parts = append(parts, fmt.Sprintf("--%s=%.0f", k, num))
			} else {
				parts = append(parts, fmt.Sprintf("--%s=%v", k, val))
			}
		}
	}
	return strings.Join(parts, " ")
}
