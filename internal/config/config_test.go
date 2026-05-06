package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLoadConfig_ValidFile tests loading a valid models.json file.
func TestLoadConfig_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "models.json")

	configData := `{
  "version": "1.0",
  "scan_paths": ["/models/main", "/models/backup"],
  "models": [
    {
      "name": "gemma-2b",
      "model_path": "/models/main/gemma-2b-q4_k.gguf",
      "mmproj_path": "/models/main/mmproj-gemma-2b.bin",
      "mmproj_on": true,
      "size": 1500000000,
      "last_scan": "2024-05-01T12:00:00Z",
      "flags": {
        "model": "/models/main/gemma-2b-q4_k.gguf",
        "ctx_len": 4096,
        "gpu_layers": 24
      }
    }
  ]
}`

	if err := os.WriteFile(path, []byte(configData), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Version != "1.0" {
		t.Errorf("expected version '1.0', got '%s'", cfg.Version)
	}

	if len(cfg.ScanPaths) != 2 {
		t.Errorf("expected 2 scan_paths, got %d", len(cfg.ScanPaths))
	}

	if len(cfg.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(cfg.Models))
	}

	model := cfg.Models[0]
	if model.Name != "gemma-2b" {
		t.Errorf("expected name 'gemma-2b', got '%s'", model.Name)
	}

	if model.MMProjPath != "/models/main/mmproj-gemma-2b.bin" {
		t.Errorf("expected mmproj_path '/models/main/mmproj-gemma-2b.bin', got '%s'", model.MMProjPath)
	}

	if !model.MMProjOn {
		t.Error("expected mmproj_on to be true")
	}

	if model.Size != 1500000000 {
		t.Errorf("expected size 1500000000, got %d", model.Size)
	}

	if model.LastScan != "2024-05-01T12:00:00Z" {
		t.Errorf("expected last_scan '2024-05-01T12:00:00Z', got '%s'", model.LastScan)
	}

	if model.Flags["ctx_len"] != float64(4096) {
		t.Errorf("expected ctx_len 4096, got %v", model.Flags["ctx_len"])
	}
}

// TestLoadConfig_EmptyPath tests loading with an empty path.
func TestLoadConfig_EmptyPath(t *testing.T) {
	_, err := LoadConfig("")
	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
}

// TestLoadConfig_NonExistentFile tests loading a non-existent file.
func TestLoadConfig_NonExistentFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/models.json")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

// TestSaveConfig_ValidConfig tests saving a valid config.
func TestSaveConfig_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "models.json")

	cfg := &Config{
		Version:   "1.0",
		ScanPaths: []string{"/models/test"},
		Models: []ModelConfig{
			{
				Name:       "test-model",
				ModelPath:  "/models/test/model.gguf",
				Size:       1000000000,
				MMProjOn:   false,
				LastScan:   time.Now().Format(time.RFC3339),
				Flags:      map[string]interface{}{"ctx_len": int64(2048)},
			},
		},
	}

	if err := SaveConfig(cfg, path); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file was created and can be loaded back
	loadedCfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("failed to reload saved config: %v", err)
	}

	if loadedCfg.Version != cfg.Version {
		t.Errorf("expected version '%s', got '%s'", cfg.Version, loadedCfg.Version)
	}

	if len(loadedCfg.Models) != 1 {
		t.Fatalf("expected 1 model after reload, got %d", len(loadedCfg.Models))
	}

	if loadedCfg.Models[0].Name != "test-model" {
		t.Errorf("expected name 'test-model', got '%s'", loadedCfg.Models[0].Name)
	}
}

// TestSaveConfig_NilConfig tests saving with a nil config.
func TestSaveConfig_NilConfig(t *testing.T) {
	err := SaveConfig(nil, "/tmp/test.json")
	if err == nil {
		t.Error("expected error for nil config, got nil")
	}
}

// TestSaveConfig_EmptyPath tests saving with an empty path.
func TestSaveConfig_EmptyPath(t *testing.T) {
	cfg := &Config{Version: "1.0"}
	err := SaveConfig(cfg, "")
	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
}

