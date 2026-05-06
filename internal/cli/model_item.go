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

// ListItem представляет элемент списка моделей с 3-строчным форматом отображения.
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

// Description возвращает описание элемента (используется list.Item interface;
// фактический рендеринг идёт через ListItem.Render).
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

// Render рендерит элемент списка с 3-строчным форматом.
func (l *ListItem) Render(w io.Writer, m list.Model, index int, item list.Item) {
	listItem := item.(*ListItem)
	model := listItem.model

	st := GetStyles()
	if st == nil {
		st = &StyleConfig{}
	}

	selected := m.Index() == index

	// Строка 1: имя модели
	var nameLine string
	if selected {
		nameLine = st.ItemSelectedStyle().Render(model.Name)
	} else {
		nameLine = st.ItemNormalStyle().Render(model.Name)
	}

	// Строка 2: путь с truncation слева (TextSecondary, отступ 4 символа)
	pathStr := truncatePathLeft(model.Path, computePathWidth(m))
	pathLine := st.ItemNormalStyle().
		Foreground(lipgloss.Color(st.TextSecondary)).
		Render("    " + pathStr)

	// Блок бейджей: bordered → 3 строки. Отступ 4 символа применяется
	// через padding-обёртку, чтобы JoinHorizontal не сломал многострочный layout.
	badgeBlock := lipgloss.NewStyle().PaddingLeft(4).Render(formatMetadataBadges(model, st))

	card := lipgloss.JoinVertical(lipgloss.Left, nameLine, pathLine, badgeBlock)
	fmt.Fprint(w, card)
}

func computePathWidth(m list.Model) int {
	totalW := m.Width()
	// Вычитаем: prefix(2) + padding left(2) + right padding(2) + list border(1) + margins
	const reserved = 7
	if totalW > reserved {
		return totalW - reserved
	}
	return 30 // fallback
}

// truncatePathLeft обрезает путь слева, оставляя видимым конец пути.
// Работает с rune-counts, не байтами — UTF-8 safe для путей с кириллицей.
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

// formatMetadataBadges формирует строку бейджей метаданных модели.
// Bordered бейджи — многострочные блоки, поэтому склеиваются через
// lipgloss.JoinHorizontal, а не strings.Join (иначе они рендерятся лесенкой).
// Порядок: Size / mmproj / Quant.
func formatMetadataBadges(m *modelscan.Model, st *StyleConfig) string {
	if st == nil {
		parts := "[" + formatSize(m.Size) + "]"
		if len(m.MMProjPaths) > 0 {
			parts += " [mmproj]"
		}
		parts += " [" + extractQuantization(m.Name) + "]"
		return parts
	}
	sep := "  "
	parts := []string{
		st.SizeBadgeStyle().Render(formatSize(m.Size)),
		sep,
	}
	if len(m.MMProjPaths) > 0 {
		parts = append(parts, st.MMProjBadgeStyle().Render("mmproj"), sep)
	}
	parts = append(parts, st.QuantizationBadgeStyle().Render(extractQuantization(m.Name)))
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
// StyledDelegate — кастомный delegate с 3-строчным форматом отображения
// ============================================================================

type StyledDelegate struct {
	base   list.DefaultDelegate
	styles *StyleConfig
}

func (d *StyledDelegate) Height() int {
	// 1 (name) + 1 (path) + 3 (bordered badges) = 5 строк
	return 5
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
	return d.styles.ItemActiveBorderStyle().Foreground(lipgloss.Color(d.styles.TextPrimary))
}

func (d *StyledDelegate) DescStyleSelected() lipgloss.Style {
	return d.styles.ItemNormalStyle().Foreground(lipgloss.Color(d.styles.TextSecondary))
}

func (d *StyledDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

