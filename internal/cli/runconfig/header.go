package runconfig

import (
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/uistyle"
	"llama-server-loader/pkg/modelscan"
)

// RenderHeader рендерит верхний блок экрана: имя модели, путь, бэйджи.
func RenderHeader(m *modelscan.Model, st *uistyle.StyleConfig, innerW int) string {
	if m == nil {
		return ""
	}

	name := strings.TrimSuffix(filepath.Base(m.Path), ".gguf")

	nameStr := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(st.NeonGreen)).
		Background(lipgloss.Color(st.BgPanel)).
		Render(name)

	// innerW минус padding блока (2*2=4) и рамка (2)
	pathMaxW := innerW - 8
	if pathMaxW < 1 {
		pathMaxW = 1
	}
	pathStr := lipgloss.NewStyle().
		Foreground(lipgloss.Color(st.TextMuted)).
		Background(lipgloss.Color(st.BgPanel)).
		Render(truncatePath(m.Path, pathMaxW))

	sizeBadge := st.SizeBadgeStyle(st.BgPanel).Render(formatSizeLocal(m.Size))

	var mmprojBadge string
	if len(m.MMProjPaths) > 0 {
		mmprojBadge = st.MMProjBadgeStyle(st.BgPanel).Render("mmproj")
	} else {
		mmprojBadge = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), true, true, true, true).
			BorderForeground(lipgloss.Color(st.TextMuted)).
			BorderBackground(lipgloss.Color(st.BgPanel)).
			Background(lipgloss.Color(st.BgPanel)).
			Foreground(lipgloss.Color(st.TextMuted)).
			Render("no-mmproj")
	}

	ggufBadge := lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(st.TextSecondary)).
		BorderBackground(lipgloss.Color(st.BgPanel)).
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextSecondary)).
		Render("gguf")

	badges := lipgloss.JoinHorizontal(lipgloss.Center, sizeBadge, " ", mmprojBadge, " ", ggufBadge)

	leftBlock := lipgloss.JoinVertical(lipgloss.Left,
		nameStr,
		pathStr,
		badges,
	)

	runBtn := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(st.NeonGreen)).
		Foreground(lipgloss.Color(st.DarkBg)).
		Padding(0, 2).
		Render("▶ ЗАПУСК (r)")
	btnW := lipgloss.Width(runBtn)
	bgCell := lipgloss.NewStyle().Background(lipgloss.Color(st.BgPanel))
	emptyBtnLine := bgCell.Width(btnW).Render("")
	btnBlock := lipgloss.JoinVertical(lipgloss.Left, emptyBtnLine, runBtn, emptyBtnLine)

	// innerW уже минус padding(2,2)=4 и рамку(2). leftBlock + gap + btn должны влезть.
	innerContentW := innerW - 6
	leftW := lipgloss.Width(leftBlock)
	gapW := innerContentW - leftW - btnW
	if gapW < 1 {
		gapW = 1
	}
	gapStrip := bgCell.Width(gapW).Render("")
	gapBlock := lipgloss.JoinVertical(lipgloss.Left, gapStrip, gapStrip, gapStrip)

	block := lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, gapBlock, btnBlock)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(st.AccentPurple)).
		BorderBackground(lipgloss.Color(st.BgPanel)).
		Background(lipgloss.Color(st.BgPanel)).
		Padding(0, 2).
		Width(innerW).
		Render(block)
}

// formatSizeLocal форматирует размер файла в человекочитаемый вид.
func formatSizeLocal(size int64) string {
	switch {
	case size < 1024:
		return fmt.Sprintf("%dB", size)
	case size < 1024*1024:
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	case size < 1024*1024*1024:
		return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
	default:
		return fmt.Sprintf("%.1fGB", float64(size)/(1024*1024*1024))
	}
}
