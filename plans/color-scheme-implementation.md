# План внедрения зеленой цветовой схемы

## Приоритет: CLI UI (TUI) > Web UI

---

## 🎯 Цель

Внедрить расширенную зеленую цветовую схему в проект llama-server-loader с соблюдением требований доступности WCAG AA.

### Цветовая палитра

| Название | HEX | Использование | Контраст на #0f172a |
|----------|-----|---------------|----------------------|
| **Зеленый темный** | `#064e3b` | Текст на зеленом фоне | 8.2:1 ✅ WCAG AAA |
| **Зеленый основной** | `#34d399` | Кнопки CTA, активные элементы | 5.8:1 ✅ WCAG AA |
| **Зеленый яркий** | `#6ee7b7` | Hover-эффекты, границы | 4.2:1 ⚠️ Крупный текст |
| **Зеленый неоновый** | `#00ff88` | Декоративные акценты (без текста) | 3.5:1 ❌ Только фон |

---

## 📋 Этапы реализации

### Этап 1: CLI TUI — Создание модуля стилей (Приоритет: Высокий)

**Файл:** `internal/cli/styles.go` (новый файл)

#### Задачи:
1. Создать новый файл `styles.go` в пакете `cli`
2. Определить структуру `StyleConfig` с цветовыми токенами
3. Реализовать функции для создания стилей lipgloss:
   - `HeaderStyle()` — стиль заголовка (зеленый фон, темный текст)
   - `FooterStyle()` — стиль футера (серый текст)
   - `ItemSelectedStyle()` — стиль выбранного элемента списка
   - `StatusSuccessStyle()` — стиль успешных сообщений

#### Пример реализации:
```go
package cli

import "charm.land/lipgloss/v2"

// StyleConfig определяет цветовую схему TUI
type StyleConfig struct {
    GreenDark     lipgloss.Color // #064e3b - текст на зеленом
    GreenPrimary  lipgloss.Color // #34d399 - основной CTA
    GreenBright   lipgloss.Color // #6ee7b7 - hover, границы
    TextPrimary   lipgloss.Color // #f8fafc - основной текст
    TextSecondary lipgloss.Color // #cbd5e1 - вторичный текст
}

// GetStyles возвращает конфигурацию стилей
func GetStyles() *StyleConfig {
    return &StyleConfig{
        GreenDark:     lipgloss.Color("#064e3b"),
        GreenPrimary:  lipgloss.Color("#34d399"),
        GreenBright:   lipgloss.Color("#6ee7b7"),
        TextPrimary:   lipgloss.Color("#f8fafc"),
        TextSecondary: lipgloss.Color("#cbd5e1"),
    }
}

// HeaderStyle возвращает стиль заголовка
func (s *StyleConfig) HeaderStyle() lipgloss.Style {
    return lipgloss.NewStyle().
        Bold(true).
        Padding(1, 2).
       Background(s.GreenPrimary).
       Foreground(s.GreenDark) // Темный текст на зеленом!
}

// FooterStyle возвращает стиль футера
func (s *StyleConfig) FooterStyle() lipgloss.Style {
    return lipgloss.NewStyle().
        PaddingLeft(1).
        Foreground(s.TextSecondary)
}
```

---

### Этап 2: CLI TUI — Обновление компонента App

**Файл:** `internal/cli/cli.go` (строки 478-567)

#### Задачи:
1. Добавить поле `styles *StyleConfig` в структуру `App`
2. Инициализировать стили в функции `NewApp()`
3. Обновить метод `View()` для использования стилей из `StyleConfig`:
   - Применить `HeaderStyle()` к заголовку
   - Применить `FooterStyle()` к статусному сообщению
   - Добавить цветные границы для списка моделей

#### Изменения в коде:
```go
// В структуре App добавить поле styles
type App struct {
    list      list.Model
    selected  *modelscan.Model
    statusMsg string
    width     int
    height    int
    title     string
    filtering bool
    styles    *StyleConfig // Добавить это поле
}

// В функции NewApp() инициализировать стили
func NewApp(models []*modelscan.Model) *App {
    // ... существующий код создания списка ...
    
    return &App{
        list:      l,
        statusMsg: "Используйте стрелки для навигации, Enter для выбора",
        title:     "llama-server-loader - Model Selector",
        styles:    GetStyles(), // Добавить эту строку
    }
}

// В методе View() использовать стили
func (a *App) View() tea.View {
    header := a.styles.HeaderStyle().Render(a.title)
    footer := a.styles.FooterStyle().Render(a.statusMsg)
    
    content := lipgloss.JoinVertical(
        lipgloss.Center,
        header,
        a.list.View(),
        footer,
    )
    
    v := tea.NewView(content)
    v.AltScreen = true
    return v
}
```

---

### Этап 3: CLI TUI — Стилизация списка моделей

**Файл:** `internal/cli/cli.go` (строки 478-505)

#### Задачи:
1. Настроить delegate списка для применения стилей к элементам
2. Добавить подсветку выбранного элемента зеленым цветом
3. Обновить стиль фильтрации (когда пользователь вводит текст поиска)

