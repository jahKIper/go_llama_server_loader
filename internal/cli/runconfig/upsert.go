package runconfig

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"llama-server-loader/internal/config"
	"llama-server-loader/internal/ggufmeta"
	"llama-server-loader/pkg/modelscan"
)

// fillParams читает GGUF и заполняет []ModelParam. При ошибке логирует и возвращает nil.
func fillParams(path string) []config.ModelParam {
	ps, err := ggufmeta.ExtractParams(path)
	if err != nil {
		log.Printf("ggufmeta: skip %s: %v", path, err)
		return nil
	}
	out := make([]config.ModelParam, len(ps))
	for i, p := range ps {
		out[i] = config.ModelParam{Key: p.Key, Value: p.Value, DescriptionRU: p.DescriptionRU}
	}
	return out
}

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
		// Бэкфилл: если у существующей записи нет GGUF-параметров — заполняем.
		if len(existing.Params) == 0 {
			for i := range cfg.Models {
				if cfg.Models[i].Name == name {
					cfg.Models[i].Params = fillParams(cfg.Models[i].ModelPath)
					return &cfg.Models[i], false
				}
			}
		}
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
		Params:     fillParams(m.Path),
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
