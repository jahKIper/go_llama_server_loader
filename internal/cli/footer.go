package cli

import (
	"strings"

	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/uistyle"
)

// HelpRow представляет одну строку подсказок в футере.
type HelpRow struct {
	Keys  string
	Label string
}

// Footer — минималистичный однострочный footer с двумя группами биндингов.
type Footer struct {
	leftRows   []HelpRow
	rightRows  []HelpRow
	styles     *uistyle.StyleConfig
	width      int
	primaryCTA string
}

// NewFooter создаёт Footer с дефолтными биндингами.
func NewFooter(styles *uistyle.StyleConfig) *Footer {
	return &Footer{
		leftRows: []HelpRow{
			{Keys: "↑↓", Label: "навигация"},
			{Keys: "/", Label: "фильтр"},
			{Keys: "?", Label: "помощь"},
			{Keys: "Tab", Label: "таб"},
		},
		rightRows: []HelpRow{
			{Keys: "^q", Label: "выход"},
		},
		styles: styles,
	}
}

// SetWidth устанавливает ширину футера.
func (f *Footer) SetWidth(w int) {
	f.width = w
}

// SetPrimaryCTA задаёт текст primary-кнопки footer'а. Пустая строка — кнопка скрыта.
func (f *Footer) SetPrimaryCTA(text string) {
	f.primaryCTA = text
}

// Render рендерит однострочный футер: [CTA] [hints-left] [spacer] [hints-right].
func (f *Footer) Render() string {
	if f.styles == nil {
		return f.renderFallback()
	}

	hintsLeftStr := f.renderGroup(f.leftRows)
	rightStr := f.renderGroup(f.rightRows)

	var ctaStr, ctaSepStr string
	if f.primaryCTA != "" {
		ctaStr = f.styles.FooterPrimaryCTAStyle().Render(f.primaryCTA)
		ctaSepStr = lipgloss.NewStyle().
			Background(lipgloss.Color(f.styles.DarkBg)).
			Render("  ")
	}

	ctaW := lipgloss.Width(ctaStr) + lipgloss.Width(ctaSepStr)
	hintsLeftW := lipgloss.Width(hintsLeftStr)
	rightW := lipgloss.Width(rightStr)

	innerW := f.width - f.styles.FooterContainerStyle().GetHorizontalFrameSize()
	if innerW < 1 {
		innerW = 1
	}
	spacerW := innerW - ctaW - hintsLeftW - rightW
	if spacerW < 1 {
		spacerW = 1
	}
	spacer := lipgloss.NewStyle().
		Background(lipgloss.Color(f.styles.DarkBg)).
		Render(strings.Repeat(" ", spacerW))

	parts := make([]string, 0, 5)
	if ctaStr != "" {
		parts = append(parts, ctaStr, ctaSepStr)
	}
	parts = append(parts, hintsLeftStr, spacer, rightStr)
	line := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
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
func RenderHeader(title, version string, styles *uistyle.StyleConfig, width int, tabsStr ...string) string {
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
