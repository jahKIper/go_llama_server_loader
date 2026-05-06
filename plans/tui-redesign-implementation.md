# План: редизайн TUI экрана выбора моделей

## Context

Текущий TUI ([internal/cli/cli.go](d:\src\go_llama_server_loader\internal\cli\cli.go), [internal/cli/styles.go](d:\src\go_llama_server_loader\internal\cli\styles.go)) уже использует зелёную палитру и 3-блочную раскладку (header / content / footer), но интерфейс остаётся «плоским»: нет визуальной иерархии активных элементов, поле фильтра в активном режиме не выделено, список не имеет индикатора прокрутки, и в шапке нет навигации между разделами.

Пользователь предоставил референс (Bagels TUI) и потребовал:
1. Перенести стилистику референса: скруглённые рамки с встроенным title-лейблом, двухстрочные элементы списка с цветным маркером, плотный однострочный футер.
2. Добавить **scrollbar** справа от списка и **tab bar** в правой части header (Models / Running / Logs — последние два как визуальный каркас).
3. **Неоновый эффект** на активные элементы (выбранная строка, активная рамка, активный таб).
4. **Выделить поле ввода фильтра** в активном состоянии.

Цель — визуально различимая иерархия (где фокус, что активно, где скроллится) при сохранении WCAG AA контраста и существующей зелёной палитры.

## Палитра (расширение [internal/cli/styles.go](d:\src\go_llama_server_loader\internal\cli\styles.go))

**Принципы:**
- Без дубликатов имён под одним цветом (`GreenPrimary` используется напрямую вместо алиаса `NeonGlow`).
- Сжатая сетка фонов — только 3 уровня глубины.
- Неон — мягкий Dracula-оттенок `#50fa7b`, не агрессивный «hacker green».
- WCAG AA проверен для всех текстовых пар (≥4.5:1).

**Новые токены в `StyleConfig`:**

| Токен | HEX | Назначение |
|---|---|---|
| `BgPanel` | `#0d1320` | Фон интерьера content-блока (приподнятая панель над экраном) |
| `BgSelected` | `#1a2e25` | Заливка выбранной строки списка (зелёный тинт) |
| `BorderIdle` | `#1f2937` | Приглушённые рамки неактивных элементов (idle filter, scrollbar track, разделитель в футере) |
| `NeonGreen` | `#50fa7b` | Неоновый акцент: bracket selected-строки, рамка активного фильтра, scrollbar thumb, halo, активная рамка content-блока |
| `TextMuted` | `#94a3b8` | Приглушённый текст: placeholder фильтра, label хоткеев в футере, disabled tab |
| `KeyHint` | `#6ee7b7` | Цвет клавиш в футере (новый цвет, отличается от существующего `GreenBright #4ade80`) |

**Удалены из плана:** `NeonGlow` (дубль `GreenPrimary`), `BgFilterActive` (используем `BgPanel` + неоновую рамку для отличия от idle), `BgFooter` (футер использует тот же `DarkBg` что и экран — отделение через зазор/типографику, без bg-контраста).

**Существующие токены сохраняются:** `GreenDark` `#064e3b`, `GreenPrimary` `#34d399`, `TextPrimary` `#ffffff`, `TextSecondary` `#e2e8f0`, `DarkBg` `#0a0f18`.

**Удаляются** из `StyleConfig`: `GreenBright` `#4ade80` и `ListBorder` `#4ade80` — сводим к 2 зелёным акцентам:
- `NeonGreen` `#50fa7b` — активность/фокус (рамка content-блока, активный фильтр, bracket selected-строки, scrollbar thumb, help popup, курсор).
- `GreenPrimary` `#34d399` — статичные акценты (header version badge, активный таб, halo selected-строки, граница и заливка mmproj-бейджа, верхняя граница ранее использовавшая `GreenBright` в footer/badges).

