package runconfig

import (
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/uistyle"
)

type footerHint struct {
	key   string
	label string
}

var footerHints = []footerHint{
	{"Bksp", "← модели"},
	{"Tab", "переключить"},
	{"↑↓", "выбрать"},
	{"Enter", "add/edit"},
	{"e", "редакт."},
	{"d", "удалить"},
	{"?", "описание"},
	{"/", "поиск"},
	{"r", "Run"},
	{"q", "выход"},
}

// RenderFooter рендерит однострочный footer с подсказками клавиш.
// Если содержимое шире w — обрезает хинты справа до тех пор, пока не вместится;
// если даже один хинт не влезает — обрезает строку с "…".
func RenderFooter(st *uistyle.StyleConfig, w int) string {
	if w <= 0 {
		return ""
	}

	sep := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Foreground(lipgloss.Color(st.BorderIdle)).
		Render(" · ")
	sepW := utf8.RuneCountInString(" · ")

	renderHint := func(h footerHint) string {
		k := lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color(st.DarkBg)).
			Foreground(lipgloss.Color(st.KeyHint)).
			Render(h.key)
		l := lipgloss.NewStyle().
			Background(lipgloss.Color(st.DarkBg)).
			Foreground(lipgloss.Color(st.TextMuted)).
			Render(" " + h.label)
		return lipgloss.JoinHorizontal(lipgloss.Top, k, l)
	}

	outer := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Width(w)

	// Строим line, прогрессивно добавляя хинты; как только не влезает — останавливаемся.
	type built struct {
		rendered string
		visW     int
	}
	var segments []built
	usedW := 0

	for i, h := range footerHints {
		seg := renderHint(h)
		segW := lipgloss.Width(seg)
		extra := segW
		if i > 0 {
			extra += sepW
		}
		if usedW+extra > w {
			break
		}
		segments = append(segments, built{seg, segW})
		usedW += extra
	}

	var parts []string
	for i, s := range segments {
		if i > 0 {
			parts = append(parts, sep)
		}
		parts = append(parts, s.rendered)
	}

	var line string
	if len(parts) == 0 {
		// Вырожденный случай: вообще ничего не влезло — показываем обрезанный текст.
		fallback := "Bksp · Tab · ↑↓ · Enter · d · ? · / · r · q"
		runes := []rune(fallback)
		if len(runes) > w-1 {
			runes = runes[:w-1]
			fallback = string(runes) + "…"
		}
		line = lipgloss.NewStyle().
			Background(lipgloss.Color(st.DarkBg)).
			Foreground(lipgloss.Color(st.TextMuted)).
			Render(fallback)
	} else {
		line = lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	}

	return outer.Render(strings.TrimRight(line, " "))
}
