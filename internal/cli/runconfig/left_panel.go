package runconfig

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/modelparams"
	"llama-server-loader/internal/cli/uistyle"
	"llama-server-loader/internal/config"
)

// LeftPanel — левая панель: список выбранных параметров для запуска.
type LeftPanel struct {
	rows    []ParamRow
	cursor  int
	offset  int
	w, h    int
	st      *uistyle.StyleConfig
	editing bool
	input   textinput.Model

	// GGUF-параметры для inline-подсказок «из модели: …» и apply-from-model (m).
	params    *modelparams.Lookup
	modelPath string
}

// SetParams задаёт источник GGUF-параметров и путь модели для inline-подсказок.
func (p *LeftPanel) SetParams(lookup *modelparams.Lookup, modelPath string) {
	p.params = lookup
	p.modelPath = modelPath
}

// ApplyModelValue подставляет GGUF-значение в значение текущей строки, если
// для её --long-флага есть маппинг в FlagApplyMap. Возвращает true при успехе.
func (p *LeftPanel) ApplyModelValue() bool {
	if p.params == nil || p.modelPath == "" {
		return false
	}
	if p.cursor < 0 || p.cursor >= len(p.rows) {
		return false
	}
	flag := "--" + strings.TrimPrefix(p.rows[p.cursor].Long, "--")
	_, disp, ok := p.params.ResolveGGUFKeyForFlag(p.modelPath, flag)
	if !ok {
		return false
	}
	p.rows[p.cursor].Value = disp
	return true
}

// NewLeftPanel создаёт пустую левую панель.
func NewLeftPanel(st *uistyle.StyleConfig, w, h int) *LeftPanel {
	ti := textinput.New()
	ti.Placeholder = "значение..."
	return &LeftPanel{
		rows:  []ParamRow{},
		st:    st,
		w:     w,
		h:     h,
		input: ti,
	}
}

// IsEditing сообщает, активен ли режим редактирования значения.
func (p *LeftPanel) IsEditing() bool {
	return p.editing
}

// StartEdit входит в режим редактирования для текущей строки.
func (p *LeftPanel) StartEdit() tea.Cmd {
	if len(p.rows) == 0 {
		return nil
	}
	p.editing = true
	p.input.SetValue(p.rows[p.cursor].Value)
	p.input.CursorEnd()
	return p.input.Focus()
}

// ConfirmEdit сохраняет значение из input и выходит из режима редактирования.
func (p *LeftPanel) ConfirmEdit() {
	if !p.editing || len(p.rows) == 0 {
		return
	}
	p.rows[p.cursor].Value = p.input.Value()
	p.editing = false
	p.input.Blur()
}

// CancelEdit отменяет редактирование без сохранения.
func (p *LeftPanel) CancelEdit() {
	p.editing = false
	p.input.Blur()
}

// UpdateInput форвардит сообщения в textinput когда идёт редактирование.
func (p *LeftPanel) UpdateInput(msg tea.Msg) tea.Cmd {
	if !p.editing {
		return nil
	}
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	return cmd
}

// SetSize обновляет размеры и корректирует прокрутку.
func (p *LeftPanel) SetSize(w, h int) {
	p.w = w
	p.h = h
	p.clampOffset()
}

// Rows возвращает текущий список параметров (копия не нужна — caller не мутирует).
func (p *LeftPanel) Rows() []ParamRow {
	return p.rows
}

// Add добавляет параметр. Если параметр с таким Long уже есть — пропускается.
func (p *LeftPanel) Add(meta *config.ParamMeta) {
	if meta == nil {
		return
	}
	long := stripFlagArg(meta.LongFlag)
	if long == "" {
		long = stripFlagArg(meta.ShortFlag)
	}
	if long == "" {
		return
	}
	for _, r := range p.rows {
		if r.Long == long {
			return
		}
	}
	p.rows = append(p.rows, ParamRow{
		Long:  long,
		Short: stripFlagArg(meta.ShortFlag),
		Key:   ParamKey(meta),
		Value: "",
		Meta:  meta,
	})
	p.cursor = len(p.rows) - 1
	p.clampOffset()
}

