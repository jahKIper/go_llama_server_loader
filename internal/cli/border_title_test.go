package cli

import (
	"strings"
	"testing"
)

// makeBorder создаёт строку, имитирующую верхний бордер lipgloss RoundedBorder шириной w.
func makeBorder(w int) string {
	if w < 2 {
		return "╭╮"
	}
	return "╭" + strings.Repeat("─", w-2) + "╮"
}

func TestInjectBorderTitle_Basic(t *testing.T) {
	// 3 строки: верхняя граница + контент + нижняя граница
	top := makeBorder(40)
	rendered := top + "\n│ content │\n╰" + strings.Repeat("─", 38) + "╯"

	out := InjectBorderTitle(rendered, "Models", "10 / 25")

	lines := strings.Split(out, "\n")
	topOut := lines[0]

	if !strings.Contains(topOut, "Models") {
		t.Errorf("left label missing in top border: %q", topOut)
	}
	if !strings.Contains(topOut, "10 / 25") {
		t.Errorf("right label missing in top border: %q", topOut)
	}
	// Остальные строки не тронуты
	if lines[1] != "│ content │" {
		t.Errorf("line 1 should be unchanged, got %q", lines[1])
	}
}

func TestInjectBorderTitle_LabelsTooWide(t *testing.T) {
	// Граница слишком узкая — лейблы не помещаются, возвращаем без изменений
	top := makeBorder(10)
	rendered := top + "\n│ hi │\n╰────────╯"

	out := InjectBorderTitle(rendered, "Very Long Left Label", "Very Long Right Label")
	if out != rendered {
		t.Errorf("expected unchanged rendered when labels are too wide")
	}
}

func TestInjectBorderTitle_EmptyLabels(t *testing.T) {
	top := makeBorder(30)
	rendered := top + "\n│ content │"

	out := InjectBorderTitle(rendered, "", "")
	// Пустые лейблы вставляются как " " + "" + " " = "  " с каждой стороны
	lines := strings.Split(out, "\n")
	if !strings.HasPrefix(lines[0], "╭") {
		t.Errorf("first line should still start with ╭, got %q", lines[0])
	}
}

func TestInjectBorderTitle_NoBorderCorners(t *testing.T) {
	// Строка без угловых символов — возвращаем без изменений
	rendered := "plain text\nsecond line"
	out := InjectBorderTitle(rendered, "L", "R")
	if out != rendered {
		t.Errorf("no border corners: expected unchanged rendered")
	}
}

func TestInjectBorderTitle_EmptyInput(t *testing.T) {
	out := InjectBorderTitle("", "L", "R")
	if out != "" {
		t.Errorf("empty input should return empty, got %q", out)
	}
}

func TestInjectBorderTitle_Unicode(t *testing.T) {
	// Кириллические лейблы — rune-safe обработка
	top := makeBorder(50)
	rendered := top + "\n│ content │"

	out := InjectBorderTitle(rendered, "Модели", "10 / 25")
	lines := strings.Split(out, "\n")
	if !strings.Contains(lines[0], "Модели") {
		t.Errorf("unicode left label missing: %q", lines[0])
	}
}

func TestInjectBorderTitle_PreservesOtherLines(t *testing.T) {
	top := makeBorder(40)
	line2 := "│ line 2 content │"
	line3 := "╰" + strings.Repeat("─", 38) + "╯"
	rendered := top + "\n" + line2 + "\n" + line3

	out := InjectBorderTitle(rendered, "L", "R")
	outLines := strings.Split(out, "\n")
	if len(outLines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(outLines))
	}
	if outLines[1] != line2 {
		t.Errorf("line 2 changed: %q", outLines[1])
	}
	if outLines[2] != line3 {
		t.Errorf("line 3 changed: %q", outLines[2])
	}
}
