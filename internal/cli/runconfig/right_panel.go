package runconfig

import (
	"strings"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/uistyle"
)

// RightPanel — правая панель: каталог параметров с фильтром и скроллом.
type RightPanel struct {
	all      []CatalogEntry
	filtered []CatalogEntry
	cursor   int
	offset   int

	filterText   string
	filterCursor int
	filterActive bool

	st *uistyle.StyleConfig
	w  int
	h  int
}

// NewRightPanel создаёт правую панель с заданными записями и размерами.
func NewRightPanel(entries []CatalogEntry, st *uistyle.StyleConfig, w, h int) *RightPanel {
	p := &RightPanel{
		all:      entries,
		filtered: entries,
		st:       st,
		w:        w,
		h:        h,
	}
	return p
}

// SetSize обновляет размеры панели и корректирует смещение прокрутки.
func (p *RightPanel) SetSize(w, h int) {
	p.w = w
	p.h = h
	p.clampOffset()
}

// Selected возвращает текущую выбранную запись или nil.
func (p *RightPanel) Selected() *CatalogEntry {
	if len(p.filtered) == 0 || p.cursor < 0 || p.cursor >= len(p.filtered) {
		return nil
	}
	e := p.filtered[p.cursor]
	return &e
}

// IsFilterActive сообщает, открыт ли сейчас режим фильтра.
func (p *RightPanel) IsFilterActive() bool {
	return p.filterActive
}

// Update обрабатывает сообщения; действует только когда focused=true.
func (p *RightPanel) Update(msg tea.Msg, focused bool) tea.Cmd {
	if !focused {
		return nil
	}
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	if p.filterActive {
		return p.handleFilterKey(keyMsg.String())
	}
	switch keyMsg.String() {
	case "up":
		if p.cursor > 0 {
			p.cursor--
			p.clampOffset()
		}
	case "down":
		if p.cursor < len(p.filtered)-1 {
			p.cursor++
			p.clampOffset()
		}
	case "/":
		p.filterActive = true
	case "esc":
		// Если есть текст фильтра — очищаем (filter не активен, но текст мог остаться).
		if p.filterText != "" {
			p.filterText = ""
			p.filterCursor = 0
			p.applyFilter()
		}
	}
	return nil
}

func (p *RightPanel) handleFilterKey(key string) tea.Cmd {
	switch key {
	case "esc":
		p.filterActive = false
		p.filterText = ""
		p.filterCursor = 0
		p.applyFilter()
	case "enter":
		p.filterActive = false
	case "backspace", "ctrl+h":
		if p.filterCursor > 0 {
			_, size := utf8.DecodeLastRuneInString(p.filterText[:p.filterCursor])
			p.filterText = p.filterText[:p.filterCursor-size] + p.filterText[p.filterCursor:]
			p.filterCursor -= size
			p.applyFilter()
		}
	case "delete", "ctrl+d":
		if p.filterCursor < len(p.filterText) {
			_, size := utf8.DecodeRuneInString(p.filterText[p.filterCursor:])
			p.filterText = p.filterText[:p.filterCursor] + p.filterText[p.filterCursor+size:]
			p.applyFilter()
		}
	case "left":
		if p.filterCursor > 0 {
			_, size := utf8.DecodeLastRuneInString(p.filterText[:p.filterCursor])
			p.filterCursor -= size
		}
	case "right":
		if p.filterCursor < len(p.filterText) {
			_, size := utf8.DecodeRuneInString(p.filterText[p.filterCursor:])
			p.filterCursor += size
		}
	default:
		if utf8.RuneCountInString(key) == 1 {
			r, _ := utf8.DecodeRuneInString(key)
			if unicode.IsPrint(r) {
				p.filterText = p.filterText[:p.filterCursor] + key + p.filterText[p.filterCursor:]
				p.filterCursor += len(key)
				p.applyFilter()
			}
		}
	}
	return nil
}

