# llama-server-loader — План доработок

## 📊 Текущий статус

| Компонент | Статус | Оценка |
|-----------|--------|--------|
| `pkg/modelscan/*` | ✅ Полностью реализован | 100% |
| `pkg/servercmd/*` | ✅ Полностью реализован | 100% |
| `internal/config/*` | ✅ Полностью реализован | 100% |
| `internal/cli/*` | ⚠️ Частично (нет UI) | 30% |
| `internal/webui/*` | ⚠️ Частично (нет servercmd) | 40% |
| `cmd/llama-server-loader/main.go` | ⚠️ Нет интеграции | 20% |
| Документация | ✅ Полностью реализована | 100% |

**Общая оценка: ~75%**

---

## 📋 Задача 1: Добавить зависимости charmbracelet/* в go.mod

### Описание
Для реализации CLI UI на charmbracelet необходимо добавить зависимости в `go.mod`.

### Список зависимостей
```go
require (
    github.com/charmbracelet/bubbles v0.20.0
    github.com/charmbracelet/bubbletea v0.27.0
    github.com/charmbracelet/lipgloss v1.0.0
)
```

### Команды для выполнения
```bash
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go mod tidy
```

### Критерии готовности
- [ ] `go.mod` содержит все 3 зависимости charmbracelet
- [ ] `go.sum` обновлён
- [ ] `go mod tidy` проходит без ошибок

---

## 📋 Задача 2: Реализовать CLI UI на charm (internal/cli)

### Описание
Заменить TODO-заглушку в `runInteractive()` на реальный терминальный UI с использованием `charmbracelet/bubbles/list` + `tea`.

### Файл для редактирования
- [`internal/cli/cli.go`](internal/cli/cli.go:218) — метод `runInteractive()` (строки 218-227)

### Что реализовать

#### 2.1. Добавить структуру CLI с полями для UI
```go
package cli

import (
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"

    "github.com/charmbracelet/bubbles/list"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss/v2"
)

// ListItem представляет элемент списка моделей.
type ListItem struct {
    Name     string
    Path     string
    Size     int64
    IsMMProj bool
    Selected bool
    Index    int
}

func (l ListItem) Title() string       { return FormatModelName(l.Name, l.Path, l.Size) }
func (l ListItem) Description() string { return l.Path }
func (l ListItem) Value() interface{}  { return l.Index }
```

#### 2.2. Реализовать ModelList — обёртка над bubbles/list
```go
type ModelList struct {
    list *list.Model
    items []ListItem
}

func NewModelList(models []*modelscan.Model) *ModelList {
    var items []ListItem
    for i, m := range models {
        items = append(items, ListItem{
            Name:     m.Name,
            Path:     m.Path,
            Size:     m.Size,
            IsMMProj: m.IsMMProj,
            Index:    i,
        })
    }
    
    ls := list.New(list.Defaults())
    for _, item := range items {
        ls.InsertItem(0, item)
    }
    ls.SetFilter("", false, true)
    
    return &ModelList{list: ls, items: items}
}
```

#### 2.3. Реализовать ModelView — компонент для отображения
```go
type ModelView struct {
    list *ModelList
    selected *ListItem
}

func (m ModelView) Init() tea.Cmd { return nil }

func (m ModelView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.list.list.SetWidth(msg.Width - 4)
        m.list.list.SetHeight(msg.Height - 4)
        return m, nil
    case tea.KeyMsg:
        // Обработка клавиш
    }
    
    var cmd tea.Cmd
    m.list.list, cmd = m.list.list.Update(msg)
    return m, cmd
}
```

