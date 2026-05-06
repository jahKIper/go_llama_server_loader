package cli

import (
	"strings"
	"unicode/utf8"
)

// InjectBorderTitle вставляет лейблы в верхнюю (rounded) границу блока.
// Формат верхней границы lipgloss RoundedBorder: "╭─────────────────╮"
// Лейбл слева вставляется после "╭", лейбл справа — перед "╮".
// Если лейблы не помещаются, возвращает rendered без изменений.
func InjectBorderTitle(rendered, leftLabel, rightLabel string) string {
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

	// Ищем позиции угловых символов rounded border (в rune-индексах)
	leftIdx := -1
	rightIdx := -1
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

	// Внутренняя часть — руны между углами (без самих углов)
	innerRunes := runes[leftIdx+1 : rightIdx]
	innerLen := len(innerRunes)

	leftWrapped := " " + leftLabel + " "
	rightWrapped := " " + rightLabel + " "

	leftR := utf8.RuneCountInString(leftWrapped)
	rightR := utf8.RuneCountInString(rightWrapped)

	minDashes := 2 // минимум ─ между лейблами
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
