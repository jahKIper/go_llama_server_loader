package runconfig

import (
	"llama-server-loader/internal/config"
	"llama-server-loader/pkg/modelscan"
)

// ParamRow — одна строка в левой панели (выбранный флаг).
type ParamRow struct {
	Long      string            // "--ctx-size" (без NAME/N/STR)
	Short     string            // "-c" (может быть пустым)
	Key       string            // "ctx_size" — ключ в ModelConfig.Flags
	Value     string            // ввод пользователя (raw)
	Meta      *config.ParamMeta // ссылка в каталог (nil если параметр не из каталога)
	IsDefault bool              // строка добавлена автодефолтами (ComputeModelDefaults), сбрасывается при ручном редактировании
}

// RunConfigAction — что вернуть наружу после закрытия экрана.
type RunConfigAction int

const (
	ActionCancel RunConfigAction = iota // q — выйти из приложения
	ActionRun                           // r — Save+Run
	ActionBack                          // Backspace / клик на таб Models — вернуться к первому экрану
)

// RunConfigResult — результат экрана для caller-а в internal/cli/cli.go.
type RunConfigResult struct {
	Action  RunConfigAction
	Rows    []ParamRow
	Model   *modelscan.Model
	Comment string
}
