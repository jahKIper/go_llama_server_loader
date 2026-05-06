package cli

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"llama-server-loader/pkg/modelscan"
)

func TestRenderVerification(t *testing.T) {
	models := []*modelscan.Model{
		{Name: "mistral-7b-instruct-v0.2.Q4_K_M.gguf", Path: "/models/mistral.gguf", Size: 4_200_000_000},
		{Name: "llama-3-8b-q5_k_m.gguf", Path: "/models/llama.gguf", Size: 5_300_000_000},
		{Name: "gemma-2-9b-it-Q8_0.gguf", Path: "/models/gemma.gguf", Size: 9_800_000_000},
		{Name: "deepseek-r1-distill-qwen-7b.f16.gguf", Path: "/models/deepseek.gguf", Size: 14_000_000_000},
		{Name: "qwen2.5-7b-instruct-q3_k_m.gguf", Path: "/models/qwen.gguf", Size: 3_900_000_000},
		{
			Name: "llava-v1.6-mistral-7b.Q4_K_M.gguf", Path: "/models/llava.gguf",
			Size: 4_100_000_000, MMProjPaths: []string{"/models/mmproj.gguf"},
		},
	}

	a := NewApp(models)
	m, _ := a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app := m.(*App)
	content := app.View().Content

	checks := []struct{ name, want string }{
		{"version badge",   "v0.1.0"},
		{"title",           "llama-server-loader"},
		{"tab Models",      "Models"},
		{"tab Running",     "Running"},
		{"tab Logs",        "Logs"},
		{"counter 6/6",     "6 / 6"},
		{"dot indicator",   "●"},
		{"filter idle",     "поиск..."},
		{"footer ↑↓",       "↑↓"},
		{"footer ^q",       "^q"},
		{"rounded border ╭","╭"},
		{"rounded border ╰","╰"},
		{"scrollbar track", "│"},
	}

	t.Log("=== Main view checks ===")
	for _, c := range checks {
		if strings.Contains(content, c.want) {
			t.Logf("  ✓  %-22s %q", c.name, c.want)
		} else {
			t.Errorf("  ✗  %-22s MISSING %q", c.name, c.want)
		}
	}

	// Filter active
	m2, _ := app.Update(tea.KeyPressMsg{Text: "/"})
	app2 := m2.(*App)
	c2 := app2.View().Content
	t.Log("=== Filter active ===")
	if strings.Contains(c2, "▌") {
		t.Logf("  ✓  filter cursor ▌")
	} else {
		t.Errorf("  ✗  filter cursor ▌ MISSING")
	}

	// Help popup — свежий app чтобы filterState был Idle
	aHelp := NewApp(models)
	mHelp, _ := aHelp.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	appHelp := mHelp.(*App)
	mHelp2, _ := appHelp.Update(tea.KeyPressMsg{Text: "?"})
	appHelp2 := mHelp2.(*App)
	c3 := appHelp2.View().Content
	t.Log("=== Help popup ===")
	for _, want := range []string{"Справка", "навигация", "Esc или ?"} {
		if strings.Contains(c3, want) {
			t.Logf("  ✓  %q", want)
		} else {
			t.Errorf("  ✗  MISSING %q", want)
		}
	}

	// Tiny window — не паникует
	t.Log("=== Tiny window ===")
	aTiny := NewApp(models[:2])
	mTiny, _ := aTiny.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	appTiny := mTiny.(*App)
	_ = appTiny.View()
	t.Logf("  ✓  40×12 window renders without panic")
	_ = fmt.Sprint("done")
}
