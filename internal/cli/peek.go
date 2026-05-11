package cli

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/modelparams"
	"llama-server-loader/internal/cli/uistyle"
)

// peekPanelHeight — фиксированная высота peek-блока в строках (вкл. рамку).
const peekPanelHeight = 9

// renderPeekPanel рендерит inline-peek с курируемой сводкой GGUF-параметров.
// Если params == nil или у модели нет записи в models.json — возвращает короткую заглушку.
func renderPeekPanel(c modelparams.Curated, st *uistyle.StyleConfig, w int) string {
	if w < 30 {
		return ""
	}

	innerW := w - 2 // рамка
	if innerW < 10 {
		innerW = 10
	}

	rowFill := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Width(innerW)
	keyStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextMuted))
	valStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextSecondary))
	hintStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextMuted))

	// Двухколоночные пары "ключ — значение"
	colW := innerW / 2
	if colW < 16 {
		colW = innerW
	}

	renderCell := func(k, v string) string {
		if v == "" {
			v = "—"
		}
		key := keyStyle.Render(k)
		val := valStyle.Render(v)
		gap := colW - lipgloss.Width(key) - lipgloss.Width(val)
		if gap < 1 {
			gap = 1
		}
		spacer := lipgloss.NewStyle().
			Background(lipgloss.Color(st.BgPanel)).
			Render(strings.Repeat(" ", gap))
		return lipgloss.JoinHorizontal(lipgloss.Top, key, spacer, val)
	}

	row := func(a, b string) string {
		if colW == innerW {
			return rowFill.Render(a)
		}
		return rowFill.Render(lipgloss.JoinHorizontal(lipgloss.Top, a, b))
	}

	emptyCell := strings.Repeat(" ", colW)

	pairs := [][2]string{
		{
			renderCell("Архитектура", emptyOr(c.Architecture, "—")),
			renderCell("Контекст", formatInt64(c.ContextLength)),
		},
		{
			renderCell("Размер", emptyOr(c.SizeLabel, "—")),
			renderCell("Блоков", formatInt64(c.BlockCount)),
		},
		{
			renderCell("Embedding", formatInt64(c.EmbeddingLength)),
			renderCell("FFN", formatInt64(c.FFNLength)),
		},
		{
			renderCell("Heads (Q/KV)", formatHeads(c.HeadCount, c.HeadCountKV)),
			renderCell("Tokenizer", emptyOr(c.TokenizerModel, "—")),
		},
		{
			renderCell("Sampling", formatSampling(c)),
			renderCell("Chat template", boolMark(c.HasChatTemplate)),
		},
	}
	_ = emptyCell

	hint := fmt.Sprintf("Enter → все %d ✎", c.TotalCount)
	hintLine := rowFill.Render(hintStyle.Align(lipgloss.Right).Width(innerW).Render(hint))

	var lines []string
	for _, p := range pairs {
		lines = append(lines, row(p[0], p[1]))
	}
	lines = append(lines, hintLine)

	body := lipgloss.JoinVertical(lipgloss.Left, lines...)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(st.AccentPurpleMuted)).
		BorderBackground(lipgloss.Color(st.BgPanel)).
		Background(lipgloss.Color(st.BgPanel)).
		Width(innerW).
		Render(body)

	return injectBorderTitlePeek(box, "Параметры модели", st)
}

// renderPeekEmpty — заглушка для модели без params (peek открыт, данных нет).
func renderPeekEmpty(st *uistyle.StyleConfig, w int) string {
	if w < 30 {
		return ""
	}
	innerW := w - 2
	if innerW < 10 {
		innerW = 10
	}
	msg := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextMuted)).
		Width(innerW).
		Render("GGUF-параметры не загружены. Нажмите Enter — после открытия модели они запишутся в models.json.")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(st.BorderIdle)).
		BorderBackground(lipgloss.Color(st.BgPanel)).
		Background(lipgloss.Color(st.BgPanel)).
		Width(innerW).
		Render(msg)
}

func emptyOr(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

func formatInt64(n int64) string {
	if n <= 0 {
		return "—"
	}
	return fmt.Sprintf("%d", n)
}

func formatHeads(q, kv int64) string {
	if q <= 0 && kv <= 0 {
		return "—"
	}
	return fmt.Sprintf("%d / %d", q, kv)
}

func formatSampling(c modelparams.Curated) string {
	parts := []string{}
	if c.Temp > 0 {
		parts = append(parts, fmt.Sprintf("temp %s", trimFloat(c.Temp)))
	}
	if c.TopP > 0 {
		parts = append(parts, fmt.Sprintf("top_p %s", trimFloat(c.TopP)))
	}
	if c.TopK > 0 {
		parts = append(parts, fmt.Sprintf("top_k %d", c.TopK))
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, " · ")
}

func trimFloat(f float64) string {
	if f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}
	return fmt.Sprintf("%g", f)
}

func boolMark(b bool) string {
	if b {
		return "✓"
	}
	return "—"
}

// injectBorderTitlePeek встраивает заголовок «Параметры модели» в верхнюю границу.
// Используем существующий helper из этого пакета, если он есть, иначе fallback.
func injectBorderTitlePeek(rendered, title string, st *uistyle.StyleConfig) string {
	// Используем общий helper InjectBorderTitle (без счётчика).
	return InjectBorderTitle(rendered, title, "")
}
