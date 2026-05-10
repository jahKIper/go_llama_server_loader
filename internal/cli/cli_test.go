package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		validate func(*testing.T, *Flags)
	}{
		{
			name:     "empty args",
			args:     []string{},
			wantErr:  false,
			validate: nil,
		},
		{
			name:   "scan-dir flag",
			args:   []string{"--scan-dir", "/models"},
			wantErr: false,
			validate: func(t *testing.T, f *Flags) {
				if f.ScanDir != "/models" {
					t.Errorf("expected scan-dir=/models, got %s", f.ScanDir)
				}
			},
		},
		{
			name:   "model flag",
			args:   []string{"--model", "gemma-4"},
			wantErr: false,
			validate: func(t *testing.T, f *Flags) {
				if f.Model != "gemma-4" {
					t.Errorf("expected model=gemma-4, got %s", f.Model)
				}
			},
		},
		{
			name:   "threads flag valid",
			args:   []string{"--threads", "16"},
			wantErr: false,
			validate: func(t *testing.T, f *Flags) {
				if f.Threads != 16 {
					t.Errorf("expected threads=16, got %d", f.Threads)
				}
			},
		},
		{
			name:      "threads flag invalid",
			args:      []string{"--threads", "abc"},
			wantErr:   true,
			validate:  nil,
		},
		{
			name:   "temperature flag valid",
			args:   []string{"--temperature", "0.8"},
			wantErr: false,
			validate: func(t *testing.T, f *Flags) {
				if f.Temperature != 0.8 {
					t.Errorf("expected temperature=0.8, got %f", f.Temperature)
				}
			},
		},
		{
			name:      "temperature flag invalid",
			args:      []string{"--temperature", "abc"},
			wantErr:   true,
			validate:  nil,
		},
		{
			name:    "start-webui flag",
			args:    []string{"--start-webui"},
			wantErr: false,
			validate: func(t *testing.T, f *Flags) {
				if !f.StartWebUI {
					t.Error("expected start-webui to be true")
				}
			},
		},
		{
			name:   "port flag",
			args:   []string{"--port", "9090"},
			wantErr: false,
			validate: func(t *testing.T, f *Flags) {
				if f.WebPort != 9090 {
					t.Errorf("expected port=9090, got %d", f.WebPort)
				}
			},
		},
		{
			name:   "save-config flag",
			args:   []string{"--save-config", "models.json"},
			wantErr: false,
			validate: func(t *testing.T, f *Flags) {
				if f.SaveConfig != "models.json" {
					t.Errorf("expected save-config=models.json, got %s", f.SaveConfig)
				}
			},
		},
		{
			name:   "generate-params flag",
			args:   []string{"--generate-params"},
			wantErr: false,
			validate: func(t *testing.T, f *Flags) {
				if !f.GenerateParams {
					t.Error("expected generate-params to be true")
				}
			},
		},
		{
			name:   "output flag",
			args:   []string{"--output", "custom.json"},
			wantErr: false,
			validate: func(t *testing.T, f *Flags) {
				if f.Output != "custom.json" {
					t.Errorf("expected output=custom.json, got %s", f.Output)
				}
			},
		},
		{
			name:      "missing value for scan-dir",
			args:      []string{"--scan-dir"},
			wantErr:   true,
			validate:  nil,
		},
		{
			name:      "unknown flag",
			args:      []string{"--unknown-flag"},
			wantErr:   true,
			validate:  nil,
		},
		{
			name:    "params-file flag space",
			args:    []string{"--params-file", "/data/params_ru.json"},
			wantErr: false,
			validate: func(t *testing.T, f *Flags) {
				if f.ParamsFile != "/data/params_ru.json" {
					t.Errorf("expected params-file=/data/params_ru.json, got %s", f.ParamsFile)
				}
			},
		},
		{
			name:    "params-file flag equals",
			args:    []string{"--params-file=/data/params_ru.json"},
			wantErr: false,
			validate: func(t *testing.T, f *Flags) {
				if f.ParamsFile != "/data/params_ru.json" {
					t.Errorf("expected params-file=/data/params_ru.json, got %s", f.ParamsFile)
				}
			},
		},
		{
			name:    "models-config flag space",
			args:    []string{"--models-config", "/data/models.json"},
			wantErr: false,
			validate: func(t *testing.T, f *Flags) {
				if f.ModelsConfig != "/data/models.json" {
					t.Errorf("expected models-config=/data/models.json, got %s", f.ModelsConfig)
				}
			},
		},
		{
			name:    "models-config flag equals",
			args:    []string{"--models-config=/data/models.json"},
			wantErr: false,
			validate: func(t *testing.T, f *Flags) {
				if f.ModelsConfig != "/data/models.json" {
					t.Errorf("expected models-config=/data/models.json, got %s", f.ModelsConfig)
				}
			},
		},
		{
			name:      "missing value for params-file",
			args:      []string{"--params-file"},
			wantErr:   true,
			validate:  nil,
		},
		{
			name:      "missing value for models-config",
			args:      []string{"--models-config"},
			wantErr:   true,
			validate:  nil,
		},
		{
			name:    "params-file and models-config together",
			args:    []string{"--params-file", "./params_ru.json", "--models-config", "./models.json"},
			wantErr: false,
			validate: func(t *testing.T, f *Flags) {
				if f.ParamsFile != "./params_ru.json" {
					t.Errorf("expected params-file=./params_ru.json, got %s", f.ParamsFile)
				}
				if f.ModelsConfig != "./models.json" {
					t.Errorf("expected models-config=./models.json, got %s", f.ModelsConfig)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, err := ParseFlags(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tt.validate != nil {
				tt.validate(t, flags)
			}
		})
	}
}

func TestNewCLI(t *testing.T) {
	tests := []struct {
		name               string
		flags              *Flags
		wantScanDir        string
		wantModel          string
		wantThreads        int
		wantTemp           float64
		wantWebPort        int
		wantParamsFile     string
		wantModelsConfig   string
	}{
		{
			name: "all flags set",
			flags: &Flags{
				ScanDir: "/models",
				Model: "gemma-4",
				Threads: 16,
				Temperature: 0.9,
				StartWebUI: true,
				WebPort: 9090,
			},
			wantScanDir: "/models",
			wantModel:   "gemma-4",
			wantThreads: 16,
			wantTemp:    0.9,
			wantWebPort: 9090,
		},
		{
			name: "default web port",
			flags: &Flags{
				StartWebUI: true,
			},
			wantWebPort: 8080,
		},
		{
			name: "params-file and models-config propagated",
			flags: &Flags{
				ParamsFile:   "/data/params_ru.json",
				ModelsConfig: "/data/models.json",
			},
			wantWebPort:      8080,
			wantParamsFile:   "/data/params_ru.json",
			wantModelsConfig: "/data/models.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCLI(tt.flags)
			if c.scanDir != tt.wantScanDir {
				t.Errorf("scanDir = %s, want %s", c.scanDir, tt.wantScanDir)
			}
			if c.modelName != tt.wantModel {
				t.Errorf("modelName = %s, want %s", c.modelName, tt.wantModel)
			}
			if c.threads != tt.wantThreads {
				t.Errorf("threads = %d, want %d", c.threads, tt.wantThreads)
			}
			if c.temperature != tt.wantTemp {
				t.Errorf("temperature = %.1f, want %.1f", c.temperature, tt.wantTemp)
			}
			if c.webPort != tt.wantWebPort {
				t.Errorf("webPort = %d, want %d", c.webPort, tt.wantWebPort)
			}
			if c.paramsFile != tt.wantParamsFile {
				t.Errorf("paramsFile = %s, want %s", c.paramsFile, tt.wantParamsFile)
			}
			if c.modelsConfigPath != tt.wantModelsConfig {
				t.Errorf("modelsConfigPath = %s, want %s", c.modelsConfigPath, tt.wantModelsConfig)
			}
		})
	}
}

