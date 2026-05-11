package cli

import (
	"strings"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/uistyle"
	"llama-server-loader/pkg/modelscan"
)

// FilterState определяет состояние режима фильтрации.
type FilterState int

const (
	// FilterIdle — фильтр не активен.
	FilterIdle FilterState = iota
	// FilterActive — поле ввода активно, но пустое.
	FilterActive
	// Filtering — пользователь вводит текст, фильтрация применена.
	Filtering
)

func (f FilterState) String() string {
	switch f {
	case FilterIdle:
		return "idle"
	case FilterActive:
		return "active"
	case Filtering:
		return "filtering"
	default:
		return "unknown"
	}
}

// filterInput представляет поле ввода фильтра в TUI.
type filterInput struct {
	visible bool
	state   FilterState
	text    string
	cursor  int
	styles  *uistyle.StyleConfig
}

func newFilterInput(styles *uistyle.StyleConfig) *filterInput {
	return &filterInput{
		visible: false,
		state:   FilterIdle,
		styles:  styles,
		cursor:  0,
	}
}

// Toggle переключает состояние фильтра.
func (f *filterInput) Toggle() {
	switch f.state {
	case FilterIdle:
		f.visible = true
		f.state = FilterActive
	case FilterActive, Filtering:
		if f.text == "" || f.state == Filtering {
			f.text = ""
			f.cursor = 0
		}
		f.visible = false
		f.state = FilterIdle
	default:
		f.visible = !f.visible
		if f.visible && f.state == FilterIdle {
			f.state = FilterActive
		}
	}
}

// Text возвращает текущий текст фильтра.
func (f *filterInput) Text() string {
	return f.text
}

// Clear сбрасывает текст и позицию курсора.
func (f *filterInput) Clear() {
	f.text = ""
	f.cursor = 0
}

// HandleKey обрабатывает нажатия клавиш в поле ввода фильтра.
func (f *filterInput) HandleKey(key string) tea.Cmd {
	switch key {
	case "backspace", "ctrl+h":
		if f.cursor > 0 {
			r, size := utf8.DecodeLastRuneInString(f.text[:f.cursor])
			_ = r
			f.text = f.text[:f.cursor-size] + f.text[f.cursor:]
			f.cursor -= size
		}
	case "delete", "ctrl+d":
		if f.cursor < len(f.text) {
			_, size := utf8.DecodeRuneInString(f.text[f.cursor:])
			f.text = f.text[:f.cursor] + f.text[f.cursor+size:]
		}
	case "left":
		if f.cursor > 0 {
			_, size := utf8.DecodeLastRuneInString(f.text[:f.cursor])
			f.cursor -= size
		}
	case "right":
		if f.cursor < len(f.text) {
			_, size := utf8.DecodeRuneInString(f.text[f.cursor:])
			f.cursor += size
		}
	case "home":
		f.cursor = 0
	case "end":
		f.cursor = len(f.text)
	default:
		if utf8.RuneCountInString(key) == 1 {
			r, _ := utf8.DecodeRuneInString(key)
			if unicode.IsPrint(r) {
				f.text = f.text[:f.cursor] + key + f.text[f.cursor:]
				f.cursor += len(key)
			}
		}
	}
	return nil
}

// Render — единая точка рендера: возвращает idle или active вид по состоянию.
func (f *filterInput) Render(blockWidth int) string {
	if f.state == FilterIdle {
		return f.RenderFilterIdle(blockWidth)
	}
	return f.RenderFilterActive(blockWidth)
}

// RenderFilterIdle рендерит поле в idle-состоянии.
func (f *filterInput) RenderFilterIdle(blockWidth int) string {
	if f.styles == nil {
		return "[/ поиск...]"
	}
	st := f.styles.FilterInputIdleStyle()
	if blockWidth > 0 {
		st = st.Width(blockWidth)
	}
	placeholder := lipgloss.NewStyle().
		Background(lipgloss.Color(f.styles.BgPanel)).
		Foreground(lipgloss.Color(f.styles.TextMuted)).
		Render("поиск...")
	return st.Render("/  " + placeholder)
}

// RenderFilterActive рендерит поле в активном состоянии.
func (f *filterInput) RenderFilterActive(blockWidth int) string {
	if f.styles == nil {
		cursorChar := "│"
		if f.cursor < len(f.text) {
			return "/  " + f.text[:f.cursor] + cursorChar + f.text[f.cursor:]
		}
		return "/  " + f.text + cursorChar
	}

	cursor := lipgloss.NewStyle().
		Foreground(lipgloss.Color(f.styles.NeonGreen)).
		Render("▌")

	var display string
	if f.cursor < len(f.text) {
		display = f.text[:f.cursor] + cursor + f.text[f.cursor:]
	} else {
		display = f.text + cursor
	}

	st := f.styles.FilterInputActiveStyle()
	if blockWidth > 0 {
		st = st.Width(blockWidth)
	}
	return st.Render("/  " + display)
}

// RenderFilterBadge — backward-compat обёртка над RenderFilterIdle(0).
func (f *filterInput) RenderFilterBadge() string {
	return f.RenderFilterIdle(0)
}

// RenderFilterInput — backward-compat обёртка над RenderFilterActive(0).
func (f *filterInput) RenderFilterInput() string {
	return f.RenderFilterActive(0)
}

// FilterModels применяет текстовый фильтр к списку моделей.
func FilterModels(models []*modelscan.Model, text string) []*modelscan.Model {
	if text == "" {
		return models
	}

	lowerText := strings.ToLower(text)
	result := make([]*modelscan.Model, 0)

	for _, m := range models {
		if strings.Contains(strings.ToLower(m.Name), lowerText) ||
			strings.Contains(strings.ToLower(m.Path), lowerText) {
			result = append(result, m)
		}
	}

	return result
}
