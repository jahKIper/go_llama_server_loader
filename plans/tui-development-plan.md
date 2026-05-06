# План разработки TUI — llama-server-loader

> Составлен: 2026-05-05  
> На основе аудита кода и схемы `tui-model-list-screen-schema.md`

---

## Текущее состояние

**Код не компилируется** — дублирующиеся типы + отсутствующие импорты.  
Инфраструктурные компоненты (filter, footer, dimensions, styles) определены, но **не подключены** к `App`.

---

## Граф зависимостей фаз

```
Фаза 1 (compile errors) → Фаза 2 (App struct)
                               ↓          ↓
                           Фаза 3     Фаза 4    ← параллельно
                           (Update)   (View)
                               └────┬────┘
                                    ↓
Фаза 1 → Фаза 5 (badges) ──→ Фаза 6 (integration)
```

---

## Фаза 1 — Устранение ошибок компиляции

**Цель:** `go build ./...` проходит без ошибок.  
**Риск:** Низкий. Только удаление кода и добавление импортов.

### Задачи

| # | Файл | Действие |
|---|------|----------|
| 1.1 | `cli.go` строки 458–484 | Удалить дублирующий `ListItem` struct + методы |
| 1.2 | `cli.go` строки 499–568 | Удалить дублирующий `StyledDelegate` struct + методы |
| 1.3 | `model_item.go` | Добавить импорты: `llama-server-loader/pkg/modelscan`, `charm.land/lipgloss/v2` |
| 1.4 | `filter.go` | Добавить импорт: `llama-server-loader/pkg/modelscan` |

### Acceptance criteria
- `go build ./internal/cli/...` → exit 0
- `go vet ./internal/cli/...` → exit 0
- Тесты `TestParseFlags`, `TestNewCLI`, `TestCLIValidate` проходят

---

## Фаза 2 — Обновление App struct и конструктора NewApp

**Цель:** `App` struct соответствует схеме §5.1. `NewApp()` инициализирует все компоненты.

### App struct (целевой)

```go
type App struct {
    list        list.Model
    selected    *modelscan.Model
    filterState FilterState
    filterInput *filterInput
    filterText  string
    allModels   []*modelscan.Model
    countShown  int
    countTotal  int
    title       string
    version     string
    width       int
    height      int
    styles      *StyleConfig
    dims        *Dimensions
    footer      *Footer
}
```

**Удалить:** `statusMsg string`, `filtering bool`

### Задачи

| # | Действие |
|---|----------|
| 2.1 | Заменить `App` struct в `cli.go` на целевой |
| 2.2 | Переписать `NewApp()`: инициализировать `filterInput`, `Footer`, `Dimensions`, `allModels`, `countShown/Total`, `version="0.1.0"`, отключить встроенный bubbles-фильтр (`SetFilteringEnabled(false)`) |
| 2.3 | Передавать `&StyledDelegate{...}` (pointer) в `list.New()` — у `model_item.go` pointer receivers |
| 2.4 | Удалить неиспользуемый `var keysChoose` (строка 593) |

### Acceptance criteria
- `NewApp(models)` возвращает `*App` со всеми полями non-nil
- Встроенная фильтрация bubbles/v2 отключена

---

## Фаза 3 — Переписать App.Update() с state machine фильтра

**Цель:** Полная обработка клавиш. Filter state machine работает. Счётчики обновляются. Dimensions управляют размером списка.

### Структура Update()

```
KeyPressMsg:
  "ctrl+c", "q"  → tea.Quit
  "enter"        → выбрать модель, tea.Quit
  "esc"          → отмена фильтра (→ FilterIdle, очистить, восстановить список)
  "/"            → если Idle: активировать фильтр; иначе: передать filterInput
  default:
    если filterState Active/Filtering: передать filterInput.HandleKey()
    иначе: передать list.Update()
WindowSizeMsg:
    dims.ClampSize() → list.SetSize()
    footer.SetWidth()
```

### Задачи

| # | Действие |
|---|----------|
| 3.1 | `WindowSizeMsg`: заменить хардкод `msg.Width-4, msg.Height-8` на `dims.ClampSize()` с вычетом зарезервированных строк для header/filter/count/footer |
| 3.2 | Обработать клавишу `/` → `FilterActive` |
| 3.3 | Обработать `Esc` → `FilterIdle`, очистить фильтр, восстановить полный список |
| 3.4 | Роутинг символов → `filterInput.HandleKey()` → `FilterModels()` → обновить `countShown` |
| 3.5 | `Enter` → `item.Model()` (accessor из `model_item.go`), убрать `statusMsg` |
| 3.6 | Добавить helper `setListItems(a *App, models []*modelscan.Model)` |

### Acceptance criteria
- `/` активирует поле ввода
- Ввод текста фильтрует список в реальном времени
- `Esc` восстанавливает полный список
- `↑↓` работают в обоих режимах
- `countShown` всегда согласован с содержимым списка

---

## Фаза 4 — Переписать App.View() с 3-блочной раскладкой

**Цель:** Экран рендерит точно 3 блока по схеме.

### Сборка экрана

