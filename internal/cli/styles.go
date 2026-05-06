package cli

import "charm.land/lipgloss/v2"

// === Palette ===

// StyleConfig определяет цветовую схему TUI интерфейса (WCAG AA ≥ 4.5:1).
type StyleConfig struct {
	// Existing palette — сохранены
	GreenDark     string // #064e3b — тёмный зелёный (текст на зелёном фоне)
	GreenPrimary  string // #34d399 — статичные акценты: active tab, version badge, halo
	TextPrimary   string // #ffffff — чистый белый
	TextSecondary string // #e2e8f0 — светло-серый вторичный текст
	DarkBg        string // #0a0f18 — экран и футер

	// New palette tokens
	BgPanel    string // #0d1320 — интерьер content-блока, активного фильтра
	BgSelected string // #1a2e25 — заливка выбранной строки (зелёный тинт)
	BorderIdle string // #1f2937 — рамки неактивных элементов, scrollbar track
	NeonGreen  string // #50fa7b — активность/фокус: рамки, cursor, scrollbar thumb
	TextMuted  string // #94a3b8 — placeholder, label хоткеев, disabled tab
	KeyHint    string // #6ee7b7 — клавиши в футере

	// Accent для dot-индикатора (комплементарный к зелёной палитре)
	AccentPurple      string // #a78bfa — selected dot, в тон с NeonGreen по яркости
	AccentPurpleMuted string // #6b5b95 — unselected dot, в тон с TextMuted
}

// GetStyles возвращает конфигурацию стилей с зелёной цветовой схемой.
func GetStyles() *StyleConfig {
	return &StyleConfig{
		GreenDark:     "#064e3b",
		GreenPrimary:  "#34d399",
		TextPrimary:   "#ffffff",
		TextSecondary: "#e2e8f0",
		DarkBg:        "#0a0f18",

		BgPanel:    "#0d1320",
		BgSelected: "#1a2e25",
		BorderIdle: "#1f2937",
		NeonGreen:  "#50fa7b",
		TextMuted:  "#94a3b8",
		KeyHint:    "#6ee7b7",

		AccentPurple:      "#a78bfa",
		AccentPurpleMuted: "#6b5b95",
	}
}

// === Header / Tabs ===

// TitleStyle возвращает стиль заголовка экрана.
func (s *StyleConfig) TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(s.DarkBg)).
		Foreground(lipgloss.Color(s.TextPrimary))
}

// VersionBadgeStyle возвращает стиль бейджа версии приложения.
// Фон GreenPrimary, текст GreenDark — AA контраст ~5.2:1.
func (s *StyleConfig) VersionBadgeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Padding(0, 2).
		Background(lipgloss.Color(s.GreenPrimary)).
		Foreground(lipgloss.Color(s.GreenDark))
}

// HeaderBlockStyle возвращает стиль блока заголовка.
func (s *StyleConfig) HeaderBlockStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Padding(0, 2).
		Background(lipgloss.Color(s.DarkBg)).
		Foreground(lipgloss.Color(s.TextPrimary))
}

// TabActiveStyle — активный (выбранный) таб: заливка GreenPrimary, текст GreenDark.
func (s *StyleConfig) TabActiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Background(lipgloss.Color(s.GreenPrimary)).
		Foreground(lipgloss.Color(s.GreenDark))
}

// TabInactiveStyle — неактивный enabled таб.
func (s *StyleConfig) TabInactiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Background(lipgloss.Color(s.DarkBg)).
		Foreground(lipgloss.Color(s.TextSecondary))
}

// TabDisabledStyle — отключённый таб (disabled).
func (s *StyleConfig) TabDisabledStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Background(lipgloss.Color(s.DarkBg)).
		Foreground(lipgloss.Color(s.TextMuted))
}

// TabSeparatorStyle — разделитель │ между табами.
func (s *StyleConfig) TabSeparatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(s.DarkBg)).
		Foreground(lipgloss.Color(s.BorderIdle))
}

// === Content block ===

// ContentBlockStyle — rounded рамка content-блока (AccentPurple) + BgPanel внутри.
// Фиолетовая рамка связывает контейнер с dot-индикатором selected (тоже AccentPurple),
// зелёный остаётся за «действия» — CTA, left-bracket, scrollbar thumb.
func (s *StyleConfig) ContentBlockStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(s.AccentPurple)).
		BorderBackground(lipgloss.Color(s.BgPanel)).
		Padding(1, 2).
		Background(lipgloss.Color(s.BgPanel))
}

// CountLabelStyle возвращает стиль лейбла счётчика моделей.
func (s *StyleConfig) CountLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(s.BgPanel)).
		Foreground(lipgloss.Color(s.TextSecondary))
}

// === Filter input ===

// FilterBadgeStyle — обратная совместимость (idle state badge).
func (s *StyleConfig) FilterBadgeStyle() lipgloss.Style {
	return s.FilterInputIdleStyle()
}

// FilterInputIdleStyle — поле фильтра в состоянии ожидания: приглушённая рамка.
func (s *StyleConfig) FilterInputIdleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(s.BorderIdle)).
		BorderBackground(lipgloss.Color(s.BgPanel)).
		Background(lipgloss.Color(s.BgPanel)).
		Foreground(lipgloss.Color(s.TextMuted)).
		Padding(0, 1)
}