func (p *RightPanel) applyFilter() {
	p.cursor = 0
	p.offset = 0
	if p.filterText == "" {
		p.filtered = p.all
		return
	}
	q := strings.ToLower(p.filterText)
	result := make([]CatalogEntry, 0, len(p.all))
	for _, e := range p.all {
		long := strings.ToLower(e.Meta.LongFlag)
		short := strings.ToLower(e.Meta.ShortFlag)
		desc := strings.ToLower(e.Meta.DescRU)
		if strings.Contains(long, q) || strings.Contains(short, q) || strings.Contains(desc, q) {
			result = append(result, e)
		}
	}
	p.filtered = result
}

// visibleListHeight — количество строк, доступных для списка элементов.
// Вычитаем: 2 (рамка), 1 (строка фильтра), 1 (разделитель).
func (p *RightPanel) visibleListHeight() int {
	v := p.h - 4
	if v < 1 {
		v = 1
	}
	return v
}

func (p *RightPanel) clampOffset() {
	vis := p.visibleListHeight()
	if p.cursor < p.offset {
		p.offset = p.cursor
	}
	if p.cursor >= p.offset+vis {
		p.offset = p.cursor - vis + 1
	}
	if p.offset < 0 {
		p.offset = 0
	}
}

// Render рендерит правую панель шириной p.w и высотой p.h.
func (p *RightPanel) Render(focused bool) string {
	st := p.st
	if p.w < 10 || p.h < 5 {
		return lipgloss.NewStyle().
			Width(p.w).
			Height(p.h).
			Background(lipgloss.Color(st.BgPanel)).
			Render("")
	}

	// contentW — ширина внутри рамки (рамка добавляет 2).
	contentW := p.w - 2
	listH := p.visibleListHeight()

	// scrollbar занимает правые ScrollbarWidth колонок внутри списка
	scrollW := uistyle.ScrollbarWidth
	itemW := contentW - scrollW
	if itemW < 4 {
		itemW = 4
		scrollW = contentW - 4
		if scrollW < 0 {
			scrollW = 0
			itemW = contentW
		}
	}

	// ── Строка фильтра ────────────────────────────────────────────────────────
	filterLine := p.renderFilterLine(contentW)

	// ── Разделитель ───────────────────────────────────────────────────────────
	sep := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.BorderIdle)).
		Render(strings.Repeat("─", contentW))

	// ── Строки списка + скроллбар ─────────────────────────────────────────────
	var listBlock string
	if len(p.all) == 0 {
		listBlock = lipgloss.NewStyle().
			Background(lipgloss.Color(st.BgPanel)).
			Foreground(lipgloss.Color(st.TextMuted)).
			Width(contentW).
			Height(listH).
			Render("каталог не найден")
	} else {
		scrollLines := renderScrollbarLines(p.offset, listH, len(p.filtered), listH, scrollW, st)
		rowLines := make([]string, listH)
		for i := 0; i < listH; i++ {
			idx := p.offset + i
			var itemLine string
			if idx < len(p.filtered) {
				itemLine = p.renderItem(p.filtered[idx], idx == p.cursor, itemW)
			} else {
				itemLine = lipgloss.NewStyle().
					Background(lipgloss.Color(st.BgPanel)).
					Width(itemW).
					Render("")
			}
			if scrollW > 0 {
				rowLines[i] = lipgloss.JoinHorizontal(lipgloss.Top, itemLine, scrollLines[i])
			} else {
				rowLines[i] = itemLine
			}
		}
		listBlock = strings.Join(rowLines, "\n")
	}

	// ── Сборка внутреннего содержимого ────────────────────────────────────────
	inner := strings.Join([]string{filterLine, sep, listBlock}, "\n")

	// ── Рамка с цветом по фокусу ──────────────────────────────────────────────
	borderColor := st.BorderIdle
	if focused {
		borderColor = st.NeonGreen
	}
	rendered := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(borderColor)).
		BorderBackground(lipgloss.Color(st.BgPanel)).
		Background(lipgloss.Color(st.BgPanel)).
		Width(contentW).
		Render(inner)

	// ── Title в рамку ─────────────────────────────────────────────────────────
	titleText := "Все параметры"
	if len(p.filtered) < len(p.all) && len(p.all) > 0 {
		titleText = "Поиск"
	}
	countText := ""
	if len(p.all) > 0 {
		if len(p.filtered) < len(p.all) {
			countText = strings.Repeat("", 0) // пусто справа, count в левом лейбле через пробел
		}
		_ = countText
	}
	return injectBorderTitle(rendered, titleText, "")
}

