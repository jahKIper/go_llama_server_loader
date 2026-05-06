package cli

import "charm.land/lipgloss/v2"

// Tab описывает один таб в панели табов.
type Tab struct {
	Label   string
	Enabled bool
}

// TabBar — панель табов для header правой части.
type TabBar struct {
	tabs   []Tab
	active int
	styles *StyleConfig
}

// NewTabBar создаёт TabBar с тремя табами: Models (enabled), Running, Logs (disabled).
func NewTabBar(st *StyleConfig) *TabBar {
	return &TabBar{
		tabs: []Tab{
			{Label: "Models", Enabled: true},
			{Label: "Running", Enabled: false},
			{Label: "Logs", Enabled: false},
		},
		active: 0,
		styles: st,
	}
}

// SetActive устанавливает активный таб по индексу.
func (tb *TabBar) SetActive(i int) {
	if i >= 0 && i < len(tb.tabs) {
		tb.active = i
	}
}

// Next переключает на следующий enabled таб. Если enabled только один — no-op.
func (tb *TabBar) Next() {
	for step := 1; step < len(tb.tabs); step++ {
		idx := (tb.active + step) % len(tb.tabs)
		if tb.tabs[idx].Enabled {
			tb.active = idx
			return
		}
	}
}

// Prev переключает на предыдущий enabled таб. Если enabled только один — no-op.
func (tb *TabBar) Prev() {
	for step := 1; step < len(tb.tabs); step++ {
		idx := (tb.active - step + len(tb.tabs)) % len(tb.tabs)
		if tb.tabs[idx].Enabled {
			tb.active = idx
			return
		}
	}
}

// Active возвращает индекс активного таба.
func (tb *TabBar) Active() int {
	return tb.active
}

// Render возвращает отрендеренную строку табов.
func (tb *TabBar) Render() string {
	if tb.styles == nil {
		return tb.renderFallback()
	}

	sep := tb.styles.TabSeparatorStyle().Render(" │ ")
	parts := make([]string, 0, len(tb.tabs)*2-1)

	for i, tab := range tb.tabs {
		if i > 0 {
			parts = append(parts, sep)
		}
		var rendered string
		switch {
		case i == tb.active:
			rendered = tb.styles.TabActiveStyle().Render(tab.Label)
		case !tab.Enabled:
			rendered = tb.styles.TabDisabledStyle().Render(tab.Label)
		default:
			rendered = tb.styles.TabInactiveStyle().Render(tab.Label)
		}
		parts = append(parts, rendered)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func (tb *TabBar) renderFallback() string {
	result := ""
	for i, tab := range tb.tabs {
		if i > 0 {
			result += " │ "
		}
		if i == tb.active {
			result += "[" + tab.Label + "]"
		} else {
			result += tab.Label
		}
	}
	return result
}