// TestValidateConfig_Valid tests validation of a valid config.
func TestValidateConfig_Valid(t *testing.T) {
	cfg := &Config{
		Version:   "1.0",
		ScanPaths: []string{"/models/test"},
		Models: []ModelConfig{
			{
				Name:       "valid-model",
				ModelPath:  "/models/test/model.gguf",
				Size:       1000000000,
				MMProjOn:   false,
				LastScan:   time.Now().Format(time.RFC3339),
			},
		},
	}

	err := ValidateConfig(cfg)
	if err != nil {
		t.Fatalf("ValidateConfig failed for valid config: %v", err)
	}
}

// TestValidateConfig_Nil tests validation with a nil config.
func TestValidateConfig_Nil(t *testing.T) {
	err := ValidateConfig(nil)
	if err == nil {
		t.Error("expected error for nil config, got nil")
	}
}

// TestValidateConfig_EmptyVersion tests validation with empty version.
func TestValidateConfig_EmptyVersion(t *testing.T) {
	cfg := &Config{
		ScanPaths: []string{"/models/test"},
		Models:    []ModelConfig{{Name: "test", ModelPath: "/test.gguf"}},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for empty version, got nil")
	}
}

// TestValidateConfig_EmptyScanPaths tests validation with no scan_paths.
func TestValidateConfig_EmptyScanPaths(t *testing.T) {
	cfg := &Config{
		Version: "1.0",
		Models:  []ModelConfig{{Name: "test", ModelPath: "/test.gguf"}},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for empty scan_paths, got nil")
	}
}

// TestValidateConfig_EmptyScanPathEntry tests validation with empty entry in scan_paths.
func TestValidateConfig_EmptyScanPathEntry(t *testing.T) {
	cfg := &Config{
		Version:   "1.0",
		ScanPaths: []string{"/models/test", ""},
		Models:    []ModelConfig{{Name: "test", ModelPath: "/test.gguf"}},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for empty scan_path entry, got nil")
	}
}

// TestValidateConfig_NoModels tests validation with no models.
func TestValidateConfig_NoModels(t *testing.T) {
	cfg := &Config{
		Version:   "1.0",
		ScanPaths: []string{"/models/test"},
		Models:    []ModelConfig{},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for no models, got nil")
	}
}

// TestValidateConfig_EmptyModelName tests validation with empty model name.
func TestValidateConfig_EmptyModelName(t *testing.T) {
	cfg := &Config{
		Version:   "1.0",
		ScanPaths: []string{"/models/test"},
		Models:    []ModelConfig{{Name: "", ModelPath: "/test.gguf"}},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for empty model name, got nil")
	}
}

// TestValidateConfig_EmptyModelPath tests validation with empty model_path.
func TestValidateConfig_EmptyModelPath(t *testing.T) {
	cfg := &Config{
		Version:   "1.0",
		ScanPaths: []string{"/models/test"},
		Models:    []ModelConfig{{Name: "test"}},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for empty model_path, got nil")
	}
}

// TestValidateConfig_NegativeSize tests validation with negative size.
func TestValidateConfig_NegativeSize(t *testing.T) {
	cfg := &Config{
		Version:   "1.0",
		ScanPaths: []string{"/models/test"},
		Models:    []ModelConfig{{Name: "test", ModelPath: "/test.gguf", Size: -1}},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for negative size, got nil")
	}
}

// TestValidateConfig_InvalidLastScan tests validation with invalid last_scan format.
func TestValidateConfig_InvalidLastScan(t *testing.T) {
	cfg := &Config{
		Version:   "1.0",
		ScanPaths: []string{"/models/test"},
		Models:    []ModelConfig{{Name: "test", ModelPath: "/test.gguf", LastScan: "not-a-date"}},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid last_scan, got nil")
	}
}

