package cli

import (
	"regexp"
	"strings"
	"testing"

	"llama-server-loader/internal/cli/uistyle"
)

func testStyles() *uistyle.StyleConfig {
	return uistyle.GetStyles()
}

func countLines(s string) int {
	return strings.Count(s, "\n") + 1
}

// ansiRe — для удаления ANSI-кодов из тестовых строк (стиль фона на пустой
// scrollbar-колонке делает строку не «голой», но визуально она остаётся пустой).
var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

// hasThumbBg — проверка, что в строке встречается ANSI-код фона thumb
// (заливка AccentPurple = #a78bfa = 167;139;250). Track использует BorderIdle
// = #1f2937 = 31;41;55, и не должен срабатывать на этот тест.
func hasThumbBg(line string) bool {
	return strings.Contains(line, "48;2;167;139;250")
}

func TestRenderScrollbar_HidesWhenFits(t *testing.T) {
	st := testStyles()
	// total == visible — scrollbar не нужен, возвращаем пробелы
	out := RenderScrollbar(0, 10, 10, 10, st)
	lines := strings.Split(out, "\n")
	if len(lines) != 10 {
		t.Fatalf("expected 10 lines, got %d", len(lines))
	}
	for i, l := range lines {
		if strings.TrimSpace(stripANSI(l)) != "" {
			t.Errorf("line %d should be blank, got %q", i, l)
		}
	}
}

func TestRenderScrollbar_HidesWhenLess(t *testing.T) {
	st := testStyles()
	// total < visible — тоже прячем
	out := RenderScrollbar(0, 15, 5, 8, st)
	lines := strings.Split(out, "\n")
	if len(lines) != 8 {
		t.Fatalf("expected 8 lines, got %d", len(lines))
	}
	for i, l := range lines {
		if strings.TrimSpace(stripANSI(l)) != "" {
			t.Errorf("line %d should be blank, got %q", i, l)
		}
	}
}

func TestRenderScrollbar_ThumbAtTop(t *testing.T) {
	st := testStyles()
	// offset=0, visible=3, total=10, height=10 → thumb сверху
	out := RenderScrollbar(0, 3, 10, 10, st)
	lines := strings.Split(out, "\n")
	if len(lines) != 10 {
		t.Fatalf("expected 10 lines, got %d", len(lines))
	}
	// thumbSize = 10*3/10 = 3, thumbPos = 0
	// строки 0,1,2 — thumb, остальные — track
	for i, l := range lines {
		hasThumb := hasThumbBg(l)
		if i < 3 && !hasThumb {
			t.Errorf("line %d: expected thumb, got %q", i, l)
		}
		if i >= 3 && hasThumb {
			t.Errorf("line %d: unexpected thumb, got %q", i, l)
		}
	}
}

func TestRenderScrollbar_ThumbAtBottom(t *testing.T) {
	st := testStyles()
	// offset=7, visible=3, total=10, height=10 → thumb внизу
	out := RenderScrollbar(7, 3, 10, 10, st)
	lines := strings.Split(out, "\n")
	if len(lines) != 10 {
		t.Fatalf("expected 10 lines, got %d", len(lines))
	}
	// thumbSize=3, maxOffset=7, thumbPos=7*(10-3)/7=7
	// строки 7,8,9 — thumb
	for i, l := range lines {
		hasThumb := hasThumbBg(l)
		if i >= 7 && !hasThumb {
			t.Errorf("line %d: expected thumb, got %q", i, l)
		}
		if i < 7 && hasThumb {
			t.Errorf("line %d: unexpected thumb, got %q", i, l)
		}
	}
}

func TestRenderScrollbar_ThumbMiddle(t *testing.T) {
	st := testStyles()
	// offset=5, visible=4, total=20, height=10
	// thumbSize = 10*4/20 = 2
	// thumbPos = 5 * (10-2) / (20-4) = 5*8/16 = 2
	out := RenderScrollbar(5, 4, 20, 10, st)
	lines := strings.Split(out, "\n")
	if len(lines) != 10 {
		t.Fatalf("expected 10 lines, got %d", len(lines))
	}
	thumbCount := 0
	for _, l := range lines {
		if hasThumbBg(l) {
			thumbCount++
		}
	}
	if thumbCount != 2 {
		t.Errorf("expected 2 thumb lines, got %d", thumbCount)
	}
}

func TestRenderScrollbar_SmallList(t *testing.T) {
	st := testStyles()
	// total = visible + 1 — граничный случай
	out := RenderScrollbar(0, 4, 5, 5, st)
	if countLines(out) != 5 {
		t.Fatalf("expected 5 lines, got %d", countLines(out))
	}
	// total = 1 — прячем
	out2 := RenderScrollbar(0, 1, 1, 3, st)
	for _, l := range strings.Split(out2, "\n") {
		if strings.TrimSpace(stripANSI(l)) != "" {
			t.Errorf("expected blank line, got %q", stripANSI(l))
		}
	}
}

func TestRenderScrollbar_ZeroHeight(t *testing.T) {
	st := testStyles()
	out := RenderScrollbar(0, 5, 20, 0, st)
	if out != "" {
		t.Errorf("zero height: expected empty string, got %q", out)
	}
}
