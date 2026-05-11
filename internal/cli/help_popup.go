package cli

import (
	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/uistyle"
)

// HelpPopup рендерит полноэкранный centered popup со всеми биндингами.
type HelpPopup struct {
	styles *uistyle.StyleConfig
}

// NewHelpPopup создаёт HelpPopup.
func NewHelpPopup(st *uistyle.StyleConfig) *HelpPopup {
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

	popupW := screenWidth / 2
	if popupW < 40 {
		popupW = 40
	}
	if popupW > 80 {
		popupW = 80
	}

	innerW := popupW - st.HelpPopupStyle().GetHorizontalFrameSize()
	if innerW < 10 {
		innerW = 10
	}

	rowFill := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Width(innerW)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.GreenPrimary))

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.KeyHint)).
		Width(14)

	descStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextSecondary))

	closeStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextMuted))

	empty := rowFill.Render("")

	rows := make([]string, 0, len(helpBindings)+4)
	rows = append(rows, rowFill.Render(titleStyle.Render("Справка")), empty)
	for _, b := range helpBindings {
		inline := lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Render(b.key),
			descStyle.Render(b.desc),
		)
		rows = append(rows, rowFill.Render(inline))
	}
	rows = append(rows, empty, rowFill.Render(closeStyle.Render("Esc или ? — закрыть")))

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)

	popup := st.HelpPopupStyle().Width(innerW).Render(body)

	return lipgloss.Place(
		screenWidth, screenHeight,
		lipgloss.Center, lipgloss.Center,
		popup,
	)
}