func TestCLIValidate(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		cli       *CLI
		wantErr   bool
	}{
		{
			name: "valid scan dir",
			cli: &CLI{scanDir: tmpDir},
			wantErr: false,
		},
		{
			name: "invalid scan dir",
			cli: &CLI{scanDir: "/nonexistent/path"},
			wantErr: true,
		},
		{
			name: "file instead of directory",
			cli: &CLI{scanDir: filepath.Join(t.TempDir(), "file.txt")},
			wantErr: true,
		},
		{
			name: "valid threads",
			cli: &CLI{threads: 16},
			wantErr: false,
		},
		{
			name: "zero threads no model ok",
			cli: &CLI{scanDir: tmpDir, threads: 0},
			wantErr: false,
		},
		{
			name: "invalid threads zero with model",
			cli: &CLI{scanDir: tmpDir, modelName: "test-model", threads: 0},
			wantErr: true,
		},
		{
			name: "valid temperature low",
			cli: &CLI{scanDir: tmpDir, temperature: 0.1},
			wantErr: false,
		},
		{
			name: "valid temperature high",
			cli: &CLI{scanDir: tmpDir, temperature: 2.0},
			wantErr: false,
		},
		{
			name: "temperature too low",
			cli: &CLI{scanDir: tmpDir, temperature: -0.1},
			wantErr: true,
		},
		{
			name: "temperature too high",
			cli: &CLI{scanDir: tmpDir, temperature: 2.1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cli.validate()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{"bytes", 512, "512B"},
		{"kilobytes", 10240, "10.0KB"},
		{"megabytes", 10485760, "10.0MB"},
		{"gigabytes", 10737418240, "10.0GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.size)
			if result != tt.expected {
				t.Errorf("formatSize(%d) = %s, want %s", tt.size, result, tt.expected)
			}
		})
	}
}