#### 2.4. Обновить runInteractive()
```go
func (c *CLI) runInteractive() error {
    fmt.Println("Scanning directory:", c.scanDir)
    
    // Интеграция с pkg/modelscan
    scanResult, err := modelscan.ScanDir(c.scanDir)
    if err != nil {
        return fmt.Errorf("ошибка сканирования: %w", err)
    }
    
    fmt.Printf("Найдено моделей: %d\n", len(scanResult.Models))
    fmt.Printf("Найдено mmproj: %d\n", len(scanResult.MMModels))
    
    // Match models with mmproj
    enrichedModels, err := modelscan.MatchMMProj(scanResult.Models)
    if err != nil {
        return fmt.Errorf("ошибка сопоставления mmproj: %w", err)
    }
    
    // Create UI
    modelList := NewModelList(enrichedModels)
    m := ModelView{list: modelList}
    
    p := tea.NewProgram(m, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        return fmt.Errorf("ошибка запуска UI: %w", err)
    }
    
    // После выбора модели — запустить llama-server или Web UI
    if c.selected != nil {
        fmt.Printf("Выбрана модель: %s\n", c.selected.Path)
        // TODO: Запуск llama-server или передача в Web UI
    }
    
    return nil
}
```

### Критерии готовности
- [ ] `runInteractive()` сканирует директорию через `modelscan.ScanDir()`
- [ ] `runInteractive()` сопоставляет модели с mmproj через `modelscan.MatchMMProj()`
- [ ] Отображается интерактивный список на charmbracelet
- [ ] Пользователь может выбрать модель (стрелки + Enter)
- [ ] Покрытие тестами ≥70%

---

## 📋 Задача 3: Интегрировать pkg/modelscan в main.go и cli.go

### Описание
Обновить `main.go` и `cli.go` для использования `pkg/modelscan.Scanner`.

### Файлы для редактирования

#### 3.1. [`cmd/llama-server-loader/main.go`](cmd/llama-server-loader/main.go:104) — функция `runInteractive()`
**Текущий код (строки 107-114):**
```go
// TODO: Интеграция с pkg/modelscan.Scanner
// scanResult, err := modelscan.ScanDir(c.ScanDir())
// if err != nil {
//     return fmt.Errorf("ошибка сканирования: %w", err)
// }

log.Println("Сканирование завершено. Список моделей будет отображён в терминальном UI.")
```

**Что изменить:**
- Раскомментировать и интегрировать `modelscan.ScanDir()`
- Добавить обработку результатов сканирования
- Передать модели в CLI для отображения

#### 3.2. [`internal/cli/cli.go`](internal/cli/cli.go:184) — метод `Run()`
**Текущий код (строки 184-187):**
```go
// Interactive mode: scan and show model list
if c.scanDir != "" {
    return c.runInteractive()
}
```

**Что добавить:**
- Вызов `modelscan.ScanDir()` перед `runInteractive()`
- Передача результатов в UI компонент

### Критерии готовности
- [ ] `main.go` вызывает `modelscan.ScanDir(c.ScanDir())`
- [ ] Результаты сканирования передаются в CLI UI
- [ ] Ошибки сканирования обрабатываются gracefully

---

## 📋 Задача 4: Интегрировать pkg/servercmd в webui.go

### Описание
Обновить `webui.go` для использования `pkg/servercmd.BuildCommand()` при запуске llama-server.

### Файл для редактирования
- [`internal/webui/webui.go`](internal/webui/webui.go:191) — метод `handleStartServer()` (строки 191-218)

### Текущий код (строки 204-217):
```go
s.status = "starting"
s.logger.Printf("Starting server for model: %s", p.ModelPath)

// TODO: Integrate with pkg/servercmd to actually start llama-server
// For now, just update status
if p.Port > 0 {
    s.status = fmt.Sprintf("running on port %d", p.Port)
} else {
    s.status = "running"
}
```

### Что реализовать

#### 4.1. Добавить поле для хранения процесса сервера
```go
type Server struct {
    addr     string
    server   *http.Server
    logger   *log.Logger
    models   []ModelInfo
    status   string
    shutdown chan struct{}
    cmd      *exec.Cmd       // Process for llama-server
    cmdCtx   context.Context // Context for command cancellation
}
```

