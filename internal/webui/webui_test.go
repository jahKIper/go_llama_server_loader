package webui

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name       string
		addr       string
		wantStatus string
	}{
		{
			name:       "default address",
			addr:       ":8080",
			wantStatus: "idle",
		},
		{
			name:       "custom address",
			addr:       ":9090",
			wantStatus: "idle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewServer(tt.addr)
			if srv == nil {
				t.Fatal("NewServer returned nil")
			}
			if srv.addr != tt.addr {
				t.Errorf("addr = %q, want %q", srv.addr, tt.addr)
			}
			if srv.status != tt.wantStatus {
				t.Errorf("status = %q, want %q", srv.status, tt.wantStatus)
			}
		})
	}
}

func TestServer_SetModels(t *testing.T) {
	srv := NewServer(":8080")

	models := []ModelInfo{
		{Name: "model1.gguf", Path: "/models/model1.gguf", Size: 4200000000},
		{Name: "model2.gguf", Path: "/models/model2.gguf", Size: 3500000000, IsMMProj: true},
	}

	srv.SetModels(models)

	if len(srv.models) != 2 {
		t.Errorf("got %d models, want 2", len(srv.models))
	}

	if srv.models[0].Name != "model1.gguf" {
		t.Errorf("models[0].Name = %q, want %q", srv.models[0].Name, "model1.gguf")
	}

	if srv.models[1].IsMMProj != true {
		t.Errorf("models[1].IsMMProj = %v, want true", srv.models[1].IsMMProj)
	}
}

func TestHandleGetModels(t *testing.T) {
	srv := NewServer(":8080")
	srv.SetModels([]ModelInfo{
		{Name: "test-model", Path: "/models/test.gguf", Size: 1000000},
	})

	result := srv.handleGetModels()

	res, ok := result.(GetModelsResult)
	if !ok {
		t.Fatalf("expected GetModelsResult, got %T", result)
	}

	if res.Count != 1 {
		t.Errorf("Count = %d, want 1", res.Count)
	}

	if len(res.Models) != 1 {
		t.Errorf("got %d models in result, want 1", len(res.Models))
	}
}

func TestHandleGetStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		wantAddr string
	}{
		{name: "idle status", status: "idle", wantAddr: ":8080"},
		{name: "running status", status: "running on port 8080", wantAddr: ":8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewServer(tt.wantAddr)
			srv.status = tt.status

			result := srv.handleGetStatus()
			res, ok := result.(StatusResult)
			if !ok {
				t.Fatalf("expected StatusResult, got %T", result)
			}

			if res.Status != tt.status {
				t.Errorf("status = %q, want %q", res.Status, tt.status)
			}

			if res.Addr != tt.wantAddr {
				t.Errorf("addr = %q, want %q", res.Addr, tt.wantAddr)
			}
		})
	}
}

