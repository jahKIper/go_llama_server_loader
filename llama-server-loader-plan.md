# 🦞 llama-server-loader — Проектный План (v1)

## 🏁 Вехи (Milestones)
| Веха | Deliverable | Критерий готовности |
|------|-------------|-------------------|
| **M1** | `params_ru.json`, структура | JSON валиден, 120+ параметров |
| **M2** | `pkg/modelscan`, `pkg/servercmd` | `go test` pass, coverage ≥70% |
| **M3** | `internal/config`, `models.json` | Сохранение без потерь |
| **M4** | CLI + Web, smoke test | `go build` clean |
| **M5** | Docs, README, build check | Coverage ≥80%, docs ready |

## 📋 Примечания перед началом реализации

1. **params_ru.json** — перед началом разработки M1 выписать все параметры из официальной документации llama.cpp, сформировать JSON с 120+ параметрами, перевести описания на русский
2. **Go модуль** — проинициализировать `go.mod` перед началом разработки
3. **Зависимости** — описать все зависимости: `github.com/tidwall/jsonv`, `golang.org/x/sync`, `github.com/charmbracelet/bubbletea`, `testify` и другие
4. **Параллелизм** — сканер использует ровно 4 горутины (golang.org/x/sync/errgroup)

## 🎯 Цель
Собрать Go-приложение (`llama-server-loader`), которое:
1. Сканирует папку на `.gguf` / `mmproj` файлы
2. Загружает 120+ параметров `llama-server` с переводом на русский
3. Предоставляет CLI + embedded Web UI для выбора и настройки
4. Сохраняет конфиги в `models.json`, валидирует JSON перед запуском
5. Формирует команду `llama-server` и запускает её

## 🛠️ Стек технологий
| Компонент | Выбор | Обоснование |
|-----------|-------|-------------|
| **Язык** | Go 1.21+ | Нативная работа с JSON, горутин, лёгкий бинарник |
| **CLI UI** | `charm` | Terminal UI, идеален для интерактивных меню |
| **Web UI** | `net/http` + embed | Встроенный сервер, фронт на чистом HTML/JS (без сборщиков) |
| **Конфиг** | `params_ru.json` + `models.json` | Разделение справочника и пользовательских данных |
| **Тесты** | `testing` + `testify` | Юниты, интеграция с `llama-server` (mock/real) |

## 📦 Архитектура

### Структура проекта
```
llama-server-loader/
├── cmd/llama-server-loader/main.go (Точка входа)
├── pkg/modelscan (Сканер моделей)
├── pkg/servercmd (Формирование команды)
├── internal/
│   ├── config (Конфигурация, JSON)
│   ├── cli (Terminal UI)
│   └── webui (Embedded Web Server)
├── docs/
└── models/
```