// FilterInputActiveStyle — поле фильтра в активном состоянии: неоновая рамка.
func (s *StyleConfig) FilterInputActiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(s.NeonGreen)).
		BorderBackground(lipgloss.Color(s.BgPanel)).
		Background(lipgloss.Color(s.BgPanel)).
		Foreground(lipgloss.Color(s.TextPrimary)).
		Padding(0, 1)
}

// === List item ===

// ItemNormalStyle — обычный (не выбранный) элемент списка.
func (s *StyleConfig) ItemNormalStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(s.DarkBg)).
		Foreground(lipgloss.Color(s.TextPrimary)).
		Padding(0, 0, 0, 2)
}

// ItemSelectedStyle — выбранный элемент: только левая полоса NeonGreen + BgSelected.
func (s *StyleConfig) ItemSelectedStyle() lipgloss.Style {
	b := lipgloss.NormalBorder()
	return lipgloss.NewStyle().
		Bold(true).
		Border(b, false, false, false, true).
		BorderForeground(lipgloss.Color(s.NeonGreen)).
		BorderBackground(lipgloss.Color(s.BgSelected)).
		Background(lipgloss.Color(s.BgSelected)).
		Foreground(lipgloss.Color(s.TextPrimary)).
		Padding(0, 1)
}

// SizeBadgeStyle возвращает стиль бейджа размера файла.
// rowBg — фон строки-родителя (DarkBg / BgSelected) для прозрачной интеграции.
// BorderBackground важен: без него ячейки border'а останутся без bg → дырки терминального фона.
func (s *StyleConfig) SizeBadgeStyle(rowBg ...string) lipgloss.Style {
	bg := s.DarkBg
	if len(rowBg) > 0 && rowBg[0] != "" {
		bg = rowBg[0]
	}
	return lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(s.NeonGreen)).
		BorderBackground(lipgloss.Color(bg)).
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(s.GreenPrimary))
}

// MMProjBadgeStyle — стиль рамки бейджа mmproj (контур + padding на фоне строки).
// Заливной зелёный фон применяется отдельно к самой надписи (см. MMProjLabelStyle).
func (s *StyleConfig) MMProjBadgeStyle(rowBg ...string) lipgloss.Style {
	bg := s.DarkBg
	if len(rowBg) > 0 && rowBg[0] != "" {
		bg = rowBg[0]
	}
	return lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(s.GreenPrimary)).
		BorderBackground(lipgloss.Color(bg)).
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(s.GreenPrimary))
}

// MMProjLabelStyle — стиль самой надписи «mmproj»: заливной зелёный фон + тёмный текст.
func (s *StyleConfig) MMProjLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(s.GreenPrimary)).
		Foreground(lipgloss.Color(s.GreenDark))
}

// QuantizationBadgeStyle возвращает стиль бейджа квантования.
func (s *StyleConfig) QuantizationBadgeStyle(rowBg ...string) lipgloss.Style {
	bg := s.DarkBg
	if len(rowBg) > 0 && rowBg[0] != "" {
		bg = rowBg[0]
	}
	return lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(s.TextSecondary)).
		BorderBackground(lipgloss.Color(bg)).
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(s.TextSecondary))
}

// === Scrollbar ===

// ScrollbarTrackStyle — стиль дорожки scrollbar (заливной BorderIdle, 2 cols).
// Используется как bg-fill, без glyph'ов — стабильнее в разных терминалах.
func (s *StyleConfig) ScrollbarTrackStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(s.BorderIdle))
}

// ScrollbarThumbStyle — стиль ползунка scrollbar (заливной GreenPrimary, 2 cols).
func (s *StyleConfig) ScrollbarThumbStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(s.GreenPrimary))
}

// ScrollbarWidth — ширина scrollbar в колонках. Используется и в рендере,
// и в layout (recomputeListSize).
const ScrollbarWidth = 2

// === Footer ===

// FooterContainerStyle — минималистичный однострочный футер без top border.
func (s *StyleConfig) FooterContainerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 2).
		Background(lipgloss.Color(s.DarkBg)).
		Foreground(lipgloss.Color(s.TextSecondary))
}

// FooterPrimaryCTAStyle — стиль primary-кнопки footer'а («pill» Enter — Запустить).
// Заливной зелёный + тёмный текст, по аналогии с VersionBadge / TabActive.
func (s *StyleConfig) FooterPrimaryCTAStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Padding(0, 2).
		Background(lipgloss.Color(s.GreenPrimary)).
		Foreground(lipgloss.Color(s.GreenDark))
}

// FooterKeyStyle — клавиша в футере: KeyHint bold, subtle bg BgPanel.
func (s *StyleConfig) FooterKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Background(lipgloss.Color(s.BgPanel)).
		Foreground(lipgloss.Color(s.KeyHint))
}

// FooterLabelStyle — подпись клавиши в футере.
func (s *StyleConfig) FooterLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(s.DarkBg)).
		Foreground(lipgloss.Color(s.TextMuted))
}

// FooterSeparatorStyle — разделитель │ между парами key+label в футере.
func (s *StyleConfig) FooterSeparatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(s.DarkBg)).
		Foreground(lipgloss.Color(s.BorderIdle))
}

// === Help popup ===

// HelpPopupStyle — popup полной справки: NeonGreen rounded border, BgPanel фон.
func (s *StyleConfig) HelpPopupStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(s.NeonGreen)).
		BorderBackground(lipgloss.Color(s.BgPanel)).
		Background(lipgloss.Color(s.BgPanel)).
		Padding(1, 2)
}
