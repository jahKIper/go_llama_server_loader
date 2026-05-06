package cli

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// HelpRow представляет одну строку подсказок в футере.
type HelpRow struct {
	Keys  string // Например "↑↓"
	Label string // Например "навигация"
}

// Footer отвечает за рендеринг подсказок клавиш в подвале приложения.
// Поддерживает два режима: compact (одна строка) и expanded (две строки)
// — переключение по клавише "?".
type Footer struct {
	compactRows []HelpRow // всегда видны
	extraRows   []HelpRow // видны только в expanded
	expanded    bool
	styles      *StyleConfig
	width       int
}

// NewFooter создает новый Footer с дефолтными подсказками.
func NewFooter(styles *StyleConfig) *Footer {
	return &Footer{
		compactRows: []HelpRow{
			{Keys: "↑↓", Label: "навигация"},
			{Keys: "Enter", Label: "выбор"},
			{Keys: "/", Label: "фильтр"},
			{Keys: "q", Label: "выход"},
			{Keys: "?", Label: "помощь"},
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

// SetExpanded переключает развёрнутый режим (показ extraRows).
// В compact метка кнопки "?" — "помощь", в expanded — "скрыть".
func (f *Footer) SetExpanded(v bool) {
	f.expanded = v
	for i := range f.compactRows {
		if f.compactRows[i].Keys == "?" {
			if v {
				f.compactRows[i].Label = "скрыть"
			} else {
				f.compactRows[i].Label = "помощь"
			}
			break
		}
	}
}

// Render рендерит футер. В expanded режиме два ряда подсказок.
func (f *Footer) Render() string {
	if f.styles == nil {
		return f.renderFallback()
	}

	line1 := f.renderLine(f.compactRows)
	if !f.expanded {
		return f.applyContainer(line1)
	}

	line2 := f.renderLine(f.extraRows)
	combined := lipgloss.JoinVertical(lipgloss.Center, line1, line2)
	return f.applyContainer(combined)
}

// applyContainer оборачивает контент в FooterContainerStyle с учётом ширины.
func (f *Footer) applyContainer(content string) string {
	style := f.styles.FooterContainerStyle()
	if f.width > 0 {
		style = style.Width(f.width).Align(lipgloss.Center)
	}
	return style.Render(content)
}

// renderLine склеивает ряд подсказок через разделитель "│".
func (f *Footer) renderLine(rows []HelpRow) string {
	var parts []string
	for _, row := range rows {
		parts = append(parts, f.renderRow(row))
	}
	return strings.Join(parts, "  │  ")
}

// renderRow рендерит одну пару key+label.
func (f *Footer) renderRow(row HelpRow) string {
	keyStyle := f.styles.VersionBadgeStyle()
	labelStyle := f.styles.CountLabelStyle()
	return keyStyle.Render(row.Keys) + " " + labelStyle.Render(row.Label)
}

// renderFallback рендерит футер без стилей (для nil styles).
func (f *Footer) renderFallback() string {
	rows := f.compactRows
	if f.expanded {
		rows = append(rows, f.extraRows...)
	}
	var parts []string
	for _, row := range rows {
		parts = append(parts, row.Keys+" "+row.Label)
	}
	return strings.Join(parts, " │ ")
}

// ============================================================================
// Header рендеринг
// ============================================================================

// RenderHeader рендерит шапку приложения: бейдж версии слева, title по центру строки.
// width — полная ширина терминала (для расчёта центровки).
func RenderHeader(title, version string, styles *StyleConfig, width int) string {
	if styles == nil {
		return title
	}

	versionBadge := styles.VersionBadgeStyle().Render("v" + version)
	badgeW := lipgloss.Width(versionBadge)

	// Title центрирован относительно полной ширины. Чтобы визуальный центр
	// совпал с центром экрана, занимаем всю оставшуюся ширину после бейджа
	// и центрируем title внутри. Резерв справа = badgeW для симметрии.
	available := width - badgeW*2
	if available < 1 {
		return versionBadge + "  " + styles.TitleStyle().Render(title)
	}
	titleCentered := styles.TitleStyle().
		Width(available).
		Align(lipgloss.Center).
		Render(title)

	return lipgloss.JoinHorizontal(lipgloss.Top, versionBadge, titleCentered)
}
