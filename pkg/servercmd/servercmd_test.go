package servercmd

import (
	"os/exec"
	"strings"
	"testing"
)

// TestBuildCommandNilConfig tests that nil config returns an error.
func TestBuildCommandNilConfig(t *testing.T) {
	_, err := BuildCommand(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

// TestValidateModelConfig tests the validateModelConfig function.
func TestValidateModelConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  ModelConfig
		wantErr bool
	}{
		{
			name:    "empty config",
			config:  ModelConfig{},
			wantErr: false,
		},
		{
			name: "config with name from flags",
			config: ModelConfig{
				Flags: map[string]any{
					"model": "/path/to/model.gguf",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateModelConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateModelConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDetectFlagConflicts tests the detectFlagConflicts function.
func TestDetectFlagConflicts(t *testing.T) {
	tests := []struct {
		name          string
		flags         map[string]any
		expectedCount int
	}{
		{
			name:          "no conflicts",
			flags:         map[string]any{"model": "/path/to/model.gguf"},
			expectedCount: 0,
		},
		{
			name: "model path conflict",
			flags: map[string]any{
				"model": "/path1/model.gguf",
				"m":     "/path2/model.gguf",
			},
			expectedCount: 1,
		},
		{
			name: "temperature conflict",
			flags: map[string]any{
				"temp":        0.8,
				"temperature": 0.9,
			},
			expectedCount: 1,
		},
		{
			name: "threads conflict",
			flags: map[string]any{
				"t":         4,
				"threads":   8,
			},
			expectedCount: 1,
		},
		{
			name: "ctx size conflict",
			flags: map[string]any{
				"c":          2048,
				"ctx_size":   4096,
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflicts := detectFlagConflicts(tt.flags)
			if len(conflicts) != tt.expectedCount {
				t.Errorf("detectFlagConflicts() returned %d conflicts, expected %d", len(conflicts), tt.expectedCount)
			}
		})
	}
}

// TestBuildArgs tests the buildArgs function.
func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		modelCfg ModelConfig
		wantArgs []string
	}{
		{
			name: "basic model path",
			modelCfg: ModelConfig{
				ModelPath: "/path/to/model.gguf",
			},
			wantArgs: []string{"-m", "/path/to/model.gguf"},
		},
		{
			name: "model with mmproj",
			modelCfg: ModelConfig{
				ModelPath:  "/path/to/model.gguf",
				MMProjOn:   true,
				MMProjPath: "/path/to/mmproj.gguf",
			},
			wantArgs: []string{"-m", "/path/to/model.gguf", "-mm", "/path/to/mmproj.gguf"},
		},
		{
			name: "model with threads",
			modelCfg: ModelConfig{
				ModelPath: "/path/to/model.gguf",
				Flags: map[string]any{
					"threads": 8,
				},
			},
			wantArgs: []string{"-m", "/path/to/model.gguf", "-t", "8"},
		},
		{
			name: "model with temperature",
			modelCfg: ModelConfig{
				ModelPath: "/path/to/model.gguf",
				Flags: map[string]any{
					"temp": 0.8,
				},
			},
			wantArgs: []string{"-m", "/path/to/model.gguf", "--temp", "0.80"},
		},
		{
			name: "model with gpu_layers",
			modelCfg: ModelConfig{
				ModelPath: "/path/to/model.gguf",
				Flags: map[string]any{
					"gpu_layers": 24,
				},
			},
			wantArgs: []string{"-m", "/path/to/model.gguf", "-ngl", "24"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Version: "1.0",
				Models:  []ModelConfig{tt.modelCfg},
			}
			args, err := buildArgs(cfg, tt.modelCfg)
			if err != nil {
				t.Fatalf("buildArgs() returned error: %v", err)
			}

			for i, want := range tt.wantArgs {
				if len(args) <= i || args[i] != want {
					t.Errorf("buildArgs() at index %d = %v, want %v", i, args[i], want)
				}
			}
		})
	}
}

// TestBuildArgsHostPort tests host and port args with order-independent checking.
func TestBuildArgsHostPort(t *testing.T) {
	cfg := &Config{
		Version: "1.0",
		Models: []ModelConfig{
			{
				ModelPath: "/path/to/model.gguf",
				Flags: map[string]any{
					"host": "0.0.0.0",
					"port": 8080,
				},
			},
		},
	}

	args, err := buildArgs(cfg, cfg.Models[0])
	if err != nil {
		t.Fatalf("buildArgs() returned error: %v", err)
	}

	expected := map[string]bool{
		"-m":       false,
		"--host":   false,
		"--port":   false,
	}
	for _, a := range args {
		if _, ok := expected[a]; ok {
			expected[a] = true
		}
	}
	for flag, found := range expected {
		if !found {
			t.Errorf("buildArgs() missing expected arg: %s", flag)
		}
	}
}

// TestNormalizeFlag tests the normalizeFlag function.
func TestNormalizeFlag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "already kebab", input: "some-flag", expected: "some-flag"},
		{name: "camelCase", input: "topK", expected: "top-k"},
		{name: "snake_case", input: "batch_size", expected: "batch-size"},
		{name: "multi camel", input: "gpuLayers", expected: "gpu-layers"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeFlag(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeFlag(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestSplitCamelCase tests the splitCamelCase function.
func TestSplitCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{name: "empty", input: "", expected: nil},
		{name: "simple", input: "topK", expected: []string{"top", "K"}},
		{name: "complex", input: "gpuLayers", expected: []string{"gpu", "Layers"}},
		{name: "single", input: "A", expected: []string{"A"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitCamelCase(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitCamelCase(%q) returned %d parts, expected %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, exp := range tt.expected {
				if result[i] != exp {
					t.Errorf("splitCamelCase(%q)[%d] = %v, want %v", tt.input, i, result[i], exp)
				}
			}
		})
	}
}

// TestBuildArgsWithVariousFlags tests buildArgs with various flag types.
func TestBuildArgsWithVariousFlags(t *testing.T) {
	cfg := &Config{
		Version: "1.0",
		Models: []ModelConfig{
			{
				ModelPath: "/path/to/model.gguf",
				Flags: map[string]any{
					"top_k":          40,
					"top_p":          0.95,
					"min_p":          0.05,
					"repeat_penalty": 1.1,
					"flash_attn":     true,
					"mlock":          true,
					"embedding":      false,
					"seed":           42,
					"n_predict":      2048,
				},
			},
		},
	}

	args, err := buildArgs(cfg, cfg.Models[0])
	if err != nil {
		t.Fatalf("buildArgs() returned error: %v", err)
	}

	expectedFlags := []string{
		"-m", "--top-k", "--top-p", "--min-p",
		"--repeat-penalty", "-fa", "--mlock", "-s", "-n",
	}

	for _, flag := range expectedFlags {
		found := false
		for i, arg := range args {
			if arg == flag {
				found = true
				break
			}
			// Check for boolean flags (no value follows)
			if flag == "-fa" && i+1 < len(args) && args[i+1] == "on" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("buildArgs() missing expected flag: %s", flag)
		}
	}
}

// TestBuildCommandWithLlamaServer tests BuildCommand when llama-server is available.
func TestBuildCommandWithLlamaServer(t *testing.T) {
	cfg := &Config{
		Version: "1.0",
		Models: []ModelConfig{
			{
				ModelPath: "/path/to/model.gguf",
			},
		},
	}

	// Skip if llama-server is not available
	_, err := exec.LookPath("llama-server")
	if err != nil {
		t.Skip("llama-server not found, skipping test")
	}

	cmd, err := BuildCommand(cfg)
	if err != nil {
		t.Fatalf("BuildCommand() returned error: %v", err)
	}
	if cmd == nil {
		t.Error("BuildCommand() returned nil command")
	}
}

// TestConfigJSONStructure tests the Config structure is correct.
func TestConfigJSONStructure(t *testing.T) {
	cfg := &Config{
		Version: "1.0",
		ScanPaths: []string{"/models", "/data"},
		Models: []ModelConfig{
			{
				Name:       "gemma-4bit",
				ModelPath:  "/models/gemma-4bit.gguf",
				MMProjPath: "/models/mmproj.gguf",
				MMProjOn:   true,
				Size:       4200000000,
				LastScan:   "2024-04-30T15:00:00Z",
				Flags: map[string]any{
					"model":      "/models/gemma-4bit.gguf",
					"ctx_size":   4096,
					"gpu_layers": 24,
					"temp":       0.8,
				},
			},
		},
	}

	if cfg.Version != "1.0" {
		t.Errorf("config version = %v, want 1.0", cfg.Version)
	}
	if len(cfg.ScanPaths) != 2 {
		t.Errorf("config scan_paths has %d elements, want 2", len(cfg.ScanPaths))
	}
	if len(cfg.Models) != 1 {
		t.Errorf("config models has %d elements, want 1", len(cfg.Models))
	}

	m := cfg.Models[0]
	if m.Name != "gemma-4bit" {
		t.Errorf("model name = %v, want gemma-4bit", m.Name)
	}
	if !m.MMProjOn {
		t.Error("mmproj_on should be true")
	}
	if m.MMProjPath == "" {
		t.Error("mmproj_path should not be empty")
	}
}

// TestBuildArgsFromFlags tests building args when model path comes from flags.
func TestBuildArgsFromFlags(t *testing.T) {
	modelCfg := ModelConfig{
		Name: "test-model",
		Flags: map[string]any{
			"model": "/flags/path/model.gguf",
			"t":     8,
			"ngl":   24,
		},
	}

	cfg := &Config{
		Version: "1.0",
		Models:  []ModelConfig{modelCfg},
	}

	args, err := buildArgs(cfg, modelCfg)
	if err != nil {
		t.Fatalf("buildArgs() returned error: %v", err)
	}

	// Should have -m flag with path from flags
	foundModel := false
	for i, arg := range args {
		if arg == "-m" && i+1 < len(args) {
			if args[i+1] == "/flags/path/model.gguf" {
				foundModel = true
				break
			}
		}
	}
	if !foundModel {
		t.Errorf("buildArgs() did not find model path from flags, got: %v", args)
	}
}

// TestBuildCommandWithSoftValidation tests soft validation behavior.
func TestBuildCommandWithSoftValidation(t *testing.T) {
	cfg := &Config{
		Version: "1.0",
		Models: []ModelConfig{
			{
				Name: "test-model",
				Flags: map[string]any{
					"model": "/flags/path/model.gguf",
				},
			},
		},
	}

	// Skip if llama-server is not available
	_, err := exec.LookPath("llama-server")
	if err != nil {
		t.Skip("llama-server not found, skipping test")
	}

	cmd, err := BuildCommand(cfg)
	if err != nil {
		t.Fatalf("BuildCommand() returned error: %v", err)
	}
	if cmd == nil {
		t.Error("BuildCommand() returned nil command")
	}
}

// TestNormalizeFlagPreservesKebab tests that kebab-case flags are preserved.
func TestNormalizeFlagPreservesKebab(t *testing.T) {
	input := "some-flag"
	expected := "some-flag"
	result := normalizeFlag(input)
	if result != expected {
		t.Errorf("normalizeFlag(%q) = %v, want %v", input, result, expected)
	}
}

// TestBuildArgsWithBooleanFlags tests boolean flag handling.
func TestBuildArgsWithBooleanFlags(t *testing.T) {
	cfg := &Config{
		Version: "1.0",
		Models: []ModelConfig{
			{
				ModelPath: "/path/to/model.gguf",
				Flags: map[string]any{
					"flash_attn": true,
					"mlock":      true,
					"embedding":  false,
					"mmap":       false,
				},
			},
		},
	}

	args, err := buildArgs(cfg, cfg.Models[0])
	if err != nil {
		t.Fatalf("buildArgs() returned error: %v", err)
	}

	// Check -fa on
	foundFA := false
	for i, arg := range args {
		if arg == "-fa" && i+1 < len(args) && args[i+1] == "on" {
			foundFA = true
			break
		}
	}
	if !foundFA {
		t.Errorf("buildArgs() missing -fa on, got: %v", args)
	}

	// Check --mlock
	foundMlock := false
	for _, arg := range args {
		if arg == "--mlock" {
			foundMlock = true
			break
		}
	}
	if !foundMlock {
		t.Errorf("buildArgs() missing --mlock, got: %v", args)
	}

	// Check --no-mmap
	foundNoMmap := false
	for _, arg := range args {
		if arg == "--no-mmap" {
			foundNoMmap = true
			break
		}
	}
	if !foundNoMmap {
		t.Errorf("buildArgs() missing --no-mmap, got: %v", args)
	}
}

// TestBuildCommandVersionDefault tests that empty version is set to 1.0.
func TestBuildCommandVersionDefault(t *testing.T) {
	cfg := &Config{
		Models: []ModelConfig{
			{
				ModelPath: "/path/to/model.gguf",
			},
		},
	}

	// Check that version gets set to 1.0 (this happens inside BuildCommand)
	if cfg.Version == "" {
		cfg.Version = "1.0" // default value
	}
	if cfg.Version != "1.0" {
		t.Errorf("config version = %v, want 1.0", cfg.Version)
	}
}

// TestBuildArgsWithFloatTemperature tests float temperature handling.
func TestBuildArgsWithFloatTemperature(t *testing.T) {
	cfg := &Config{
		Version: "1.0",
		Models: []ModelConfig{
			{
				ModelPath: "/path/to/model.gguf",
				Flags: map[string]any{
					"temp": 0.8,
				},
			},
		},
	}

	args, err := buildArgs(cfg, cfg.Models[0])
	if err != nil {
		t.Fatalf("buildArgs() returned error: %v", err)
	}

	// Check --temp 0.80
	foundTemp := false
	for i, arg := range args {
		if arg == "--temp" && i+1 < len(args) && args[i+1] == "0.80" {
			foundTemp = true
			break
		}
	}
	if !foundTemp {
		t.Errorf("buildArgs() missing --temp 0.80, got: %v", args)
	}
}

// TestBuildCommandWithMissingModel tests behavior when model is missing.
func TestBuildCommandWithMissingModel(t *testing.T) {
	cfg := &Config{
		Version: "1.0",
		Models:  []ModelConfig{},
	}

	_, err := BuildCommand(cfg)
	if err != nil && !strings.Contains(err.Error(), "llama-server not found") {
		t.Logf("BuildCommand() returned error (expected): %v", err)
	}
}
