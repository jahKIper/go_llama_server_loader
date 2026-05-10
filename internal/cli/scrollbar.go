package cli

import (
	"strings"

	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/uistyle"
)

// ScrollbarWidth — ширина scrollbar в колонках.
const ScrollbarWidth = uistyle.ScrollbarWidth

// RenderScrollbar возвращает вертикальный scrollbar шириной ScrollbarWidth (2 col).
func RenderScrollbar(offset, visible, total, height int, st *uistyle.StyleConfig) string {
	lines := make([]string, height)
	cell := strings.Repeat(" ", uistyle.ScrollbarWidth)

	if total <= visible || height == 0 {
		emptyCell := lipgloss.NewStyle().
			Background(lipgloss.Color(st.BgPanel)).
			Render(cell)
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

	trackCell := st.ScrollbarTrackStyle().Render(cell)
	thumbCell := st.ScrollbarThumbStyle().Render(cell)

	for i := range lines {
		if i >= thumbPos && i < thumbPos+thumbSize {
			lines[i] = thumbCell
		} else {
			lines[i] = trackCell
		}
	}

	return strings.Join(lines, "\n")
}