```go
// Block 1: Header
header := RenderHeader(a.title, a.version, a.styles)

// Block 2: Content
filterRow := a.filterInput.RenderFilterBadge() // или RenderFilterInput()
countLabel := a.styles.CountLabelStyle().Render(
    fmt.Sprintf("Показано: %d / Всего: %d", a.countShown, a.countTotal))
listBlock := a.styles.ListBlockBorderStyle().Render(a.list.View())

content := lipgloss.JoinVertical(lipgloss.Left,
    filterRow, countLabel, listBlock)

// Block 3: Footer
footerLine := a.footer.Render()

// Полный экран
screen := lipgloss.JoinVertical(lipgloss.Left, header, content, footerLine)

v := tea.NewView(screen)
v.AltScreen = true
v.BackgroundColor = lipgloss.Color(a.styles.DarkBg)
return v
```

### Задачи

| # | Действие |
|---|----------|
| 4.1 | Header: вызвать `RenderHeader()` из `footer.go` |
| 4.2 | Filter row: переключать `RenderFilterBadge()` / `RenderFilterInput()` по `filterState` |
| 4.3 | Count label: `"Показано: %d / Всего: %d"` через `CountLabelStyle()` |
| 4.4 | List block: `ListBlockBorderStyle().Render()` — учесть frame size при расчёте высоты в фазе 3 |
| 4.5 | Footer: `a.footer.Render()` |
| 4.6 | Удалить fallback ветку `a.styles == nil` (styles всегда non-nil после фазы 2) |
| 4.7 | Guard для первого рендера (width==0): вернуть `tea.NewView("Loading...")` |

### Acceptance criteria
- Видны: `[v0.1.0]` badge + title
- Видны: `[FILTER]` / input field
- Видна строка `Показано: N / Всего: M`
- Список в rounded border (#4ade80)
- Footer с key hints: `↑↓ навигация │ Enter выбор │ / фильтр │ q выход`
- Фон `#0a0f18`

---

## Фаза 5 — Стилизованные бейджи в model_item

**Цель:** 3-я строка элемента показывает styled pills вместо plain строк.  
**Зависит:** Фаза 1 (импорт lipgloss в model_item.go).

### Целевой `formatMetadataBadges()`

```go
func formatMetadataBadges(m *modelscan.Model, st *StyleConfig) string {
    var parts []string
    parts = append(parts, st.SizeBadgeStyle().Render(formatSize(m.Size)))
    parts = append(parts, st.QuantizationBadgeStyle().Render(extractQuantization(m.Name)))
    if len(m.MMProjPaths) > 0 {
        parts = append(parts, st.MMProjBadgeStyle().Render("mmproj"))
    }
    return strings.Join(parts, " ")
}
```

### Задачи

| # | Действие |
|---|----------|
| 5.1 | Переименовать `formatQuantizationBadge` → `extractQuantization` (возвращает строку без скобок) |
| 5.2 | Применить стили: `SizeBadgeStyle`, `QuantizationBadgeStyle`, `MMProjBadgeStyle` |
| 5.3 | Удалить `formatSizeBadge()` если больше не используется |

### Acceptance criteria
- Бейджи рендерятся с рамками и фоном по схеме
- Тесты проходят

---

## Фаза 6 — Интеграционное тестирование

**Цель:** End-to-end проверка на реальном PowerShell / Windows Terminal.

### Smoke test сценарии

| # | Сценарий | Ожидаемый результат |
|---|----------|---------------------|
| 1 | Запуск с реальной папкой моделей | Экран с тёмным фоном, header, список |
| 2 | Нажать `/` | `[FILTER]` → поле ввода с курсором |
| 3 | Ввести часть имени | Список фильтруется, счётчик обновляется |
| 4 | Нажать `Esc` | Полный список восстановлен |
| 5 | `↑↓` | Курсор перемещается, активный элемент — с левым бордером |
| 6 | `Enter` | Выход из TUI, имя файла выводится в консоль |
| 7 | `q` | Чистый выход |
| 8 | Resize окна | Список адаптируется без артефактов |

---

## Реестр рисков

| Риск | Вероятность | Митигация |
|------|-------------|-----------|
| Pointer receiver `StyledDelegate` не удовлетворяет `list.ItemDelegate` | Низкая | Проверить интерфейс bubbles/v2; pointer всегда удовлетворяет interface |
| bubbles/v2 `list` перехватывает `/` несмотря на `SetFilteringEnabled(false)` | Средняя | Перехватывать `KeyPressMsg` до вызова `list.Update()` |
| `ListBlockBorderStyle` Padding+Margin не учтены → переполнение | Средняя | Использовать `style.GetVerticalFrameSize()` динамически |
| `v.BackgroundColor` не работает в legacy conhost | Низкая (pre-existing) | `SetTerminalBackground()` как fallback уже есть |

---

## Сводная таблица фаз

| Фаза | Файлы | Объём изменений | Риск | После фазы |
|------|-------|-----------------|------|-----------|
| 1 | cli.go, model_item.go, filter.go | ~120 строк удалить, ~4 добавить | Низкий | Компилируется |
| 2 | cli.go | ~50 строк переписать | Низкий | App struct готов |
| 3 | cli.go | ~80 строк переписать | Средний | Клавиши работают |
| 4 | cli.go | ~60 строк переписать | Низкий | Визуал готов |
| 5 | model_item.go | ~20 строк переписать | Низкий | Бейджи стилизованы |
| 6 | — | 0 кода | Низкий | MVP готов |
