package runconfig

import (
	"testing"

	"llama-server-loader/internal/cli/uistyle"
	"llama-server-loader/internal/config"
)

func newTestLeftPanel() *LeftPanel {
	return NewLeftPanel(uistyle.GetStyles(), 60, 20)
}

func makeMeta(long, short string) *config.ParamMeta {
	return &config.ParamMeta{
		LongFlag:  long,
		ShortFlag: short,
		DescRU:    "описание " + long,
	}
}

// ── Add ───────────────────────────────────────────────────────────────────────

func TestLeftPanel_Add_Basic(t *testing.T) {
	p := newTestLeftPanel()
	p.Add(makeMeta("--ctx-size N", "-c N"))

	rows := p.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].Long != "--ctx-size" {
		t.Errorf("expected Long=--ctx-size, got %q", rows[0].Long)
	}
	if rows[0].Key != "ctx_size" {
		t.Errorf("expected Key=ctx_size, got %q", rows[0].Key)
	}
}

func TestLeftPanel_Add_NoDuplicate(t *testing.T) {
	p := newTestLeftPanel()
	meta := makeMeta("--ctx-size N", "-c N")
	p.Add(meta)
	p.Add(meta)
	p.Add(meta)

	if len(p.Rows()) != 1 {
		t.Fatalf("expected 1 row after duplicate adds, got %d", len(p.Rows()))
	}
}

func TestLeftPanel_Add_MultipleDifferent(t *testing.T) {
	p := newTestLeftPanel()
	p.Add(makeMeta("--ctx-size N", "-c"))
	p.Add(makeMeta("--temp F", "-t"))
	p.Add(makeMeta("--seed N", "-s"))

	if len(p.Rows()) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(p.Rows()))
	}
}

func TestLeftPanel_Add_NilMeta(t *testing.T) {
	p := newTestLeftPanel()
	p.Add(nil) // не должно паниковать

	if len(p.Rows()) != 0 {
		t.Fatal("expected 0 rows after nil add")
	}
}

func TestLeftPanel_Add_EmptyFlags(t *testing.T) {
	p := newTestLeftPanel()
	p.Add(&config.ParamMeta{LongFlag: "", ShortFlag: ""})

	if len(p.Rows()) != 0 {
		t.Fatal("expected 0 rows for empty flags")
	}
}

func TestLeftPanel_Add_CursorMovesToNew(t *testing.T) {
	p := newTestLeftPanel()
	p.Add(makeMeta("--ctx-size N", ""))
	p.Add(makeMeta("--temp F", ""))

	if p.cursor != 1 {
		t.Errorf("expected cursor=1 after second add, got %d", p.cursor)
	}
}

// ── Remove ────────────────────────────────────────────────────────────────────

func TestLeftPanel_Remove_First(t *testing.T) {
	p := newTestLeftPanel()
	p.Add(makeMeta("--ctx-size N", ""))
	p.Add(makeMeta("--temp F", ""))

	p.Remove(0)

	rows := p.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row after remove, got %d", len(rows))
	}
	if rows[0].Long != "--temp" {
		t.Errorf("expected --temp after remove of first, got %q", rows[0].Long)
	}
}

func TestLeftPanel_Remove_Last(t *testing.T) {
	p := newTestLeftPanel()
	p.Add(makeMeta("--ctx-size N", ""))
	p.Add(makeMeta("--temp F", ""))

	p.cursor = 1
	p.Remove(1)

	if len(p.Rows()) != 1 {
		t.Fatalf("expected 1 row, got %d", len(p.Rows()))
	}
	// Курсор должен быть скорректирован (не выходить за границы)
	if p.cursor >= len(p.Rows()) {
		t.Errorf("cursor out of bounds: cursor=%d len=%d", p.cursor, len(p.Rows()))
	}
}

func TestLeftPanel_Remove_OutOfBounds(t *testing.T) {
	p := newTestLeftPanel()
	p.Add(makeMeta("--ctx-size N", ""))

	p.Remove(-1)  // не паникует
	p.Remove(100) // не паникует

	if len(p.Rows()) != 1 {
		t.Fatal("row count should not change on out-of-bounds remove")
	}
}

func TestLeftPanel_Remove_ToEmpty(t *testing.T) {
	p := newTestLeftPanel()
	p.Add(makeMeta("--ctx-size N", ""))
	p.Remove(0)

	if len(p.Rows()) != 0 {
		t.Fatal("expected 0 rows after removing only element")
	}
	if p.cursor != 0 {
		t.Errorf("cursor should be 0 after empty list, got %d", p.cursor)
	}
}

// ── Navigation ────────────────────────────────────────────────────────────────

func TestLeftPanel_Navigate_EmptyList(t *testing.T) {
	p := newTestLeftPanel()
	p.MoveUp()   // не паникует
	p.MoveDown() // не паникует

	if p.cursor != 0 {
		t.Errorf("cursor should stay 0 on empty list, got %d", p.cursor)
	}
}

func TestLeftPanel_Navigate_Bounds(t *testing.T) {
	p := newTestLeftPanel()
	p.Add(makeMeta("--ctx-size N", ""))
	p.Add(makeMeta("--temp F", ""))
	p.Add(makeMeta("--seed N", ""))

	// Up на первом не уходит ниже 0
	p.cursor = 0
	p.MoveUp()
	if p.cursor != 0 {
		t.Errorf("cursor should stay 0 at top, got %d", p.cursor)
	}

	// Down до конца и ещё раз
	for i := 0; i < 10; i++ {
		p.MoveDown()
	}
	if p.cursor != 2 {
		t.Errorf("cursor should be at last index 2, got %d", p.cursor)
	}
}

func TestLeftPanel_Navigate_UpDown(t *testing.T) {
	p := newTestLeftPanel()
	p.Add(makeMeta("--ctx-size N", ""))
	p.Add(makeMeta("--temp F", ""))

	p.MoveDown()
	if p.cursor != 1 {
		t.Errorf("expected cursor=1 after MoveDown, got %d", p.cursor)
	}
	p.MoveUp()
	if p.cursor != 0 {
		t.Errorf("expected cursor=0 after MoveUp, got %d", p.cursor)
	}
}

// ── Render (smoke) ────────────────────────────────────────────────────────────

func TestLeftPanel_Render_Empty(t *testing.T) {
	p := newTestLeftPanel()
	out := p.Render(false)
	if out == "" {
		t.Error("Render should return non-empty string")
	}
}

func TestLeftPanel_Render_WithRows(t *testing.T) {
	p := newTestLeftPanel()
	p.Add(makeMeta("--ctx-size N", "-c N"))
	out := p.Render(true)
	if out == "" {
		t.Error("Render should return non-empty string with rows")
	}
}

func TestLeftPanel_Render_TooSmall(t *testing.T) {
	p := NewLeftPanel(uistyle.GetStyles(), 5, 2)
	p.Add(makeMeta("--ctx-size N", ""))
	// Не должно паниковать на маленьком размере
	_ = p.Render(false)
}
