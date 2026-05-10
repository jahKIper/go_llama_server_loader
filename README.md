# 🦞 llama-server-loader

Go-приложение для сканирования, настройки и запуска `llama-server` с поддержкой CLI и Web UI.

## 📋 Содержание

- [Возможности](#-возможности)
- [Установка](#-установка)
- [Использование](#-использование)
- [Форматы конфигурации](#-форматы-конфигурации)
- [JSON-RPC 2.0 API](#-json-rpc-20-api)
- [Архитектура проекта](#-архитектура-проекта)
- [Разработка](#-разработка)
- [Тестирование](#-тестирование)

## ✨ Возможности

- 🔍 **Рекурсивное сканирование** директорий на `.gguf` и `mmproj` файлы (4 горутины)
- 🎯 **Автоматическое связывание** моделей с мультимодальными промптами (`mmproj`)
- 🌐 **Embedded Web UI** для визуальной настройки через браузер
- 💻 **CLI интерфейс** с парсингом флагов
- 💾 **Сохранение конфигураций** в `models.json`
- 📖 **120+ параметров** `llama-server` с переводом на русский
- ⚡ **Формирование команды** запуска `llama-server` с валидацией

## 📦 Установка

### Требования

- Go 1.21+
- `llama-server` в `PATH`

### Из исходников

```bash
# Клонирование репозитория
git clone <repository-url>
cd llama-server-loader

# Сборка
go build -o llama-server-loader ./cmd/llama-server-loader

# Установка
go install ./cmd/llama-server-loader
```

### Зависимости

```go
require (
    github.com/stretchr/testify v1.11.1
    github.com/tidwall/gjson v1.18.0
    golang.org/x/sync v0.20.0
)
```

## 🚀 Использование

### Базовый запуск

```bash
# Сканирование директории с моделями
./llama-server-loader --scan-dir=/path/to/models

# Запуск с указанием модели
./llama-server-loader --scan-dir=/models --model=gemma-4

# Настройка потоков и температуры
./llama-server-loader --scan-dir=/models --threads=16 --temperature=0.8
```

### CLI флаги

| Флаг | Описание | Пример |
|------|----------|--------|
| `--scan-dir` | Каталог для сканирования | `--scan-dir=/models` |
| `--model` | Имя модели для запуска | `--model=gemma-4` |
| `--threads` | Количество потоков CPU | `--threads=16` |
| `--temperature` | Температура генерации | `--temperature=0.8` |
| `--start-webui` | Запуск Web UI | `--start-webui --port=8080` |
| `--save-config` | Сохранение конфига | `--save-config=models.json` |
| `--generate-params` | Генерация params файла | `--generate-params --output=custom.json` |

### Запуск Web UI

```bash
# Запуск с встроенным веб-интерфейсом
./llama-server-loader --start-webui --port=8080

# Сканирование + Web UI
./llama-server-loader --scan-dir=/models --start-webui --port=8080
```

### Генерация параметр-файла

```bash
# Генерация JSON файла с параметрами
./llama-server-loader --generate-params --output=custom.json
```

## 📁 Форматы конфигурации

### params_ru.json

Справочник параметров `llama-server` с русскими описаниями. Подробнее в [docs/params_format.md](docs/params_format.md).

### models.json

Пользовательская конфигурация для сохранения выбранных моделей и их настроек. Подробнее в [docs/models_json_format.md](docs/models_json_format.md).

### Экран «Параметры запуска модели»

TUI-экран для интерактивного подбора флагов `llama-server` после выбора модели. Подробнее в [docs/run_config_screen.md](docs/run_config_screen.md).

## 🔌 JSON-RPC 2.0 API

Web UI использует JSON-RPC 2.0 над HTTP для коммуникации с сервером.

### Методы

| Method | Direction | Описание |
|--------|-----------|----------|
| `getModels` | Client → Server | Запрос списка моделей |
| `startServer` | Client → Server | Запуск llama-server |
| `getStatus` | Client → Server | Получение статуса сервера |
| `shutdown` | Client → Server | Корректное завершение работы |

Подробнее в [docs/jsonrpc_api.md](docs/jsonrpc_api.md).

## 🏗️ Архитектура проекта

```
llama-server-loader/
├── cmd/llama-server-loader/main.go       # Точка входа
├── pkg/modelscan/
│   ├── scanner.go                        # Рекурсивный сканер (4 горутины)
│   └── matcher.go                        # Связка .gguf <-> mmproj
├── pkg/servercmd/
│   └── servercmd.go                      # BuildCommand()
├── internal/
│   ├── config/
│   │   ├── loader.go                     # LoadParams / LoadConfig
│   │   ├── validator.go                  # ValidateParams
│   │   └── config.go                     # LoadConfig/SaveConfig
│   ├── cli/
│   │   └── cli.go                        # Terminal UI на charm
│   └── webui/
│       ├── embed.go                      # Встраивание статических файлов
│       └── webui.go                      # Embedded HTTP + JSON-RPC 2.0
├── docs/
│   ├── params_format.md                  # Формат params_ru.json
│   ├── models_json_format.md             # Формат models.json
│   └── jsonrpc_api.md                    # JSON-RPC 2.0 API
├── params_ru.json                        # Справочник параметров (120+)
├── go.mod                                # Модуль Go
└── README.md                             # Этот файл
```

### Ключевые модули

| Модуль | Назначение |
|--------|------------|
| `pkg/modelscan` | Рекурсивный поиск `.gguf` и `mmproj` файлов, группировка по модели |
| `pkg/servercmd` | Формирование `exec.Command` для запуска `llama-server` |
| `internal/config` | Загрузка и валидация конфигураций |
| `internal/cli` | Terminal UI для интерактивного выбора моделей |
| `internal/webui` | Embedded HTTP сервер с Web UI и JSON-RPC 2.0 |

## 🧪 Тестирование

```bash
# Запуск всех тестов
go test ./... -v

# Покрытие тестами
go test ./... -cover

# Детальная покрытие по модулям
go test ./pkg/modelscan -cover
go test ./pkg/servercmd -cover
go test ./internal/config -cover
go test ./internal/cli -cover
```

### Целевое покрытие

| Модуль | Тип тестов | Покрытие |
|--------|------------|----------|
| `pkg/modelscan` | Юниты (mock filesystem) | 80%+ |
| `pkg/servercmd` | Юниты (mock exec.Command) | 75%+ |
| `internal/config` | Валидация схем, edge-case | 70%+ |
| **Итого** | **≥70%** | **go test ./... -cover** |

## 🛡️ Риски и Edge Cases

| Риск | Влияние | Митигация |
|------|---------|-----------|
| `llama-server` нет в PATH | High | Pre-flight check + graceful error |
| JSON невалидный | High | Schema validation + auto-correct |
| Web UI конфликт портов | Med | Раздельный интерфейс + CLI-only mode |
| ARM64 совместимость | High | Cross-compilation support |

## 📜 Вехи проекта

| Веха | Статус | Описание |
|------|--------|----------|
| M1 | ✅ Выполнено | `params_ru.json`, структура проекта, 120+ параметров |
| M2 | ⚠️ В процессе | `pkg/modelscan` и `pkg/servercmd` |
| M3 | ⚠️ В процессе | `internal/config`, `models.json` |
| M4 | 🟡 В процессе | CLI + Web UI |
| M5 | ✅ Выполнено | Docs и README (Этот этап) |

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте feature branch (`git checkout -b feature/feature-name`)
3. Commit изменения (`git commit -am 'Add feature'`)
4. Push на branch (`git push origin feature/feature-name`)
5. Создайте Pull Request

## 📄 Лиценция

MIT License. См. [`LICENSE`](LICENSE) файл для подробностей.

## 📞 Поддержка

- Issues: GitHub Issues
- Документация: [docs/](docs/)
