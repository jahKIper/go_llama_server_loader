package cli

import (
	"strings"
	"testing"
)

func TestNewTabBar_DefaultState(t *testing.T) {
	tb := NewTabBar(GetStyles())
	if tb.Active() != 0 {
		t.Errorf("expected active=0, got %d", tb.Active())
	}
	if len(tb.tabs) != 3 {
		t.Fatalf("expected 3 tabs, got %d", len(tb.tabs))
	}
	if !tb.tabs[0].Enabled {
		t.Error("tab 0 (Models) should be enabled")
	}
	if tb.tabs[1].Enabled || tb.tabs[2].Enabled {
		t.Error("tabs 1,2 should be disabled")
	}
}

func TestTabBar_SetActive(t *testing.T) {
	tb := NewTabBar(GetStyles())
	tb.SetActive(2)
	if tb.Active() != 2 {
		t.Errorf("expected active=2, got %d", tb.Active())
	}
	// out of range — no-op
	tb.SetActive(99)
	if tb.Active() != 2 {
		t.Errorf("out-of-range SetActive should be no-op, got %d", tb.Active())
	}
}

func TestTabBarNext_SkipsDisabled(t *testing.T) {
	tb := NewTabBar(GetStyles())
	// Только один enabled таб (Models) — Next() должен быть no-op
	tb.Next()
	if tb.Active() != 0 {
		t.Errorf("Next() with one enabled tab should be no-op, got active=%d", tb.Active())
	}
}

func TestTabBarPrev_SkipsDisabled(t *testing.T) {
	tb := NewTabBar(GetStyles())
	tb.Prev()
	if tb.Active() != 0 {
		t.Errorf("Prev() with one enabled tab should be no-op, got active=%d", tb.Active())
	}
}

func TestTabBarNext_MultipleEnabled(t *testing.T) {
	tb := NewTabBar(GetStyles())
	// Включаем все табы вручную
	tb.tabs[1].Enabled = true
	tb.tabs[2].Enabled = true

	tb.SetActive(0)
	tb.Next()
	if tb.Active() != 1 {
		t.Errorf("expected active=1, got %d", tb.Active())
	}
	tb.Next()
	if tb.Active() != 2 {
		t.Errorf("expected active=2, got %d", tb.Active())
	}
	// Wrap around
	tb.Next()
	if tb.Active() != 0 {
		t.Errorf("expected wrap to 0, got %d", tb.Active())
	}
}

func TestTabBarPrev_MultipleEnabled(t *testing.T) {
	tb := NewTabBar(GetStyles())
	tb.tabs[1].Enabled = true
	tb.tabs[2].Enabled = true

	tb.SetActive(0)
	tb.Prev()
	if tb.Active() != 2 {
		t.Errorf("expected wrap to 2, got %d", tb.Active())
	}
}

func TestTabBar_Render_ContainsLabels(t *testing.T) {
	tb := NewTabBar(GetStyles())
	out := tb.Render()
	for _, label := range []string{"Models", "Running", "Logs"} {
		if !strings.Contains(out, label) {
			t.Errorf("Render() missing label %q", label)
		}
	}
}

func TestTabBar_Render_NilStyles(t *testing.T) {
	tb := NewTabBar(nil)
	out := tb.Render()
	if !strings.Contains(out, "[Models]") {
		t.Errorf("fallback render should wrap active tab in brackets, got %q", out)
	}
}
