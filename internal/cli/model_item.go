package cli

import (
	"fmt"
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"

	"llama-server-loader/pkg/modelscan"
)

// ListItem представляет элемент списка моделей.
type ListItem struct {
	model *modelscan.Model
}

// NewListItem создает новый ListItem.
func NewListItem(m *modelscan.Model) *ListItem {
	return &ListItem{model: m}
}

// Title возвращает заголовок элемента (для совместимости с list.Interface).
func (l *ListItem) Title() string {
	return l.model.Name
}

// Description возвращает описание элемента.
func (l *ListItem) Description() string {
	if len(l.model.MMProjPaths) > 0 {
		return l.model.Path + "\nMMProj: " + l.model.MMProjPaths[0]
	}
	return l.model.Path
}

// FilterValue возвращает значение для фильтрации.
func (l *ListItem) FilterValue() string {
	return l.model.Name
}

// Model возвращает структуру модели.
func (l *ListItem) Model() *modelscan.Model {
	return l.model
}

// Render рендерит элемент списка.
// Структура (Height=7):
//   - Selected:   F-bracket с border top+bottom (содержит name+path+badges, 5 строк) = 7
//   - Unselected: пустая строка(1) + name+path+badges(5) + пустая строка(1) = 7
func (l *ListItem) Render(w io.Writer, m list.Model, index int, item list.Item) {
	listItem := item.(*ListItem)
	model := listItem.model

	st := GetStyles()
	if st == nil {
		st = &StyleConfig{}
	}

	selected := m.Index() == index
	itemWidth := m.Width()

	quant := extractQuantization(model.Name)
	// Точка-индикатор: фон делаем зависящим от состояния (selected/normal),
	// чтобы при склейке в строку имени не образовывалась «дырка» терминального фона.
	dotStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(quantColor(quant, st)))
	if selected {
		dotStyle = dotStyle.Background(lipgloss.Color(st.BgSelected))
	} else {
		dotStyle = dotStyle.Background(lipgloss.Color(st.DarkBg))
	}
	dot := dotStyle.Render("●")

	if selected {
		fmt.Fprint(w, renderSelected(model, dot, quant, itemWidth, st))
	} else {
		fmt.Fprint(w, renderNormal(model, dot, quant, itemWidth, st))
	}
}

// renderSelected рендерит выбранный item: halo-top + F-bracket(name+path+badges) + halo-bottom.
func renderSelected(model *modelscan.Model, dot, quant string, itemWidth int, st *StyleConfig) string {
	bracketStyle := st.ItemSelectedStyle()
	innerW := itemWidth - bracketStyle.GetHorizontalFrameSize()
	if innerW < 1 {
		innerW = 1
	}

	// rowFill — стиль с фоном BgSelected и Width=innerW. Им оборачиваем
	// уже отстайленный inline-контент, чтобы добить пустоту справа фоном.
	// ВАЖНО: lipgloss теряет внешний bg после inner ANSI-reset, если внутри
	// строки есть plain-текст после вложенного отстайленного куска. Поэтому
	// все вложения собираем через JoinHorizontal предстайленных кусков
	// (у каждого свой bg), а не конкатенацией с plain-string.
	rowFill := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgSelected)).
		Width(innerW)

	nameRest := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(st.BgSelected)).
		Foreground(lipgloss.Color(st.TextPrimary)).
		Render(" " + model.Name)
	nameLine := rowFill.Render(lipgloss.JoinHorizontal(lipgloss.Top, dot, nameRest))

	pathStr := truncatePathLeft(model.Path, itemWidth-4)
	pathLine := rowFill.
		Foreground(lipgloss.Color(st.TextSecondary)).
		Render("  " + pathStr)

	badgeBlock := rowFill.
		PaddingLeft(2).
		Render(formatMetadataBadgesSelected(model, st))

	content := lipgloss.JoinVertical(lipgloss.Left, nameLine, pathLine, badgeBlock)

	// F-bracket: top + left + bottom (no right), NeonGreen, BgSelected.
	// lipgloss v2 .Width() — total width: задаём itemWidth, чтобы занять всю ширину item'а.
	if itemWidth > 0 {
		bracketStyle = bracketStyle.Width(itemWidth)
	}
	return bracketStyle.Render(content)
}

// renderNormal рендерит обычный (не выбранный) item: пустая строка + контент + пустая строка.
func renderNormal(model *modelscan.Model, dot, quant string, itemWidth int, st *StyleConfig) string {
	if itemWidth < 1 {
		itemWidth = 1
	}

	// rowFill — фон DarkBg + Width=itemWidth для добивки пустоты справа.
	// Inline-контент с предстайленными вставками собираем через JoinHorizontal,
	// чтобы lipgloss не терял bg после внутренних ANSI-reset.
	rowFill := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Width(itemWidth)

	// nameLine: левый padding (2) + dot + " " + имя — каждый кусок отстайлен.
	leftPad2 := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Render("  ")
	nameRest := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Foreground(lipgloss.Color(st.TextPrimary)).
		Render(" " + model.Name)
	nameLine := rowFill.Render(lipgloss.JoinHorizontal(lipgloss.Top, leftPad2, dot, nameRest))

	pathStr := truncatePathLeft(model.Path, itemWidth-4)
	pathLine := rowFill.
		PaddingLeft(2).
		Foreground(lipgloss.Color(st.TextSecondary)).
		Render("  " + pathStr)

	badgeBlock := rowFill.
		PaddingLeft(4).
		Render(formatMetadataBadges(model, st))

	content := lipgloss.JoinVertical(lipgloss.Left, nameLine, pathLine, badgeBlock)

	// Пустые строки сверху и снизу для стабильной высоты = 7 строк.
	empty := rowFill.Render("")
	return lipgloss.JoinVertical(lipgloss.Left, empty, content, empty)
}

