package runconfig

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"llama-server-loader/internal/cli/uistyle"
	"llama-server-loader/internal/config"
)

// keyMsg строит KeyPressMsg по строковому имени клавиши.
func keyMsg(key string) tea.KeyPressMsg {
	switch key {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEsc}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	default:
		if len(key) == 1 {
			return tea.KeyPressMsg{Code: rune(key[0]), Text: key}
		}
		return tea.KeyPressMsg{Text: key}
	}
}

func makeEntries() []CatalogEntry {
	return []CatalogEntry{
		{Category: "Основные", Meta: &config.ParamMeta{
			LongFlag:  "--ctx-size N",
			ShortFlag: "-c N",
			DescRU:    "Размер контекста модели",
		}},
		{Category: "Основные", Meta: &config.ParamMeta{
			LongFlag:  "--temp N",
			ShortFlag: "-t N",
			DescRU:    "Температура сэмплирования",
		}},
		{Category: "Мультимодальные", Meta: &config.ParamMeta{
			LongFlag:  "--mmproj FNAME",
			ShortFlag: "",
			DescRU:    "Путь к мультимодальному проектору",
		}},
		{Category: "Основные", Meta: &config.ParamMeta{
			LongFlag:  "--seed N",
			ShortFlag: "-s N",
			DescRU:    "Начальное значение генератора",
		}},
	}
}

func newTestPanel(entries []CatalogEntry) *RightPanel {
	return NewRightPanel(entries, uistyle.GetStyles(), 50, 20)
}

// ── Фильтрация ────────────────────────────────────────────────────────────────

func TestRightPanel_FilterByLong(t *testing.T) {
	p := newTestPanel(makeEntries())
	p.filterText = "ctx"
	p.applyFilter()

	if len(p.filtered) != 1 {
		t.Fatalf("expected 1 result for 'ctx', got %d", len(p.filtered))
	}
	if p.filtered[0].Meta.LongFlag != "--ctx-size N" {
		t.Errorf("unexpected entry: %v", p.filtered[0].Meta.LongFlag)
	}
}

func TestRightPanel_FilterByShort(t *testing.T) {
	p := newTestPanel(makeEntries())
	p.filterText = "-t"
	p.applyFilter()

	if len(p.filtered) != 1 {
		t.Fatalf("expected 1 result for '-t', got %d", len(p.filtered))
	}
	if p.filtered[0].Meta.LongFlag != "--temp N" {
		t.Errorf("unexpected entry: %v", p.filtered[0].Meta.LongFlag)
	}
}

func TestRightPanel_FilterByDesc(t *testing.T) {
	p := newTestPanel(makeEntries())
	p.filterText = "мультимодальн"
	p.applyFilter()

	if len(p.filtered) != 1 {
		t.Fatalf("expected 1 result for 'мультимодальн', got %d", len(p.filtered))
	}
	if p.filtered[0].Meta.LongFlag != "--mmproj FNAME" {
		t.Errorf("unexpected entry: %v", p.filtered[0].Meta.LongFlag)
	}
}

func TestRightPanel_FilterEmpty(t *testing.T) {
	p := newTestPanel(makeEntries())
	p.filterText = "ctx"
	p.applyFilter()
	p.filterText = ""
	p.applyFilter()

	if len(p.filtered) != len(makeEntries()) {
		t.Fatalf("expected all %d entries after clear, got %d", len(makeEntries()), len(p.filtered))
	}
}

func TestRightPanel_FilterNoMatch(t *testing.T) {
	p := newTestPanel(makeEntries())
	p.filterText = "xyznonexistent"
	p.applyFilter()

	if len(p.filtered) != 0 {
		t.Fatalf("expected 0 results, got %d", len(p.filtered))
	}
}

func TestRightPanel_FilterCaseInsensitive(t *testing.T) {
	p := newTestPanel(makeEntries())
	p.filterText = "CTX"
	p.applyFilter()

	if len(p.filtered) != 1 {
		t.Fatalf("expected 1 result for 'CTX', got %d", len(p.filtered))
	}
}

