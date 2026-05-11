package ggufmeta

import "fmt"

// Param — формат записи, который сохраняется в models.json.
type Param struct {
	Key           string `json:"key"`
	Value         any    `json:"value"`
	DescriptionRU string `json:"description_ru"`
}

// ExtractParams парсит GGUF и возвращает плоский список параметров
// с русскими описаниями. Подходит для прямой записи в ModelConfig.Params.
func ExtractParams(path string) ([]Param, error) {
	f, err := ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := make([]Param, 0, len(f.KV))
	for _, kv := range f.KV {
		val := kv.Value
		if kv.Key == "general.file_type" {
			if u, ok := toUint64(val); ok {
				if name, ok := FileTypeName[u]; ok {
					val = fmt.Sprintf("%d (%s)", u, name)
				}
			}
		}
		out = append(out, Param{
			Key:           kv.Key,
			Value:         val,
			DescriptionRU: Describe(kv.Key),
		})
	}
	return out, nil
}

func toUint64(v any) (uint64, bool) {
	switch x := v.(type) {
	case uint64:
		return x, true
	case uint32:
		return uint64(x), true
	case uint16:
		return uint64(x), true
	case uint8:
		return uint64(x), true
	case int64:
		if x >= 0 {
			return uint64(x), true
		}
	case int32:
		if x >= 0 {
			return uint64(x), true
		}
	case int:
		if x >= 0 {
			return uint64(x), true
		}
	}
	return 0, false
}
