# 📁 Формат models.json

## Описание

`models.json` — пользовательский конфигурационный файл для сохранения выбранных моделей, их настроек и путей сканирования. Файл поддерживает lossless сохранение JSON без потери данных.

## Структура файла

```json
{
  "version": "1.0",
  "scan_paths": ["/home/user/models", "/opt/llm"],
  "models": [
    {
      "name": "gemma-2b-it-q4k",
      "model_path": "/home/user/models/gemma-2b-it.Q4_K_M.gguf",
      "mmproj_path": "/home/user/models/mmproj-fp16.gguf",
      "mmproj_on": true,
      "size": 1719900672,
      "last_scan": "2024-05-03T12:00:00Z",
      "flags": {
        "model": "/home/user/models/gemma-2b-it.Q4_K_M.gguf",
        "ctx_len": 4096,
        "gpu_layers": 24,
        "threads": 8,
        "temperature": 0.7,
        "top_k": 120,
        "top_p": 0.95
      }
    },
    {
      "name": "llama-3.1-8b-instruct",
      "model_path": "/opt/llm/llama-3.1-8b-instruct.Q4_K_M.gguf",
      "mmproj_path": "",
      "mmproj_on": false,
      "size": 5399006720,
      "last_scan": "2024-05-03T12:00:00Z",
      "flags": {
        "model": "/opt/llm/llama-3.1-8b-instruct.Q4_K_M.gguf",
        "ctx_len": 8192,
        "gpu_layers": 35,
        "threads": 16,
        "temperature": 0.5
      }
    }
  ]
}
```

## Поля структуры

### Корневой уровень

| Поле | Тип | Обязательный | Описание |
|------|-----|:-----------:|----------|
| `version` | string | Да | Версия схемы конфигурации (текущая: "1.0") |
| `scan_paths` | string[] | Нет | Список директорий для рекурсивного сканирования `.gguf` файлов |
| `models` | array[] | Нет | Массив сохранённых конфигураций моделей |

### model

| Поле | Тип | Обязательный | Описание |
|------|-----|:-----------:|----------|
| `name` | string | Да | Имя модели (читаемое, без расширений) |
| `model_path` | string | Да | Полный путь к `.gguf` файлу модели |
| `mmproj_path` | string | Нет | Путь к мультимодальному проектору. Может быть пустым если `mmproj_on: false` |
| `mmproj_on` | boolean | Нет | Флаг использования мультимодального проектора (по умолчанию: false) |
| `size` | integer | Да | Размер файла модели в байтах |
| `last_scan` | string | Нет | ISO 8601 timestamp последнего сканирования |
| `flags` | object | Нет | Карта флагов для передачи в `llama-server` |

### flags (внутри model)

Это произвольная карта, где ключи соответствуют именам параметров `llama-server`. Значения могут быть любого типа (`string`, `number`, `boolean`).

**Стандартные флаги:**

| Ключ | Тип | Описание |
|------|-----|----------|
| `model` | string | Путь к файлу модели (дублирует `model_path`) |
| `ctx_len` / `ctx-size` | integer | Размер контекстного окна |
| `gpu_layers` / `n-gpu-layers` | integer | Количество слоёв для GPU |
| `threads` | integer | Количество CPU-потоков |
| `temperature` | number | Температура генерации |
| `top_k` | integer | Top-k семплирование |
| `top_p` | number | Top-p семплирование |
| `repeat_penalty` | number | Штраф за повторения |

## API использования

### Загрузка конфигурации

```go
// Загрузка модели из файла
cfg, err := config.LoadConfig("models.json")
if err != nil {
    log.Fatal(err)
}

fmt.Println(cfg.Version)           // "1.0"
fmt.Println(len(cfg.ScanPaths))     // 2
fmt.Println(len(cfg.Models))        // 2
```

### Сохранение конфигурации

```go
// Создание новой модели
newModel := &config.ModelItem{
    Name:       "gemma-2b-it-q4k",
    ModelPath:  "/models/gemma-2b-it.Q4_K_M.gguf",
    MMProjPath: "",
    MMProjOn:   false,
    Size:       1719900672,
    Flags: map[string]interface{}{
        "model":      "/models/gemma-2b-it.Q4_K_M.gguf",
        "ctx_len":    4096,
        "gpu_layers": 24,
        "threads":    8,
        "temperature": 0.7,
    },
}

// Сохранение без потерь
err := config.SaveConfig(cfg, "models.json")
```

### Валидация конфигурации

```go
// Проверка корректности структуры
err := config.ValidateConfig(cfg)
if err != nil {
    log.Printf("Ошибка валидации: %v", err)
}
```

## Правила генерации имени модели

Имя модели формируется на основе префикса имени файла `.gguf` без расширений и квантизации:

| Файл | Имя модели |
|------|------------|
| `llama-3.1-8b-instruct.Q4_K_M.gguf` | `llama-3.1-8b-instruct` |
| `gemma-2b-it-Q5_K_S.gguf` | `gemma-2b-it` |
| `mmproj-fp16.gguf` | `mmproj` (мультимодальный) |

## Нормализация путей

Все пути нормализуются через `filepath.ToSlash()` для кроссплатформенности:

```go
// Windows: "D:\\models\\model.gguf" -> "D:/models/model.gguf"
normalized := filepath.ToSlash(filepath)
```

## Версионирование

Текущая версия схемы: `1.0`

При изменении структуры файла необходимо:
1. Увеличить версию в поле `version`
2. Обеспечить backward compatibility для старых версий
3. Документировать breaking changes
