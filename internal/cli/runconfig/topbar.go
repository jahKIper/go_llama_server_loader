package runconfig

import (
	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/uistyle"
)

// Индексы табов — должны совпадать с internal/cli TabModels/TabParams/…
const (
	tabModels  = 0
	tabParams  = 1
	tabRunning = 2
	tabLogs    = 3
)

var topBarTabs = []string{"Models", "Params", "Running", "Logs"}

// enabledTabs — какие табы можно сделать активными визуально. Должно
// соответствовать NewTabBar в internal/cli/tabs.go.
var enabledTabs = []bool{true, true, false, false}

// renderTabBar рендерит ту же ленту табов, что используется на первом экране,
// с активным табом по индексу active.
func renderTabBar(st *uistyle.StyleConfig, active int) string {
	if st == nil {
		out := ""
		for i, lbl := range topBarTabs {
			if i > 0 {
				out += " │ "
			}
			if i == active {
				out += "[" + lbl + "]"
			} else {
				out += lbl
			}
		}
		return out
	}

	sep := st.TabSeparatorStyle().Render(" │ ")
	parts := make([]string, 0, len(topBarTabs)*2-1)
	for i, lbl := range topBarTabs {
		if i > 0 {
			parts = append(parts, sep)
		}
		var rendered string
		switch {
		case i == active:
			rendered = st.TabActiveStyle().Render(lbl)
		case !enabledTabs[i]:
			rendered = st.TabDisabledStyle().Render(lbl)
		default:
			rendered = st.TabInactiveStyle().Render(lbl)
		}
		parts = append(parts, rendered)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// RenderTopBar рендерит шапку второго экрана: version-бейдж слева,
// title по центру, tabs справа — визуально идентично первому экрану.
func RenderTopBar(title, version string, st *uistyle.StyleConfig, width int) string {
	if st == nil {
		return title
	}
	versionBadge := st.VersionBadgeStyle().Render("v" + version)
	badgeW := lipgloss.Width(versionBadge)

	tabs := renderTabBar(st, tabParams)
	tabsW := lipgloss.Width(tabs)

	available := width - badgeW - tabsW
	if available < 1 {
		return versionBadge + "  " + st.TitleStyle().Render(title)
	}
	titleCentered := st.TitleStyle().
		Width(available).
		Align(lipgloss.Center).
		Render(title)

	return lipgloss.JoinHorizontal(lipgloss.Top, versionBadge, titleCentered, tabs)
}