### Ключевые модули
1. **cmd/**: Оркестрация, парсинг флагов CLI.
2. **pkg/modelscan**: Рекурсивный поиск, метаданные, типизация (.gguf / мультимодальные файлы). 4 горутины. Без хэшей.
3. **pkg/servercmd**: Парсинг JSON → формирование `exec.Command`, **мягкая** валидация.
4. **internal/config**: Загрузка `params_ru.json` и `models.json`. Валидация JSON.
5. **internal/cli**: Terminal UI (`charm`) для выбора модели.
6. **internal/webui**: Embedded HTTP сервер для Web UI.

## 📜 Вехи M1 — Core Logic & Scanning

### `pkg/modelscan` (Сканер моделей)
- Рекурсивный поиск `.gguf` файлов в указанном каталоге
- Автоматическое обнаружение мультимодальных файлов (имя файла содержит 'mmproj') в каталоге модели
- Группировка по имени модели (префиксы совпадают)
- Параллелим: ровно 4 горутины (`golang.org/x/sync/errgroup`)
- **Без хэшей** — только метаданные и пути

### `pkg/modelscan` API
| Функция | Описание |
|---------|----------|
| `ScanDir(path string) ([]*Model, error)` | Рекурсивный поиск. Ищем `.gguf` и файлы с именем, содержащим 'mmproj' (мультимодальные файлы, не имеющие собственного расширения), в одном каталоге с моделью. |
| **Без хэшей** | Вычисление `SHA256` удалено из требований. |
| **Параллелим** | 4 горутины для сканирования. |

### `internal/config` (М1 — JSON парсер)
| Функция | Описание |
|---------|----------|
| `ScanDir(path string) ([]*Model, error)` | Рекурсивный поиск. Ищем `.gguf` и файлы с именем, содержащим 'mmproj' (мультимодальные файлы, не имеющие собственного расширения), в одном каталоге с моделью. |
| **Без хэшей** | Вычисление `SHA256` удалено из требований. |
| **Параллелим** | 4 горутины для сканирования. |

### Технические решения M1
| Параметр | Выбор | Обоснование |
|----------|-------|-------------|
| Парсер JSON | `github.com/tidwall/jsonv` | Выдача точной строки/колонны при ошибке парсинга |
| Связка `.gguf` ↔ `mmproj` | Поиск по каталогу + имя (содержит 'mmproj') | Семантическая связь в одном каталоге |
| Нормализация путей | `filepath.ToSlash()` | Кроссплатформенная (Unix/Windows) без костылей |
| `pkg/servercmd` | **Вынесено на M2** | Не входит в scope M1 |

### Задачи на M1 (Ready to Code)
| Задача | Описание | Приоритет |
|--------|----------|-----------|
| **M1-1** | `pkg/modelscan/scanner.go` | 🔴 Critical |
| **M1-2** | `pkg/modelscan/matcher.go` (по каталогу и имени) | 🔴 Critical |
| **M1-3** | `internal/config/loader.go` (jsonv) | 🔴 Critical |
| **M1-4** | `internal/config/validator.go` | 🟡 High |
| **M1-5** | Unit-тесты (mock fs + jsonv errors) | 🟡 High |

### Текущий статус М1

| Задача | Статус | Примечание |
|--------|--------|------------|
| M1-1 | ✅ Выполнена | `pkg/modelscan/scanner.go` — 96.2% покрытие |
| M1-2 | ✅ Выполнена | `pkg/modelscan/matcher.go` — 85.7% покрытие |
| M1-3 | ✅ Выполнена | `internal/config/loader.go` — 100% покрытие |
| M1-4 | ✅ Выполнена | `internal/config/validator.go` — 80% покрытие |
| M1-5 | ✅ Выполнена | 13+9 тестов, 94.6% / 88% |

### Критерии готовности M1
* [x] Сканер рекурсивно проходит папку (4 горутины) — ✅
* [x] `mmproj` найден по каталогу и имени, связан с моделью — ✅
* [x] JSON валидируется с выводом позиции ошибки — ✅ (gjson, jsonv недоступен)
* [x] `pkg/servercmd` НЕ в scope (перенесён в M2) — ✅
* [x] Покрытие ≥70% (`go test ./...`) — ✅ 91.9%

## 📜 M1 (дополнение) — Структура `params_ru.json`

Файл `params_ru.json` не редактируется пользователем — это неизменяемый справочник параметров `llama-server`.

### Структура файла

```json
{
  "version": "1.0",
  "categories": [
    {
      "name": "Загрузка модели",
      "params": [
        {
          "short_flag": "-m",
          "long_flag": "--model FNAME",
          "description_ru": "Путь к файлу модели для загрузки"
        },
        {
          "short_flag": "-t",
          "long_flag": "--threads N",
          "description_ru": "Количество CPU-потоков для генерации (по умолчанию: -1)"
        }
      ]
    }
  ],
  "total_params_count": 120,
  "source_docs": [
    "https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md",
    "https://github.com/ggml-org/llama.cpp/blob/master/tools/cli/README.md",
    "https://github.com/ggml-org/llama.cpp/blob/master/docs/multimodal.md"
  ]
}
```

### Структура `params_ru.json` (не редактируется пользователем)

Файл содержит все параметры llama-server, переведённые на русский язык. Формат:

- **version** — версия схемы JSON-справочника
- **categories[].name** — категория параметров (на русском)
- **categories[].params[]** — массив параметров: `short_flag` (короткий флаг, может быть пустым), `long_flag` (полный флаг с аргументом), `description_ru` (русское описание)
- **total_params_count** — общее число параметров во всех категориях
- **source_docs** — ссылки на официальную документацию llama.cpp, из которой вычитаны параметры

**Важное замечание:** Все параметры необходимо заранее вычитать из официальной документации llama.cpp и записать в этот файл перед началом разработки M1.

## 📜 Вехи M2 — Scanners & Command Builder

### `pkg/modelscan` API
| Функция | Описание |
|---------|----------|
| `ScanDir(path string) ([]*Model, error)` | Рекурсивный поиск. Ищем `.gguf` и файлы с именем, содержащим 'mmproj' (мультимодальные файлы, не имеющие собственного расширения), в одном каталоге с моделью. |

### `pkg/servercmd` API
| Функция | Описание |
|---------|----------|
| `BuildCommand(cfg *Config) (*exec.Cmd, error)` | Парсинг JSON → сборка `os/exec.Command`. |
| **Мягкая валидация** | Только парсинг + базовые checks (конфликты, дубли). |

### Критерии готовности M2
* [ ] `pkg/modelscan` — рекурсивный сканер, 4 горутины, без хэшей
* [ ] `pkg/servercmd` — JSON → exec.Command
* [ ] `go test` pass
* [ ] Coverage ≥70%

## 📜 Вехи M3 — Config & JSON

### `internal/config` API
| Функция | Описание |
|---------|----------|
| `LoadConfig(path string) (*Config, error)` | Загрузка и парсинг `models.json`. |
| `SaveConfig(cfg *Config, path string) error` | Сохранение структуры конфига. |
| `ValidateConfig(cfg *Config) error` | Валидация структуры конфига. |

### `models.json` (Пользовательский конфиг)
```json
{
  "version": "1.0",
  "scan_paths": ["/home/user/models", "/opt/llm"],
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

### `models.json` (Структура)
| Поле | Тип | Обязательность | Описание |
|------|-----|----------------|----------|
| `name` | string | Обязательно | Имя модели (извлекается из имени файла без расширения). |
| `mmproj_on` | boolean | Обязательно | `true`, если mmproj-файл найден, иначе `false`. |
| `mmproj_path` | string | Опционально | Путь к mmproj-файлу (если найден). |
| `flags` | object | Обязательно | Профиль запуска (CLI-аргументы для `llama-server`). |

### Критерии готовности M3
* [ ] `LoadConfig` — загрузка и парсинг `models.json`
* [ ] `SaveConfig` — сохранение без потерь
* [ ] `ValidateConfig` — валидация структуры
* [ ] Сохранение без потерь

## 📜 Вехи M4 — CLI + Web UI

### CLI API

| Команда/Флаг | Описание | Пример |
|-------------|----------|--------|
| `--scan-dir` | Каталог для сканирования моделей | `--scan-dir=/models` |
| `--model` | Имя модели для запуска | `--model=gemma-4-...` |
| `--threads` | Количество потоков | `--threads=16` |
| `--temperature` | Температура генерации | `--temperature=0.8` |
| `--start-webui` | Запуск Web UI отдельно | `--start-webui --port=8080` |
| `--save-config` | Сохранение конфига в файл | `--save-config=models.json` |
| `--generate-params` | Генерация params из docs | `--generate-params --output=custom_params.json` |

### Архитектура Decoupled
CLI и Web-сервер работают независимо друг от друга.
CLI не зависит от Web, Web не зависит от CLI.
Web-сервер запускается отдельно, CLI только запускает его при выборе модели.

### `internal/cli`
- Интерфейс: `charm` (не bubbletea)
- Интерактивное меню выбора модели
- При выборе модели — запуск Web-сервера

### `internal/webui`
- Встроенный HTTP сервер (`net/http` + embed)
- Фронт на чистом HTML/JS/CSS (без сборщиков)
- Раздельный интерфейс от CLI
- JSON-RPC 2.0 над HTTP

### Черновик API (JSON-RPC 2.0)
| Method | Direction | Description |
|--------|-----------|-------------|
| `getModels` | Client → Server | Запрос списка моделей |
| `startServer` | Client → Server | Запуск `llama-server` с выбранными флагами |
| `getStatus` | Client → Server | Статус сервера (running/stopped/error) |
| `shutdown` | Client → Server | Корректное завершение сервера |

### Smoke test
- `go build` чистый
- Smoke test CLI + Web
- Тесты UI убраны (по запросу)

### Критерии готовности M4
- [] CLI на `charm`
- [] Web-сервер standalone (не завязан на CLI), запускается самостоятельно отдельной командой
- [] JSON-RPC 2.0 над HTTP
- [] Фронт на чистом стеке (HTML/JS/CSS)

## 📜 Вехи M5 — Docs & Final Build

### Документация
- `README.md` в корне проекта
- `docs/` — дополнительные docs файлы

### Качество
- Покрытие тестами ≥80% (`go test ./... -cover`)

### ARM64 Build Check
- `go build` на ARM64 (только build, без тестов)

### Задачи M5 (Ready to Code)
| Задача | Описание | Приоритет |
|--------|----------|-----------|
| **M5-1** | `README.md` — Setup, Usage, API docs | 🔴 Critical |
| **M5-2** | `docs/` — форматы JSON | 🔴 High |
| **M5-3** | `go test -cover` для достижения 80% | 🟡 High |
| **M5-4** | `go build` на ARM64 | 🟡 High |

### Критерии готовности M5
* [ ] README.md с Setup, Usage, API docs
* [ ] docs/ с форматами JSON
* [ ] `go test -cover` ≥80%
* [ ] `go build` на ARM64 успешен

## 🧪 Стратегия тестов

### Юнит-тесты
| Модуль | Тип тестов | Покрытие |
|--------|------------|----------|
| `pkg/modelscan` | Юниты (mock filesystem `os.TempDir`) | 80% |
| `pkg/servercmd` | Юниты (mock `exec.Command`), парсинг JSON | 75% |
| `internal/config` | Валидация схем, edge-case JSON | 70% |
| **Итого** | **≥70%** | **go test ./... -cover** |

**Инструменты:** `testing`, `testify`, `go mock` (internal), `subprocess` для smoke-testов.

## ⚠️ Риски и Edge Cases
| Риск | Влияние | Митигация | Fallback |
|------|---------|-----------|----------|
| `llama-server` нет в PATH | High | Pre-flight check | Graceful error |
| JSON невалидный | High | Schema validation | Auto-correct |
| Web UI конфликт | Med | Раздельный интерфейс | CLI-only mode |
| ARM64 compatibility | High | Cross-compilation | Report gap |
