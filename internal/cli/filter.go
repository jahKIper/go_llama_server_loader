package cli

import (
	"strings"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"llama-server-loader/pkg/modelscan"
)

// FilterState определяет состояние режима фильтрации.
type FilterState int

const (
	// FilterIdle — фильтр не активен, показываем бейдж [FILTER].
	FilterIdle FilterState = iota
	// FilterActive — поле ввода активно, но пустое, ждём ввод.
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
	styles  *StyleConfig
}

func newFilterInput(styles *StyleConfig) *filterInput {
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

// HandleKey обрабатывает нажатия клавиш в поле ввода фильтра.
// Cursor работает с байтовыми смещениями, но всегда на rune-границах (UTF-8 safe).
// Esc перехватывается на уровне App.Update, сюда не доходит.
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
		// Вставляем печатный rune на позицию курсора. UTF-8: kириллица, эмодзи и др.
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

// RenderFilterBadge рендерит бейдж FILTER (в режиме FilterIdle).
func (f *filterInput) RenderFilterBadge() string {
	if f.styles == nil {
		return "[FILTER]"
	}
	return f.styles.FilterBadgeStyle().Render("[FILTER]")
}

// RenderFilterInput рендерит поле ввода фильтра.
func (f *filterInput) RenderFilterInput() string {
	if f.styles == nil {
		cursorChar := "│"
		if f.cursor < len(f.text) {
			return f.text[:f.cursor] + cursorChar + f.text[f.cursor:]
		}
		return f.text + cursorChar
	}

	prompt := f.styles.FilterBadgeStyle().Render("[FILTER] ")
	cursorChar := "▌"

	var display string
	if f.cursor < len(f.text) {
		display = f.text[:f.cursor] + cursorChar + f.text[f.cursor:]
	} else {
		display = f.text + cursorChar
	}

	return prompt + f.styles.CountLabelStyle().Render(display)
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
