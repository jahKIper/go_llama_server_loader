package runconfig

import (
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/uistyle"
)

// wordWrap переносит текст s по словам, чтобы каждая строка не превышала w rune.
func wordWrap(s string, w int) string {
	if w <= 0 {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}
	var lines []string
	current := ""
	for _, word := range words {
		wordW := utf8.RuneCountInString(word)
		if current == "" {
			if wordW > w {
				lines = append(lines, string([]rune(word)[:w]))
				current = ""
			} else {
				current = word
			}
		} else {
			candidate := current + " " + word
			if utf8.RuneCountInString(candidate) <= w {
				current = candidate
			} else {
				lines = append(lines, current)
				if wordW > w {
					lines = append(lines, string([]rune(word)[:w]))
					current = ""
				} else {
					current = word
				}
			}
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return strings.Join(lines, "\n")
}

// truncatePath обрезает строку s до w rune, добавляя "…" если строка длиннее.
func truncatePath(s string, w int) string {
	if utf8.RuneCountInString(s) <= w {
		return s
	}
	if w <= 1 {
		return "…"
	}
	runes := []rune(s)
	return string(runes[:w-1]) + "…"
}

// stripFlagArg убирает аргументный суффикс из флага: "--ctx-size N" → "--ctx-size".
func stripFlagArg(flag string) string {
	if idx := strings.IndexByte(flag, ' '); idx >= 0 {
		return flag[:idx]
	}
	return flag
}

// injectBorderTitle вставляет лейбл в верхнюю (rounded) границу блока.
// Ожидает формат lipgloss RoundedBorder: первая строка начинается с "╭".
func injectBorderTitle(rendered, leftLabel, rightLabel string) string {
	if rendered == "" {
		return rendered
	}
	lines := strings.Split(rendered, "\n")
	if len(lines) == 0 {
		return rendered
	}
	top := lines[0]
	runes := []rune(top)
	n := len(runes)
	if n < 4 {
		return rendered
	}
	leftIdx, rightIdx := -1, -1
	for i, r := range runes {
		if r == '╭' {
			leftIdx = i
		}
		if r == '╮' {
			rightIdx = i
		}
	}
	if leftIdx < 0 || rightIdx < 0 || rightIdx <= leftIdx+1 {
		return rendered
	}
	innerRunes := runes[leftIdx+1 : rightIdx]
	innerLen := len(innerRunes)

	leftWrapped := ""
	rightWrapped := ""
	if leftLabel != "" {
		leftWrapped = " " + leftLabel + " "
	}
	if rightLabel != "" {
		rightWrapped = " " + rightLabel + " "
	}

	leftR := utf8.RuneCountInString(leftWrapped)
	rightR := utf8.RuneCountInString(rightWrapped)
	minDashes := 2
	if leftR+rightR+minDashes > innerLen {
		return rendered
	}
	dashesCount := innerLen - leftR - rightR
	dashes := strings.Repeat("─", dashesCount)
	newInner := leftWrapped + dashes + rightWrapped
	newTopRunes := make([]rune, 0, n)
	newTopRunes = append(newTopRunes, runes[:leftIdx+1]...)
	newTopRunes = append(newTopRunes, []rune(newInner)...)
	newTopRunes = append(newTopRunes, runes[rightIdx:]...)
	lines[0] = string(newTopRunes)
	return strings.Join(lines, "\n")
}

// renderScrollbarLines возвращает срез строк высотой height для вертикального скроллбара.
// offset — первый видимый элемент, visible — количество видимых, total — всего элементов.
func renderScrollbarLines(offset, visible, total, height, scrollW int, st *uistyle.StyleConfig) []string {
	cell := strings.Repeat(" ", scrollW)
	lines := make([]string, height)

	if total <= visible || height == 0 {
		empty := lipgloss.NewStyle().
			Background(lipgloss.Color(st.BgPanel)).
			Render(cell)
		for i := range lines {
			lines[i] = empty
		}
		return lines
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

	track := st.ScrollbarTrackStyle().Render(cell)
	thumb := st.ScrollbarThumbStyle().Render(cell)
	for i := range lines {
		if i >= thumbPos && i < thumbPos+thumbSize {
			lines[i] = thumb
		} else {
			lines[i] = track
		}
	}
	return lines
}
