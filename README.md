# 🦞 llama-server-loader

Go-приложение для сканирования GGUF-моделей, интерактивного выбора и запуска `llama-server` из llama.cpp. Включает терминальный UI на Bubble Tea, embedded Web UI с JSON-RPC 2.0, парсер GGUF-метаданных и две утилиты для обогащения и анализа моделей.

## 📋 Содержание

- [Возможности](#-возможности)
- [Установка](#-установка)
- [Бинарники](#-бинарники)
- [Использование](#-использование)
- [Форматы конфигурации](#-форматы-конфигурации)
- [JSON-RPC 2.0 API](#-json-rpc-20-api)
- [Архитектура проекта](#-архитектура-проекта)
- [Тестирование](#-тестирование)

## ✨ Возможности

- 🔍 **Параллельное рекурсивное сканирование** директорий на `.gguf` и `mmproj`-файлы
- 🎯 **Автоматическое связывание** обычных моделей с мультимодальными проекторами (`mmproj`)
- 🖥️ **Терминальный UI** на [Bubble Tea](https://github.com/charmbracelet/bubbletea) v2: список моделей, фильтрация, peek, конфигуратор запуска
- 🌐 **Embedded Web UI** для визуальной настройки через браузер
- 🧠 **Парсер GGUF-метаданных** на базе `gpustack/gguf-parser-go` с русскими описаниями ключей
- 💾 **Сохранение/загрузка конфигураций** в `models.json` (готовые наборы параметров для каждой модели)
- 📖 **120+ параметров** `llama-server` с переводом на русский (`params_ru.json`)
- 🧱 **Утилита `gguf-layers`** для послойного разбора тензоров GGUF (типы квантизации, размерности, байты)
- 🧪 **Утилита `gguf-enrich`** для заполнения `models.json` метаданными GGUF

## 📦 Установка

### Требования

- Go **1.26+** (см. `go.mod`)
- `llama-server` в `PATH`

### Сборка

```bash
git clone https://github.com/jahKIper/go_llama_server_loader
cd go_llama_server_loader

# Все три бинарника
go build -o bin/llama-server-loader ./cmd/llama-server-loader
go build -o bin/gguf-enrich         ./cmd/gguf-enrich
go build -o bin/gguf-layers         ./cmd/gguf-layers
```

Или установить в `$GOBIN`:

```bash
go install ./cmd/...
```

## 🧰 Бинарники

| Бинарь | Назначение |
|--------|------------|
| `llama-server-loader` | Главное приложение: TUI/Web UI для сканирования и запуска моделей |
| `gguf-enrich` | Парсит GGUF-файлы и заполняет поле `params` в `models.json` |
| `gguf-layers` | Печатает послойный разбор тензоров одного GGUF-файла |

## 🚀 Использование

### Главное приложение

```bash
# Интерактивный TUI: сканирование + выбор + конфигуратор запуска
./llama-server-loader --scan-dir=/path/to/models

# Старт сразу с указанной моделью
./llama-server-loader --scan-dir=/models --model=gemma-4 --threads=16 --temperature=0.8

# Только Web UI (без сканирования из CLI)
./llama-server-loader --start-webui --port=8080

# Сканирование + Web UI
./llama-server-loader --scan-dir=/models --start-webui --port=8080

# Сгенерировать каталог параметров
./llama-server-loader --generate-params --output=params_ru.json
```

### CLI флаги

| Флаг | Описание |
|------|----------|
| `--scan-dir <path>` | Каталог для рекурсивного сканирования `.gguf` / `mmproj` |
| `--model <name>` | Имя модели для запуска (из найденных в `--scan-dir`) |
| `--threads <n>` | Количество CPU-потоков (по умолчанию — авто-определение) |
| `--temperature <f>` | Температура сэмплирования |
| `--start-webui` | Запустить встроенный Web UI |
| `--port <n>` | Порт Web UI (по умолчанию 8080) |
| `--save-config <file>` | Сохранить конфигурацию в файл |
| `--generate-params` | Сгенерировать каталог параметров |
| `--output <file>` | Куда писать сгенерированные параметры |
| `--params-file <file>` | Путь к `params_ru.json` |
| `--models-config <file>` | Путь к `models.json` (по умолчанию `./models.json`) |
| `-h`, `--help` | Справка |

Поддерживаются обе формы: `--flag value` и `--flag=value`.

### `gguf-enrich`

```bash
# Перечитать все модели из models.json, дополнить поле "params" метаданными GGUF
./gguf-enrich -f models.json -o models.json
```

### `gguf-layers`

```bash
# Полная разбивка по блокам blk.N.* + non-block тензоры
./gguf-layers ./models/gemma-4-it-q4_k_m.gguf

# Подробно: каждый тензор слоя
./gguf-layers ./models/gemma-4-it-q4_k_m.gguf -v

# Только агрегированная сводка
./gguf-layers ./models/gemma-4-it-q4_k_m.gguf -summary
```

## 📁 Форматы конфигурации

| Файл | Назначение | Документация |
|------|------------|--------------|
| `params_ru.json` | Каталог параметров `llama-server` с русскими описаниями | [docs/params_format.md](docs/params_format.md) |
| `models.json` | Сохранённые пресеты моделей и их параметров запуска | [docs/models_json_format.md](docs/models_json_format.md) |

## 🔌 JSON-RPC 2.0 API

Web UI общается с сервером по JSON-RPC 2.0 над HTTP.

| Метод | Описание |
|-------|----------|
| `getModels` | Список найденных моделей |
| `startServer` | Запуск `llama-server` с выбранной конфигурацией |
| `getStatus` | Текущий статус сервера |
| `shutdown` | Корректное завершение |

Полная спецификация — [docs/jsonrpc_api.md](docs/jsonrpc_api.md).

## 🏗️ Архитектура проекта

```
go_llama_server_loader/
├── cmd/
│   ├── llama-server-loader/   # Главный бинарник (TUI + Web UI)
│   ├── gguf-enrich/           # Обогащение models.json метаданными GGUF
│   └── gguf-layers/           # Анализ тензоров GGUF по слоям
├── pkg/
│   ├── modelscan/             # Параллельный сканер .gguf и mmproj + matcher
│   └── servercmd/             # Формирование exec.Command для llama-server
├── internal/
│   ├── cli/                   # Парсинг флагов + Bubble Tea TUI
│   │   ├── cli.go             # Точка входа CLI, ParseFlags
│   │   ├── modelparams/       # Lookup по каталогу параметров
│   │   ├── runconfig/         # Экран конфигурации запуска модели
│   │   └── uistyle/           # Layout и стили lipgloss
│   ├── config/                # Loader / Validator / SaveConfig
│   ├── ggufmeta/              # Обёртка над gguf-parser-go + рус. описания
│   └── webui/                 # Embedded HTTP + JSON-RPC 2.0 + static UI
├── docs/                      # Спецификации форматов и API
├── plans/                     # Планы/RFC по развитию
├── doc_cashe/                 # Кэш документации llama.cpp
├── params_ru.json             # Каталог параметров (120+)
├── models.json                # Пресеты моделей (рантайм-конфиг)
├── go.mod
└── README.md
```

### Ключевые модули

| Модуль | Назначение |
|--------|------------|
| `pkg/modelscan` | Параллельный поиск GGUF/mmproj, группировка по модели |
| `pkg/servercmd` | Сборка `exec.Command` для `llama-server` с валидацией |
| `internal/cli` | Парсинг флагов, TUI на Bubble Tea v2 (список, фильтр, peek, runconfig) |
| `internal/cli/runconfig` | Экран настройки параметров запуска конкретной модели |
| `internal/config` | Загрузка/валидация `params_ru.json` и `models.json` |
| `internal/ggufmeta` | Чтение GGUF KV-метаданных, описания ключей на русском |
| `internal/webui` | Embedded HTTP-сервер + JSON-RPC 2.0 + статический Web UI |

### Основные зависимости

- [`github.com/charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea), `bubbles`, `lipgloss` — терминальный UI
- [`github.com/gpustack/gguf-parser-go`](https://github.com/gpustack/gguf-parser-go) — парсер GGUF
- [`github.com/tidwall/gjson`](https://github.com/tidwall/gjson) — выборки из JSON
- [`github.com/stretchr/testify`](https://github.com/stretchr/testify) — тесты
- `golang.org/x/sync` — примитивы синхронизации

## 🧪 Тестирование

```bash
# Все тесты
go test ./... -v

# Покрытие по проекту
go test ./... -cover

# По модулям
go test ./pkg/modelscan      -cover
go test ./pkg/servercmd      -cover
go test ./internal/config    -cover
go test ./internal/cli       -cover
go test ./internal/cli/runconfig -cover
go test ./internal/webui     -cover
```

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте feature-ветку (`git checkout -b feature/xxx`)
3. Закоммитьте изменения (`git commit -am 'Add xxx'`)
4. Запушьте (`git push origin feature/xxx`)
5. Откройте Pull Request

## 📄 Лицензия

MIT License. См. файл [`LICENSE`](LICENSE).

## 📞 Поддержка

- Issues: [GitHub Issues](https://github.com/jahKIper/go_llama_server_loader/issues)
- Документация: [docs/](docs/)
