package runconfig

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"llama-server-loader/internal/config"
	"llama-server-loader/pkg/modelscan"
)

func makeTestModel(path string, size int64, mmprojPaths []string) *modelscan.Model {
	return &modelscan.Model{
		Path:        path,
		Size:        size,
		MMProjPaths: mmprojPaths,
	}
}

func makeTestConfig() *config.Config {
	return &config.Config{Version: "1.0"}
}

// TestUpsertAddsNewModel проверяет, что новая модель добавляется в пустой конфиг.
func TestUpsertAddsNewModel(t *testing.T) {
	cfg := makeTestConfig()
	m := makeTestModel("/models/llama-7b.gguf", 4_000_000_000, nil)
	rows := []ParamRow{
		{Long: "--ctx-size", Short: "-c", Key: "ctx_size", Value: "4096"},
	}
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	mc, added := UpsertModelInConfig(cfg, m, rows, now)

	if !added {
		t.Fatal("expected added=true для новой модели")
	}
	if mc == nil {
		t.Fatal("expected non-nil *ModelConfig")
	}
	if mc.Name != "llama-7b" {
		t.Errorf("expected name=llama-7b, got %s", mc.Name)
	}
	if mc.ModelPath != "/models/llama-7b.gguf" {
		t.Errorf("expected ModelPath=/models/llama-7b.gguf, got %s", mc.ModelPath)
	}
	if mc.Size != 4_000_000_000 {
		t.Errorf("expected Size=4000000000, got %d", mc.Size)
	}
	if mc.MMProjOn {
		t.Error("expected MMProjOn=false при отсутствии mmprojPaths")
	}
	if mc.LastScan != "2024-01-15T12:00:00Z" {
		t.Errorf("expected LastScan=2024-01-15T12:00:00Z, got %s", mc.LastScan)
	}
	if v, ok := mc.Flags["ctx_size"]; !ok || v != 4096 {
		t.Errorf("expected Flags[ctx_size]=4096, got %v", mc.Flags["ctx_size"])
	}
	if v, ok := mc.Flags["model"]; !ok || v != "/models/llama-7b.gguf" {
		t.Errorf("expected Flags[model]=/models/llama-7b.gguf, got %v", mc.Flags["model"])
	}
	if len(cfg.Models) != 1 {
		t.Errorf("expected 1 model in config, got %d", len(cfg.Models))
	}
}

// TestUpsertUpdatesExistingWithUserRows проверяет, что повторный upsert
// обновляет Flags существующей записи пользовательскими (не-default) значениями.
func TestUpsertUpdatesExistingWithUserRows(t *testing.T) {
	cfg := makeTestConfig()
	existingFlags := map[string]interface{}{"ctx_size": 2048, "model": "/models/llama-7b.gguf"}
	cfg.Models = []config.ModelConfig{
		{
			Name:      "llama-7b",
			ModelPath: "/models/llama-7b.gguf",
			Size:      4_000_000_000,
			LastScan:  "2023-06-01T10:00:00Z",
			Flags:     existingFlags,
		},
	}

	m := makeTestModel("/models/llama-7b.gguf", 4_000_000_000, nil)
	newRows := []ParamRow{
		{Long: "--ctx-size", Short: "-c", Key: "ctx_size", Value: "8192"},
		{Long: "--threads", Short: "-t", Key: "threads", Value: "16"},
	}

	mc, added := UpsertModelInConfig(cfg, m, newRows, time.Now())

	if added {
		t.Fatal("expected added=false для уже существующей модели")
	}
	if mc.Flags["ctx_size"] != 8192 {
		t.Errorf("Flags должны обновиться: expected ctx_size=8192, got %v", mc.Flags["ctx_size"])
	}
	if mc.Flags["threads"] != 16 {
		t.Errorf("новый флаг должен сохраниться: expected threads=16, got %v", mc.Flags["threads"])
	}
	if len(cfg.Models) != 1 {
		t.Errorf("количество моделей не должно измениться, got %d", len(cfg.Models))
	}
}

// TestUpsertMMProjFields проверяет корректное заполнение MMProj-полей.
func TestUpsertMMProjFields(t *testing.T) {
	cfg := makeTestConfig()
	m := makeTestModel("/models/llava.gguf", 5_000_000_000, []string{"/models/llava-mmproj.gguf", "/models/llava-mmproj2.gguf"})
	now := time.Now()

	mc, added := UpsertModelInConfig(cfg, m, nil, now)

	if !added {
		t.Fatal("expected added=true")
	}
	if !mc.MMProjOn {
		t.Error("expected MMProjOn=true при наличии mmprojPaths")
	}
	if mc.MMProjPath != "/models/llava-mmproj.gguf" {
		t.Errorf("expected MMProjPath первый из среза, got %s", mc.MMProjPath)
	}
}