// renderFilterLine рендерит строку фильтра шириной w.
func (p *RightPanel) renderFilterLine(w int) string {
	st := p.st
	if p.filterActive {
		cursorStr := lipgloss.NewStyle().
			Foreground(lipgloss.Color(st.NeonGreen)).
			Background(lipgloss.Color(st.BgPanel)).
			Render("▌")
		var display string
		if p.filterCursor < len(p.filterText) {
			display = p.filterText[:p.filterCursor] + cursorStr + p.filterText[p.filterCursor:]
		} else {
			display = p.filterText + cursorStr
		}
		return lipgloss.NewStyle().
			Background(lipgloss.Color(st.BgPanel)).
			Foreground(lipgloss.Color(st.TextPrimary)).
			Width(w).
			Render("/ " + display)
	}
	// idle: отображаем подсказку
	keyPart := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.KeyHint)).
		Render("/")
	hintPart := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextMuted)).
		Render(" поиск...")
	hint := lipgloss.JoinHorizontal(lipgloss.Top, keyPart, hintPart)
	return lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Width(w).
		Render(hint)
}

// renderItem рендерит одну строку списка шириной w.
func (p *RightPanel) renderItem(e CatalogEntry, selected bool, w int) string {
	st := p.st
	const indicW = 2 // "▶ " или "  "

	longF := stripFlagArg(e.Meta.LongFlag)
	shortF := stripFlagArg(e.Meta.ShortFlag)

	flagPart := longF
	if shortF != "" && shortF != longF {
		flagPart += " " + shortF
	}
	const maxFlagW = 24
	if utf8.RuneCountInString(flagPart) > maxFlagW {
		flagPart = truncatePath(flagPart, maxFlagW)
	}
	flagPartW := utf8.RuneCountInString(flagPart)

	const addStr = "[ADD]"
	const addStrW = 6 // "[ADD] "
	addPartW := 0
	if selected {
		addPartW = addStrW
	}

	descMaxW := w - indicW - flagPartW - 1 - addPartW // 1 for space after flag
	if descMaxW < 0 {
		descMaxW = 0
	}

	var bg string
	if selected {
		bg = st.BgSelected
	} else {
		bg = st.BgPanel
	}

	// Индикатор
	var indicator string
	if selected {
		indicator = lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Foreground(lipgloss.Color(st.NeonGreen)).
			Render("▶ ")
	} else {
		indicator = lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Render("  ")
	}

	// Флаг
	var flagRendered string
	if selected {
		flagRendered = lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color(bg)).
			Foreground(lipgloss.Color(st.NeonGreen)).
			Render(flagPart)
	} else {
		flagRendered = lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Foreground(lipgloss.Color(st.NeonGreen)).
			Render(flagPart)
	}

	// [ADD] бейдж
	var addRendered string
	if selected {
		addRendered = " " + lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color(st.NeonGreen)).
			Foreground(lipgloss.Color(st.GreenDark)).
			Render(addStr)
	}

	// Описание
	var descRendered string
	if descMaxW > 0 && e.Meta.DescRU != "" {
		descText := truncatePath(e.Meta.DescRU, descMaxW)
		fg := st.TextMuted
		if selected {
			fg = st.TextSecondary
		}
		descRendered = lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Foreground(lipgloss.Color(fg)).
			Render(" " + descText)
	}

	line := lipgloss.JoinHorizontal(lipgloss.Top, indicator, flagRendered, addRendered, descRendered)

	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Width(w).
		Render(line)
}
