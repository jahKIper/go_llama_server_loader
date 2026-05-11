package runconfig

import (
	"strings"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/modelparams"
	"llama-server-loader/internal/cli/uistyle"
	"llama-server-loader/internal/config"
)


// RightTab — переключаемые вкладки правой панели.
type RightTab int

const (
	RightTabCLI  RightTab = iota // CLI-параметры (каталог llama-server)
	RightTabGGUF                 // GGUF-параметры выбранной модели
)

// RightPanel — правая панель: каталог параметров с фильтром и скроллом.
type RightPanel struct {
	all      []CatalogEntry
	filtered []CatalogEntry
	cursor   int
	offset   int

	// GGUF-таб: параметры модели (read-only).
	ggufAll      []config.ModelParam
	ggufFiltered []config.ModelParam
	ggufCursor   int
	ggufOffset   int

	tab RightTab

	filterText   string
	filterCursor int
	filterActive bool

	st *uistyle.StyleConfig
	w  int
	h  int
}

// SetGGUFParams задаёт срез GGUF-параметров для второй вкладки.
func (p *RightPanel) SetGGUFParams(params []config.ModelParam) {
	p.ggufAll = params
	p.ggufFiltered = params
	p.ggufCursor = 0
	p.ggufOffset = 0
}

// Tab возвращает активную вкладку.
func (p *RightPanel) Tab() RightTab {
	return p.tab
}