// TestUpsertRoundTrip проверяет, что SaveConfig/LoadConfig сохраняет данные без потерь.
func TestUpsertRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "models.json")

	cfg := makeTestConfig()
	m := makeTestModel("/models/mistral.gguf", 3_500_000_000, nil)
	rows := []ParamRow{
		{Long: "--ctx-size", Key: "ctx_size", Value: "8192"},
		{Long: "--temp", Key: "temp", Value: "0.7"},
	}
	now := time.Date(2024, 3, 10, 8, 0, 0, 0, time.UTC)

	mc, _ := UpsertModelInConfig(cfg, m, rows, now)

	if err := config.SaveConfig(cfg, cfgPath); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	loaded, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	loadedMC, ok := loaded.GetModel("mistral")
	if !ok {
		t.Fatal("модель mistral не найдена после round-trip")
	}
	if loadedMC.Name != mc.Name {
		t.Errorf("Name: got %s, want %s", loadedMC.Name, mc.Name)
	}
	if loadedMC.ModelPath != mc.ModelPath {
		t.Errorf("ModelPath: got %s, want %s", loadedMC.ModelPath, mc.ModelPath)
	}
	if loadedMC.LastScan != "2024-03-10T08:00:00Z" {
		t.Errorf("LastScan: got %s", loadedMC.LastScan)
	}

	// После JSON-round-trip числа приходят как float64
	ctxRaw := loadedMC.Flags["ctx_size"]
	ctxJSON, _ := json.Marshal(ctxRaw)
	if string(ctxJSON) != "8192" {
		t.Errorf("ctx_size after round-trip: %s", ctxJSON)
	}
}

// TestUpsertEmptyRows проверяет, что при пустых rows добавляется только "model".
func TestUpsertEmptyRows(t *testing.T) {
	cfg := makeTestConfig()
	m := makeTestModel("/models/phi.gguf", 1_500_000_000, nil)

	mc, added := UpsertModelInConfig(cfg, m, []ParamRow{}, time.Now())

	if !added {
		t.Fatal("expected added=true")
	}
	if len(mc.Flags) != 1 {
		t.Errorf("expected только Flags[model], got %d entries: %v", len(mc.Flags), mc.Flags)
	}
	if mc.Flags["model"] != "/models/phi.gguf" {
		t.Errorf("expected Flags[model]=/models/phi.gguf, got %v", mc.Flags["model"])
	}
}

// TestFirstOrEmpty проверяет вспомогательную функцию.
func TestFirstOrEmpty(t *testing.T) {
	if got := firstOrEmpty(nil); got != "" {
		t.Errorf("nil slice: expected empty, got %q", got)
	}
	if got := firstOrEmpty([]string{}); got != "" {
		t.Errorf("empty slice: expected empty, got %q", got)
	}
	if got := firstOrEmpty([]string{"a", "b"}); got != "a" {
		t.Errorf("expected a, got %q", got)
	}
}

// TestUpsertUpdatesExistingFlags проверяет, что повторный upsert обновляет
// Flags существующей записи пользовательскими (не-default) значениями.
func TestUpsertUpdatesExistingFlags(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "models.json")

	// Первый upsert: пользователь задал ctx_size=4096
	cfg := makeTestConfig()
	m := makeTestModel("/models/qwen.gguf", 2_000_000_000, nil)
	rows := []ParamRow{{Long: "--ctx-size", Key: "ctx_size", Value: "4096"}}
	UpsertModelInConfig(cfg, m, rows, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	if err := config.SaveConfig(cfg, cfgPath); err != nil {
		t.Fatalf("первый SaveConfig: %v", err)
	}

	// Второй upsert: пользователь поменял значение на 8192
	cfg2, _ := config.LoadConfig(cfgPath)
	newRows := []ParamRow{{Long: "--ctx-size", Key: "ctx_size", Value: "8192"}}
	mc, added := UpsertModelInConfig(cfg2, m, newRows, time.Now())
	if added {
		t.Fatal("повторный upsert не должен добавлять модель")
	}
	if mc.Flags["ctx_size"] != 8192 {
		t.Errorf("Flags[ctx_size] expected 8192, got %v", mc.Flags["ctx_size"])
	}
}

// TestUpsertSkipsDefaultRows проверяет, что строки с IsDefault=true не попадают
// в сохранённые Flags.
func TestUpsertSkipsDefaultRows(t *testing.T) {
	cfg := makeTestConfig()
	m := makeTestModel("/models/qwen.gguf", 2_000_000_000, nil)
	rows := []ParamRow{
		{Long: "--ctx-size", Key: "ctx_size", Value: "4096"},                  // пользователь
		{Long: "--gpu-layers", Key: "gpu_layers", Value: "40", IsDefault: true}, // дефолт
	}
	mc, _ := UpsertModelInConfig(cfg, m, rows, time.Now())
	if _, ok := mc.Flags["gpu_layers"]; ok {
		t.Errorf("IsDefault строка не должна попадать в Flags, got %v", mc.Flags)
	}
	if mc.Flags["ctx_size"] != 4096 {
		t.Errorf("Flags[ctx_size] expected 4096, got %v", mc.Flags["ctx_size"])
	}
}