Все ссылки на `GreenBright` и `ListBorder` в существующих стилях ([styles.go:23](d:\src\go_llama_server_loader\internal\cli\styles.go#L23), [styles.go:27](d:\src\go_llama_server_loader\internal\cli\styles.go#L27), `VersionBadgeStyle`, `FilterBadgeStyle`, `ItemActiveBorderStyle`, `ItemSelectedStyle`, `ContentBlockStyle`, `FooterContainerStyle`, `SizeBadgeStyle`, `MMProjBadgeStyle`) — заменить на `NeonGreen` или `GreenPrimary` согласно карте выше.

**Сетка фонов (3 уровня глубины):**
1. `DarkBg` `#0a0f18` — экран и футер.
2. `BgPanel` `#0d1320` — интерьер content-блока, активного поля фильтра.
3. `BgSelected` `#1a2e25` — выбранная строка списка.

**Контрастные пары (WCAG AA):**
- `TextPrimary` на `BgSelected` → ~12:1 ✅ AAA
- `TextPrimary` на `BgPanel` → ~16:1 ✅ AAA
- `TextMuted` `#94a3b8` на `BgPanel` → ~5.6:1 ✅ AA
- `TextMuted` на `DarkBg` (футер) → ~6.4:1 ✅ AA
- `GreenDark` на `NeonGreen` `#50fa7b` → ~6.5:1 ✅ AAA
- `GreenDark` на `GreenPrimary` `#34d399` (active tab, version badge) → ~5.2:1 ✅ AA
- `KeyHint` `#6ee7b7` на `BgPanel` (subtle bg клавиш в футере) → ~9:1 ✅ AAA

**Открытые вопросы (решаются после реализации):**
- Контраст бейджей `Size`/`mmproj`/`Quant` **внутри** selected-строки на фоне `BgSelected` — проверить эмпирически и при необходимости подкорректировать `BgSelected` или сами бейджи (E из обсуждения пункта 2).

## Неоновый эффект (TUI-приём)

В терминале нет настоящего glow, поэтому имитируем через комбинацию **3 приёмов**:
1. **Halo-строки** — символы `▔` сверху и `▁` снизу выбранной строки цветом `GreenPrimary` создают «свечение» по вертикали (1 строка сверху + 1 снизу, не больше — иначе item «прыгает» при навигации).
2. **Bold + bright text** (`TextPrimary` + `Bold(true)`) для содержимого активного элемента.
3. **Tinted background** (`BgSelected`) — мягкий зелёный фон под текстом, имитирует рассеянный свет.

Плюс F-bracket обрамление (top + left + bottom border `NeonGreen`) — несущая граница активной зоны.

**Карта неона** (где `NeonGreen #50fa7b` появляется):

| Элемент | Эффект |
|---|---|
| Selected list item — bracket | Border NeonGreen (top + left + bottom, полная ширина) |
| Selected list item — halo | `▔`/`▁` цвета `GreenPrimary` (приём #1 — мягче неона) |
| Selected list item — фон | `BgSelected` + Bold TextPrimary (приёмы #2, #3) |
| Активный фильтр | RoundedBorder NeonGreen + курсор `▌` NeonGreen |
| Content-блок | RoundedBorder NeonGreen (всегда — главный экран) |
| Scrollbar thumb | `┃` NeonGreen |
| Help popup | RoundedBorder NeonGreen |

**Что НЕ неон** (зарезервировано для статичных акцентов): активный таб, header version badge, halo selected-строки, активные бейджи внутри item — используют `GreenPrimary` `#34d399`.

## Изменения по файлам

### 1. [internal/cli/styles.go](d:\src\go_llama_server_loader\internal\cli\styles.go) — расширение токенов и стилей

**Структура файла** — один файл с явными секциями-разделителями:
```
// === Palette ===
// === Header / Tabs ===
// === Content block ===
// === Filter input ===
// === List item ===
// === Scrollbar ===
// === Footer ===
// === Help popup ===
```

**Backward compatibility — hard break:**
- Удалить из `StyleConfig` поля `GreenBright` (`#4ade80`) и `ListBorder` (`#4ade80`). Все их использования заменить на `NeonGreen` или `GreenPrimary` (см. карту неона).
- Удалить `ItemActiveBorderStyle()` ([styles.go:67-72](d:\src\go_llama_server_loader\internal\cli\styles.go#L67-L72)) — мёртвый после редизайна (selected обрабатывает `ItemSelectedStyle` с F-bracket; ссылка в [model_item.go:189](d:\src\go_llama_server_loader\internal\cli\model_item.go#L189) `TitleStyleSelected` тоже удаляется, т.к. фактический рендер идёт через `ListItem.Render`).
- Тесты в [cli_test.go](d:\src\go_llama_server_loader\internal\cli\cli_test.go) и [app_integration_test.go](d:\src\go_llama_server_loader\internal\cli\app_integration_test.go), ссылающиеся на удалённые поля/методы, обновляются в этом же PR.

**Динамическая ширина:** стили остаются «чистыми» (без зашитой ширины). Компоненты применяют `.Width(w)` на месте рендера (паттерн как в текущем `Footer.SetWidth` + `applyContainer`).

**Изменения / новые стили:**

- Добавить поля палитры из таблицы выше в `StyleConfig` и `GetStyles()`.
- **Переписать `ItemSelectedStyle()`** — неоновая выбранная строка с **«F-bracket» обрамлением (top + left + bottom, полная ширина)**:
  - Border: `lipgloss.NormalBorder()` с флагами `(top=true, right=false, bottom=true, left=true)` — `[`-скобка во всю ширину item: top-граница `┌─────...─` тянется до правого края, vertical `│`, bottom-граница `└─────...─` тоже на всю ширину. Без правой границы — открыто справа.
  - `BorderForeground`: `NeonGreen` (`#00ff88`) — неоновый акцент.
  - Background: `BgSelected` (на всю область внутри bracket).
  - Foreground: `TextPrimary` + `Bold(true)`.
  - Padding `0 1` (внутри bracket).
- Добавить `ItemHaloStyle()` — рендерит строку из `▔`/`▁` цветом `GreenPrimary` для halo сверху/снизу bracket-обрамления (полный halo: bracket + halo-строки выше top-границы и ниже bottom-границы).
- **Переписать `ContentBlockStyle()`** — двойная неоновая рамка активного блока:
  - Внешняя — `lipgloss.RoundedBorder()` цвета `NeonGreen`.
  - Внутренняя имитация — padding `1 2` с фоном `BgPanel` для контраста с внешним `DarkBg`.
- Добавить `FilterInputActiveStyle()` — стиль активного поля ввода:
  - `lipgloss.RoundedBorder()` цвета `NeonGreen`.
  - Background `BgPanel` (тот же что у content-блока — отличие от idle обеспечивается неоновой рамкой), Foreground `TextPrimary`.
  - Padding `0 1`, ширина = ширина content-блока минус отступы.
- Добавить `FilterInputIdleStyle()` — для не-активного поля (тонкая рамка `BorderIdle`).
- Добавить scrollbar-стили: `ScrollbarTrackStyle()` (`│` цвет `BorderIdle`) и `ScrollbarThumbStyle()` (`┃` цвет `NeonGreen`).
- Добавить tab-стили: `TabActiveStyle()` (заливка `GreenPrimary` + текст `GreenDark` + `Bold` — **потише, без неона**; неон зарезервирован только для selected-строки списка), `TabInactiveStyle()` (текст `TextSecondary`), `TabDisabledStyle()` (текст `TextMuted`), `TabSeparatorStyle()` (`│` цвет `BorderIdle`).
- Обновить `FooterContainerStyle()` — **убрать top border полностью** (минимализм, отделение от content-блока обеспечивается зазором). Background `DarkBg` (тот же что у экрана — без bg-контраста), padding `0 2`.
- Добавить `FooterKeyStyle()` — клавиша: bold-текст цвета `KeyHint` на фоне `BgPanel` (`#0d1320`, subtle подсветка), padding `0 1`, без рамки.
- Добавить `FooterLabelStyle()` — label: `TextMuted` без фона.
- Добавить `FooterSeparatorStyle()` — `│` цвета `BorderIdle` для разделения пар key+label.
- Добавить `HelpPopupStyle()` — стиль popup-панели полной справки: `RoundedBorder` цвета `NeonGreen`, фон `BgPanel`, padding `1 2`, рендерится по центру экрана поверх content-блока.

### 2. [internal/cli/scrollbar.go](d:\src\go_llama_server_loader\internal\cli\scrollbar.go) — новый файл

**Положение:** последняя колонка внутри списка (между списком и правой рамкой content-блока — впритык, без пробела). Без модификации border-символов content-блока.

Функция `RenderScrollbar(offset, visible, total, height int, st *StyleConfig) string`:
- Возвращает многострочный блок шириной 1 (для `lipgloss.JoinHorizontal` со списком).
- **Скрытие:** если `total <= visible` — возвращает колонку пробелов высотой `height` (UI не «прыгает», но визуально scrollbar отсутствует).
- Размер thumb: `max(1, height * visible / total)`.
- Позиция thumb: `offset * (height - thumbSize) / max(1, total - visible)`.
- Каждая строка: `track` (`│` цвет `BorderIdle`) или `thumb` (`┃` цвет `NeonGreen`).
- В `App.recomputeListSize()` вычесть 1 из `listW` под scrollbar; параметры расчёта: `visible = listH / itemHeight` где `itemHeight` берётся из `StyledDelegate.Height()` (после изменений = 7).

**Источник offset/visible/total для вызова из App** — исследовать в момент реализации: bubbles/v2 `list.Model` API на наличие `ScrollOffset()` или эквивалента. Fallback при отсутствии прямого метода: аппроксимация `offset ≈ clamp(Index() - visible/2, 0, total - visible)`. Аппроксимация остаётся в `App.View()` (или helper), сама `RenderScrollbar` принимает уже посчитанные числа — это позволяет тестировать её детерминированно.

**Тесты:** `TestRenderScrollbar_HidesWhenFits`, `TestRenderScrollbar_ThumbAtTop`, `TestRenderScrollbar_ThumbAtBottom`, `TestRenderScrollbar_ThumbMiddle`, `TestRenderScrollbar_SmallList` (граничные случаи `total = visible + 1`, `total = 1`).

### 3. [internal/cli/tabs.go](d:\src\go_llama_server_loader\internal\cli\tabs.go) — новый файл

```go
type Tab struct {
    Label   string
    Enabled bool
}

type TabBar struct {
    tabs   []Tab
    active int
    styles *StyleConfig
}
```

Методы:
- `NewTabBar(st *StyleConfig) *TabBar` — создаёт `[{Models, true}, {Running, false}, {Logs, false}]`, `active=0`.
- `Render() string` — собирает строку: `TabActiveStyle("Models") │ TabDisabledStyle("Running") │ TabDisabledStyle("Logs")`.
- `SetActive(i int)`, `Next()`, `Prev()` — переключение между **enabled** табами; если enabled-таб только один, `Next()`/`Prev()` **молча no-op** (ошибки не возвращаем, фокус не меняем).

### 4. [internal/cli/footer.go](d:\src\go_llama_server_loader\internal\cli\footer.go) — рестайл + popup

**Footer (одна строка, минималистичный):**
- В `renderRow()` ([internal/cli/footer.go:102-106](d:\src\go_llama_server_loader\internal\cli\footer.go#L102-L106)) — заменить `VersionBadgeStyle` на новый `FooterKeyStyle()` (subtle bg `BgPanel` + `KeyHint` bold), label через `FooterLabelStyle()`.
- Разделитель между парами в `renderLine()` ([internal/cli/footer.go:93-99](d:\src\go_llama_server_loader\internal\cli\footer.go#L93-L99)) — `FooterSeparatorStyle().Render(" │ ")`.
- **Раздельные группы:** `compactRows` разделяется на `leftRows` (всё кроме `q`) и `rightRows` (только `^q выход`). В `Render()` рендерим обе группы и склеиваем через `lipgloss.JoinHorizontal` со spacer-ом, выравнивающим правую группу к правому краю (`width - leftWidth - rightWidth` пробелов посередине). Биндинг изменить с `q` на `^q` (Ctrl+Q) для согласованности с конвенцией Quit.
- Убрать top border (см. изменения в `FooterContainerStyle`).
- Expanded режим (через `?`) **больше не используется в footer** — переезжает в popup (см. ниже). Поле `expanded` и `extraRows` из `Footer` можно удалить, либо оставить как fallback. Метод `SetExpanded()` удалить.

**Help popup (новый):**
- Новый файл [internal/cli/help_popup.go](d:\src\go_llama_server_loader\internal\cli\help_popup.go).
- Структура `HelpPopup` со списком всех биндингов (compact + extra), методом `Render(width, height int) string` — рендерит centered popup через `HelpPopupStyle()`. Размеры: ширина ~50% экрана (clamp 40-80 cols), высота автоматическая по содержимому.
- Содержимое: заголовок «Справка», далее таблица `key  │  description` для всех бинд (включая `↑↓`, `Enter`, `/`, `Esc`, `?`, `^q`, `1`/`2`/`3`, `Tab`, `g/G`, `Home/End`, `PgUp/PgDn`).
- Внизу popup строка `Esc или ? — закрыть` цветом `TextMuted`.
- **Рендеринг — полноэкранный режим (без overlay):** в `App.View()` если `helpExpanded == true`, возвращается экран только с centered popup через `lipgloss.Place(width, height, Center, Center, popup, lipgloss.WithWhitespaceBackground(DarkBg))`. Список и footer на это время не видны — пользователь читает справку. По `Esc`/`?` popup закрывается, обычный рендер возвращается. Это упрощает реализацию (~5 строк вместо ~30 со склейкой строк) ценой потери фонового контекста, что для help-popup приемлемо.

### 5. [internal/cli/filter.go](d:\src\go_llama_server_loader\internal\cli\filter.go) — выделение активного поля

**Поле всегда видимо** (idle и active), стабильная форма UI. Префикс `/  ` слева, ширина = ширина content-блока.

- `RenderFilterBadge()` ([internal/cli/filter.go:127-132](d:\src\go_llama_server_loader\internal\cli\filter.go#L127-L132)) переименовать в `RenderFilterIdle()`:
  - Возвращает `FilterInputIdleStyle().Width(blockWidth).Render("/  поиск...")`.
  - Стиль: `RoundedBorder` цвета `BorderIdle` (приглушённый), фон `BgPanel`, текст placeholder цветом `TextMuted`.
- `RenderFilterInput()` ([internal/cli/filter.go:135-155](d:\src\go_llama_server_loader\internal\cli\filter.go#L135-L155)):
  - Возвращает `FilterInputActiveStyle().Width(blockWidth).Render("/  " + text + cursor)`.
  - Стиль: `RoundedBorder` цвета `NeonGreen`, фон `BgPanel`, текст `TextPrimary`.
  - Префикс `/  ` (два пробела для дыхания) — статичная часть, не редактируется.
  - Курсор `▌` цветом `NeonGreen` (через `lipgloss.NewStyle().Foreground(NeonGreen).Render("▌")`).
  - **Без inline-счётчика** — счётчик `Показано/Всего` показывается только в верхней границе content-блока (см. [cli.go](d:\src\go_llama_server_loader\internal\cli\cli.go) интеграция).
- Метод `filterInput.Render(blockWidth int) string` — единая точка входа: возвращает idle или active в зависимости от `f.state`. App вызывает `f.Render(width)` вместо текущей развилки `RenderFilterBadge()` / `RenderFilterInput()`.
- Сигнатуры существующих методов `RenderFilterBadge()` / `RenderFilterInput()` сохранить как обёртки над `Render()` для совместимости с тестами, либо обновить тесты.

### 6. [internal/cli/model_item.go](d:\src\go_llama_server_loader\internal\cli\model_item.go) — обновление рендера элемента

**Структура остаётся 5-строчной** (текущая): name(1) + path(1) + badges-bordered(3) = 5 строк, `Height()=5`, `Spacing()=1`. Не меняем.

Изменения:
- Добавить **точку-индикатор** `●` слева перед именем модели, цвет по квантованию (новая функция `quantColor(quant string, st *StyleConfig) lipgloss.Color`):
  - `Q5_*`, `Q6_*`, `Q8`, `F16`, `F32` → `NeonGreen`
  - `Q4_*` → `GreenPrimary`
  - `Q3_*`, `Q2_*` → `TextSecondary`
  - Реализация: рендер `●` через `lipgloss.NewStyle().Foreground(quantColor(...))` + пробел перед `model.Name`.
- Если `selected` — **полный halo + F-bracket обрамление**:
  - Весь блок (name + path + badges) заворачивается в `ItemSelectedStyle()` с `[`-обрамлением (top + left + bottom border неоновым `NeonGreen`).
  - Над bracket рисуется halo-строка `▔` цветом `GreenPrimary` (ширина = ширина item).
  - Под bracket — halo-строка `▁` цветом `GreenPrimary`.
  - Имя внутри bracket — `Bold(true)` + `TextPrimary`.
  - Path — остаётся `TextSecondary` (без поднятия яркости).
  - Bracket добавляет 2 строки (top + bottom border), halo добавляет 2 строки → selected item визуально занимает `5 + 2 + 2 = 9` строк. Базовая `Height()=5`, `Spacing()=1`.
  - **Решение по геометрии:** увеличить `Height()` до `7` (5 контента + 1 halo top + 1 halo bottom), `Spacing()` оставить `1`. Bracket top/bottom рисуются на тех же 2-х halo-строках (символы `┌─...─` и `└─...─` совмещаются с halo `▔`/`▁` — для selected рисуется bracket с halo по краям, для unselected — пустые строки). Это даёт стабильную высоту item = 7 строк независимо от выбора, что укладывается в API `bubbles/list`. Trade-off: список вмещает на ~30% меньше моделей, чем при `Height=5`. Альтернатива — позже добавить compact mode (toggle через хоткей).
- Путь — `TextSecondary`, truncation слева через `truncatePathLeft` (без изменений).
- mmproj badge — текст `mmproj` (без счётчика).

### 7. [internal/cli/cli.go](d:\src\go_llama_server_loader\internal\cli\cli.go) — интеграция

В структуре `App` добавить поля:
```go
tabs      *TabBar
helpPopup *HelpPopup
```
Существующее `helpExpanded` остаётся как флаг видимости popup.

В `NewApp()` ([internal/cli/cli.go:476-506](d:\src\go_llama_server_loader\internal\cli\cli.go#L476-L506)) — `tabs: NewTabBar(st)`, `helpPopup: NewHelpPopup(st)`.

**`Update()` ([internal/cli/cli.go:514-594](d:\src\go_llama_server_loader\internal\cli\cli.go#L514-L594)) — расширение существующего switch** (по аналогии с текущим паттерном для `q`/`?`):

Новые кейсы вне filter mode:
- `"1"` / `"2"` / `"3"` → `a.tabs.SetActive(n-1)`.
- `"tab"` → `a.tabs.Next()`, `"shift+tab"` → `a.tabs.Prev()`.
- `"?"` → toggle `a.helpExpanded`. Старый вызов `a.footer.SetExpanded(...)` удалить (метод удалён по решению пункта 6).

**Esc-иерархия** (явные проверки в начале обработки `"esc"`):
1. Если `helpExpanded == true` → `helpExpanded = false`, return.
2. Иначе если `filterState != FilterIdle` → выйти из фильтра (текущая логика).
3. Иначе → no-op.

**`View()` ([internal/cli/cli.go:656-705](d:\src\go_llama_server_loader\internal\cli\cli.go#L656-L705)):**

- **Полноэкранный help-popup** (early return в начале `View()` если `helpExpanded`):
  ```go
  if a.helpExpanded {
      popup := a.helpPopup.Render(a.width, a.height)
      screen := lipgloss.Place(a.width, a.height,
          lipgloss.Center, lipgloss.Center, popup,
          lipgloss.WithWhitespaceBackground(lipgloss.Color(a.styles.DarkBg)))
      v := tea.NewView(screen)
      v.AltScreen = true
      v.BackgroundColor = lipgloss.Color(a.styles.DarkBg)
      return v
  }
  ```
- **Header:** `lipgloss.JoinHorizontal(versionBadge, titleCentered, tabsRendered)` — version слева, title по центру, tabs справа. **Filter-state badge удалён.** Существующую логику центрирования из `RenderHeader()` ([footer.go:127-148](d:\server\internal\cli\footer.go#L127-L148)) обновить: `available = width - badgeW - tabsW`, title центрируется в этой полосе.
- **Content:** счётчик `Показано: X / Всего: Y` инжектится в верхнюю границу content-блока через `injectBorderTitle(rendered, "Models", fmt.Sprintf("%d / %d", shown, total))`. Реализация — манипуляция строк: `lines := strings.Split(rendered, "\n")`, в `lines[0]` (верхняя граница) находится последовательность `─` между угловыми символами `╭` и `╮`; вычисляются позиции вставки (после `╭`, перед `╮`), участки `─` заменяются на ` ─ Models ─ ` и ` ─ N / M ─ `. Утилита в новом файле [internal/cli/border_title.go](d:\src\go_llama_server_loader\internal\cli\border_title.go) с runes-safe обработкой (UTF-8) и обработкой случая, когда лейблы не помещаются (border слишком узкий — рендерим без лейблов).
- **Список + scrollbar:** `lipgloss.JoinHorizontal(Top, a.list.View(), RenderScrollbar(offset, visible, total, listH, st))`. Параметры:
  - `total = len(a.list.Items())`
  - `visible = listH / itemHeight` где `itemHeight = StyledDelegate.Height()` (`= 7` после изменений)
  - `offset` — через bubbles API (исследуется при реализации); fallback — аппроксимация `clamp(Index() - visible/2, 0, total - visible)`.

В `recomputeListSize()` ([internal/cli/cli.go:598-630](d:\src\go_llama_server_loader\internal\cli\cli.go#L598-L630)):
- Вычесть 1 колонку из `listW` под scrollbar.
- `listH` пересчитать с учётом нового `itemHeight = 7`.

## Verification

1. **Сборка:** `go build ./...` из корня проекта.
2. **Юнит-тесты:** `go test ./internal/cli/...` — проверить, что существующие тесты ([internal/cli/cli_test.go](d:\src\go_llama_server_loader\internal\cli\cli_test.go), [internal/cli/app_integration_test.go](d:\src\go_llama_server_loader\internal\cli\app_integration_test.go)) проходят. Добавить тесты:
   - `TestRenderScrollbar` — позиции thumb для разных offset/total/visible.
   - `TestTabBarNext_SkipsDisabled` — `Next()` пропускает disabled табы.
   - `TestInjectBorderTitle` — корректная вставка лейблов в верхнюю границу.
3. **Визуальная проверка:**
   - Запуск: `go run ./cmd/llama-server-loader --scan-dir=./testdata` (или путь с .gguf).
   - Проверить:
     - В шапке справа видны табы — `Models` подсвечен неоном, `Running`/`Logs` приглушены.
     - Content-блок имеет неоновую рамку, в верхней границе слева `─ Models ─`, справа `─ N / M ─`.
     - Поле фильтра в idle — тонкая рамка с placeholder; нажатие `/` → яркая неоновая рамка + tinted фон + курсор `▌`.
     - Выбранная строка списка — неоновый `┃` слева, фон с зелёным тинтом, halo-строки `▔`/`▁` сверху/снизу.
     - Справа от списка вертикальный scrollbar с зелёным thumb (если моделей больше, чем влезает).
     - Footer — одна строка, клавиши цветные без бейджа, подписи приглушённые.
   - Проверить навигацию: `↑↓`, `Enter`, `/`, `Esc`, `1`/`2`/`3` (последние два не должны падать — disabled, но не ошибка), `Tab`, `?`, `q`.
4. **Контраст:** все текстовые комбинации (TextPrimary на BgSelected, TextMuted на DarkBg/BgPanel, GreenDark на NeonGreen, KeyHint на BgPanel) проверить через [WebAIM Contrast Checker](https://webaim.org/resources/contrastchecker/) — должно быть ≥ 4.5:1 (целевые значения зафиксированы в секции «Палитра»).

## Файлы

**Изменяются:**
- [internal/cli/styles.go](d:\src\go_llama_server_loader\internal\cli\styles.go)
- [internal/cli/cli.go](d:\src\go_llama_server_loader\internal\cli\cli.go)
- [internal/cli/footer.go](d:\src\go_llama_server_loader\internal\cli\footer.go)
- [internal/cli/filter.go](d:\src\go_llama_server_loader\internal\cli\filter.go)
- [internal/cli/model_item.go](d:\src\go_llama_server_loader\internal\cli\model_item.go)

**Новые:**
- [internal/cli/scrollbar.go](d:\src\go_llama_server_loader\internal\cli\scrollbar.go)
- [internal/cli/tabs.go](d:\src\go_llama_server_loader\internal\cli\tabs.go)
- [internal/cli/border_title.go](d:\src\go_llama_server_loader\internal\cli\border_title.go) (или функция внутри cli.go)
- [internal/cli/help_popup.go](d:\src\go_llama_server_loader\internal\cli\help_popup.go)

## Порядок реализации

1. Расширение палитры и стилей в [styles.go](d:\src\go_llama_server_loader\internal\cli\styles.go) (фундамент).
2. Изолированные компоненты: [scrollbar.go](d:\src\go_llama_server_loader\internal\cli\scrollbar.go), [tabs.go](d:\src\go_llama_server_loader\internal\cli\tabs.go), `injectBorderTitle` — каждый с тестами.
3. Рестайл существующих компонентов: [footer.go](d:\src\go_llama_server_loader\internal\cli\footer.go), [filter.go](d:\src\go_llama_server_loader\internal\cli\filter.go), [model_item.go](d:\src\go_llama_server_loader\internal\cli\model_item.go).
4. Интеграция в [cli.go](d:\src\go_llama_server_loader\internal\cli\cli.go) (`View()`, `Update()`, `recomputeListSize()`).
5. Прогон тестов и визуальная валидация.
