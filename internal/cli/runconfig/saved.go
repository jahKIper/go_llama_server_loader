package runconfig

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"llama-server-loader/internal/config"
	"llama-server-loader/pkg/modelscan"
)

// LoadSavedFlagsForModel читает models.json и возвращает сохранённую карту флагов
// для модели m (поиск по имени = basename без ".gguf"). Возвращает (nil, false),
// если файл не существует, нечитаем или модель в нём отсутствует.
func LoadSavedFlagsForModel(modelsCfgPath string, m *modelscan.Model) (map[string]any, bool) {
	if modelsCfgPath == "" || m == nil {
		return nil, false
	}
	cfg, err := config.LoadConfig(modelsCfgPath)
	if err != nil {
		return nil, false
	}
	name := strings.TrimSuffix(filepath.Base(m.Path), ".gguf")
	mc, ok := cfg.GetModel(name)
	if !ok || mc == nil {
		return nil, false
	}
	return mc.Flags, true
}

// LoadCommentForModel читает comment из models.json для модели m.
// Возвращает пустую строку, если файл не существует или модель не найдена.
func LoadCommentForModel(modelsCfgPath string, m *modelscan.Model) string {
	if modelsCfgPath == "" || m == nil {
		return ""
	}
	cfg, err := config.LoadConfig(modelsCfgPath)
	if err != nil {
		return ""
	}
	name := strings.TrimSuffix(filepath.Base(m.Path), ".gguf")
	mc, ok := cfg.GetModel(name)
	if !ok || mc == nil {
		return ""
	}
	return mc.Comment
}

// MergeWithSavedFlags берёт предзаполненные строки (--model/--mmproj/…) и
// накатывает на них сохранённые значения из models.json. Если сохранённый
// флаг отсутствует в prefilled — он добавляется в хвост (по алфавиту ключа).
func MergeWithSavedFlags(catalog []CatalogEntry, prefilled []ParamRow, saved map[string]any) []ParamRow {
	out := append([]ParamRow(nil), prefilled...)
	longIdx := make(map[string]int, len(out))
	for i, r := range out {
		longIdx[r.Long] = i
	}

	keys := make([]string, 0, len(saved))
	for k := range saved {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		meta := findMetaByKey(catalog, key)
		if meta == nil {
			continue
		}
		long := stripFlagArg(meta.LongFlag)
		if long == "" {
			long = stripFlagArg(meta.ShortFlag)
		}
		if long == "" {
			continue
		}
		valStr := formatFlagValue(saved[key])
		if idx, ok := longIdx[long]; ok {
			out[idx].Value = valStr
			continue
		}
		out = append(out, ParamRow{
			Long:  long,
			Short: stripFlagArg(meta.ShortFlag),
			Key:   ParamKey(meta),
			Value: valStr,
			Meta:  meta,
		})
		longIdx[long] = len(out) - 1
	}
	return out
}

// findMetaByKey ищет каталожный мета-объект по нормализованному Key (см. ParamKey).
func findMetaByKey(catalog []CatalogEntry, key string) *config.ParamMeta {
	for i := range catalog {
		if ParamKey(catalog[i].Meta) == key {
			return catalog[i].Meta
		}
	}
	return nil
}

// formatFlagValue приводит значение из map[string]any (как читает encoding/json)
// к строке, которую можно отобразить в инпуте и снова распарсить через ParseFlagValue.
func formatFlagValue(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		// encoding/json по умолчанию декодирует числа в float64.
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}
