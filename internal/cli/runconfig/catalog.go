package runconfig

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"llama-server-loader/internal/config"
)

// CatalogEntry — одна запись в правой панели (каталог параметров).
type CatalogEntry struct {
	Category string
	Meta     *config.ParamMeta
}

// ResolveParamsFile ищет params_ru.json в порядке:
//  1. explicit (если непустой и файл существует)
//  2. рядом с бинарником
//  3. ./params_ru.json
//
// Если ни один вариант не подошёл — ("", os.ErrNotExist).
func ResolveParamsFile(explicit string) (string, error) {
	if explicit != "" {
		if _, err := os.Stat(explicit); err == nil {
			return explicit, nil
		}
	}

	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "params_ru.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	candidate := "./params_ru.json"
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	return "", os.ErrNotExist
}

// LoadCatalog читает и парсит params_ru.json по заданному пути.
func LoadCatalog(path string) (*config.ParamFile, error) {
	return config.LoadParams(path)
}

// FlattenCatalog возвращает плоский срез всех параметров с категориями.
// Записи, у которых и LongFlag, и ShortFlag пусты после нормализации — отфильтровываются.
// Записи только с ShortFlag (пустой LongFlag) — сохраняются: Key формируется из ShortFlag.
func FlattenCatalog(pf *config.ParamFile) []CatalogEntry {
	if pf == nil {
		return nil
	}
	entries := make([]CatalogEntry, 0, int(pf.TotalParamsCount))
	for i := range pf.Categories {
		cat := &pf.Categories[i]
		for j := range cat.Params {
			meta := &cat.Params[j]
			longOK := strings.TrimSpace(stripFlagArg(meta.LongFlag)) != ""
			shortOK := strings.TrimSpace(stripFlagArg(meta.ShortFlag)) != ""
			if !longOK && !shortOK {
				// Нет ни long, ни short флага — пропускаем
				continue
			}
			entries = append(entries, CatalogEntry{
				Category: cat.Name,
				Meta:     meta,
			})
		}
	}
	return entries
}

// ParamKey возвращает нормализованный ключ для ModelConfig.Flags.
// Пример: "--ctx-size N" → "ctx_size", "--temp" → "temp".
// Если LongFlag пустой — нормализует ShortFlag.
func ParamKey(meta *config.ParamMeta) string {
	src := meta.LongFlag
	if src == "" {
		src = meta.ShortFlag
	}
	// Отбросить хвост после первого пробела: "--ctx-size N" → "--ctx-size"
	if idx := strings.IndexByte(src, ' '); idx >= 0 {
		src = src[:idx]
	}
	// Убрать ведущие тире
	src = strings.TrimLeft(src, "-")
	// Заменить тире на подчёркивание
	return strings.ReplaceAll(src, "-", "_")
}

// errNotFound возвращает true, если ошибка означает «файл не найден».
func errNotFound(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
