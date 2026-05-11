// Package modelparams предоставляет lookup по GGUF-параметрам моделей из
// models.json. Используется UI-слоем для обогащения карточек моделей и
// inline-подсказок на экране запуска.
package modelparams

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"llama-server-loader/internal/config"
	"llama-server-loader/internal/ggufmeta"
)

// Lookup — индексированный реестр GGUF-параметров для моделей.
type Lookup struct {
	byName map[string][]config.ModelParam
}

// NewLookup строит индекс из загруженного config.Config. nil-safe.
func NewLookup(cfg *config.Config) *Lookup {
	l := &Lookup{byName: make(map[string][]config.ModelParam)}
	if cfg == nil {
		return l
	}
	for i := range cfg.Models {
		l.byName[cfg.Models[i].Name] = cfg.Models[i].Params
	}
	return l
}

// LoadFromFile читает models.json и возвращает Lookup. При ошибке — пустой Lookup.
func LoadFromFile(path string) *Lookup {
	if path == "" {
		return NewLookup(nil)
	}
	cfg, err := config.LoadConfig(path)
	if err != nil {
		return NewLookup(nil)
	}
	return NewLookup(cfg)
}

// nameFromPath нормализует имя модели по полному пути к .gguf.
func nameFromPath(path string) string {
	return strings.TrimSuffix(filepath.Base(path), ".gguf")
}

// ForPath возвращает параметры для модели по полному пути к .gguf.
func (l *Lookup) ForPath(path string) []config.ModelParam {
	if l == nil || l.byName == nil {
		return nil
	}
	return l.byName[nameFromPath(path)]
}

// HasModel сообщает, есть ли в индексе запись для пути.
func (l *Lookup) HasModel(path string) bool {
	if l == nil || l.byName == nil {
		return false
	}
	_, ok := l.byName[nameFromPath(path)]
	return ok
}

// Get возвращает значение по ключу для модели по пути.
func (l *Lookup) Get(path, key string) (any, bool) {
	for _, p := range l.ForPath(path) {
		if p.Key == key {
			return p.Value, true
		}
	}
	return nil, false
}

// Curated — курируемая выжимка ключевых полей для шапки/peek.
type Curated struct {
	Architecture    string
	ContextLength   int64
	SizeLabel       string
	BlockCount      int64
	EmbeddingLength int64
	FFNLength       int64
	HeadCount       int64
	HeadCountKV     int64
	KeyLength       int64
	TopK            int64
	TopP            float64
	Temp            float64
	RopeFreqBase    float64
	ExpertCount     int64
	FileType        string // "Q6_K", "IQ3_M", … (из general.file_type)
	TokenizerModel  string
	HasChatTemplate bool
	HasTokens       bool
	TotalCount      int
	// Architecture-prefix (`gemma4`, `qwen35`, …) — пуст, если general.architecture отсутствует.
	archPrefix string
}

// ExtractCurated собирает Curated из набора параметров. Безопасен к nil.
func ExtractCurated(params []config.ModelParam) Curated {
	c := Curated{TotalCount: len(params)}
	if len(params) == 0 {
		return c
	}
	// 1-й проход: architecture (нужен как префикс)
	for i := range params {
		if params[i].Key == "general.architecture" {
			if s, ok := params[i].Value.(string); ok {
				c.Architecture = s
				c.archPrefix = s
			}
			break
		}
	}
	for i := range params {
		k := params[i].Key
		v := params[i].Value
		switch k {
		case "general.size_label":
			if s, ok := v.(string); ok {
				c.SizeLabel = s
			}
		case "general.sampling.top_k":
			c.TopK = asInt(v)
		case "general.sampling.top_p":
			c.TopP = asFloat(v)
		case "general.sampling.temp":
			c.Temp = asFloat(v)
		case "general.file_type":
			c.FileType = parseFileType(v)
		case "tokenizer.ggml.model":
			if s, ok := v.(string); ok {
				c.TokenizerModel = s
			}
		case "tokenizer.chat_template":
			c.HasChatTemplate = true
		case "tokenizer.ggml.tokens":
			c.HasTokens = true
		}
		// Архитектурно-зависимые ключи: префиксная развязка
		if c.archPrefix != "" && strings.HasPrefix(k, c.archPrefix+".") {
			sub := strings.TrimPrefix(k, c.archPrefix+".")
			switch sub {
			case "context_length":
				c.ContextLength = asInt(v)
			case "block_count":
				c.BlockCount = asInt(v)
			case "embedding_length":
				c.EmbeddingLength = asInt(v)
			case "feed_forward_length":
				c.FFNLength = asInt(v)
			case "attention.head_count":
				c.HeadCount = asInt(v)
			case "attention.head_count_kv":
				c.HeadCountKV = asInt(v)
			case "attention.key_length":
				c.KeyLength = asInt(v)
			case "rope.freq_base":
				c.RopeFreqBase = asFloat(v)
			case "expert_count":
				c.ExpertCount = asInt(v)
			}
		}
	}
	return c
}