// ToggleTab переключает активную вкладку. Сбрасывает текст фильтра.
func (p *RightPanel) ToggleTab() {
	if p.tab == RightTabCLI {
		p.tab = RightTabGGUF
	} else {
		p.tab = RightTabCLI
	}
	p.filterText = ""
	p.filterCursor = 0
	p.filterActive = false
	p.applyFilter()
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

// Selected возвращает текущую выбранную CLI-запись или nil.
// Для GGUF-таба возвращает nil (там нет CatalogEntry).
func (p *RightPanel) Selected() *CatalogEntry {
	if p.tab != RightTabCLI {
		return nil
	}
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
		if p.cursor < p.currentLen()-1 {
			p.cursor++
			p.clampOffset()
		}
	case "pgup":
		step := p.visibleListHeight() / 2
		if step < 1 {
			step = 1
		}
		p.cursor -= step
		if p.cursor < 0 {
			p.cursor = 0
		}
		p.clampOffset()
	case "pgdown":
		step := p.visibleListHeight() / 2
		if step < 1 {
			step = 1
		}
		p.cursor += step
		if p.cursor >= p.currentLen() {
			p.cursor = p.currentLen() - 1
		}
		p.clampOffset()
	case "home":
		p.cursor = 0
		p.clampOffset()
	case "end":
		p.cursor = p.currentLen() - 1
		if p.cursor < 0 {
			p.cursor = 0
		}
		p.clampOffset()
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
	if p.tab == RightTabGGUF {
		if p.filterText == "" {
			p.ggufFiltered = p.ggufAll
			return
		}
		q := strings.ToLower(p.filterText)
		res := make([]config.ModelParam, 0, len(p.ggufAll))
		for _, m := range p.ggufAll {
			k := strings.ToLower(m.Key)
			d := strings.ToLower(m.DescriptionRU)
			v := strings.ToLower(modelparams.FormatValue(m.Value, 64))
			if strings.Contains(k, q) || strings.Contains(d, q) || strings.Contains(v, q) {
				res = append(res, m)
			}
		}
		p.ggufFiltered = res
		return
	}
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

// currentLen возвращает длину видимого списка с учётом активного таба.
func (p *RightPanel) currentLen() int {
	if p.tab == RightTabGGUF {
		return len(p.ggufFiltered)
	}
	return len(p.filtered)
}

// visibleListHeight — количество строк, доступных для списка элементов.
// Вычитаем: 2 (внешняя рамка панели), 3 (бордюрное поле фильтра).
func (p *RightPanel) visibleListHeight() int {
	v := p.h - 5
	if v < 1 {
		v = 1
	}
	return v
}

// itemW вычисляет ширину одного item-блока (без скроллбара) для текущей панели.
func (p *RightPanel) itemW() int {
	contentW := p.w - 2
	scrollW := uistyle.ScrollbarWidth
	w := contentW - scrollW
	if w < 4 {
		w = 4
	}
	return w
}

// itemHeight возвращает суммарную высоту в строках для одного элемента списка:
// 1 строка для флага + перенесённое описание + 1 строка-разделитель.
func (p *RightPanel) itemHeight(e CatalogEntry, w int) int {
	const indicW = 2
	descMaxW := w - indicW
	if descMaxW < 1 {
		descMaxW = 1
	}
	descLines := 0
	if e.Meta != nil && e.Meta.DescRU != "" {
		wrapped := wordWrap(e.Meta.DescRU, descMaxW)
		descLines = strings.Count(wrapped, "\n") + 1
	}
	return 1 + descLines + 1 // флаг + описание + пустая строка-разделитель
}

// fitsCount возвращает количество подряд идущих элементов, начиная с idx,
// которые целиком помещаются в listH строк.
func (p *RightPanel) fitsCount(idx, listH, w int) int {
	used := 0
	count := 0
	for i := idx; i < len(p.filtered); i++ {
		h := p.itemHeight(p.filtered[i], w)
		if used+h > listH {
			break
		}
		used += h
		count++
	}
	return count
}

// ggufItemHeight возвращает высоту GGUF-элемента в строках:
// 1 (ключ+значение) + строки описания (только для выбранного элемента).
func (p *RightPanel) ggufItemHeight(idx int) int {
	if idx < 0 || idx >= len(p.ggufFiltered) {
		return 1
	}
	if idx != p.cursor {
		return 1
	}
	m := p.ggufFiltered[idx]
	if m.DescriptionRU == "" {
		return 1
	}
	const indicW = 2
	descMaxW := p.itemW() - indicW
	if descMaxW < 1 {
		descMaxW = 1
	}
	wrapped := wordWrap(m.DescriptionRU, descMaxW)
	return 1 + strings.Count(wrapped, "\n") + 1
}

// ggufFitsCount возвращает количество GGUF-элементов, начиная с idx,
// которые целиком помещаются в listH строк.
func (p *RightPanel) ggufFitsCount(idx, listH int) int {
	used := 0
	count := 0
	for i := idx; i < len(p.ggufFiltered); i++ {
		h := p.ggufItemHeight(i)
		if used+h > listH {
			break
		}
		used += h
		count++
	}
	return count
}

func (p *RightPanel) clampOffset() {
	listH := p.visibleListHeight()
	w := p.itemW()
	if p.cursor < p.offset {
		p.offset = p.cursor
	}
	if p.tab == RightTabGGUF {
		for p.offset < p.cursor {
			fits := p.ggufFitsCount(p.offset, listH)
			if p.offset+fits > p.cursor {
				break
			}
			p.offset++
		}
		if p.offset < 0 {
			p.offset = 0
		}
		return
	}
	// Прокручиваем offset вперёд, пока курсор не помещается в видимой области.
	for p.offset < p.cursor {
		fits := p.fitsCount(p.offset, listH, w)
		if p.offset+fits > p.cursor {
			break
		}
		p.offset++
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

	// ── Поле фильтра (бордюрное, как на первом экране) ───────────────────────
	filterBlock := p.renderFilterLine(contentW)

	// ── Строки списка + скроллбар ─────────────────────────────────────────────
	var listBlock string
	if p.tab == RightTabGGUF {
		listBlock = p.renderGGUFList(listH, itemW, scrollW)
	} else if len(p.all) == 0 {
		listBlock = lipgloss.NewStyle().
			Background(lipgloss.Color(st.BgPanel)).
			Foreground(lipgloss.Color(st.TextMuted)).
			Width(contentW).
			Height(listH).
			Render("каталог не найден")
	} else {
		// Собираем последовательно линии всех видимых элементов до заполнения listH.
		itemLines := make([]string, 0, listH)
		linesUsed := 0
		idx := p.offset
		for linesUsed < listH && idx < len(p.filtered) {
			h := p.itemHeight(p.filtered[idx], itemW)
			if linesUsed+h > listH {
				break
			}
			block := p.renderItem(p.filtered[idx], idx == p.cursor, itemW)
			for _, ln := range strings.Split(block, "\n") {
				itemLines = append(itemLines, ln)
			}
			linesUsed += h
			idx++
		}
		// Добиваем пустыми строками
		emptyLine := lipgloss.NewStyle().
			Background(lipgloss.Color(st.BgPanel)).
			Width(itemW).
			Render("")
		for linesUsed < listH {
			itemLines = append(itemLines, emptyLine)
			linesUsed++
		}

		visibleItems := idx - p.offset
		scrollLines := renderScrollbarLines(p.offset, visibleItems, len(p.filtered), listH, scrollW, st)
		rowLines := make([]string, listH)
		for i := 0; i < listH; i++ {
			if scrollW > 0 {
				rowLines[i] = lipgloss.JoinHorizontal(lipgloss.Top, itemLines[i], scrollLines[i])
			} else {
				rowLines[i] = itemLines[i]
			}
		}
		listBlock = strings.Join(rowLines, "\n")
	}

	// ── Сборка внутреннего содержимого ────────────────────────────────────────
	inner := strings.Join([]string{filterBlock, listBlock}, "\n")

	// ── Рамка с цветом по фокусу ──────────────────────────────────────────────
	borderColor := st.BorderIdle
	if focused {
		borderColor = st.NeonGreen
	}
	// Высота content-блока внутри рамки = p.h - 2 (две строки на рамку).
	innerH := p.h - 2
	if innerH < 1 {
		innerH = 1
	}
	rendered := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(borderColor)).
		BorderBackground(lipgloss.Color(st.BgPanel)).
		Background(lipgloss.Color(st.BgPanel)).
		Width(contentW).
		Height(innerH).
		Render(inner)

	// ── Title в рамку ─────────────────────────────────────────────────────────
	var titleText string
	switch p.tab {
	case RightTabGGUF:
		titleText = "▣ Параметры модели  │  CLI"
	default:
		titleText = "▣ CLI  │  Параметры модели"
	}
	return injectBorderTitle(rendered, titleText, "")
}

// renderGGUFList рендерит список GGUF-параметров: «ключ  значение».
// Описание показывается под выбранной строкой (multi-line wordWrap).
func (p *RightPanel) renderGGUFList(listH, itemW, scrollW int) string {
	st := p.st
	if len(p.ggufAll) == 0 {
		empty := lipgloss.NewStyle().
			Background(lipgloss.Color(st.BgPanel)).
			Foreground(lipgloss.Color(st.TextMuted)).
			Width(itemW + scrollW).
			Height(listH).
			Render("GGUF-параметры не найдены. Откройте модель — они запишутся в models.json.")
		return empty
	}

	lines := make([]string, 0, listH)
	idx := p.offset
	visibleItems := 0
	for len(lines) < listH && idx < len(p.ggufFiltered) {
		itemLines := p.renderGGUFItem(p.ggufFiltered[idx], idx == p.cursor, itemW)
		for _, l := range itemLines {
			if len(lines) < listH {
				lines = append(lines, l)
			}
		}
		visibleItems++
		idx++
	}
	emptyLine := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Width(itemW).
		Render("")
	for len(lines) < listH {
		lines = append(lines, emptyLine)
	}

	scrollLines := renderScrollbarLines(p.offset, visibleItems, len(p.ggufFiltered), listH, scrollW, st)
	rowLines := make([]string, listH)
	for i := 0; i < listH; i++ {
		if scrollW > 0 {
			rowLines[i] = lipgloss.JoinHorizontal(lipgloss.Top, lines[i], scrollLines[i])
		} else {
			rowLines[i] = lines[i]
		}
	}
	return strings.Join(rowLines, "\n")
}

// renderGGUFItem — строки «ключ  значение» (+ описание под выделенной).
// Возвращает срез строк: 1 строка для обычного элемента, 2+ для выделенного с описанием.
func (p *RightPanel) renderGGUFItem(m config.ModelParam, selected bool, w int) []string {
	st := p.st
	bg := st.BgPanel
	if selected {
		bg = st.BgSelected
	}
	const indicW = 2
	keyMaxW := w / 2
	if keyMaxW > 36 {
		keyMaxW = 36
	}
	if keyMaxW < 8 {
		keyMaxW = 8
	}
	key := m.Key
	if utf8.RuneCountInString(key) > keyMaxW {
		key = truncatePath(key, keyMaxW)
	}
	// Бюджет ширины для значения: вся ширина w минус индикатор, ключ и 2 пробела между ними.
	// Жёстко обрезаем, чтобы lipgloss не переносил строку на второй ряд (что ломает учёт высоты).
	valBudget := w - indicW - keyMaxW - 2
	if valBudget < 1 {
		valBudget = 1
	}
	valStr := modelparams.FormatValue(m.Value, valBudget)
	if utf8.RuneCountInString(valStr) > valBudget {
		valStr = truncatePath(valStr, valBudget)
	}
	valStyleFg := st.TextSecondary
	if !selected {
		valStyleFg = st.TextMuted
	}
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
	keyStyled := lipgloss.NewStyle().
		Bold(selected).
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(st.NeonGreen)).
		Width(keyMaxW).
		Render(key)
	valStyled := lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(valStyleFg)).
		Render("  " + valStr)
	mainLine := lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Width(w).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, indicator, keyStyled, valStyled))

	result := []string{mainLine}

	if selected && m.DescriptionRU != "" {
		descMaxW := w - indicW
		if descMaxW < 1 {
			descMaxW = 1
		}
		indent := lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Render(strings.Repeat(" ", indicW))
		descStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Foreground(lipgloss.Color(st.TextMuted))
		rowFill := lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Width(w)
		wrapped := wordWrap(m.DescriptionRU, descMaxW)
		for _, dl := range strings.Split(wrapped, "\n") {
			dl := lipgloss.JoinHorizontal(lipgloss.Top, indent, descStyle.Render(dl))
			result = append(result, rowFill.Render(dl))
		}
	}

	return result
}

