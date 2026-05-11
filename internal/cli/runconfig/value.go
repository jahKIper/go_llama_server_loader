package runconfig

import (
	"strconv"
	"strings"
)

// ParseFlagValue парсит строковое значение флага в типизированное.
// Порядок: bool ("true"/"false") → int → float64 → string. Пустая строка → "".
func ParseFlagValue(raw string) any {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	switch strings.ToLower(s) {
	case "true":
		return true
	case "false":
		return false
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

// BuildFlagsMap собирает все флаги (включая IsDefault) для запуска llama-server.
// Гарантированно добавляет "model": modelPath, если не задан явно через Key.
func BuildFlagsMap(rows []ParamRow, modelPath string) map[string]any {
	m := make(map[string]any, len(rows)+1)
	for _, r := range rows {
		if r.Key == "" {
			continue
		}
		m[r.Key] = ParseFlagValue(r.Value)
	}
	if _, ok := m["model"]; !ok {
		m["model"] = modelPath
	}
	return m
}

// BuildSavedFlagsMap собирает флаги для сохранения в models.json: пропускает
// строки с IsDefault=true (нетронутые автодефолты — они и так подмешаются
// при следующем открытии экрана). "model" добавляется всегда.
func BuildSavedFlagsMap(rows []ParamRow, modelPath string) map[string]any {
	m := make(map[string]any, len(rows)+1)
	for _, r := range rows {
		if r.Key == "" || r.IsDefault {
			continue
		}
		m[r.Key] = ParseFlagValue(r.Value)
	}
	if _, ok := m["model"]; !ok {
		m["model"] = modelPath
	}
	return m
}
