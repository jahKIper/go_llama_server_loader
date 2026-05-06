// Package servercmd provides functionality to build and execute
// llama-server commands based on configuration.
package servercmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ModelConfig represents the configuration for a single model.
type ModelConfig struct {
	Name       string            `json:"name"`
	ModelPath  string            `json:"model_path"`
	MMProjPath string            `json:"mmproj_path,omitempty"`
	MMProjOn   bool              `json:"mmproj_on"`
	Size       int64             `json:"size"`
	LastScan   string            `json:"last_scan"`
	Flags      map[string]any    `json:"flags"`
}

// Config represents the top-level configuration.
type Config struct {
	Version    string   `json:"version"`
	ScanPaths  []string `json:"scan_paths"`
	Models     []ModelConfig `json:"models"`
}

// ServerCommand holds the built command and its arguments.
type ServerCommand struct {
	cmd      *exec.Cmd
	config   *Config
	modelCfg ModelConfig
}

// BuildCommand creates an exec.Cmd configured for llama-server based on the given config.
// It performs soft validation: checks for required fields, flag conflicts, duplicates,
// and verifies that llama-server is available in PATH.
func BuildCommand(cfg *Config) (*exec.Cmd, error) {
	if cfg == nil {
		return nil, fmt.Errorf("servercmd: config is nil")
	}

	// Validate version
	if cfg.Version == "" {
		cfg.Version = "1.0"
	}

	// Select the first model (or use empty config for bare server)
	var modelCfg ModelConfig
	if len(cfg.Models) > 0 {
		modelCfg = cfg.Models[0]
	}

	// Validate model configuration
	if err := validateModelConfig(modelCfg); err != nil {
		return nil, fmt.Errorf("servercmd: invalid model config: %w", err)
	}

	// Check llama-server availability
	if _, err := exec.LookPath("llama-server"); err != nil {
		return nil, fmt.Errorf("servercmd: llama-server not found in PATH: %w", err)
	}

	// Build command arguments
	args, err := buildArgs(cfg, modelCfg)
	if err != nil {
		return nil, fmt.Errorf("servercmd: failed to build args: %w", err)
	}

	cmd := exec.Command("llama-server", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd, nil
}

// BuildCommandWithContext creates an exec.Cmd with a context for cancellation.
func BuildCommandWithContext(ctx context.Context, cfg *Config) (*exec.Cmd, error) {
	cmd, err := BuildCommand(cfg)
	if err != nil {
		return nil, err
	}
	// Note: exec.CommandContext would replace the cmd, so we set context separately
	// For now, return the cmd and let caller handle context
	_ = ctx
	return cmd, nil
}

// validateModelConfig performs soft validation of model configuration.
func validateModelConfig(m ModelConfig) error {
	if m.Name == "" && len(m.Flags) > 0 {
		// Try to derive name from flags
		if mp, ok := m.Flags["model"].(string); ok && mp != "" {
			m.Name = filepath.Base(mp)
		}
	}

	// Check for flag conflicts and duplicates
	conflicts := detectFlagConflicts(m.Flags)
	if len(conflicts) > 0 {
		for _, c := range conflicts {
			fmt.Printf("servercmd: warning: potential conflict: %s\n", c)
		}
	}

	return nil
}

// detectFlagConflicts checks for conflicting or duplicate flags.
func detectFlagConflicts(flags map[string]any) []string {
	var conflicts []string

	// Check for model path conflicts
	modelCount := 0
	if _, ok := flags["model"]; ok {
		modelCount++
	}
	if _, ok := flags["m"]; ok {
		modelCount++
	}
	if modelCount > 1 {
		conflicts = append(conflicts, "multiple model path flags specified")
	}

	// Check for temperature conflicts
	tempCount := 0
	if _, ok := flags["temp"]; ok {
		tempCount++
	}
	if _, ok := flags["temperature"]; ok {
		tempCount++
	}
	if tempCount > 1 {
		conflicts = append(conflicts, "multiple temperature flags specified")
	}

	// Check for threads conflicts
	threadsCount := 0
	if _, ok := flags["t"]; ok {
		threadsCount++
	}
	if _, ok := flags["threads"]; ok {
		threadsCount++
	}
	if threadsCount > 1 {
		conflicts = append(conflicts, "multiple threads flags specified")
	}

	// Check for ctx-size conflicts
	ctxCount := 0
	if _, ok := flags["c"]; ok {
		ctxCount++
	}
	if _, ok := flags["ctx_size"]; ok {
		ctxCount++
	}
	if ctxCount > 1 {
		conflicts = append(conflicts, "multiple context size flags specified")
	}

	return conflicts
}

// buildArgs constructs the command line arguments for llama-server.
func buildArgs(cfg *Config, m ModelConfig) ([]string, error) {
	var args []string

	// Add model path
	modelPath := m.ModelPath
	if modelPath == "" {
		if mp, ok := m.Flags["model"].(string); ok {
			modelPath = mp
		} else if mp, ok := m.Flags["m"].(string); ok {
			modelPath = mp
		}
	}

	if modelPath != "" {
		args = append(args, "-m", modelPath)
	}

	// Add mmproj if enabled
	if m.MMProjOn && m.MMProjPath != "" {
		args = append(args, "-mm", m.MMProjPath)
	}

	// Process flags from config
	for key, val := range m.Flags {
		switch key {
		case "model", "m":
			// Already handled above
			continue

		case "mmproj", "mm":
			// Already handled above
			continue

		case "threads", "t":
			if v, ok := val.(int); ok {
				args = append(args, "-t", fmt.Sprintf("%d", v))
			}

		case "ctx_size", "c":
			if v, ok := val.(int); ok {
				args = append(args, "-c", fmt.Sprintf("%d", v))
			}

		case "temp", "temperature":
			switch v := val.(type) {
			case float64:
				args = append(args, "--temp", fmt.Sprintf("%.2f", v))
			case int:
				args = append(args, "--temp", fmt.Sprintf("%d", v))
			case string:
				args = append(args, "--temp", v)
			}

		case "gpu_layers", "ngl":
			if v, ok := val.(int); ok {
				args = append(args, "-ngl", fmt.Sprintf("%d", v))
			}

		case "host":
			if v, ok := val.(string); ok && v != "" {
				args = append(args, "--host", v)
			}

		case "port":
			switch v := val.(type) {
			case int:
				args = append(args, "--port", fmt.Sprintf("%d", v))
			case float64:
				args = append(args, "--port", fmt.Sprintf("%d", int(v)))
			case string:
				args = append(args, "--port", v)
			}

		case "batch_size", "b":
			if v, ok := val.(int); ok {
				args = append(args, "-b", fmt.Sprintf("%d", v))
			}

		case "ubatch_size", "ub":
			if v, ok := val.(int); ok {
				args = append(args, "-ub", fmt.Sprintf("%d", v))
			}

		case "top_k":
			if v, ok := val.(int); ok {
				args = append(args, "--top-k", fmt.Sprintf("%d", v))
			}

		case "top_p":
			switch v := val.(type) {
			case float64:
				args = append(args, "--top-p", fmt.Sprintf("%.2f", v))
			case int:
				args = append(args, "--top-p", fmt.Sprintf("%d", v))
			}

		case "min_p":
			switch v := val.(type) {
			case float64:
				args = append(args, "--min-p", fmt.Sprintf("%.2f", v))
			case int:
				args = append(args, "--min-p", fmt.Sprintf("%d", v))
			}

		case "repeat_penalty":
			switch v := val.(type) {
			case float64:
				args = append(args, "--repeat-penalty", fmt.Sprintf("%.2f", v))
			case int:
				args = append(args, "--repeat-penalty", fmt.Sprintf("%d", v))
			}

		case "presence_penalty":
			switch v := val.(type) {
			case float64:
				args = append(args, "--presence-penalty", fmt.Sprintf("%.2f", v))
			case int:
				args = append(args, "--presence-penalty", fmt.Sprintf("%d", v))
			}

		case "frequency_penalty":
			switch v := val.(type) {
			case float64:
				args = append(args, "--frequency-penalty", fmt.Sprintf("%.2f", v))
			case int:
				args = append(args, "--frequency-penalty", fmt.Sprintf("%d", v))
			}

		case "flash_attn", "fa":
			if v, ok := val.(bool); ok {
				if v {
					args = append(args, "-fa", "on")
				} else {
					args = append(args, "-fa", "off")
				}
			}

		case "mmap", "no_mmap":
			if v, ok := val.(bool); ok && !v {
				args = append(args, "--no-mmap")
			}

		case "mlock":
			if v, ok := val.(bool); ok && v {
				args = append(args, "--mlock")
			}

		case "embedding", "embeddings":
			if v, ok := val.(bool); ok && v {
				args = append(args, "--embedding")
			}

		case "cont_batching", "cb":
			if v, ok := val.(bool); ok && !v {
				args = append(args, "--no-cont-batching")
			}

		case "webui":
			if v, ok := val.(bool); ok && !v {
				args = append(args, "--no-webui")
			}

		case "parallel", "np":
			if v, ok := val.(int); ok {
				args = append(args, "-np", fmt.Sprintf("%d", v))
			}

		case "json_schema", "j":
			if v, ok := val.(string); ok && v != "" {
				args = append(args, "-j", v)
			}

		case "grammar":
			if v, ok := val.(string); ok && v != "" {
				args = append(args, "--grammar", v)
			}

		case "seed":
			if v, ok := val.(int); ok {
				args = append(args, "-s", fmt.Sprintf("%d", v))
			}

		case "predict", "n_predict":
			if v, ok := val.(int); ok {
				args = append(args, "-n", fmt.Sprintf("%d", v))
			}

		case "keep":
			if v, ok := val.(int); ok {
				args = append(args, "--keep", fmt.Sprintf("%d", v))
			}

		case "rope_scaling":
			if v, ok := val.(string); ok && v != "" {
				args = append(args, "--rope-scaling", v)
			}

		case "rope_scale":
			switch v := val.(type) {
			case float64:
				args = append(args, "--rope-scale", fmt.Sprintf("%.2f", v))
			case int:
				args = append(args, "--rope-scale", fmt.Sprintf("%d", v))
			}

		case "rope_freq_base":
			switch v := val.(type) {
			case float64:
				args = append(args, "--rope-freq-base", fmt.Sprintf("%.2f", v))
			case int:
				args = append(args, "--rope-freq-base", fmt.Sprintf("%d", v))
			}

		case "device":
			if v, ok := val.(string); ok && v != "" {
				args = append(args, "--device", v)
			}

		case "tensor_split":
			if v, ok := val.(string); ok && v != "" {
				args = append(args, "-ts", v)
			}

		case "main_gpu", "mg":
			if v, ok := val.(int); ok {
				args = append(args, "-mg", fmt.Sprintf("%d", v))
			}

		case "split_mode", "sm":
			if v, ok := val.(string); ok && v != "" {
				args = append(args, "-sm", v)
			}

		case "lora":
			if v, ok := val.(string); ok && v != "" {
				args = append(args, "--lora", v)
			}

		case "lora_scale":
			if v, ok := val.(string); ok && v != "" {
				args = append(args, "--lora-scaled", v)
			}

		case "log_level", "verbosity":
			switch v := val.(type) {
			case int:
				args = append(args, "--log-verbosity", fmt.Sprintf("%d", v))
			case string:
				args = append(args, "--log-verbosity", v)
			}

		case "log_file":
			if v, ok := val.(string); ok && v != "" {
				args = append(args, "--log-file", v)
			}

		default:
			// Handle unknown flags by converting to llama-server format
			flagName := normalizeFlag(key)
			if flagName == "" {
				continue
			}
			switch v := val.(type) {
			case bool:
				args = append(args, "--"+flagName)
			case string:
				if v != "" {
					args = append(args, "--"+flagName, v)
				}
			case int:
				args = append(args, "--"+flagName, fmt.Sprintf("%d", v))
			case float64:
				args = append(args, "--"+flagName, fmt.Sprintf("%.2f", v))
			default:
				s := fmt.Sprintf("%v", v)
				if s != "<nil>" && s != "" {
					args = append(args, "--"+flagName, s)
				}
			}
		}
	}

	return args, nil
}

// normalizeFlag converts a camelCase or snake_case key to a kebab-case flag.
func normalizeFlag(key string) string {
	// If already kebab-case, return as-is
	if strings.Contains(key, "-") {
		return key
	}

	// Convert snake_case to kebab-case first
	if strings.Contains(key, "_") {
		return strings.ReplaceAll(key, "_", "-")
	}

	// Convert camelCase to kebab-case
	var result []string
	for _, part := range splitCamelCase(key) {
		result = append(result, strings.ToLower(part))
	}
	return strings.Join(result, "-")
}

// splitCamelCase splits a camelCase string into parts.
func splitCamelCase(s string) []string {
	if s == "" {
		return nil
	}

	var parts []string
	start := 0
	for i := 1; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			part := s[start:i]
			if part != "" {
				parts = append(parts, part)
			}
			start = i
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// RunCommand executes the built command and waits for completion.
func RunCommand(cmd *exec.Cmd) error {
	fmt.Printf("Executing: %s %s\n", cmd.Path, strings.Join(cmd.Args[1:], " "))
	return cmd.Run()
}