func TestFormatModelName(t *testing.T) {
	tests := []struct {
		desc     string
		model    string
		path     string
		size     int64
		expected string
	}{
		{
			desc:     "with name",
			model:    "gemma-4",
			path:     "/models/gemma-4.gguf",
			size:     4291768320,
			expected: "gemma-4 (4.0GB)",
		},
		{
			desc:     "empty name uses path base",
			model:    "",
			path:     "/models/llama-2.gguf",
			size:     3221225472,
			expected: "llama-2.gguf (3.0GB)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := FormatModelName(tt.model, tt.path, tt.size)
			if result != tt.expected {
				t.Errorf("FormatModelName(%q, %q, %d) = %s, want %s",
					tt.model, tt.path, tt.size, result, tt.expected)
			}
		})
	}
}

func TestBuildFlagsString(t *testing.T) {
	tests := []struct {
		name        string
		scanDir     string
		modelName   string
		threads     int
		temperature float64
		expected    string
	}{
		{
			name:        "all flags",
			scanDir:     "/models",
			modelName:   "gemma-4",
			threads:     16,
			temperature: 0.9,
			expected:    "--scan-dir=/models --model=gemma-4 --threads=16 --temperature=0.90",
		},
		{
			name:        "empty scan dir",
			scanDir:     "",
			modelName:   "llama-2",
			threads:     8,
			temperature: 0.5,
			expected:    "--model=llama-2 --threads=8 --temperature=0.50",
		},
		{
			name:        "only model",
			scanDir:     "",
			modelName:   "test-model",
			threads:     0,
			temperature: 0,
			expected:    "--model=test-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildFlagsString(tt.scanDir, tt.modelName, tt.threads, tt.temperature)
			if result != tt.expected {
				t.Errorf("BuildFlagsString() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestGenerateParameters(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output", "params.json")

	c := &CLI{
		generateParams: true,
		output:         outputPath,
	}

	err := c.generateParameters()
	if err != nil {
		t.Errorf("generateParameters() returned error: %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Errorf("cannot read generated params file: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("generated file is not valid JSON: %v", err)
	}
	if parsed["output_file"] != outputPath {
		t.Errorf("output_file = %v, want %s", parsed["output_file"], outputPath)
	}
	if ts, ok := parsed["generated_at"].(string); !ok || ts == "" || ts == "TODO_TIMESTAMP" {
		t.Errorf("generated_at must be a real timestamp, got %v", parsed["generated_at"])
	}
	if threads, ok := parsed["default_threads"].(float64); !ok || threads < 1 {
		t.Errorf("default_threads must be >= 1 (auto-detected), got %v", parsed["default_threads"])
	}
	if temp, ok := parsed["default_temperature"].(float64); !ok || temp <= 0 {
		t.Errorf("default_temperature must be > 0, got %v", parsed["default_temperature"])
	}
}

func TestGenerateParametersDefaultOutput(t *testing.T) {
	tmpDir := t.TempDir()
	// Change working directory to tmpDir
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer func() {
		os.Chdir(oldWd)
	}()

	c := &CLI{
		generateParams: true,
		output:         "",
	}

	err := c.generateParameters()
	if err != nil {
		t.Errorf("generateParameters() returned error: %v", err)
	}

	// Verify default file was created
	content, err := os.ReadFile("generated_params.json")
	if err != nil {
		t.Errorf("cannot read default generated params file: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("default generated file is not valid JSON: %v", err)
	}
	if parsed["output_file"] != "generated_params.json" {
		t.Errorf("output_file = %v, want generated_params.json", parsed["output_file"])
	}
}

func TestCLIRun(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		cli       *CLI
		wantErr   bool
	}{
		{
			name: "empty run shows help",
			cli:  &CLI{},
			wantErr: false,
		},
		{
			name: "generate params only",
			cli: &CLI{
				generateParams: true,
				output:         filepath.Join(tmpDir, "params.json"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cli.Run()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDetectCPUCores(t *testing.T) {
	result := DetectCPUCores()
	if result < 1 {
		t.Errorf("DetectCPUCores() = %d, expected >= 1", result)
	}
}