// quantColor возвращает hex-строку цвета точки-индикатора по квантованию.
func quantColor(quant string, st *StyleConfig) string {
	switch {
	case strings.HasPrefix(quant, "Q5"),
		strings.HasPrefix(quant, "Q6"),
		strings.HasPrefix(quant, "Q8"),
		quant == "F16", quant == "F32", quant == "F64":
		return st.NeonGreen
	case strings.HasPrefix(quant, "Q4"):
		return st.GreenPrimary
	default:
		return st.TextSecondary
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func computePathWidth(m list.Model) int {
	totalW := m.Width()
	const reserved = 7
	if totalW > reserved {
		return totalW - reserved
	}
	return 30
}

// truncatePathLeft обрезает путь слева, оставляя видимым конец пути. UTF-8 safe.
func truncatePathLeft(path string, maxLen int) string {
	if maxLen <= 0 {
		return "..."
	}
	runes := []rune(path)
	if len(runes) <= maxLen {
		return path
	}
	suffix := string(runes[len(runes)-maxLen+3:])
	return "..." + suffix
}

// formatMetadataBadges формирует строку бейджей для обычного (unselected) item.
func formatMetadataBadges(m *modelscan.Model, st *StyleConfig) string {
	return formatMetadataBadgesWithBg(m, st, "")
}

// formatMetadataBadgesSelected — то же, но separator красится в BgSelected.
func formatMetadataBadgesSelected(m *modelscan.Model, st *StyleConfig) string {
	if st == nil {
		return formatMetadataBadges(m, st)
	}
	return formatMetadataBadgesWithBg(m, st, st.BgSelected)
}

// formatMetadataBadgesWithBg — общая реализация с настраиваемым фоном separator'а.
// parentBg — цвет фона строки-родителя (DarkBg / BgSelected) для прокраски пробелов
// между бейджами; пустая строка → DarkBg по умолчанию.
func formatMetadataBadgesWithBg(m *modelscan.Model, st *StyleConfig, parentBg string) string {
	if st == nil {
		parts := "[" + formatSize(m.Size) + "]"
		if len(m.MMProjPaths) > 0 {
			parts += " [mmproj]"
		}
		parts += " [" + extractQuantization(m.Name) + "]"
		return parts
	}
	if parentBg == "" {
		parentBg = st.DarkBg
	}
	// sep — 3 строки высотой (как у бейджа с border'ом), чтобы JoinHorizontal
	// не достраивал короткий sep голыми пробелами без bg.
	sepCell := lipgloss.NewStyle().
		Background(lipgloss.Color(parentBg)).
		Render("  ")
	sep := sepCell + "\n" + sepCell + "\n" + sepCell
	parts := []string{
		st.SizeBadgeStyle(parentBg).Render(formatSize(m.Size)),
		sep,
	}
	if len(m.MMProjPaths) > 0 {
		// Сначала надпись с зелёной заливкой, затем оборачиваем в рамку с фоном строки.
		mmLabel := st.MMProjLabelStyle().Render("mmproj")
		parts = append(parts, st.MMProjBadgeStyle(parentBg).Render(mmLabel), sep)
	}
	parts = append(parts, st.QuantizationBadgeStyle(parentBg).Render(extractQuantization(m.Name)))
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// extractQuantization извлекает квантование из имени файла.
func extractQuantization(name string) string {
	quantPatterns := []string{
		"q5_k_m", "q5_k_s", "q5_0", "q5_1",
		"q4_k_m", "q4_k_s", "q4_0", "q4_1",
		"q3_k_m", "q3_k_s", "q3_k_l", "q2_k",
		"q2_0", "q2_1",
		"q8_0",
		"f16", "f32", "f64",
		"q6_k",
	}

	nameLower := strings.ToLower(name)
	for _, pattern := range quantPatterns {
		if strings.Contains(nameLower, pattern) {
			return strings.ToUpper(pattern)
		}
	}

	return "Q4_K_M"
}

// ============================================================================
// StyledDelegate — кастомный delegate с 7-строчным форматом отображения
// ============================================================================

type StyledDelegate struct {
	base   list.DefaultDelegate
	styles *StyleConfig
}

func (d *StyledDelegate) Height() int {
	// border-top/empty(1) + name(1) + path(1) + badges(3) + border-bottom/empty(1) = 7
	return 7
}

func (d *StyledDelegate) Spacing() int {
	return 1
}

func (d *StyledDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	listItem, ok := item.(*ListItem)
	if !ok {
		d.base.Render(w, m, index, item)
		return
	}
	listItem.Render(w, m, index, item)
}

func (d *StyledDelegate) TitleStyle() lipgloss.Style {
	return d.styles.ItemNormalStyle().Foreground(lipgloss.Color(d.styles.TextPrimary))
}

func (d *StyledDelegate) DescStyle() lipgloss.Style {
	return d.styles.ItemNormalStyle().Foreground(lipgloss.Color(d.styles.TextSecondary))
}

func (d *StyledDelegate) TitleStyleSelected() lipgloss.Style {
	return d.styles.ItemSelectedStyle().Foreground(lipgloss.Color(d.styles.TextPrimary))
}

func (d *StyledDelegate) DescStyleSelected() lipgloss.Style {
	return d.styles.ItemNormalStyle().Foreground(lipgloss.Color(d.styles.TextSecondary))
}

func (d *StyledDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