// renderFilterLine рендерит бордюрное поле фильтра шириной w —
// визуально совпадает с фильтром первого экрана (uistyle.FilterInput*Style).
func (p *RightPanel) renderFilterLine(w int) string {
	st := p.st

	// Внутренняя ширина поля — w минус рамка (2) и горизонтальный padding (2).
	innerW := w - 4
	if innerW < 1 {
		innerW = 1
	}

	if p.filterActive {
		cursor := lipgloss.NewStyle().
			Foreground(lipgloss.Color(st.NeonGreen)).
			Background(lipgloss.Color(st.BgPanel)).
			Render("▌")
		var display string
		if p.filterCursor < len(p.filterText) {
			display = p.filterText[:p.filterCursor] + cursor + p.filterText[p.filterCursor:]
		} else {
			display = p.filterText + cursor
		}
		return st.FilterInputActiveStyle().Width(innerW).Render("/  " + display)
	}

	// idle
	placeholder := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextMuted)).
		Render("поиск...")
	return st.FilterInputIdleStyle().Width(innerW).Render("/  " + placeholder)
}

// renderItem рендерит элемент списка шириной w. Высота элемента — переменная:
// строка с флагом + перенесённое описание + 1 пустая строка-разделитель.
func (p *RightPanel) renderItem(e CatalogEntry, selected bool, w int) string {
	st := p.st
	const indicW = 2 // "▶ " или "  "

	longF := stripFlagArg(e.Meta.LongFlag)
	shortF := stripFlagArg(e.Meta.ShortFlag)

	flagPart := longF
	if shortF != "" && shortF != longF {
		flagPart += "  " + shortF
	}
	flagMaxW := w - indicW
	if flagMaxW < 1 {
		flagMaxW = 1
	}
	if utf8.RuneCountInString(flagPart) > flagMaxW {
		flagPart = truncatePath(flagPart, flagMaxW)
	}

	bg := st.BgPanel
	if selected {
		bg = st.BgSelected
	}

	rowFill := lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Width(w)

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
	flagStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(st.NeonGreen)).
		Bold(selected)
	flagLine := rowFill.Render(lipgloss.JoinHorizontal(lipgloss.Top, indicator, flagStyle.Render(flagPart)))

	lines := []string{flagLine}

	// Описание — перенос по словам с отступом indicW
	if e.Meta != nil && e.Meta.DescRU != "" {
		descMaxW := w - indicW
		if descMaxW < 1 {
			descMaxW = 1
		}
		fg := st.TextMuted
		if selected {
			fg = st.TextSecondary
		}
		descStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Foreground(lipgloss.Color(fg))
		indentStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(bg))
		indent := indentStyle.Render(strings.Repeat(" ", indicW))

		wrapped := wordWrap(e.Meta.DescRU, descMaxW)
		for _, dl := range strings.Split(wrapped, "\n") {
			line := lipgloss.JoinHorizontal(lipgloss.Top, indent, descStyle.Render(dl))
			lines = append(lines, rowFill.Render(line))
		}
	}

	// Разделитель — пустая строка фоном панели (чтобы выделение не «слипалось» с соседом)
	sepBg := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Width(w).
		Render("")
	lines = append(lines, sepBg)

	return strings.Join(lines, "\n")
}