#### 4.2. Обновить handleStartServer()
```go
func (s *Server) handleStartServer(params json.RawMessage) interface{} {
    if len(params) == 0 {
        return map[string]string{"message": "No parameters provided"}
    }

    var p StartServerParams
    if err := json.Unmarshal(params, &p); err != nil {
        return map[string]string{
            "error": fmt.Sprintf("Invalid params: %v", err),
        }
    }

    // Build config for servercmd
    cfg := &servercmd.Config{
        Version: "1.0",
        Models: []servercmd.ModelConfig{
            {
                Name:       FormatModelName(p.ModelPath),
                ModelPath:  p.ModelPath,
                MMProjPath: p.MMProjPath,
                MMProjOn:   p.MMProjPath != "",
                Size:       0, // Will be populated from scan
                Flags: map[string]any{
                    "threads":     p.Threads,
                    "temperature": p.Temperature,
                    "port":        p.Port,
                },
            },
        },
    }

    // Build command using servercmd
    cmd, err := servercmd.BuildCommand(cfg)
    if err != nil {
        s.status = "error"
        return map[string]string{
            "error": fmt.Sprintf("Failed to build command: %v", err),
        }
    }

    // Start the command
    ctx, cancel := context.WithCancel(context.Background())
    s.cmdCtx = ctx
    
    if err := cmd.Start(); err != nil {
        cancel()
        s.status = "error"
        return map[string]string{
            "error": fmt.Sprintf("Failed to start server: %v", err),
        }
    }

    // Store command for shutdown
    s.cmd = cmd
    
    s.status = "running"
    if p.Port > 0 {
        s.addr = fmt.Sprintf(":%d", p.Port)
    }
    
    return map[string]string{
        "message": "Server starting...",
        "status":  s.status,
        "addr":    s.addr,
    }
}
```

#### 4.3. Обновить handleShutdown()
```go
func (s *Server) handleShutdown() interface{} {
    if s.cmd != nil && s.cmd.Process != nil {
        if s.cmdCtx != nil {
            context.CancelFunc(s.cmdCtx)()
        }
        s.cmd.Process.Signal(os.Interrupt)
        s.cmd.Wait()
    }
    
    go func() {
        time.Sleep(100 * time.Millisecond)
        s.status = "shutting_down"
    }()
    return map[string]string{
        "message": "Shutdown initiated",
    }
}
```

### Критерии готовности
- [ ] `handleStartServer()` вызывает `servercmd.BuildCommand()`
- [ ] Команда llama-server реально запускается
- [ ] `handleShutdown()` корректно останавливает сервер
- [ ] Статус сервера обновляется правильно (running/stopped/error)

---

## 📋 Задача 5: Полная интеграция модулей в main.go

### Описание
Объединить все модули в единую систему оркестрации в `main.go`.

### Файл для редактирования
- [`cmd/llama-server-loader/main.go`](cmd/llama-server-loader/main.go)

### 5.1. Добавить импорт pkg/modelscan и pkg/servercmd
```go
import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "llama-server-loader/internal/cli"
    "llama-server-loader/internal/config"
    "llama-server-loader/internal/webui"
    
    "llama-server-loader/pkg/modelscan"
    "llama-server-loader/pkg/servercmd"
)
```

