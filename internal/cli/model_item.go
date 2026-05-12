package cli

import (
	"fmt"
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/modelparams"
	"llama-server-loader/internal/cli/uistyle"
	"llama-server-loader/pkg/modelscan"
)

// ListItem представляет элемент списка моделей.
type ListItem struct {
	model   *modelscan.Model
	params  *modelparams.Lookup
	comment string
}

// NewListItem создает новый ListItem.
func NewListItem(m *modelscan.Model, params *modelparams.Lookup, comment string) *ListItem {
	return &ListItem{model: m, params: params, comment: comment}
}

// Title возвращает заголовок элемента.
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
func (l *ListItem) Render(w io.Writer, m list.Model, index int, item list.Item) {
	listItem := item.(*ListItem)
	model := listItem.model

	st := uistyle.GetStyles()
	if st == nil {
		st = &uistyle.StyleConfig{}
	}

	selected := m.Index() == index
	itemWidth := m.Width()

	quant := extractQuantization(model.Name)
	var dotFg, dotBg string
	if selected {
		dotFg = st.AccentPurple
		dotBg = st.BgSelected
	} else {
		dotFg = st.AccentPurpleMuted
		dotBg = st.DarkBg
	}
	dot := lipgloss.NewStyle().
		Background(lipgloss.Color(dotBg)).
		Foreground(lipgloss.Color(dotFg)).
		Render("●")

	if selected {
		fmt.Fprint(w, renderSelected(model, dot, quant, itemWidth, st, listItem.params, listItem.comment))
	} else {
		fmt.Fprint(w, renderNormal(model, dot, quant, itemWidth, st, listItem.params, listItem.comment))
	}
}

// renderSelected рендерит выбранный item.
func renderSelected(model *modelscan.Model, dot, quant string, itemWidth int, st *uistyle.StyleConfig, params *modelparams.Lookup, comment string) string {
	bracketStyle := st.ItemSelectedStyle()
	innerW := itemWidth - bracketStyle.GetHorizontalFrameSize()
	if innerW < 1 {
		innerW = 1
	}

	rowFill := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgSelected)).
		Width(innerW)

	nameRest := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(st.BgSelected)).
		Foreground(lipgloss.Color(st.TextPrimary)).
		Render(" " + model.Name)
	metaBracket := formatMetaBracket(model, st, st.BgSelected, params)
	nameSpacerW := innerW - lipgloss.Width(dot) - lipgloss.Width(nameRest) - lipgloss.Width(metaBracket)
	if nameSpacerW < 1 {
		nameSpacerW = 1
	}
	nameSpacer := lipgloss.NewStyle().Background(lipgloss.Color(st.BgSelected)).Render(strings.Repeat(" ", nameSpacerW))
	nameLine := rowFill.Render(lipgloss.JoinHorizontal(lipgloss.Top, dot, nameRest, nameSpacer, metaBracket))

	pathStr := truncatePathLeft(model.Path, itemWidth-4)
	pathLine := rowFill.
		Foreground(lipgloss.Color(st.TextMuted)).
		Render("  " + pathStr)

	badgeBlock := rowFill.
		PaddingLeft(2).
		Render(formatMetadataBadgesSelected(model, st))

	empty := rowFill.Render("")
	parts := []string{empty, nameLine, pathLine, badgeBlock}
	if comment != "" {
		commentLine := rowFill.
			Foreground(lipgloss.Color(st.TextMuted)).
			Italic(true).
			Render("  # " + comment)
		parts = append(parts, commentLine)
	}
	parts = append(parts, empty)
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	if itemWidth > 0 {
		bracketStyle = bracketStyle.Width(itemWidth)
	}
	return bracketStyle.Render(content)
}

// renderNormal рендерит обычный (не выбранный) item.
func renderNormal(model *modelscan.Model, dot, quant string, itemWidth int, st *uistyle.StyleConfig, params *modelparams.Lookup, comment string) string {
	if itemWidth < 1 {
		itemWidth = 1
	}

	rowFill := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Width(itemWidth)

	leftPad2 := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Render("  ")
	nameRest := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(st.DarkBg)).
		Foreground(lipgloss.Color(st.TextPrimary)).
		Render(" " + model.Name)
	metaBracket := formatMetaBracket(model, st, st.DarkBg, params)
	nameSpacerW := itemWidth - lipgloss.Width(leftPad2) - lipgloss.Width(dot) - lipgloss.Width(nameRest) - lipgloss.Width(metaBracket)
	if nameSpacerW < 1 {
		nameSpacerW = 1
	}
	nameSpacer := lipgloss.NewStyle().Background(lipgloss.Color(st.DarkBg)).Render(strings.Repeat(" ", nameSpacerW))
	nameLine := rowFill.Render(lipgloss.JoinHorizontal(lipgloss.Top, leftPad2, dot, nameRest, nameSpacer, metaBracket))

	pathStr := truncatePathLeft(model.Path, itemWidth-4)
	pathLine := rowFill.
		PaddingLeft(2).
		Foreground(lipgloss.Color(st.TextMuted)).
		Render("  " + pathStr)

	badgeBlock := rowFill.
		PaddingLeft(4).
		Render(formatMetadataBadges(model, st))

	parts := []string{nameLine, pathLine, badgeBlock}
	if comment != "" {
		commentLine := rowFill.
			PaddingLeft(4).
			Foreground(lipgloss.Color(st.TextMuted)).
			Italic(true).
			Render("# " + comment)
		parts = append(parts, commentLine)
	}
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	empty := rowFill.Render("")
	return lipgloss.JoinVertical(lipgloss.Left, empty, content, empty)
}