// Seed добавляет готовые строки в панель (используется для предзаполнения
// параметров на основе выбранной модели). Дубликаты по Long пропускаются.
func (p *LeftPanel) Seed(rows []ParamRow) {
	for _, r := range rows {
		if r.Long == "" {
			continue
		}
		dup := false
		for _, ex := range p.rows {
			if ex.Long == r.Long {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		p.rows = append(p.rows, r)
	}
	if p.cursor >= len(p.rows) {
		p.cursor = 0
	}
	p.clampOffset()
}

// Remove удаляет строку по индексу. Безопасен при выходе за границы.
func (p *LeftPanel) Remove(i int) {
	if i < 0 || i >= len(p.rows) {
		return
	}
	p.rows = append(p.rows[:i], p.rows[i+1:]...)
	if len(p.rows) == 0 {
		p.cursor = 0
	} else if p.cursor >= len(p.rows) {
		p.cursor = len(p.rows) - 1
	}
	p.clampOffset()
}

// MoveUp перемещает курсор вверх.
func (p *LeftPanel) MoveUp() {
	if p.cursor > 0 {
		p.cursor--
		p.clampOffset()
	}
}

// MoveDown перемещает курсор вниз.
func (p *LeftPanel) MoveDown() {
	if p.cursor < len(p.rows)-1 {
		p.cursor++
		p.clampOffset()
	}
}

// CursorIndex возвращает текущую позицию курсора.
func (p *LeftPanel) CursorIndex() int {
	return p.cursor
}

func (p *LeftPanel) visibleListHeight() int {
	v := p.h - 2 // две строки на рамку
	if v < 1 {
		v = 1
	}
	return v
}

func (p *LeftPanel) clampOffset() {
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

// Render рендерит левую панель шириной p.w и высотой p.h.
func (p *LeftPanel) Render(focused bool) string {
	st := p.st
	if p.w < 10 || p.h < 3 {
		return lipgloss.NewStyle().
			Width(p.w).
			Height(p.h).
			Background(lipgloss.Color(st.BgPanel)).
			Render("")
	}

	contentW := p.w - 2 // вычитаем рамку
	listH := p.visibleListHeight()

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

	var inner string
	if len(p.rows) == 0 {
		inner = p.renderEmptyState(contentW, listH)
	} else {
		inner = p.renderRows(listH, itemW, scrollW)
	}

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

	return injectBorderTitle(rendered, "Запуск", "")
}

func (p *LeftPanel) renderEmptyState(contentW, listH int) string {
	st := p.st
	msg := "Параметры не выбраны.\nПерейдите в правую панель\n(Tab) и добавьте флаг."
	return lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextMuted)).
		Width(contentW).
		Height(listH).
		Render(msg)
}

func (p *LeftPanel) renderRows(listH, itemW, scrollW int) string {
	st := p.st
	scrollLines := renderScrollbarLines(p.offset, listH, len(p.rows), listH, scrollW, st)
	rowLines := make([]string, listH)
	for i := 0; i < listH; i++ {
		idx := p.offset + i
		var line string
		if idx < len(p.rows) {
			editingThis := p.editing && idx == p.cursor
			line = p.renderItem(p.rows[idx], idx == p.cursor, itemW, editingThis)
		} else {
			line = lipgloss.NewStyle().
				Background(lipgloss.Color(st.BgPanel)).
				Width(itemW).
				Render("")
		}
		if scrollW > 0 {
			rowLines[i] = lipgloss.JoinHorizontal(lipgloss.Top, line, scrollLines[i])
		} else {
			rowLines[i] = line
		}
	}
	return strings.Join(rowLines, "\n")
}

// renderItem рендерит одну строку левой панели.
// Формат: [▶ ] --flag  value/input  [?]  [REMOVE]
func (p *LeftPanel) renderItem(row ParamRow, selected bool, w int, editing bool) string {
	st := p.st

	const indicW = 2 // "▶ " / "  "

	flagName := row.Long
	if flagName == "" {
		flagName = row.Short
	}
	const maxFlagW = 20
	if utf8.RuneCountInString(flagName) > maxFlagW {
		flagName = truncatePath(flagName, maxFlagW)
	}
	flagW := utf8.RuneCountInString(flagName)

	const hintStr = "[?]"
	const hintW = 4 // "[?] "

	// inline-подсказка «из модели: …» (если есть GGUF-значение для этого флага)
	var modelHintText string
	if p.params != nil && p.modelPath != "" {
		if _, disp, ok := p.params.ResolveGGUFKeyForFlag(p.modelPath, "--"+strings.TrimPrefix(row.Long, "--")); ok {
			modelHintText = "из модели: " + disp
		}
	}
	modelHintW := 0
	if modelHintText != "" {
		modelHintW = utf8.RuneCountInString(modelHintText) + 2 // "  prefix"
	}

	const removeStr = "[REMOVE (d)]"
	const removeW = 13
	removePad := 0
	if selected && !editing {
		removePad = removeW + 1 // пробел перед кнопкой
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
	flagRendered := lipgloss.NewStyle().
		Bold(selected).
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(st.NeonGreen)).
		Render(flagName)

	// [?] hint
	hintRendered := lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(st.TextMuted)).
		Render(" " + hintStr)

	// Value или textinput
	valueMaxW := w - indicW - flagW - 1 - hintW - removePad - modelHintW
	if valueMaxW < 0 {
		valueMaxW = 0
	}
	var valueRendered string
	if editing {
		// Показываем textinput; ширина = valueMaxW (без [REMOVE])
		p.input.SetWidth(valueMaxW - 1)
		valueRendered = " " + lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Foreground(lipgloss.Color(st.TextPrimary)).
			Render(p.input.View())
	} else if valueMaxW > 0 {
		valueText := row.Value
		if valueText == "" {
			valueText = "…"
		}
		if utf8.RuneCountInString(valueText) > valueMaxW {
			valueText = truncatePath(valueText, valueMaxW)
		}
		valueFg := st.TextMuted
		if row.Value != "" {
			valueFg = st.TextSecondary
		}
		valueRendered = lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Foreground(lipgloss.Color(valueFg)).
			Render(" " + valueText)
	}

	// [REMOVE] — только для выделенной строки вне режима редактирования
	var removeRendered string
	if selected && !editing {
		removeRendered = " " + lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("#7f1d1d")).
			Foreground(lipgloss.Color("#fca5a5")).
			Render(removeStr)
	}

	// inline-подсказка «из модели: …» — серый текст справа от значения
	var modelHintRendered string
	if modelHintText != "" {
		modelHintRendered = "  " + lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Foreground(lipgloss.Color(st.AccentPurpleMuted)).
			Render(modelHintText)
	}

	line := lipgloss.JoinHorizontal(lipgloss.Top,
		indicator, flagRendered, valueRendered, modelHintRendered, hintRendered, removeRendered)

	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Width(w).
		Render(line)
}
