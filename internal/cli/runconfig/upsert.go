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

// UpsertModelInConfig сохраняет модель в конфиг:
//   - если записи нет — создаёт её;
//   - если есть — обновляет Flags пользовательскими значениями (без IsDefault-строк)
//     и LastScan. Params (GGUF-метаданные) бэкфиллятся, если пустые.
//
// IsDefault-строки сюда не попадают: они и так автоматически подмешиваются при
// следующем открытии экрана через ComputeModelDefaults.
func UpsertModelInConfig(
	cfg *config.Config,
	m *modelscan.Model,
	rows []ParamRow,
	now time.Time,
) (mc *config.ModelConfig, added bool) {
	name := strings.TrimSuffix(filepath.Base(m.Path), ".gguf")
	savedFlags := BuildSavedFlagsMap(rows, m.Path)

	if _, ok := cfg.GetModel(name); ok {
		for i := range cfg.Models {
			if cfg.Models[i].Name != name {
				continue
			}
			cfg.Models[i].Flags = savedFlags
			cfg.Models[i].LastScan = now.UTC().Format(time.RFC3339)
			if len(cfg.Models[i].Params) == 0 {
				cfg.Models[i].Params = fillParams(cfg.Models[i].ModelPath)
			}
			return &cfg.Models[i], false
		}
	}

	newMC := config.ModelConfig{
		Name:       name,
		ModelPath:  m.Path,
		MMProjPath: firstOrEmpty(m.MMProjPaths),
		MMProjOn:   len(m.MMProjPaths) > 0,
		Size:       m.Size,
		LastScan:   now.UTC().Format(time.RFC3339),
		Flags:      savedFlags,
		Params:     fillParams(m.Path),
	}
	cfg.AddModel(newMC)

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
