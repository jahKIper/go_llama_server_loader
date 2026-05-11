package runconfig

import (
	"llama-server-loader/internal/config"
	"llama-server-loader/pkg/modelscan"
)

// ParamRow — одна строка в левой панели (выбранный флаг).
type ParamRow struct {
	Long  string            // "--ctx-size" (без NAME/N/STR)
	Short string            // "-c" (может быть пустым)
	Key   string            // "ctx_size" — ключ в ModelConfig.Flags
	Value string            // ввод пользователя (raw)
	Meta  *config.ParamMeta // ссылка в каталог (nil если параметр не из каталога)
}

// RunConfigAction — что вернуть наружу после закрытия экрана.
type RunConfigAction int

const (
	ActionCancel RunConfigAction = iota // Esc/q — ничего не делать
	ActionRun                           // r — Save+Run
)

// RunConfigResult — результат экрана для caller-а в internal/cli/cli.go.
type RunConfigResult struct {
	Action RunConfigAction
	Rows   []ParamRow
	Model  *modelscan.Model
}