#### Пример настройки delegate:
```go
// В функции NewApp() настроить delegate
d := list.NewDefaultDelegate()
d.TitleStyle = func(lm list.Item) lipgloss.Style {
    return lipgloss.NewStyle().Foreground(styles.TextPrimary)
}
d.IndicatorStyle = func(selected bool) lipgloss.Style {
    if selected {
        return lipgloss.NewStyle().Foreground(styles.GreenBright).PaddingRight(1)
    }
    return lipgloss.NewStyle().Foreground(styles.TextSecondary)
}

l := list.New(items, d, 60, 20)
```

---

### Этап 4: Web UI — Обновление CSS переменных

**Файл:** `internal/webui/static/style.css` (строки 1-223)

#### Задачи:
1. Добавить CSS кастомные свойства в селектор `:root`
2. Обновить существующие стили для использования новых цветов
3. Исправить контрастность кнопок и статусных индикаторов

#### Изменения в коде:
```css
/* В начале файла добавить :root с цветовыми токенами */
:root {
    /* Зеленая шкала (основная) */
    --green-900: #064e3b;     /* Темный — текст на зеленом */
    --green-800: #059669;     /* Вторичные кнопки, иконки */
    --green-700: #10b981;     /* Hover состояния */
    --green-600: #34d399;     /* Основной CTA (контраст 5.8:1) */
    --green-500: #6ee7b7;     /* Границы, акценты */
    
    /* Фон и текст */
    --bg-primary: #0f172a;
    --text-primary: #f8fafc;
    --text-secondary: #cbd5e1;
}

/* Обновить кнопки для использования зеленого */
.btn-success {
    background-color: var(--green-600);
    color: var(--green-900);  /* Темный текст на зеленом! */
}

.btn-primary {
    background-color: transparent;
    border-color: var(--green-700);
    color: var(--green-600);
}

/* Обновить статусные индикаторы */
.status-indicator.running {
    background-color: var(--green-800);
    color: white;  /* Белый на темном зеленом — контраст 8.2:1 ✅ */
}

/* Обновить выделение элементов списка */
.model-item.selected {
    border-left: 3px solid var(--green-600);
    background-color: rgba(52, 211, 153, 0.1);
}
```

---

### Этап 5: Тестирование и валидация

#### Задачи:
1. Проверить контрастность всех комбинаций цветов (WCAG AA)
2. Протестировать TUI интерфейс с разными размерами терминала
3. Протестировать Web UI в разных браузерах

#### Критерии успеха:
- ✅ Все текстовые элементы имеют контраст ≥ 4.5:1 на фоне
- ✅ Кнопки CTA используют зеленый цвет с темным текстом
- ✅ Выбранные элементы списка визуально выделяются
- ✅ Статусные индикаторы читаемы и соответствуют WCAG

---

## 📊 Метрики успеха

| Метрика | Целевое значение | Текущее состояние |
|---------|------------------|-------------------|
| Контрастность текста на фоне | ≥ 4.5:1 (WCAG AA) | Требуется исправление |
| Количество цветовых токенов | 5 (зеленая шкала) | Не реализовано |
| Покрытие стилей CLI TUI | 100% компонентов | Частично реализовано |
| Покрытие стилей Web UI | 100% компонентов | Требуется обновление |

---

## 🔧 Технические детали

### Зависимости для CLI:
- `charm.land/lipgloss/v2` — уже установлена (v2.0.3)
- Поддержка цветов в формате HEX через `lipgloss.Color("#xxxxxx")`

### Структура файлов после изменений:
```
internal/cli/
├── cli.go          # Обновить методы View() и NewApp()
├── styles.go       # Новый файл с StyleConfig
└── cli_test.go     # Добавить тесты для стилей

internal/webui/static/
├── style.css       # Обновить CSS переменные и стили
├── index.html      # Проверить использование цветов
└── app.js          # Проверить JS-логику статусов
```

---

## 🚀 Порядок выполнения (приоритет)

1. **Создать `styles.go`** — модуль стилей для CLI TUI
2. **Обновить `cli.go`** — интегрировать стили в App компонент
3. **Обновить `style.css`** — внедрить CSS переменные и исправить контрастность
4. **Протестировать** — запустить приложение и проверить визуальные изменения

---

## ⚠️ Риски и ограничения

| Риск | Влияние | Стратегия смягчения |
|------|----------|---------------------|
| Неподдержка HEX цветов в lipgloss | Критическое | Проверить документацию, использовать RGB если нужно |
| Низкая контрастность на светлых фонах | Высокое | Добавить темную тему как основную |
| Конфликт с существующими стилями | Среднее | Постепенное внедрение с тестированием |

---

## 📝 Примечания

- Зеленый цвет используется для позитивных действий (успех, запуск, подтверждение)
- Темный текст на ярких кнопках обеспечивает доступность
- Розовый акцент (#ec4899) сохраняется как вторичный цвет для ссылок и hover