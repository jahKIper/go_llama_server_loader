# 🦞 Детальный план разработки llama-server-loader

## 📊 Текущий статус проекта

| Веха | Статус | Описание |
|------|--------|----------|
| M1 | ✅ Выполнено | `params_ru.json` создан, `internal/config` реализован, 120+ параметров |
| M2 | ❌ Не реализовано | `pkg/modelscan` и `pkg/servercmd` отсутствуют |
| M3 | ⚠️ Частично | `internal/config` реализован, но `models.json` отсутствует |
| M4 | 🟡 В процессе | CLI частично реализован (Этап 5), Web UI отсутствует |
| M5 | ✅ Выполнено | Docs и README созданы (Этап 8) |

## ✅ Выполненные этапы

### Этап 5. internal/cli/cli.go — Выполнено

**Статус:** Реализована базовая структура CLI модуля

**Реализовано:**
- ✅ Парсинг CLI флагов: `--scan-dir`, `--model`, `--threads`, `--temperature`
- ✅ Поддержка флагов: `--start-webui`, `--port`, `--save-config`, `--generate-params`, `--output`
- ✅ Валидация входных параметров (каталог, температура 0-2, threads)
- ✅ Генерация параметр-файла (`--generate-params`)
- ✅ Форматирование размера файлов (`formatSize`)
- ✅ Форматирование имени модели (`FormatModelName`)
- ✅ Сборка строки флагов для команды (`BuildFlagsString`)
- ✅ Структура `ListItem` для будущего UI на charm
- ✅ Покрытие тестами: 82.3%

**TODO для полной реализации:**
- Интеграция с `pkg/modelscan.Scanner` для сканирования директорий
- Реализация интерактивного терминального UI на `charmbracelet/bubbles/list` + `tea`
- Запуск Web UI сервера (`--start-webui`)

## 🏗️ Архитектура проекта

```
llama-server-loader/
├── cmd/llama-server-loader/main.go          - Точка входа
├── pkg/modelscan/
│   ├── scanner.go                           - Рекурсивный сканер (4 горутины)
│   └── matcher.go                           - Связка .gguf ↔ mmproj
├── pkg/servercmd/
│   └── servercmd.go                         - BuildCommand()
├── internal/
│   ├── config/
│   │   ├── loader.go                        - LoadParams (✅ реализовано)
│   │   ├── validator.go                     - ValidateParams (✅ реализовано)
│   │   └── config.go                        - LoadConfig/SaveConfig (❌ нет)
│   ├── cli/
│   │   └── cli.go                           - Terminal UI на charm (❌ нет)
│   └── webui/
│       └── webui.go                         - Embedded HTTP + JSON-RPC 2.0 (❌ нет)
├── docs/                                    - Документация (❌ нет)
├── params_ru.json                           - Справочник параметров (✅ реализовано)
├── go.mod                                   - Модуль Go (✅ инициализирован)
└── README.md                                - Документация (❌ нет)
```

## 📋 План реализации (по приоритету)

### Этап 1. pkg/modelscan/scanner.go (Priority: 🔴 Critical)

**Цель:** Рекурсивный сканер с 4 горутинами

**Требования:**
- Использовать `golang.org/x/sync/errgroup` с `SetLimit(4)` для ровно 4 горутин
- Рекурсивное сканирование директории на `.gguf` файлы
- Без хэшей — только метаданные и пути
- Нормализация путей через `filepath.ToSlash()` для кроссплатформенности
- Структура `Model` и `ScanResult`
- Покрытие тестами ≥96%

**API:**
```go
package modelscan

type Model struct {
    Name     string `json:"name"`
    Path     string `json:"path"`
    IsMMProj bool   `json:"is_mmproj"`
    Size     int64  `json:"size"`
}

type ScanResult struct {
    Models  []*Model `json:"models"`
    MMModels []*Model `json:"mm_models"`
    Errors  []error  `json:"-"`
}

func ScanDir(root string) (*ScanResult, error)
```

**Зависимости:** `golang.org/x/sync/errgroup`, `sync`

---

### Этап 2. pkg/modelscan/matcher.go (Priority: 🔴 Critical)

**Цель:** Связка `.gguf` ↔ `mmproj` по каталогу и имени