// TestAddModel tests adding a model to the config.
func TestAddModel(t *testing.T) {
	cfg := &Config{
		Version:   "1.0",
		ScanPaths: []string{"/models/test"},
		Models:    nil,
	}

	model := ModelConfig{
		Name:      "new-model",
		ModelPath: "/models/test/new.gguf",
		Size:      500000000,
	}

	cfg.AddModel(model)

	if len(cfg.Models) != 1 {
		t.Fatalf("expected 1 model after add, got %d", len(cfg.Models))
	}

	if cfg.Models[0].Name != "new-model" {
		t.Errorf("expected name 'new-model', got '%s'", cfg.Models[0].Name)
	}
}

// TestRemoveModel tests removing a model from the config.
func TestRemoveModel(t *testing.T) {
	cfg := &Config{
		Version:   "1.0",
		ScanPaths: []string{"/models/test"},
		Models: []ModelConfig{
			{Name: "model-1", ModelPath: "/test1.gguf"},
			{Name: "model-2", ModelPath: "/test2.gguf"},
			{Name: "model-3", ModelPath: "/test3.gguf"},
		},
	}

	cfg.RemoveModel("model-2")

	if len(cfg.Models) != 2 {
		t.Fatalf("expected 2 models after remove, got %d", len(cfg.Models))
	}

	for i, m := range cfg.Models {
		if m.Name == "model-2" {
			t.Errorf("model 'model-2' should have been removed")
		}
		if i == 0 && m.Name != "model-1" {
			t.Errorf("expected first model 'model-1', got '%s'", m.Name)
		}
		if i == 1 && m.Name != "model-3" {
			t.Errorf("expected second model 'model-3', got '%s'", m.Name)
		}
	}
}

// TestRemoveModel_NotFound tests removing a non-existent model.
func TestRemoveModel_NotFound(t *testing.T) {
	cfg := &Config{
		Version:   "1.0",
		ScanPaths: []string{"/models/test"},
		Models:    []ModelConfig{{Name: "model-1", ModelPath: "/test1.gguf"}},
	}

	cfg.RemoveModel("non-existent")

	if len(cfg.Models) != 1 {
		t.Errorf("expected 1 model (no change), got %d", len(cfg.Models))
	}
}

// TestGetModel tests retrieving a model by name.
func TestGetModel(t *testing.T) {
	cfg := &Config{
		Version:   "1.0",
		ScanPaths: []string{"/models/test"},
		Models: []ModelConfig{
			{Name: "model-1", ModelPath: "/test1.gguf"},
			{Name: "model-2", ModelPath: "/test2.gguf"},
		},
	}

	model, found := cfg.GetModel("model-1")
	if !found {
		t.Error("expected model to be found")
	}

	if model.Name != "model-1" {
		t.Errorf("expected name 'model-1', got '%s'", model.Name)
	}

	_, found = cfg.GetModel("non-existent")
	if found {
		t.Error("expected model not to be found")
	}
}

// TestUpdateLastScan tests updating the last_scan field.
func TestUpdateLastScan(t *testing.T) {
	cfg := &Config{
		Version:   "1.0",
		ScanPaths: []string{"/models/test"},
		Models:    []ModelConfig{{Name: "model-1", ModelPath: "/test1.gguf"}},
	}

	newTime := time.Now().Add(24 * time.Hour)
	err := cfg.UpdateLastScan("model-1", newTime)
	if err != nil {
		t.Fatalf("UpdateLastScan failed: %v", err)
	}

	expectedFormat := newTime.Format(time.RFC3339)
	if cfg.Models[0].LastScan != expectedFormat {
		t.Errorf("expected last_scan '%s', got '%s'", expectedFormat, cfg.Models[0].LastScan)
	}
}

// TestUpdateLastScan_NotFound tests updating a non-existent model.
func TestUpdateLastScan_NotFound(t *testing.T) {
	cfg := &Config{
		Version:   "1.0",
		ScanPaths: []string{"/models/test"},
		Models:    []ModelConfig{{Name: "model-1", ModelPath: "/test1.gguf"}},
	}

	err := cfg.UpdateLastScan("non-existent", time.Now())
	if err == nil {
		t.Error("expected error for non-existent model, got nil")
	}
}

