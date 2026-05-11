package runconfig

import (
	"context"
	"fmt"
	"os"
	"time"

	"llama-server-loader/internal/config"
	"llama-server-loader/pkg/modelscan"
	"llama-server-loader/pkg/servercmd"
)

// SaveAndRun сохраняет запись о модели в models.json и запускает llama-server.
// Если файл models.json не существует — создаёт новый.
// Если модель с таким именем уже есть — не меняет запись, использует её Flags для запуска.
func SaveAndRun(modelsCfgPath string, m *modelscan.Model, rows []ParamRow, comment string) error {
	if modelsCfgPath == "" {
		modelsCfgPath = "models.json"
	}

	cfg, err := config.LoadConfig(modelsCfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = &config.Config{Version: "1.0"}
		} else {
			return fmt.Errorf("SaveAndRun: не удалось загрузить конфиг %s: %w", modelsCfgPath, err)
		}
	}

	now := time.Now()
	mc, _ := UpsertModelInConfig(cfg, m, rows, now)
	for i := range cfg.Models {
		if cfg.Models[i].Name == mc.Name {
			cfg.Models[i].LastRun = now.UTC().Format(time.RFC3339)
			cfg.Models[i].Comment = comment
			break
		}
	}

	if err := config.SaveConfig(cfg, modelsCfgPath); err != nil {
		return fmt.Errorf("SaveAndRun: не удалось сохранить конфиг: %w", err)
	}

	// Для запуска используем все флаги, включая нетронутые автодефолты
	// (которые в mc.Flags не попали — они не сохраняются).
	runFlags := BuildFlagsMap(rows, m.Path)
	runCfg := buildServercmdConfig(cfg, mc, runFlags)

	ctx := context.Background()
	cmd, err := servercmd.BuildCommandWithContext(ctx, runCfg)
	if err != nil {
		return fmt.Errorf("SaveAndRun: не удалось собрать команду: %w", err)
	}

	// Восстанавливаем нормальный экран перед запуском llama-server,
	// чтобы вывод сервера шёл в основной буфер, а не в AltScreen.
	fmt.Print("\033[?1049l")

	return servercmd.RunCommand(cmd)
}

// buildServercmdConfig конвертирует *config.ModelConfig в *servercmd.Config для запуска.
// flags — карта флагов для использования при запуске (может включать нетронутые
// автодефолты, которых нет в mc.Flags).
func buildServercmdConfig(cfg *config.Config, mc *config.ModelConfig, flags map[string]any) *servercmd.Config {
	scMC := servercmd.ModelConfig{
		Name:       mc.Name,
		ModelPath:  mc.ModelPath,
		MMProjPath: mc.MMProjPath,
		MMProjOn:   mc.MMProjOn,
		Size:       mc.Size,
		LastScan:   mc.LastScan,
		Flags:      make(map[string]any, len(flags)),
	}
	for k, v := range flags {
		scMC.Flags[k] = v
	}

	version := "1.0"
	if cfg != nil && cfg.Version != "" {
		version = cfg.Version
	}

	return &servercmd.Config{
		Version: version,
		Models:  []servercmd.ModelConfig{scMC},
	}
}
