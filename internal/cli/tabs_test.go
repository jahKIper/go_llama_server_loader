package cli

import (
	"strings"
	"testing"

	"llama-server-loader/internal/cli/uistyle"
)

func TestNewTabBar_DefaultState(t *testing.T) {
	tb := NewTabBar(uistyle.GetStyles())
	if tb.Active() != 0 {
		t.Errorf("expected active=0, got %d", tb.Active())
	}
	if len(tb.tabs) != 4 {
		t.Fatalf("expected 4 tabs, got %d", len(tb.tabs))
	}
	if !tb.tabs[TabModels].Enabled {
		t.Error("Models tab should be enabled")
	}
	if !tb.tabs[TabParams].Enabled {
		t.Error("Params tab should be enabled")
	}
	if tb.tabs[TabRunning].Enabled || tb.tabs[TabLogs].Enabled {
		t.Error("Running and Logs tabs should be disabled")
	}
}

func TestTabBar_SetActive(t *testing.T) {
	tb := NewTabBar(uistyle.GetStyles())
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
	tb := NewTabBar(uistyle.GetStyles())
	// По умолчанию enabled: Models и Params → Next() ходит между ними.
	tb.Next()
	if tb.Active() != TabParams {
		t.Errorf("expected active=Params after Next from Models, got %d", tb.Active())
	}
	tb.Next()
	if tb.Active() != TabModels {
		t.Errorf("expected wrap to Models, got %d", tb.Active())
	}
}

func TestTabBarPrev_SkipsDisabled(t *testing.T) {
	tb := NewTabBar(uistyle.GetStyles())
	tb.Prev()
	if tb.Active() != TabParams {
		t.Errorf("expected wrap to Params, got %d", tb.Active())
	}
}

func TestTabBarNext_MultipleEnabled(t *testing.T) {
	tb := NewTabBar(uistyle.GetStyles())
	tb.tabs[TabRunning].Enabled = true
	tb.tabs[TabLogs].Enabled = true

	tb.SetActive(0)
	tb.Next()
	if tb.Active() != 1 {
		t.Errorf("expected active=1, got %d", tb.Active())
	}
	tb.Next()
	if tb.Active() != 2 {
		t.Errorf("expected active=2, got %d", tb.Active())
	}
	tb.Next()
	if tb.Active() != 3 {
		t.Errorf("expected active=3, got %d", tb.Active())
	}
	// Wrap around
	tb.Next()
	if tb.Active() != 0 {
		t.Errorf("expected wrap to 0, got %d", tb.Active())
	}
}

func TestTabBarPrev_MultipleEnabled(t *testing.T) {
	tb := NewTabBar(uistyle.GetStyles())
	tb.tabs[TabRunning].Enabled = true
	tb.tabs[TabLogs].Enabled = true

	tb.SetActive(0)
	tb.Prev()
	if tb.Active() != 3 {
		t.Errorf("expected wrap to 3, got %d", tb.Active())
	}
}

func TestTabBar_Render_ContainsLabels(t *testing.T) {
	tb := NewTabBar(uistyle.GetStyles())
	out := tb.Render()
	for _, label := range []string{"Models", "Params", "Running", "Logs"} {
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