func TestHandleStartServer(t *testing.T) {
	tests := []struct {
		name        string
		params      json.RawMessage
		containsMsg bool
	}{
		{
			name:        "no params",
			params:      nil,
			containsMsg: true,
		},
		{
			name: "valid params",
			params: json.RawMessage(`{"model_path":"/models/test.gguf","threads":4,"temperature":0.8,"port":8080}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewServer(":8080")
			result := srv.handleStartServer(tt.params)

			if res, ok := result.(map[string]string); ok {
				if !containsKey(res, "message") && !containsKey(res, "status") {
					t.Error("result should contain 'message' or 'status' key")
				}
			} else {
				t.Errorf("unexpected result type: %T", result)
			}
		})
	}
}

func containsKey(m map[string]string, key string) bool {
	_, exists := m[key]
	return exists
}

func TestHandleShutdown(t *testing.T) {
	srv := NewServer(":8080")
	result := srv.handleShutdown()

	res, ok := result.(map[string]string)
	if !ok {
		t.Fatalf("expected map[string]string, got %T", result)
	}

	if _, hasMsg := res["message"]; !hasMsg {
		t.Error("shutdown result should contain 'message' key")
	}
}

func TestHandleRPC_MethodNotAllowed(t *testing.T) {
	srv := NewServer(":8080")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/rpc", nil)

	srv.handleRPC(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("unexpected status code: %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleRPC_UnknownMethod(t *testing.T) {
	srv := NewServer(":8080")
	w := httptest.NewRecorder()

	body := []byte(`{"method":"unknownMethod","id":1}`)
	r := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	srv.handleRPC(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("unexpected status code: %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleRPC_GetModels(t *testing.T) {
	srv := NewServer(":8080")
	srv.SetModels([]ModelInfo{
		{Name: "test-model", Path: "/models/test.gguf", Size: 1000},
	})

	w := httptest.NewRecorder()
	body := []byte(`{"method":"getModels","id":1}`)
	r := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	srv.handleRPC(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("unexpected status code: %d, want %d", w.Code, http.StatusOK)
	}

	var resp RPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != nil {
		t.Errorf("unexpected error in response: %v", resp.Error)
	}

	if resp.ID != 1 {
		t.Errorf("response ID = %d, want 1", resp.ID)
	}
}

func TestHandleRPC_GetStatus(t *testing.T) {
	srv := NewServer(":8080")
	srv.status = "running"

	w := httptest.NewRecorder()
	body := []byte(`{"method":"getStatus","id":2}`)
	r := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	srv.handleRPC(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("unexpected status code: %d, want %d", w.Code, http.StatusOK)
	}

	var resp RPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != 2 {
		t.Errorf("response ID = %d, want 2", resp.ID)
	}
}

func TestHandleRPC_Shutdown(t *testing.T) {
	srv := NewServer(":8080")

	w := httptest.NewRecorder()
	body := []byte(`{"method":"shutdown","id":3}`)
	r := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	srv.handleRPC(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("unexpected status code: %d, want %d", w.Code, http.StatusOK)
	}

	var resp RPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != nil {
		t.Errorf("unexpected error in response: %v", resp.Error)
	}
}

func TestHandleRPC_ParseError(t *testing.T) {
	srv := NewServer(":8080")
	w := httptest.NewRecorder()

	body := []byte(`{"method":"getModels"`) // invalid JSON
	r := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	srv.handleRPC(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("unexpected status code: %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleStatic_IndexHTML(t *testing.T) {
	srv := NewServer(":8080")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	srv.handleStatic(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("unexpected status code: %d, want %d", w.Code, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", contentType, "text/html; charset=utf-8")
	}
}

func TestHandleStatic_NotFound(t *testing.T) {
	srv := NewServer(":8080")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/nonexistent.txt", nil)

	srv.handleStatic(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("unexpected status code: %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleStatic_CSS(t *testing.T) {
	srv := NewServer(":8080")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/static/style.css", nil)

	srv.handleStatic(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("unexpected status code: %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleStatic_JS(t *testing.T) {
	srv := NewServer(":8080")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/static/app.js", nil)

	srv.handleStatic(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("unexpected status code: %d, want %d", w.Code, http.StatusOK)
	}
}

func TestServer_StartAndShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	srv := NewServer(":0") // port 0 lets the OS pick a free port

	go func() {
		if err := srv.Start(ctx); err != nil && err != http.ErrServerClosed {
			t.Errorf("server error: %v", err)
		}
	}()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		t.Errorf("shutdown error: %v", err)
	}

	cancel()
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{name: "zero bytes", size: 0, want: "0 B"},
		{name: "bytes", size: 512, want: "512.00 B"},
		{name: "kilobytes", size: 1024, want: "1.00 KB"},
		{name: "megabytes", size: 1048576, want: "1.00 MB"},
		{name: "gigabytes", size: 4294967296, want: "4.00 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSize(tt.size)
			if got != tt.want {
				t.Errorf("formatSize(%d) = %q, want %q", tt.size, got, tt.want)
			}
		})
	}
}

func TestFormatModelName(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "simple path", path: "/models/model.gguf", want: "model"},
		{name: "windows path", path: `D:\progs\models\model-v1.gguf`, want: "model"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatModelName(tt.path)
			if got != tt.want {
				t.Errorf("FormatModelName(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestBuildFlagsString(t *testing.T) {
	flags := map[string]interface{}{
		"model":      "/models/test.gguf",
		"ctx_len":    float64(4096),
		"gpu_layers": float64(24),
	}

	result := BuildFlagsString(flags)

	if result == "" {
		t.Error("BuildFlagsString should not return empty string")
	}

	// Check that all flags are present
	expectedFlags := []string{"--model=/models/test.gguf", "--ctx_len=4096", "--gpu_layers=24"}
	for _, expected := range expectedFlags {
		if !strings.Contains(result, expected) {
			t.Errorf("BuildFlagsString result %q should contain %q", result, expected)
		}
	}
}
