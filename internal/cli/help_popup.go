package cli

import (
	"charm.land/lipgloss/v2"
)

// HelpPopup рендерит полноэкранный centered popup со всеми биндингами.
type HelpPopup struct {
	styles *StyleConfig
}

// NewHelpPopup создаёт HelpPopup.
func NewHelpPopup(st *StyleConfig) *HelpPopup {
	return &HelpPopup{styles: st}
}

type helpBinding struct {
	key  string
	desc string
}

var helpBindings = []helpBinding{
	{"↑ / ↓", "навигация по списку"},
	{"Enter", "выбрать модель"},
	{"/ ", "открыть фильтр"},
	{"Esc", "закрыть фильтр / это окно"},
	{"? ", "эта справка"},
	{"Tab", "следующий таб"},
	{"Shift+Tab", "предыдущий таб"},
	{"1 / 2 / 3", "переключить таб по номеру"},
	{"g / Home", "первый элемент"},
	{"G / End", "последний элемент"},
	{"PgUp", "предыдущая страница"},
	{"PgDn", "следующая страница"},
	{"^q", "выход"},
}

// Render возвращает строку центрированного popup.
func (h *HelpPopup) Render(screenWidth, screenHeight int) string {
	st := h.styles

	// Ширина popup: ~50% экрана, clamp [40, 80]
	popupW := screenWidth / 2
	if popupW < 40 {
		popupW = 40
	}
	if popupW > 80 {
		popupW = 80
	}

	// Заголовок
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(st.GreenPrimary))
	title := titleStyle.Render("Справка")

	// Строки биндингов
	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(st.KeyHint)).
		Width(14)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(st.TextSecondary))

	rows := make([]string, 0, len(helpBindings)+3)
	rows = append(rows, title, "")
	for _, b := range helpBindings {
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Render(b.key),
			descStyle.Render(b.desc),
		)
		rows = append(rows, row)
	}

	// Footer popup
	closeHint := lipgloss.NewStyle().
		Foreground(lipgloss.Color(st.TextMuted)).
		Render("Esc или ? — закрыть")
	rows = append(rows, "", closeHint)

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)

	// Оборачиваем в HelpPopupStyle
	innerW := popupW - st.HelpPopupStyle().GetHorizontalFrameSize()
	if innerW < 10 {
		innerW = 10
	}
	popup := st.HelpPopupStyle().Width(innerW).Render(body)

	// Центрируем на экране. BackgroundColor задаётся в View() через v.BackgroundColor,
	// поэтому WithWhitespaceStyle не нужен — он ломает измерение contentWidth в v2.
	return lipgloss.Place(
		screenWidth, screenHeight,
		lipgloss.Center, lipgloss.Center,
		popup,
	)
}
