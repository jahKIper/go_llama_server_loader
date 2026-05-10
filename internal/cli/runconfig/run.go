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
func SaveAndRun(modelsCfgPath string, m *modelscan.Model, rows []ParamRow) error {
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

	mc, _ := UpsertModelInConfig(cfg, m, rows, time.Now())

	if err := config.SaveConfig(cfg, modelsCfgPath); err != nil {
		return fmt.Errorf("SaveAndRun: не удалось сохранить конфиг: %w", err)
	}

	runCfg := buildServercmdConfig(cfg, mc)

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
func buildServercmdConfig(cfg *config.Config, mc *config.ModelConfig) *servercmd.Config {
	scMC := servercmd.ModelConfig{
		Name:       mc.Name,
		ModelPath:  mc.ModelPath,
		MMProjPath: mc.MMProjPath,
		MMProjOn:   mc.MMProjOn,
		Size:       mc.Size,
		LastScan:   mc.LastScan,
		Flags:      make(map[string]any, len(mc.Flags)),
	}
	for k, v := range mc.Flags {
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