**Требования:**
- Поиск по каталогу и имени файла
- Файлы с именем, содержащим 'mmproj', считаются мультимодальными
- Связывание `.gguf` моделей с соответствующими `mmproj` файлами
- Группировка по префиксу имени (общие префиксы = одна модель)
- Покрытие тестами ≥85%

**API:**
```go
package modelscan

func MatchMMProj(models []*Model) ([]*Model, error)
func BaseContainsMMProj(base string) bool
```

**Зависимости:** нет (чистая логика)

---

### Этап 3. pkg/servercmd/servercmd.go (Priority: 🔴 Critical)

**Цель:** BuildCommand() для формирования exec.Command

**Требования:**
- Парсинг JSON → формирование `os/exec.Command`
- Мягкая валидация: парсинг + базовые checks (конфликты, дубли)
- Проверка наличия `llama-server` в PATH
- Graceful error при отсутствии сервера

**API:**
```go
package servercmd

type Config struct {
    Version string            `json:"version"`
    ScanPaths []string        `json:"scan_paths"`
    Models  []ModelConfig     `json:"models"`
}

type ModelConfig struct {
    Name      string            `json:"name"`
    ModelPath string            `json:"model_path"`
    MMProjPath string           `json:"mmproj_path,omitempty"`
    MMProjOn  bool              `json:"mmproj_on"`
    Size      int64             `json:"size"`
    LastScan  string            `json:"last_scan"`
    Flags     map[string]any   `json:"flags"`
}

func BuildCommand(cfg *Config) (*exec.Cmd, error)
```

**Зависимости:** `os/exec`

---

### Этап 4. internal/config/config.go (Priority: 🔴 Critical)

**Цель:** LoadConfig/SaveConfig для models.json

**Требования:**
- `LoadConfig(path string) (*Config, error)` — загрузка и парсинг `models.json`
- `SaveConfig(cfg *Config, path string) error` — сохранение без потерь
- `ValidateConfig(cfg *Config) error` — валидация структуры
- Сохранение JSON без потери данных (lossless)
- Покрытие тестами ≥80%

**Структура models.json:**
```json
{
  "version": "1.0",
  "scan_paths": ["/path/to/models", "/other/path"],
  "models": [
    {
      "name": "model_name",
      "model_path": "/path/to/file.gguf",
      "mmproj_path": "/path/to/mmproj.gguf",
      "mmproj_on": true,
      "size": 4200000000,
      "last_scan": "2024-04-30T15:00:00Z",
      "flags": {
        "model": "/path/to/file.gguf",
        "ctx_len": 4096,
        "gpu_layers": 24
      }
    }
  ]
}
```

**Зависимости:** `encoding/json`, `os`

---

### Этап 5. internal/cli/cli.go (Priority: 🟡 High)

**Цель:** Terminal UI на charm

**Требования:**
- Интерфейс на `charm` (не bubbletea)
- Интерактивное меню выбора модели
- При выборе модели — запуск Web-сервера
- Парсинг CLI флагов: `--scan-dir`, `--model`, `--threads`, `--temperature`
- Поддержка флагов: `--start-webui`, `--save-config`, `--generate-params`

**CLI флаги:**
| Флаг | Описание | Пример |
|------|----------|--------|
| `--scan-dir` | Каталог для сканирования | `--scan-dir=/models` |
| `--model` | Имя модели для запуска | `--model=gemma-4` |
| `--threads` | Количество потоков | `--threads=16` |
| `--temperature` | Температура | `--temperature=0.8` |
| `--start-webui` | Запуск Web UI | `--start-webui --port=8080` |
| `--save-config` | Сохранение конфига | `--save-config=models.json` |
| `--generate-params` | Генерация params | `--generate-params --output=custom.json` |

**Зависимости:** `github.com/charmbracelet/*`

---

### Этап 6. internal/webui/webui.go (Priority: 🟡 High)

**Цель:** Embedded HTTP сервер с JSON-RPC 2.0

**Требования:**
- Встроенный HTTP сервер (`net/http` + embed)
- Фронт на чистом HTML/JS/CSS (без сборщиков)
- Раздельный интерфейс от CLI
- JSON-RPC 2.0 над HTTP
- Методы: `getModels`, `startServer`, `getStatus`, `shutdown`

