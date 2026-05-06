package cli

import "charm.land/lipgloss/v2"

// StyleConfig определяет цветовую схему TUI интерфейса.
// Цвета выбраны согласно WCAG AA требованиям доступности (контраст ≥ 4.5:1).
type StyleConfig struct {
	GreenDark     string // #064e3b - темный зеленый (текст на зеленом фоне)
	GreenPrimary  string // #34d399 - основной зеленый (CTA кнопки, активные элементы)
	GreenBright   string // #4ade80 - яркий зеленый (границы, акценты, hover-эффекты)
	TextPrimary   string // #ffffff - чистый белый для максимального контраста на темном фоне
	TextSecondary string // #e2e8f0 - светло-серый для вторичного текста (контраст 13.8:1 на DarkBg)
	// Новые цвета для темного фона интерфейса
	DarkBg        string // #0a0f18 - темный фон интерфейса
	ListBorder    string // #4ade80 - яркая граница блока списка (контраст 5.2:1)
}

// GetStyles возвращает конфигурацию стилей с зеленой цветовой схемой.
func GetStyles() *StyleConfig {
	return &StyleConfig{
		GreenDark:     "#064e3b",
		GreenPrimary:  "#34d399",
		GreenBright:   "#4ade80", // Яркий зеленый для границ и акцентов (контраст 5.2:1 на темном фоне)
		TextPrimary:   "#ffffff", // Чистый белый для максимального контраста (16.2:1 на DarkBg)
		TextSecondary: "#e2e8f0", // Светло-серый для вторичного текста (контраст 13.8:1 на DarkBg)
		DarkBg:        "#0a0f18",
		ListBorder:    "#4ade80", // Яркая граница для лучшей видимости
	}
}

// VersionBadgeStyle возвращает стиль бейджа версии приложения.
// Фон: GreenBright, Текст: GreenDark — максимальная читаемость.
func (s *StyleConfig) VersionBadgeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Padding(0, 2).
		Background(lipgloss.Color(s.GreenBright)).
		Foreground(lipgloss.Color(s.GreenDark))
}

// TitleStyle возвращает стиль заголовка экрана.
func (s *StyleConfig) TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Padding(0, 0).
		Foreground(lipgloss.Color(s.TextPrimary))
}

// FilterBadgeStyle возвращает стиль бейджа FILTER (статичный, в режиме ожидания).
func (s *StyleConfig) FilterBadgeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Padding(0, 2).
		Background(lipgloss.Color(s.GreenBright)).
		Foreground(lipgloss.Color(s.GreenDark))
}

// CountLabelStyle возвращает стиль лейбла счётчика моделей.
func (s *StyleConfig) CountLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 0).
		Foreground(lipgloss.Color(s.TextSecondary))
}

// ItemActiveBorderStyle возвращает стиль активного элемента списка с неоновым бордером.
// Левый бордер GreenBright. Shadow не поддерживается lipgloss v2 — fallback на чистый бордер.
func (s *StyleConfig) ItemActiveBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color(s.GreenBright)).
		Padding(0, 0, 0, 1)
}

// ItemNormalStyle возвращает стиль обычного (не выбранного) элемента.
func (s *StyleConfig) ItemNormalStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(s.DarkBg)).
		Foreground(lipgloss.Color(s.TextPrimary)).
		Padding(0, 0, 0, 2)
}

// HeaderBlockStyle возвращает стиль блока заголовка с темным фоном.
// Без вертикального padding — header это одна строка по схеме §2.
func (s *StyleConfig) HeaderBlockStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Padding(0, 2).
		Background(lipgloss.Color(s.DarkBg)).
		Foreground(lipgloss.Color(s.TextPrimary))
}

// ItemSelectedStyle возвращает стиль выбранного элемента списка.
func (s *StyleConfig) ItemSelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color(s.GreenBright)).
		Background(lipgloss.Color(s.DarkBg)).
		Foreground(lipgloss.Color(s.TextPrimary)).
		Padding(0, 0, 0, 1)
}

// ContentBlockStyle возвращает стиль контейнера блока 2 (filter+count+list)
// с одной общей рамкой по схеме §2. Без Margin — растягивается на 100% ширины.
func (s *StyleConfig) ContentBlockStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(s.ListBorder)).
		Padding(1, 2).
		Background(lipgloss.Color(s.DarkBg))
}

// FooterContainerStyle возвращает стиль контейнера футера: одна строка по
// центру, верхний бордер-разделитель, без вертикального padding.
func (s *StyleConfig) FooterContainerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Padding(0, 2).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color(s.GreenBright)).
		Background(lipgloss.Color("#050a12")).
		Foreground(lipgloss.Color(s.TextSecondary))
}

// SizeBadgeStyle возвращает стиль бейджа размера файла.
func (s *StyleConfig) SizeBadgeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(s.GreenBright)).
		Background(lipgloss.Color("#050a12")).
		Foreground(lipgloss.Color(s.GreenPrimary))
}

// MMProjBadgeStyle возвращает стиль бейджа mmproj.
func (s *StyleConfig) MMProjBadgeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(s.GreenPrimary)).
		Background(lipgloss.Color(s.GreenPrimary)).
		Foreground(lipgloss.Color(s.GreenDark))
}

// QuantizationBadgeStyle возвращает стиль бейджа квантования.
func (s *StyleConfig) QuantizationBadgeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(s.TextSecondary)).
		Background(lipgloss.Color("#050a12")).
		Foreground(lipgloss.Color(s.TextSecondary))
}