func quantColor(quant string, st *uistyle.StyleConfig) string {
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

// truncatePathLeft обрезает путь слева, оставляя видимым конец пути.
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

// formatMetaBracket возвращает «[ gemma4 · 131K ctx · 7.5B · 38 params ]» для
// вставки в строку имени. Поля, для которых нет значения в GGUF-параметрах,
// скрываются. При полном отсутствии params возвращает прежний бейдж GGUF/quant.
func formatMetaBracket(m *modelscan.Model, st *uistyle.StyleConfig, parentBg string, params *modelparams.Lookup) string {
	if st == nil {
		return ""
	}
	mutedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(parentBg)).
		Foreground(lipgloss.Color(st.TextMuted))
	sepStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(parentBg)).
		Foreground(lipgloss.Color(st.TextSecondary))

	var parts []string
	if params != nil {
		c := params.ForPathCurated(m.Path)
		if c.Architecture != "" {
			parts = append(parts, c.Architecture)
		}
		if c.ContextLength > 0 {
			parts = append(parts, modelparams.FormatContext(c.ContextLength)+" ctx")
		}
		if c.SizeLabel != "" {
			parts = append(parts, c.SizeLabel)
		}
		if c.TotalCount > 0 {
			parts = append(parts, fmt.Sprintf("%d params", c.TotalCount))
		}
	}
	if len(parts) == 0 {
		parts = []string{"GGUF"}
		if q := extractQuantization(m.Name); q != "" {
			parts = append(parts, q)
		}
		if len(m.MMProjPaths) > 0 {
			parts = append(parts, "mmproj")
		}
	}

	pieces := make([]string, 0, len(parts)*2+1)
	pieces = append(pieces, mutedStyle.Render(" [ "))
	for i, p := range parts {
		if i > 0 {
			pieces = append(pieces, sepStyle.Render(" · "))
		}
		pieces = append(pieces, mutedStyle.Render(p))
	}
	pieces = append(pieces, mutedStyle.Render(" ]"))
	return lipgloss.JoinHorizontal(lipgloss.Top, pieces...)
}

// formatMetaLine формирует одну строку «meta» под path в карточке модели.
func formatMetaLine(m *modelscan.Model, st *uistyle.StyleConfig, parentBg string) string {
	if st == nil {
		return ""
	}
	parts := []string{"GGUF"}
	if q := extractQuantization(m.Name); q != "" {
		parts = append(parts, q)
	}
	parts = append(parts, formatSize(m.Size))
	if len(m.MMProjPaths) > 0 {
		parts = append(parts, "mmproj")
	}
	textStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(parentBg)).
		Foreground(lipgloss.Color(st.TextMuted))
	sepStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(parentBg)).
		Foreground(lipgloss.Color(st.TextSecondary))
	pieces := make([]string, 0, len(parts)*2-1)
	for i, p := range parts {
		if i > 0 {
			pieces = append(pieces, sepStyle.Render(" · "))
		}
		pieces = append(pieces, textStyle.Render(p))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, pieces...)
}

// formatMetadataBadges формирует строку бейджей для обычного (unselected) item.
func formatMetadataBadges(m *modelscan.Model, st *uistyle.StyleConfig) string {
	return formatMetadataBadgesWithBg(m, st, "")
}

// formatMetadataBadgesSelected — те же бейджи с parentBg = BgSelected.
func formatMetadataBadgesSelected(m *modelscan.Model, st *uistyle.StyleConfig) string {
	if st == nil {
		return formatMetadataBadges(m, st)
	}
	return formatMetadataBadgesWithBg(m, st, st.BgSelected)
}

// formatMetadataBadgesWithBg — общая реализация с настраиваемым фоном separator'а.
func formatMetadataBadgesWithBg(m *modelscan.Model, st *uistyle.StyleConfig, parentBg string) string {
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
	sepCell := lipgloss.NewStyle().
		Background(lipgloss.Color(parentBg)).
		Render("  ")
	sep := sepCell + "\n" + sepCell + "\n" + sepCell
	parts := []string{
		st.SizeBadgeStyle(parentBg).Render(formatSize(m.Size)),
		sep,
	}
	if len(m.MMProjPaths) > 0 {
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
		"iq4_xs", "iq4_nl",
		"iq3_xxs", "iq3_xs", "iq3_s", "iq3_m",
		"iq2_xxs", "iq2_xs", "iq2_s", "iq2_m",
		"iq1_s", "iq1_m",
	}

	nameLower := strings.ToLower(name)
	for _, pattern := range quantPatterns {
		if strings.Contains(nameLower, pattern) {
			return strings.ToUpper(pattern)
		}
	}

	return ""
}

// ============================================================================
// StyledDelegate — кастомный delegate с 7-строчным форматом отображения
// ============================================================================

type StyledDelegate struct {
	base   list.DefaultDelegate
	styles *uistyle.StyleConfig
}

func (d *StyledDelegate) Height() int {
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

func (d *StyledDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