### 5.2. Обновить main() для полной оркестрации
```go
func main() {
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    log.SetPrefix("main: ")

    if len(os.Args) < 2 {
        printUsage()
        os.Exit(0)
    }

    flags, err := cli.ParseFlags(os.Args[1:])
    if err != nil {
        log.Fatalf("Ошибка парсинга флагов: %v", err)
    }

    c := cli.NewCLI(flags)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        select {
        case <-sigChan:
            log.Println("Получен сигнал завершения...")
            cancel()
        case <-ctx.Done():
            return
        }
    }()

    // 1. Запуск Web UI сервера отдельно
    if flags.StartWebUI && c.ScanDir() == "" {
        log.Printf("Запуск Web UI сервера на %s", fmt.Sprintf(":%d", flags.WebPort))
        ws := webui.NewServer(fmt.Sprintf(":%d", flags.WebPort))
        
        // Загрузить модели если есть конфиг
        if cfgPath := getConfigPath(); cfgPath != "" {
            cfg, err := config.LoadConfig(cfgPath)
            if err != nil {
                log.Printf("Warning: cannot load config: %v", err)
            } else {
                models := convertToWebUIModels(cfg.Models)
                ws.SetModels(models)
            }
        }
        
        if err := ws.Start(ctx); err != nil {
            log.Fatalf("Ошибка запуска Web UI: %v", err)
        }
        <-ctx.Done()
        shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer shutdownCancel()
        ws.Shutdown(shutdownCtx)
        return
    }

    // 2. Генерация параметров
    if c.GenerateParams() {
        if err := c.Run(); err != nil {
            log.Fatalf("Ошибка генерации параметров: %v", err)
        }
        return
    }

    // 3. Интерактивный режим со сканированием
    if c.ScanDir() != "" {
        // Сканирование директории
        scanResult, err := modelscan.ScanDir(c.ScanDir())
        if err != nil {
            log.Fatalf("Ошибка сканирования: %v", err)
        }
        
        log.Printf("Найдено моделей: %d, mmproj: %d", 
            len(scanResult.Models), len(scanResult.MMModels))
        
        // Сопоставление моделей с mmproj
        enrichedModels, err := modelscan.MatchMMProj(scanResult.Models)
        if err != nil {
            log.Fatalf("Ошибка сопоставления mmproj: %v", err)
        }
        
        // Сохранение конфига если запрошено
        if c.SaveConfig() != "" {
            cfg := &config.Config{
                Version:   "1.0",
                ScanPaths: []string{c.ScanDir()},
                Models:    buildModelConfigs(enrichedModels),
            }
            if err := config.SaveConfig(cfg, c.SaveConfig()); err != nil {
                log.Fatalf("Ошибка сохранения конфига: %v", err)
            }
            log.Printf("Конфигурация сохранена в: %s", c.SaveConfig())
        }
        
        // Запуск llama-server если выбрана модель
        if c.ModelName() != "" {
            cfg := buildServerConfig(c, enrichedModels)
            cmd, err := servercmd.BuildCommand(cfg)
            if err != nil {
                log.Fatalf("Ошибка сборки команды: %v", err)
            }
            
            log.Printf("Запуск llama-server...")
            if err := servercmd.RunCommand(cmd); err != nil {
                log.Fatalf("Ошибка запуска сервера: %v", err)
            }
        }
        
        return
    }

    printUsage()
}
```

### 5.3. Добавить вспомогательные функции
```go
// buildModelConfigs converts modelscan.Models to config.ModelConfig
func buildModelModels(models []*modelscan.Model) []config.ModelConfig {
    var result []config.ModelConfig
    
    // Group models by directory and match with mmproj
    dirMap := make(map[string][]*modelscan.Model)
    for _, m := range models {
        dir := filepath.Dir(m.Path)
        dirMap[dir] = append(dirMap[dir], m)
    }
    
    for dir, dirModels := range dirMap {
        mmprojPaths, _ := modelscan.FindMMprojByDirAndName(dir)
        
        for _, m := range dirModels {
            if m.IsMMProj {
                continue // Skip mmproj files for now
            }
            
            mc := config.ModelConfig{
                Name:       strings.TrimSuffix(filepath.Base(m.Path), ".gguf"),
                ModelPath:  m.Path,
                Size:       m.Size,
                LastScan:   time.Now().Format(time.RFC3339),
                Flags:      make(map[string]interface{}),
            }
            
            // Find matching mmproj
            for _, mp := range mmprojPaths {
                if mc.MMProjPath == "" {
                    mc.MMProjPath = mp
                    mc.MMProjOn = true
                }
            }
            
            result = append(result, mc)
        }
    }
    
    return result
}

// buildServerConfig creates servercmd.Config from CLI flags and models
func buildServerConfig(c *cli.CLI, models []*modelscan.Model) *servercmd.Config {
    // Find the selected model
    var selectedModel *modelscan.Model
    for _, m := range models {
        if strings.Contains(m.Path, c.ModelName()) || strings.Contains(m.Name, c.ModelName()) {
            selectedModel = m
            break
        }
    }
    
    if selectedModel == nil {
        return &servercmd.Config{Version: "1.0", Models: []servercmd.ModelConfig{}}
    }
    
    cfg := &servercmd.Config{
        Version: "1.0",
        Models: []servercmd.ModelConfig{
            {
                Name:       selectedModel.Name,
                ModelPath:  selectedModel.Path,
                Size:       selectedModel.Size,
                MMProjOn:   false,
                Flags:      make(map[string]any),
            },
        },
    }
    
    // Add CLI flags to config
    if c.Threads() > 0 {
        cfg.Models[0].Flags["threads"] = c.Threads()
    }
    if c.Temperature() > 0 {
        cfg.Models[0].Flags["temperature"] = c.Temperature()
    }
    
    return cfg
}
```

