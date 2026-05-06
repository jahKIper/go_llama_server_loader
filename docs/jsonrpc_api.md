# 🔌 JSON-RPC 2.0 API

## Описание

Web UI `llama-server-loader` использует протокол **JSON-RPC 2.0** поверх HTTP для коммуникации между фронтендом и бэкендом.

## Транспорт

- **Протокол:** HTTP 1.1 / HTTPS
- **Endpoint:** `/api/jsonrpc` (или настроенный путь)
- **Content-Type:** `application/json`
- **Метод:** `POST`

## Формат запроса

```json
{
  "jsonrpc": "2.0",
  "method": "getModels",
  "params": {
    // параметры метода (если есть)
  },
  "id": 1
}
```

### Поля запроса

| Поле | Тип | Обязательный | Описание |
|------|-----|:-----------:|----------|
| `jsonrpc` | string | Да | Версия JSON-RPC (всегда `"2.0"`) |
| `method` | string | Да | Имя вызываемого метода |
| `params` | object | Нет | Параметры метода (может быть опущен если нет параметров) |
| `id` | integer\|string | Да | Идентификатор запроса для сопоставления ответа |

## Формат ответа

### Успешный ответ

```json
{
  "jsonrpc": "2.0",
  "result": {
    // данные результата
  },
  "id": 1
}
```

### Ответ с ошибкой

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32600,
    "message": "Invalid request",
    "data": "дополнительные данные"
  },
  "id": 1
}
```

### Коды ошибок

| Code | Message | Описание |
|------|---------|----------|
| `-32600` | Invalid JSON | Ответ был получен, но это не валидный JSON |
| `-32601` | Method not found | Метод не существует или недоступен |
| `-32602` | Invalid params | Невалидные параметры метода |
| `-32603` | Internal error | Внутренняя ошибка сервера |

## Методы API

### getModels

Запрос списка доступных моделей.

**Запрос:**

```json
{
  "jsonrpc": "2.0",
  "method": "getModels",
  "params": {
    "scan_dir": "/path/to/models"
  },
  "id": 1
}
```

| Параметр | Тип | Обязательный | Описание |
|----------|-----|:-----------:|----------|
| `scan_dir` | string | Нет | Директория для сканирования. Если не указан — используется сохранённая конфигурация |

**Ответ:**

```json
{
  "jsonrpc": "2.0",
  "result": {
    "models": [
      {
        "name": "gemma-2b-it-q4k",
        "path": "/models/gemma-2b-it.Q4_K_M.gguf",
        "size": 1719900672,
        "is_mmproj": false,
        "mmproj_path": "/models/mmproj-fp16.gguf",
        "mmproj_on": true
      }
    ],
    "scan_paths": ["/models"],
    "errors": []
  },
  "id": 1
}
```

### startServer

Запуск `llama-server` с выбранной конфигурацией.

**Запрос:**

```json
{
  "jsonrpc": "2.0",
  "method": "startServer",
  "params": {
    "model_name": "gemma-2b-it-q4k",
    "flags": {
      "model": "/models/gemma-2b-it.Q4_K_M.gguf",
      "ctx_len": 4096,
      "gpu_layers": 24,
      "threads": 8,
      "temperature": 0.7
    }
  },
  "id": 2
}
```

| Параметр | Тип | Обязательный | Описание |
|----------|-----|:-----------:|----------|
| `model_name` | string | Да | Имя модели из сохранённой конфигурации |
| `flags` | object | Да | Карта флагов для передачи в `llama-server` |

**Ответ:**

```json
{
  "jsonrpc": "2.0",
  "result": {
    "status": "started",
    "command": "llama-server -m /models/gemma-2b-it.Q4_K_M.gguf --ctx-size 4096 ...",
    "pid": 12345,
    "port": 8080
  },
  "id": 2
}
```

### getStatus

Получение текущего статуса сервера.

**Запрос:**

```json
{
  "jsonrpc": "2.0",
  "method": "getStatus",
  "params": {},
  "id": 3
}
```

**Ответ:**

```json
{
  "jsonrpc": "2.0",
  "result": {
    "status": "running",
    "pid": 12345,
    "uptime_seconds": 120,
    "server_url": "http://localhost:8080"
  },
  "id": 3
}
```

**Статусы:**

| Статус | Описание |
|--------|----------|
| `idle` | Сервер не запущен |
| `starting` | Инициализация сервера |
| `running` | Сервер работает |
| `error` | Ошибка при запуске или работе |
| `stopped` | Сервер остановлен |

### shutdown

Корректное завершение работы сервера.

**Запрос:**

```json
{
  "jsonrpc": "2.0",
  "method": "shutdown",
  "params": {},
  "id": 4
}
```

**Ответ:**

```json
{
  "jsonrpc": "2.0",
  "result": {
    "status": "stopped",
    "message": "Server shut down successfully"
  },
  "id": 4
}
```

## Клиентская реализация (JavaScript)

Пример использования в браузере:

```javascript
class JsonRpcClient {
  constructor(baseUrl = '/api') {
    this.baseUrl = baseUrl;
    this.id = 0;
  }

  async call(method, params = {}) {
    const requestId = ++this.id;
    const response = await fetch(`${this.baseUrl}/jsonrpc`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        jsonrpc: '2.0',
        method,
        params,
        id: requestId
      })
    });

    const data = await response.json();
    if (data.error) {
      throw new Error(`JSON-RPC Error: ${data.error.message}`);
    }
    return data.result;
  }

  async getModels(scanDir) {
    return this.call('getModels', { scan_dir: scanDir });
  }

  async startServer(modelName, flags) {
    return this.call('startServer', { model_name: modelName, flags });
  }

  async getStatus() {
    return this.call('getStatus', {});
  }

  async shutdown() {
    return this.call('shutdown', {});
  }
}

// Использование
const rpc = new JsonRpcClient();
rpc.getModels('/models')
   .then(data => console.log('Модели:', data.models))
   .catch(err => console.error(err));
```

## Интеграция с Web UI

Фронтенд (static/app.js) использует JSON-RPC для:

1. **Загрузки списка моделей** — `getModels` при инициализации
2. **Выбора модели** — отображение в интерфейсе
3. **Запуска сервера** — `startServer` при нажатии кнопки
4. **Мониторинга статуса** — периодический `getStatus` poll
5. **Остановки сервера** — `shutdown` при отключении

## Безопасность

- Web UI доступен только локально (localhost)
- CORS может быть настроен через флаги `llama-server`
- Нет аутентификации на уровне JSON-RPC (предполагается локальное использование)
