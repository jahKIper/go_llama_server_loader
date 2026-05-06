package cli

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	scrollbarTrack = "│"
	scrollbarThumb = "┃"
)

// RenderScrollbar возвращает вертикальный scrollbar шириной 1 символ.
// Если total <= visible, возвращает колонку пробелов (scrollbar не нужен,
// но высота сохраняется — UI не «прыгает»).
// Параметры: offset — первый видимый элемент, visible — кол-во видимых,
// total — кол-во всех, height — высота в строках.
func RenderScrollbar(offset, visible, total, height int, st *StyleConfig) string {
	lines := make([]string, height)

	if total <= visible || height == 0 {
		// Пустая колонка: пробел с фоном BgPanel — чтобы не было «дырки» терминального фона.
		emptyCell := lipgloss.NewStyle().
			Background(lipgloss.Color(st.BgPanel)).
			Render(" ")
		for i := range lines {
			lines[i] = emptyCell
		}
		return strings.Join(lines, "\n")
	}

	thumbSize := height * visible / total
	if thumbSize < 1 {
		thumbSize = 1
	}

	maxOffset := total - visible
	if maxOffset < 1 {
		maxOffset = 1
	}
	thumbPos := offset * (height - thumbSize) / maxOffset

	trackStyle := st.ScrollbarTrackStyle()
	thumbStyle := st.ScrollbarThumbStyle()

	for i := range lines {
		if i >= thumbPos && i < thumbPos+thumbSize {
			lines[i] = thumbStyle.Render(scrollbarThumb)
		} else {
			lines[i] = trackStyle.Render(scrollbarTrack)
		}
	}

	return strings.Join(lines, "\n")
}
