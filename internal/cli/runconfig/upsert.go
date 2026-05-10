package runconfig

import (
	"path/filepath"
	"strings"
	"time"

	"llama-server-loader/internal/config"
	"llama-server-loader/pkg/modelscan"
)

// UpsertModelInConfig сохраняет модель в конфиг по правилу upsert:
// если модель с таким именем уже есть — возвращает существующий *ModelConfig без изменений;
// если нет — формирует запись из (m, rows, now) и добавляет через AddModel.
func UpsertModelInConfig(
	cfg *config.Config,
	m *modelscan.Model,
	rows []ParamRow,
	now time.Time,
) (mc *config.ModelConfig, added bool) {
	name := strings.TrimSuffix(filepath.Base(m.Path), ".gguf")

	if existing, ok := cfg.GetModel(name); ok {
		return existing, false
	}

	newMC := config.ModelConfig{
		Name:       name,
		ModelPath:  m.Path,
		MMProjPath: firstOrEmpty(m.MMProjPaths),
		MMProjOn:   len(m.MMProjPaths) > 0,
		Size:       m.Size,
		LastScan:   now.UTC().Format(time.RFC3339),
		Flags:      BuildFlagsMap(rows, m.Path),
	}
	cfg.AddModel(newMC)

	// Получаем указатель на только что добавленный элемент среза.
	added_mc, _ := cfg.GetModel(name)
	return added_mc, true
}

// firstOrEmpty возвращает первый элемент среза или пустую строку.
func firstOrEmpty(paths []string) string {
	if len(paths) > 0 {
		return paths[0]
	}
	return ""
}