// parseFileType извлекает имя квантизации из general.file_type. Значение может
// быть строкой "16 (Q5_K_S)" (после ggufmeta.ExtractParams), голым числом
// (uint*/int*/float64) или уже именем кванта.
func parseFileType(v any) string {
	switch x := v.(type) {
	case string:
		// "16 (Q5_K_S)" → "Q5_K_S"
		if l := strings.Index(x, "("); l >= 0 {
			if r := strings.Index(x[l+1:], ")"); r >= 0 {
				return x[l+1 : l+1+r]
			}
		}
		// Чисто числовая строка — попробуем как uint64
		if u, err := strconv.ParseUint(strings.TrimSpace(x), 10, 64); err == nil {
			if name, ok := ggufmeta.FileTypeName[u]; ok {
				return name
			}
		}
		return x
	default:
		if u := uint64(asInt(v)); u != 0 {
			if name, ok := ggufmeta.FileTypeName[u]; ok {
				return name
			}
		}
	}
	return ""
}

// ForPathCurated — удобный шорткат: lookup + extract.
func (l *Lookup) ForPathCurated(path string) Curated {
	return ExtractCurated(l.ForPath(path))
}

// FormatContext форматирует длину контекста как «131K» / «8K» / «262K» / «1M».
// Делитель десятичный (1000), чтобы 131072 → "131K" — как принято в карточках
// моделей (HF, LM Studio), а не «128K» при делении на 1024.
func FormatContext(n int64) string {
	if n <= 0 {
		return ""
	}
	if n >= 1_000_000 {
		v := float64(n) / 1_000_000
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10) + "M"
		}
		return strconv.FormatFloat(v, 'f', 1, 64) + "M"
	}
	if n >= 1000 {
		// Округление до ближайшего тысячи: 131072 → 131K, 131500 → 132K.
		k := (n + 500) / 1000
		return strconv.FormatInt(k, 10) + "K"
	}
	return strconv.FormatInt(n, 10)
}

// asInt приводит значение к int64. Поддерживает float64/int/int64/json-числа.
func asInt(v any) int64 {
	switch x := v.(type) {
	case nil:
		return 0
	case int:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case float32:
		return int64(x)
	case float64:
		return int64(x)
	case string:
		if i, err := strconv.ParseInt(x, 10, 64); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(x, 64); err == nil {
			return int64(f)
		}
	}
	return 0
}

// asFloat приводит значение к float64.
func asFloat(v any) float64 {
	switch x := v.(type) {
	case nil:
		return 0
	case int:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	case float32:
		return float64(x)
	case float64:
		return x
	case string:
		if f, err := strconv.ParseFloat(x, 64); err == nil {
			return f
		}
	}
	return 0
}

// FormatValue форматирует любое значение GGUF-параметра для показа в UI.
// Длинные строки (>maxStr) сокращаются; массивы — как "<array len=N>".
func FormatValue(v any, maxStr int) string {
	if maxStr <= 0 {
		maxStr = 64
	}
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		if strings.HasPrefix(x, "<array len=") {
			return x
		}
		if len(x) > maxStr {
			return fmt.Sprintf("<text %d B>", len(x))
		}
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(x), 'f', -1, 64)
	case int, int32, int64:
		return fmt.Sprintf("%d", x)
	case []any:
		return fmt.Sprintf("<array len=%d>", len(x))
	default:
		s := fmt.Sprintf("%v", v)
		if len(s) > maxStr {
			return s[:maxStr-1] + "…"
		}
		return s
	}
}

// FlagApplyMap — таблица «GGUF-ключ → CLI long-флаг» (см. план «Apply-from-model»).
// Подставляется в inline-подсказках и при apply.
// Для архитектурно-зависимых ключей используем префикс "*." — будет подставлен на этапе lookup.
var FlagApplyMap = []struct {
	GGUFKey string // полный ключ или с префиксом "*." для подстановки архитектуры
	Flag    string
}{
	{"*.context_length", "--ctx-size"},
	{"general.sampling.temp", "--temp"},
	{"general.sampling.top_p", "--top-p"},
	{"general.sampling.top_k", "--top-k"},
	{"*.rope.freq_base", "--rope-freq-base"},
}

// ResolveGGUFKeyForFlag возвращает значение GGUF-параметра, соответствующего CLI-флагу,
// для модели по пути. Возвращает (raw, displayString, true) при успехе.
func (l *Lookup) ResolveGGUFKeyForFlag(path, flag string) (any, string, bool) {
	params := l.ForPath(path)
	if len(params) == 0 {
		return nil, "", false
	}
	curated := ExtractCurated(params)
	for _, m := range FlagApplyMap {
		if m.Flag != flag {
			continue
		}
		key := m.GGUFKey
		if strings.HasPrefix(key, "*.") {
			if curated.archPrefix == "" {
				continue
			}
			key = curated.archPrefix + "." + strings.TrimPrefix(key, "*.")
		}
		for i := range params {
			if params[i].Key == key {
				return params[i].Value, FormatValue(params[i].Value, 64), true
			}
		}
	}
	return nil, "", false
}
