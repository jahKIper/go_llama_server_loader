# 📋 Формат params_ru.json

## Описание

`params_ru.json` — неизменяемый справочник параметров `llama-server` с переводами на русский язык. Содержит все параметры CLI инструмента `llama-server` из llama.cpp.

## Структура файла

```json
{
  "version": "1.0",
  "categories": [
    {
      "name": "Категория параметров",
      "params": [
        {
          "short_flag": "-m",
          "long_flag": "--model FNAME",
          "description_ru": "Путь к файлу модели для загрузки"
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

## Поля структуры

### Корневой уровень

| Поле | Тип | Описание |
|------|-----|----------|
| `version` | string | Версия схемы JSON-справочника (текущая: "1.0") |
| `categories` | array[] | Массив категорий параметров |
| `total_params_count` | integer | Общее количество параметров во всех категориях |
| `source_docs` | string[] | Ссылки на официальную документацию llama.cpp |

### category

| Поле | Тип | Описание |
|------|-----|----------|
| `name` | string | Название категории на русском языке |
| `params` | array[] | Массив параметров в данной категории |

### param

| Поле | Тип | Обязательный | Описание |
|------|-----|:-----------:|----------|
| `short_flag` | string | Нет | Короткий флаг (например: `-m`, `-t`). Может быть пустой строкой `""` если короткий флаг отсутствует |
| `long_flag` | string | Да | Полный флаг с аргументом (например: `--model FNAME`, `--threads N`) |
| `description_ru` | string | Да | Описание параметра на русском языке |

## Категории параметров

### 1. Загрузка модели

Параметры для указания пути к модели, мультимодальным проекторам и внешним источникам.

**Примеры:**

```json
{"short_flag": "-m", "long_flag": "--model FNAME", "description_ru": "Путь к файлу модели для загрузки"}
{"short_flag": "-t", "long_flag": "--threads N", "description_ru": "Количество CPU-потоков (по умолчанию: -1)"}
```

### 2. Контекст и генерация

Параметры управления размером контекстного окна, количеством предсказанных токенов и температурой генерации.

**Примеры:**

```json
{"short_flag": "-c", "long_flag": "--ctx-size N", "description_ru": "Размер контекстного окна"}
{"short_flag": "", "long_flag": "--temperature FNUM", "description_ru": "Температура генерации (по умолчанию: 0.80)"}
```

### 3. Семплирование

Параметры для управления качеством и разнообразием генерации (top-k, top-p, temperature, repeat penalty и др.).

**Примеры:**

```json
{"short_flag": "", "long_flag": "--top-k N", "description_ru": "Топ-k семплирование"}
{"short_flag": "", "long_flag": "--top-p N", "description_ru": "Топ-p семплирование"}
```

### 4. GPU / CUDA / Metal

Параметры управления использованием GPU, количеством слоёв для оффлоада и_splits_ между несколькими GPU.

**Примеры:**

```json
{"short_flag": "-ngl", "long_flag": "--gpu-layers N", "description_ru": "Количество слоёв для хранения в VRAM"}
```

### 5. Встроенный веб-сервер

Параметры конфигурации HTTP сервера: порт, хост, CORS, API режимы.

**Примеры:**

```json
{"short_flag": "-hp", "long_flag": "--host HOST", "description_ru": "Адрес для прослушивания"}
{"short_flag": "-pp", "long_flag": "--port PORT", "description_ru": "Порт HTTP сервера"}
```

### 6. API режимы

Настройки OpenAI API, Anthropic API и других совместимых интерфейсов.

## Источники данных

Параметры вычитаны из официальной документации llama.cpp:

- [llama.cpp server README](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md)
- [llama.cpp CLI README](https://github.com/ggml-org/llama.cpp/blob/master/tools/cli/README.md)
- [llama.cpp multimodal docs](https://github.com/ggml-org/llama.cpp/blob/master/docs/multimodal.md)

## Использование в приложении

Файл загружается через `internal/config/loader.go`:

```go
// Загрузка справочника параметров
params, err := config.LoadParams("params_ru.json")
if err != nil {
    log.Fatal(err)
}

// Получение категории по индексу
category := params.Categories[0]
fmt.Println(category.Name) // "Загрузка модели"

// Перебор параметров в категории
for _, p := range category.Params {
    fmt.Printf("%s: %s\n", p.LongFlag, p.DescriptionRU)
}
```

## Версионирование

Текущая версия схемы: `1.0`

При изменении структуры файла необходимо:
1. Увеличить версию в поле `version`
2. Обновить все существующие поля при необходимости
3. Документировать изменения в changelog
