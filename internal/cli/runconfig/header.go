package runconfig

import (
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/modelparams"
	"llama-server-loader/internal/cli/uistyle"
	"llama-server-loader/pkg/modelscan"
)

// RenderHeader рендерит верхний блок экрана: имя модели, путь, бэйджи.
// curated — курируемая выжимка из GGUF-параметров (может быть нулевой Curated{}).
func RenderHeader(m *modelscan.Model, st *uistyle.StyleConfig, innerW int, curated ...modelparams.Curated) string {
	if m == nil {
		return ""
	}
	var c modelparams.Curated
	if len(curated) > 0 {
		c = curated[0]
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

	// passport — компактная строка "паспорта" модели (arch · ctx · size · quant · sampling).
	passport := formatPassport(c, st)

	leftBlockParts := []string{nameStr, pathStr, badges}
	if passport != "" {
		leftBlockParts = append(leftBlockParts, passport)
	}
	leftBlock := lipgloss.JoinVertical(lipgloss.Left, leftBlockParts...)

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

// formatPassport собирает однострочную сводку «паспорта» модели для шапки
// (arch · ctx · size · quant · sampling). Поля, для которых нет значения, пропускаются.
// Возвращает пустую строку, если ни одного поля нет.
func formatPassport(c modelparams.Curated, st *uistyle.StyleConfig) string {
	var parts []string
	if c.Architecture != "" {
		parts = append(parts, c.Architecture)
	}
	if c.ContextLength > 0 {
		parts = append(parts, modelparams.FormatContext(c.ContextLength)+" ctx")
	}
	if c.SizeLabel != "" {
		parts = append(parts, c.SizeLabel)
	}
	sampling := formatPassportSampling(c)
	if sampling != "" {
		parts = append(parts, sampling)
	}
	if len(parts) == 0 {
		return ""
	}
	textStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextMuted))
	sepStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextSecondary))
	pieces := make([]string, 0, len(parts)*2-1)
	for i, p := range parts {
		if i > 0 {
			pieces = append(pieces, sepStyle.Render(" · "))
		}
		pieces = append(pieces, textStyle.Render(p))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, pieces...)
}

func formatPassportSampling(c modelparams.Curated) string {
	var parts []string
	if c.Temp > 0 {
		if c.Temp == float64(int64(c.Temp)) {
			parts = append(parts, fmt.Sprintf("temp %d", int64(c.Temp)))
		} else {
			parts = append(parts, fmt.Sprintf("temp %g", c.Temp))
		}
	}
	if c.TopP > 0 {
		parts = append(parts, fmt.Sprintf("top_p %g", c.TopP))
	}
	if c.TopK > 0 {
		parts = append(parts, fmt.Sprintf("top_k %d", c.TopK))
	}
	return strings.Join(parts, " · ")
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
