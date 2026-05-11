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

	// Фон header'а — DarkBg (как у unselected-карточки модели на первом экране).
	headerBg := st.DarkBg

	nameStr := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(st.NeonGreen)).
		Background(lipgloss.Color(headerBg)).
		Render(name)

	// innerW минус padding блока (2*2=4) и рамка (2)
	pathMaxW := innerW - 8
	if pathMaxW < 1 {
		pathMaxW = 1
	}
	pathStr := lipgloss.NewStyle().
		Foreground(lipgloss.Color(st.TextMuted)).
		Background(lipgloss.Color(headerBg)).
		Render(truncatePath(m.Path, pathMaxW))

	sizeBadge := st.SizeBadgeStyle(headerBg).Render(formatSizeLocal(m.Size))

	// Бэйджи трёхстрочные (border сверху/середина/снизу) — делаем разделитель тоже
	// трёхстрочным с фоном, иначе по краям просвечивает фон терминала.
	badgeSepCell := lipgloss.NewStyle().Background(lipgloss.Color(headerBg)).Render(" ")
	badgeSep := badgeSepCell + "\n" + badgeSepCell + "\n" + badgeSepCell

	var mmprojBadge string
	if len(m.MMProjPaths) > 0 {
		mmprojBadge = st.MMProjBadgeStyle(headerBg).Render("mmproj")
	} else {
		mmprojBadge = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), true, true, true, true).
			BorderForeground(lipgloss.Color(st.TextMuted)).
			BorderBackground(lipgloss.Color(headerBg)).
			Background(lipgloss.Color(headerBg)).
			Foreground(lipgloss.Color(st.TextMuted)).
			Render("no-mmproj")
	}

	ggufBadge := lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(st.TextSecondary)).
		BorderBackground(lipgloss.Color(headerBg)).
		Background(lipgloss.Color(headerBg)).
		Foreground(lipgloss.Color(st.TextSecondary)).
		Render("gguf")

	badges := lipgloss.JoinHorizontal(lipgloss.Center, sizeBadge, badgeSep, mmprojBadge, badgeSep, ggufBadge)

	// passport — компактная строка "паспорта" модели (arch · ctx · size · quant · sampling).
	passport := formatPassport(c, st, headerBg)

	runBtn := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(st.NeonGreen)).
		Foreground(lipgloss.Color(st.DarkBg)).
		Padding(0, 2).
		Render("▶ ЗАПУСК (r)")
	btnW := lipgloss.Width(runBtn)
	bgCell := lipgloss.NewStyle().Background(lipgloss.Color(headerBg))
	emptyBtnLine := bgCell.Width(btnW).Render("")

	// innerW уже минус padding(2,2)=4 и рамку(2). leftBlock + gap=1 + btn должны влезть.
	innerContentW := innerW - 6
	const gapW = 1
	leftBlockW := innerContentW - btnW - gapW
	if leftBlockW < 10 {
		leftBlockW = 10
	}

	// rowFill — каждая строка leftBlock проходит через стиль с Width и Background.
	// Это гарантирует, что свободные колонки заполнены фоном headerBg, а не пробелами
	// без стиля (через которые просвечивает фон терминала).
	rowFill := lipgloss.NewStyle().
		Background(lipgloss.Color(headerBg)).
		Width(leftBlockW)

	leftRowParts := []string{
		rowFill.Render(nameStr),
		rowFill.Render(pathStr),
		rowFill.Render(badges),
	}
	if passport != "" {
		// badges — многострочный (border'ы у бэйджей дают 3 строки).
		// Каждую визуальную строку badges уже обернули rowFill'ом выше — этого мало,
		// т.к. rowFill применяется к строке целиком, lipgloss может не покрыть
		// все вертикальные сегменты. Прогоняем построчно ниже.
	}
	// Развернём каждый блок (некоторые из них многострочные) построчно, чтобы
	// каждая визуальная строка прошла через rowFill.
	expanded := make([]string, 0, 6)
	for _, part := range leftRowParts {
		for _, ln := range strings.Split(part, "\n") {
			expanded = append(expanded, rowFill.Render(ln))
		}
	}
	if passport != "" {
		expanded = append(expanded, rowFill.Render(passport))
	}
	leftBlock := lipgloss.JoinVertical(lipgloss.Left, expanded...)
	leftRows := len(expanded)

	// Выравниваем btnBlock по высоте leftBlock, добиваем пустыми строками.
	btnParts := make([]string, leftRows)
	for i := 0; i < leftRows; i++ {
		btnParts[i] = emptyBtnLine
	}
	if leftRows >= 2 {
		btnParts[1] = runBtn
	} else {
		btnParts[0] = runBtn
	}
	btnBlock := lipgloss.JoinVertical(lipgloss.Left, btnParts...)

	gapStrip := bgCell.Width(gapW).Render("")
	gapParts := make([]string, leftRows)
	for i := 0; i < leftRows; i++ {
		gapParts[i] = gapStrip
	}
	gapBlock := lipgloss.JoinVertical(lipgloss.Left, gapParts...)

	block := lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, gapBlock, btnBlock)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(st.AccentPurple)).
		BorderBackground(lipgloss.Color(headerBg)).
		Background(lipgloss.Color(headerBg)).
		Padding(0, 2).
		Width(innerW).
		Render(block)
}

// formatPassport собирает однострочную сводку «паспорта» модели для шапки
// (arch · ctx · size · quant · sampling). Поля, для которых нет значения, пропускаются.
// Возвращает пустую строку, если ни одного поля нет.
func formatPassport(c modelparams.Curated, st *uistyle.StyleConfig, bg string) string {
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
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(st.TextMuted))
	sepStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
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