### Критерии готовности
- [ ] `main.go` импортирует все пакеты: `modelscan`, `servercmd`, `config`, `cli`, `webui`
- [ ] Интерактивный режим сканирует директорию и показывает список моделей
- [ ] При выборе модели запускается llama-server через `servercmd.BuildCommand()`
- [ ] Конфиг сохраняется в файл при `--save-config`
- [ ] Graceful shutdown работает корректно

---

## 📋 Задача 6: Smoke test и проверка go build

### Описание
Выполнить финальную проверку сборки и тестов.

### Команды для выполнения
```bash
# 1. Проверка сборки
go build -o llama-server-loader.exe ./cmd/llama-server-loader/

# 2. Запуск всех тестов с покрытием
go test ./... -v -cover

# 3. Проверка покрытия
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.html
```

### Критерии готовности
- [ ] `go build` проходит без ошибок
- [ ] `go test ./...` проходит без ошибок
- [ ] Покрытие тестами ≥80% (`go test ./... -cover`)

---

## 📋 Итоговая оценка после выполнения плана

| Компонент | Текущий статус | После доработок |
|-----------|----------------|-----------------|
| `pkg/modelscan/*` | ✅ 100% | ✅ 100% |
| `pkg/servercmd/*` | ✅ 100% | ✅ 100% |
| `internal/config/*` | ✅ 100% | ✅ 100% |
| `internal/cli/*` | ⚠️ 30% | ✅ 95% (с charm UI) |
| `internal/webui/*` | ⚠️ 40% | ✅ 90% (с servercmd) |
| `main.go` интеграция | ⚠️ 20% | ✅ 100% |
| Документация | ✅ 100% | ✅ 100% |
| **Общая оценка** | **~75%** | **~98%** |

---

## 📋 Порядок выполнения задач

1. **Задача 1**: Добавить зависимости charmbracelet/* (зависит от: нет)
2. **Задача 2**: Реализовать CLI UI на charm (зависит от: Задача 1)
3. **Задача 3**: Интегрировать pkg/modelscan в main.go и cli.go (зависит от: Задача 2)
4. **Задача 4**: Интегрировать pkg/servercmd в webui.go (зависит от: нет)
5. **Задача 5**: Полная интеграция модулей в main.go (зависит от: Задача 3, Задача 4)
6. **Задача 6**: Smoke test и проверка go build (зависит от: Задача 5)

---

## ⚠️ Риски и митигация

| Риск | Влияние | Митигация | Fallback |
|------|---------|-----------|----------|
| Charmbracelet API changes | Medium | Использовать конкретные версии | Go vendor mode |
| llama-server не в PATH | High | Pre-flight check + graceful error | Показать ошибку с инструкцией |
| Кроссплатформенность путей | Medium | filepath.ToSlash() везде | Уже реализовано в scanner.go |