// TestSaveConfig_LosslessRoundTrip tests that saving and loading preserves all data.
func TestSaveConfig_LosslessRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "models.json")

	originalCfg := &Config{
		Version:   "2.0",
		ScanPaths: []string{"/models/primary", "/models/secondary"},
		Models: []ModelConfig{
			{
				Name:       "llama-3b-q5",
				ModelPath:  "/models/primary/llama-3b-q5.gguf",
				MMProjPath: "/models/primary/mmproj-3b.bin",
				MMProjOn:   true,
				Size:       2800000000,
				LastScan:   "2024-06-15T08:30:00Z",
				Flags: map[string]interface{}{
					"model":      "/models/primary/llama-3b-q5.gguf",
					"ctx_len":    int64(8192),
					"gpu_layers": int64(35),
					"temp":       float64(0.7),
				},
			},
		},
	}

	if err := SaveConfig(originalCfg, path); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loadedCfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig after save failed: %v", err)
	}

	if loadedCfg.Version != originalCfg.Version {
		t.Errorf("version mismatch: expected '%s', got '%s'", originalCfg.Version, loadedCfg.Version)
	}

	if len(loadedCfg.ScanPaths) != len(originalCfg.ScanPaths) {
		t.Fatalf("scan_paths length mismatch: expected %d, got %d", len(originalCfg.ScanPaths), len(loadedCfg.ScanPaths))
	}

	for i, sp := range originalCfg.ScanPaths {
		if loadedCfg.ScanPaths[i] != sp {
			t.Errorf("scan_paths[%d] mismatch: expected '%s', got '%s'", i, sp, loadedCfg.ScanPaths[i])
		}
	}

	if len(loadedCfg.Models) != 1 {
		t.Fatalf("models length mismatch: expected %d, got %d", 1, len(loadedCfg.Models))
	}

	origModel := originalCfg.Models[0]
	loadedModel := loadedCfg.Models[0]

	if loadedModel.Name != origModel.Name {
		t.Errorf("model name mismatch: expected '%s', got '%s'", origModel.Name, loadedModel.Name)
	}

	if loadedModel.ModelPath != origModel.ModelPath {
		t.Errorf("model_path mismatch: expected '%s', got '%s'", origModel.ModelPath, loadedModel.ModelPath)
	}

	if loadedModel.MMProjPath != origModel.MMProjPath {
		t.Errorf("mmproj_path mismatch: expected '%s', got '%s'", origModel.MMProjPath, loadedModel.MMProjPath)
	}

	if loadedModel.MMProjOn != origModel.MMProjOn {
		t.Errorf("mmproj_on mismatch: expected %v, got %v", origModel.MMProjOn, loadedModel.MMProjOn)
	}

	if loadedModel.Size != origModel.Size {
		t.Errorf("size mismatch: expected %d, got %d", origModel.Size, loadedModel.Size)
	}

	if loadedModel.LastScan != origModel.LastScan {
		t.Errorf("last_scan mismatch: expected '%s', got '%s'", origModel.LastScan, loadedModel.LastScan)
	}

	for k, v := range origModel.Flags {
		loadedVal, ok := loadedModel.Flags[k]
		if !ok {
			t.Errorf("flags[%s] missing in loaded config", k)
			continue
		}
		// JSON unmarshal converts all numbers to float64
		switch v.(type) {
		case int64:
			if loadedVal != float64(v.(int64)) {
				t.Errorf("flags[%s] mismatch: expected %v, got %v", k, v, loadedVal)
			}
		case float64:
			if loadedVal != v.(float64) {
				t.Errorf("flags[%s] mismatch: expected %v, got %v", k, v, loadedVal)
			}
		default:
			if loadedVal != v {
				t.Errorf("flags[%s] mismatch: expected %v, got %v", k, v, loadedVal)
			}
		}
	}
}
