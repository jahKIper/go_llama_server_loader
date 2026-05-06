package cli

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// HelpRow представляет одну строку подсказок в футере.
type HelpRow struct {
	Keys  string
	Label string
}

// Footer — минималистичный однострочный footer с двумя группами биндингов.
// Левая группа (навигация, фильтр, помощь) + правая (выход) — разделены spacer-ом.
type Footer struct {
	leftRows  []HelpRow
	rightRows []HelpRow
	// Backward-compat поля: expanded/extraRows оставлены но игнорируются.
	// Удаляются в шаге 4 вместе с cli.go.
	expanded  bool
	extraRows []HelpRow
	styles    *StyleConfig
	width     int
}

// NewFooter создаёт Footer с дефолтными биндингами.
func NewFooter(styles *StyleConfig) *Footer {
	return &Footer{
		leftRows: []HelpRow{
			{Keys: "↑↓", Label: "навигация"},
			{Keys: "Enter", Label: "выбор"},
			{Keys: "/", Label: "фильтр"},
			{Keys: "?", Label: "помощь"},
			{Keys: "Tab", Label: "таб"},
		},
		rightRows: []HelpRow{
			{Keys: "^q", Label: "выход"},
		},
		extraRows: []HelpRow{
			{Keys: "→/PgDn", Label: "вперёд"},
			{Keys: "←/PgUp", Label: "назад"},
			{Keys: "g/Home", Label: "начало"},
			{Keys: "G/End", Label: "конец"},
		},
		styles: styles,
	}
}

// SetWidth устанавливает ширину футера.
func (f *Footer) SetWidth(w int) {
	f.width = w
}

// SetExpanded — no-op stub для обратной совместимости с cli.go (шаг 4 удалит).
func (f *Footer) SetExpanded(v bool) {
	f.expanded = v
}

// Render рендерит однострочный футер: левая группа + spacer + правая группа.
func (f *Footer) Render() string {
	if f.styles == nil {
		return f.renderFallback()
	}

	leftStr := f.renderGroup(f.leftRows)
	rightStr := f.renderGroup(f.rightRows)

	leftW := lipgloss.Width(leftStr)
	rightW := lipgloss.Width(rightStr)

	// Inner width — без padding контейнера; spacer должен заполнять только эту область.
	innerW := f.width - f.styles.FooterContainerStyle().GetHorizontalFrameSize()
	if innerW < 1 {
		innerW = 1
	}
	spacerW := innerW - leftW - rightW
	if spacerW < 1 {
		spacerW = 1
	}
	spacer := lipgloss.NewStyle().
		Background(lipgloss.Color(f.styles.DarkBg)).
		Render(strings.Repeat(" ", spacerW))

	line := lipgloss.JoinHorizontal(lipgloss.Top, leftStr, spacer, rightStr)
	return f.applyContainer(line)
}

// renderGroup склеивает группу биндингов через FooterSeparatorStyle.
func (f *Footer) renderGroup(rows []HelpRow) string {
	sep := f.styles.FooterSeparatorStyle().Render(" │ ")
	parts := make([]string, 0, len(rows)*2-1)
	for i, row := range rows {
		if i > 0 {
			parts = append(parts, sep)
		}
		parts = append(parts, f.renderRow(row))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// renderRow рендерит одну пару key + label.
func (f *Footer) renderRow(row HelpRow) string {
	key := f.styles.FooterKeyStyle().Render(row.Keys)
	label := f.styles.FooterLabelStyle().Render(" " + row.Label)
	return lipgloss.JoinHorizontal(lipgloss.Top, key, label)
}

// applyContainer оборачивает контент в FooterContainerStyle.
// lipgloss v2 .Width() — total width, поэтому передаём f.width напрямую.
func (f *Footer) applyContainer(content string) string {
	style := f.styles.FooterContainerStyle()
	if f.width > 0 {
		style = style.Width(f.width)
	}
	return style.Render(content)
}

// renderFallback рендерит футер без стилей.
func (f *Footer) renderFallback() string {
	all := append(f.leftRows, f.rightRows...)
	parts := make([]string, 0, len(all))
	for _, row := range all {
		parts = append(parts, row.Keys+" "+row.Label)
	}
	return strings.Join(parts, " │ ")
}

// ============================================================================
// Header рендеринг
// ============================================================================

// RenderHeader рендерит шапку: version-бейдж слева, title по центру, tabs справа.
// Если tabsStr пуст — сохраняется старое поведение (симметричный отступ справа).
func RenderHeader(title, version string, styles *StyleConfig, width int, tabsStr ...string) string {
	if styles == nil {
		return title
	}

	versionBadge := styles.VersionBadgeStyle().Render("v" + version)
	badgeW := lipgloss.Width(versionBadge)

	tabs := ""
	tabsW := 0
	if len(tabsStr) > 0 && tabsStr[0] != "" {
		tabs = tabsStr[0]
		tabsW = lipgloss.Width(tabs)
	}

	// Title занимает пространство между бейджем и табами.
	available := width - badgeW - tabsW
	if available < 1 {
		return versionBadge + "  " + styles.TitleStyle().Render(title)
	}
	titleCentered := styles.TitleStyle().
		Width(available).
		Align(lipgloss.Center).
		Render(title)

	if tabs != "" {
		return lipgloss.JoinHorizontal(lipgloss.Top, versionBadge, titleCentered, tabs)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, versionBadge, titleCentered)
}