// ── Selected / навигация ──────────────────────────────────────────────────────

func TestRightPanel_Selected_Empty(t *testing.T) {
	p := newTestPanel(nil)
	if p.Selected() != nil {
		t.Error("expected nil Selected on empty panel")
	}
}

func TestRightPanel_Selected_Default(t *testing.T) {
	p := newTestPanel(makeEntries())
	sel := p.Selected()
	if sel == nil {
		t.Fatal("expected non-nil Selected")
	}
	if sel.Meta.LongFlag != "--ctx-size N" {
		t.Errorf("expected first entry, got %v", sel.Meta.LongFlag)
	}
}

func TestRightPanel_NavigationDown(t *testing.T) {
	p := newTestPanel(makeEntries())
	p.Update(tea.KeyPressMsg{Code: tea.KeyDown}, true)
	sel := p.Selected()
	if sel == nil || sel.Meta.LongFlag != "--temp N" {
		t.Errorf("expected --temp N after down, got %v", sel)
	}
}

func TestRightPanel_NavigationBounds(t *testing.T) {
	p := newTestPanel(makeEntries())
	// Нажимаем вверх когда на первом — курсор остаётся
	p.Update(tea.KeyPressMsg{Code: tea.KeyUp}, true)
	if p.cursor != 0 {
		t.Errorf("cursor should stay 0 on up at first, got %d", p.cursor)
	}
	// Навигируем до конца и ещё раз вниз
	for i := 0; i < len(makeEntries())+5; i++ {
		p.Update(tea.KeyPressMsg{Code: tea.KeyDown}, true)
	}
	if p.cursor != len(makeEntries())-1 {
		t.Errorf("cursor should be at last %d, got %d", len(makeEntries())-1, p.cursor)
	}
}

// ── Фильтр-режим ─────────────────────────────────────────────────────────────

func TestRightPanel_FilterActivate(t *testing.T) {
	p := newTestPanel(makeEntries())
	if p.IsFilterActive() {
		t.Fatal("filter should not be active initially")
	}
	p.Update(keyMsg("/"), true)
	if !p.IsFilterActive() {
		t.Error("filter should be active after '/'")
	}
}

func TestRightPanel_FilterDeactivateEsc(t *testing.T) {
	p := newTestPanel(makeEntries())
	p.Update(keyMsg("/"), true)
	p.Update(keyMsg("esc"), true)
	if p.IsFilterActive() {
		t.Error("filter should be inactive after Esc")
	}
}

func TestRightPanel_FilterTyping(t *testing.T) {
	p := newTestPanel(makeEntries())
	p.Update(keyMsg("/"), true)
	for _, ch := range "ctx" {
		p.Update(keyMsg(string(ch)), true)
	}
	if p.filterText != "ctx" {
		t.Errorf("expected filterText 'ctx', got %q", p.filterText)
	}
	if len(p.filtered) != 1 {
		t.Errorf("expected 1 filtered entry, got %d", len(p.filtered))
	}
}

// ── Не активна при focused=false ──────────────────────────────────────────────

func TestRightPanel_NotFocused(t *testing.T) {
	p := newTestPanel(makeEntries())
	p.Update(tea.KeyPressMsg{Code: tea.KeyDown}, false)
	if p.cursor != 0 {
		t.Error("cursor should not move when panel is not focused")
	}
}

// ── FilterValue через applyFilter охватывает long и desc ─────────────────────

func TestRightPanel_FilterMultiMatch(t *testing.T) {
	p := newTestPanel(makeEntries())
	// "н" встречается в описании "Начальное значение" и "Температура сэмплирования" и "Размер контекста"
	p.filterText = "начальное"
	p.applyFilter()
	if len(p.filtered) != 1 {
		t.Fatalf("expected 1 match for 'начальное', got %d: %v", len(p.filtered), p.filtered)
	}
	if p.filtered[0].Meta.LongFlag != "--seed N" {
		t.Errorf("unexpected match: %v", p.filtered[0].Meta.LongFlag)
	}
}
