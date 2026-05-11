package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config represents the top-level structure for models.json.
type Config struct {
	Version   string           `json:"version"`
	ScanPaths []string         `json:"scan_paths"`
	Models    []ModelConfig    `json:"models"`
}

// ModelConfig represents a single model configuration entry.
type ModelConfig struct {
	Name       string                 `json:"name"`
	ModelPath  string                 `json:"model_path"`
	MMProjPath string                 `json:"mmproj_path,omitempty"`
	MMProjOn   bool                   `json:"mmproj_on"`
	Size       int64                  `json:"size"`
	LastScan   string                 `json:"last_scan"`
	Flags      map[string]interface{} `json:"flags"`
	Params     []ModelParam           `json:"params,omitempty"`
}

// ModelParam — одна запись метаданных GGUF, сохраняемая в models.json.
type ModelParam struct {
	Key           string `json:"key"`
	Value         any    `json:"value"`
	DescriptionRU string `json:"description_ru"`
}

// LoadConfig loads and parses a models.json file from the given path.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("config: path cannot be empty")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: cannot read file %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("config: cannot unmarshal config: %w", err)
	}

	return &cfg, nil
}

// SaveConfig saves the Config to a JSON file at the given path.
// It uses IndentedJSON for pretty-printing and ensures no data loss.
func SaveConfig(cfg *Config, path string) error {
	if cfg == nil {
		return fmt.Errorf("config: config cannot be nil")
	}
	if path == "" {
		return fmt.Errorf("config: path cannot be empty")
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("config: cannot marshal config to JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("config: cannot write file %s: %w", path, err)
	}

	return nil
}

// ValidateConfig validates the Config structure for basic correctness.
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config: config is nil")
	}

	if cfg.Version == "" {
		return fmt.Errorf("config: version is empty")
	}

	if len(cfg.ScanPaths) == 0 {
		return fmt.Errorf("config: no scan_paths provided")
	}

	for i, path := range cfg.ScanPaths {
		if path == "" {
			return fmt.Errorf("config: empty scan_path at index %d", i)
		}
	}

	if len(cfg.Models) == 0 {
		return fmt.Errorf("config: no models provided")
	}

	for i, model := range cfg.Models {
		if err := ValidateModelConfig(&model, i); err != nil {
			return err
		}
	}

	return nil
}

// ValidateModelConfig validates a single ModelConfig entry.
func ValidateModelConfig(mc *ModelConfig, index int) error {
	if mc == nil {
		return fmt.Errorf("config: model at index %d is nil", index)
	}

	if mc.Name == "" {
		return fmt.Errorf("config: model at index %d has empty name", index)
	}

	if mc.ModelPath == "" {
		return fmt.Errorf("config: model '%s' has empty model_path", mc.Name)
	}

	if mc.Size < 0 {
		return fmt.Errorf("config: model '%s' has negative size", mc.Name)
	}

	// Validate LastScan format if provided
	if mc.LastScan != "" {
		_, err := time.Parse(time.RFC3339, mc.LastScan)
		if err != nil {
			return fmt.Errorf("config: model '%s' has invalid last_scan format: %w", mc.Name, err)
		}
	}

	return nil
}

// AddModel adds a new ModelConfig to the Config and returns the updated config.
func (cfg *Config) AddModel(model ModelConfig) {
	cfg.Models = append(cfg.Models, model)
}

// RemoveModel removes a model by name from the Config.
func (cfg *Config) RemoveModel(name string) {
	for i, model := range cfg.Models {
		if model.Name == name {
			cfg.Models = append(cfg.Models[:i], cfg.Models[i+1:]...)
			return
		}
	}
}

// GetModel retrieves a model by name from the Config.
func (cfg *Config) GetModel(name string) (*ModelConfig, bool) {
	for _, model := range cfg.Models {
		if model.Name == name {
			return &model, true
		}
	}
	return nil, false
}

// UpdateLastScan updates the LastScan field for a model by name.
func (cfg *Config) UpdateLastScan(name string, scanTime time.Time) error {
	for i, model := range cfg.Models {
		if model.Name == name {
			cfg.Models[i].LastScan = scanTime.Format(time.RFC3339)
			return nil
		}
	}
	return fmt.Errorf("config: model '%s' not found", name)
}