**JSON-RPC 2.0 API:**
| Method | Direction | Description |
|--------|-----------|-------------|
| `getModels` | Client → Server | Запрос списка моделей |
| `startServer` | Client → Server | Запуск llama-server |
| `getStatus` | Client → Server | Статус сервера |
| `shutdown` | Client → Server | Корректное завершение |

**Зависимости:** `net/http`, `embed`, `encoding/json`

---

### Этап 7. cmd/llama-server-loader/main.go (Priority: 🟡 High)

**Цель:** Точка входа

**Требования:**
- Оркестрация всех модулей
- Парсинг флагов CLI
- Интеграция с `pkg/modelscan`, `pkg/servercmd`, `internal/config`, `internal/cli`, `internal/webui`
- Обработка ошибок и graceful shutdown

**Зависимости:** все вышеперечисленные модули

---

### Этап 8. README.md и docs/ (Priority: 🟡 High)

**Требования:**
- `README.md` в корне проекта с Setup, Usage, API docs
- `docs/` — дополнительные docs файлы
- Описание форматов JSON (`params_ru.json`, `models.json`)
- Примеры использования
- Описание JSON-RPC 2.0 API

**Файлы:**
- `README.md`
- `docs/params_format.md`
- `docs/models_json_format.md`
- `docs/jsonrpc_api.md`

### ✅ Этап 8. README.md и docs/ — Выполнено

**Статус:** Документация создана полностью

**Реализовано:**
- ✅ `README.md` в корне проекта с Setup, Usage, API docs
- ✅ `docs/params_format.md` — формат params_ru.json
- ✅ `docs/models_json_format.md` — формат models.json
- ✅ `docs/jsonrpc_api.md` — JSON-RPC 2.0 API документация

**Содержание README.md:**
- Описание возможностей приложения
- Установка из исходников
- CLI флаги и примеры использования
- Архитектура проекта с диаграммой директорий
- Тестирование и целевое покрытие
- Риски и edge cases
- Вехи проекта

**Содержание docs/:**
- params_format.md — структура, поля, категории, примеры использования в Go
- models_json_format.md — структура конфигурации, API использования, правила генерации имён
- jsonrpc_api.md — формат запросов/ответов, методы API, клиентская реализация на JavaScript

---

## 📊 Стратегия тестирования

| Модуль | Тип тестов | Покрытие |
|--------|------------|----------|
| `pkg/modelscan` | Юниты (mock filesystem) | 80%+ |
| `pkg/servercmd` | Юниты (mock exec.Command) | 75%+ |
| `internal/config` | Валидация схем, edge-case | 70%+ |
| **Итого** | **≥70%** | **go test ./... -cover** |

**Инструменты:** `testing`, `testify`, `go mock`, `subprocess` для smoke-testов

---

## ⚠️ Риски и Edge Cases

| Риск | Влияние | Митигация | Fallback |
|------|---------|-----------|----------|
| `llama-server` нет в PATH | High | Pre-flight check | Graceful error |
| JSON невалидный | High | Schema validation | Auto-correct |
| Web UI конфликт | Med | Раздельный интерфейс | CLI-only mode |
| ARM64 compatibility | High | Cross-compilation | Report gap |

---

## 🚀 Порядок реализации

1. **pkg/modelscan/scanner.go** — рекурсивный сканер с 4 горутинами
2. **pkg/modelscan/matcher.go** — связка .gguf ↔ mmproj
3. **pkg/servercmd/servercmd.go** — BuildCommand()
4. **internal/config/config.go** — LoadConfig/SaveConfig для models.json
5. **internal/cli/cli.go** — Terminal UI на charm
6. **internal/webui/webui.go** — Embedded HTTP сервер
7. **cmd/llama-server-loader/main.go** — точка входа
8. **README.md и docs/** — документация

---

## 📝 Примечания

- Все пути нужно нормализовать через `filepath.ToSlash()` для кроссплатформенности
- Сканер использует ровно 4 горутины (`golang.org/x/sync/errgroup`)
- CLI и Web-сервер работают независимо друг от друга
- Web-сервер запускается отдельно, CLI только запускает его при выборе модели
- Фронтенд на чистом HTML/JS/CSS без сборщиков
- Покрытие тестами ≥80% (`go test ./... -cover`)
- Каталог с моделями `D:\progs\lmstudio\models\` можно использовать для тестов сканера моделей
