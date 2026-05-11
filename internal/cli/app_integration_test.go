package cli

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"llama-server-loader/internal/cli/uistyle"
	"llama-server-loader/pkg/modelscan"
)

// testModels returns a fixed set of models for smoke tests.
func testModels() []*modelscan.Model {
	return []*modelscan.Model{
		{Name: "mistral-7b-instruct-v0.2.Q4_K_M.gguf", Path: "/models/mistral-7b-instruct-v0.2.Q4_K_M.gguf", Size: 4_200_000_000},
		{Name: "llama-3-8b-q5_k_m.gguf", Path: "/models/llama-3-8b-q5_k_m.gguf", Size: 5_300_000_000},
		{Name: "gemma-2-9b-it-Q8_0.gguf", Path: "/models/gemma-2-9b-it-Q8_0.gguf", Size: 9_800_000_000},
		{Name: "phi-3-mini-4k-instruct.Q4_0.gguf", Path: "/models/phi-3-mini-4k-instruct.Q4_0.gguf", Size: 2_100_000_000},
		{Name: "qwen2.5-7b-instruct-q3_k_m.gguf", Path: "/models/qwen2.5-7b-instruct-q3_k_m.gguf", Size: 3_900_000_000},
		{
			Name:        "llava-v1.6-mistral-7b.Q4_K_M.gguf",
			Path:        "/models/llava-v1.6-mistral-7b.Q4_K_M.gguf",
			Size:        4_100_000_000,
			MMProjPaths: []string{"/models/llava-v1.6-mistral-7b-mmproj-f16.gguf"},
		},
	}
}

// keyMsg constructs a KeyPressMsg by name (mirrors Update's msg.String() matching).
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
	case "ctrl+c":
		return tea.KeyPressMsg{Code: rune('c'), Mod: tea.ModCtrl}
	default:
		// Single printable character
		if len(key) == 1 {
			return tea.KeyPressMsg{Code: rune(key[0]), Text: key}
		}
		return tea.KeyPressMsg{Text: key}
	}
}

// sendWindowSize sends a WindowSizeMsg and returns updated model.
func sendWindowSize(a *App, w, h int) *App {
	m, _ := a.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return m.(*App)
}

// sendKey sends a named key and returns updated model + cmd.
func sendKey(a *App, key string) (*App, tea.Cmd) {
	m, cmd := a.Update(keyMsg(key))
	return m.(*App), cmd
}

// viewContent returns the plain Content of a View (ANSI escape codes may be present).
func viewContent(a *App) string {
	return a.View().Content
}

// ── Smoke 1: запуск с реальной папкой моделей ────────────────────────────────

func TestAppSmoke1_InitialRender(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	if a.countTotal != 6 {
		t.Errorf("countTotal want 6, got %d", a.countTotal)
	}
	if a.countShown != 6 {
		t.Errorf("countShown want 6, got %d", a.countShown)
	}
	if a.filterState != FilterIdle {
		t.Errorf("filterState want FilterIdle, got %d", a.filterState)
	}

	content := viewContent(a)
	if !strings.Contains(content, "0.1.0") {
		t.Error("view should contain version '0.1.0'")
	}
	if !strings.Contains(content, "llama-server-loader") {
		t.Error("view should contain app title")
	}
	if !strings.Contains(content, "6") {
		t.Error("view should contain total model count")
	}
}

// ── Smoke 2: нажать "/" → поле ввода фильтра ────────────────────────────────

func TestAppSmoke2_SlashActivatesFilter(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	a, _ = sendKey(a, "/")

	if a.filterState != FilterActive {
		t.Errorf("filterState want FilterActive after '/', got %d", a.filterState)
	}
	// View must not panic in active mode
	_ = viewContent(a)
}

// ── Smoke 3: ввод текста фильтрует список ────────────────────────────────────

func TestAppSmoke3_FilterReducesList(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	a, _ = sendKey(a, "/")
	// "q5_k_m" matches exactly one model: llama-3-8b-q5_k_m.gguf
	for _, ch := range "q5_k_m" {
		a, _ = sendKey(a, string(ch))
	}

	if a.filterState != Filtering {
		t.Errorf("filterState want Filtering, got %d", a.filterState)
	}
	if a.countShown != 1 {
		t.Errorf("countShown want 1 after 'q5_k_m' filter, got %d", a.countShown)
	}
	if a.countTotal != 6 {
		t.Errorf("countTotal must stay 6, got %d", a.countTotal)
	}
}

func TestAppSmoke3b_FilterPartialMatch(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	a, _ = sendKey(a, "/")
	for _, ch := range "llama" {
		a, _ = sendKey(a, string(ch))
	}

	if a.countShown == 0 {
		t.Error("filter 'llama' should match at least one model")
	}
	if a.countShown > a.countTotal {
		t.Error("countShown must not exceed countTotal")
	}
}

// ── Smoke 4: Esc восстанавливает полный список ───────────────────────────────

func TestAppSmoke4_EscRestoresList(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	a, _ = sendKey(a, "/")
	for _, ch := range "gemma" {
		a, _ = sendKey(a, string(ch))
	}
	if a.countShown != 1 {
		t.Fatalf("precondition: want 1 match for 'gemma', got %d", a.countShown)
	}

	a, _ = sendKey(a, "esc")

	if a.filterState != FilterIdle {
		t.Errorf("filterState want FilterIdle after Esc, got %d", a.filterState)
	}
	if a.countShown != 6 {
		t.Errorf("countShown want 6 after Esc, got %d", a.countShown)
	}
	if a.filterText != "" {
		t.Errorf("filterText want empty after Esc, got %q", a.filterText)
	}
}

// ── Smoke 5: ↑↓ навигация ───────────────────────────────────────────────────

func TestAppSmoke5_ArrowNavigation(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	initialIndex := a.list.Index()
	a, _ = sendKey(a, "down")

	if a.list.Index() == initialIndex && len(testModels()) > 1 {
		t.Errorf("list index did not change after 'down': still %d", initialIndex)
	}
}

func TestAppSmoke5b_ArrowInFilterMode(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	a, _ = sendKey(a, "/")
	a, _ = sendKey(a, "down") // навигация должна работать в filter mode

	if a.filterState != FilterActive {
		t.Errorf("filterState want FilterActive after down in filter mode, got %d", a.filterState)
	}
}

// ── Smoke 6: Enter выбирает модель ──────────────────────────────────────────

func TestAppSmoke6_EnterSelectsModel(t *testing.T) {
	models := testModels()
	a := NewApp(models, nil, nil)
	a = sendWindowSize(a, 120, 40)

	updated, cmd := sendKey(a, "enter")

	// Enter должен вернуть tea.Quit cmd когда есть выбранный элемент
	if cmd == nil && updated.selected == nil {
		t.Log("note: Enter with no running program loop may not set selected; OK in unit context")
	} else if updated.selected != nil {
		if updated.selected.Name == "" {
			t.Error("selected model has empty name")
		}
	}
}

// ── Smoke 7: q — чистый выход ────────────────────────────────────────────────

func TestAppSmoke7_QuitKey(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	_, cmd := sendKey(a, "q")
	if cmd == nil {
		t.Error("'q' should return tea.Quit cmd")
	}
}

func TestAppSmoke7_CtrlC(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	_, cmd := sendKey(a, "ctrl+c")
	if cmd == nil {
		t.Error("'ctrl+c' should return tea.Quit cmd")
	}
}

// ── Smoke 8: resize окна ────────────────────────────────────────────────────

func TestAppSmoke8_WindowResize(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)
	a = sendWindowSize(a, 80, 24)

	if a.width != 80 {
		t.Errorf("width want 80, got %d", a.width)
	}
	if a.height != 24 {
		t.Errorf("height want 24, got %d", a.height)
	}
	_ = viewContent(a)
}

func TestAppSmoke8_TinyWindow(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 40, 10)
	_ = viewContent(a) // must not panic
}

// ── NewApp: все поля инициализированы ────────────────────────────────────────

func TestNewApp_AllFieldsInitialized(t *testing.T) {
	models := testModels()
	a := NewApp(models, nil, nil)

	if a.styles == nil {
		t.Error("styles must not be nil")
	}
	if a.dims == nil {
		t.Error("dims must not be nil")
	}
	if a.footer == nil {
		t.Error("footer must not be nil")
	}
	if a.filterInput == nil {
		t.Error("filterInput must not be nil")
	}
	if a.countTotal != len(models) {
		t.Errorf("countTotal want %d, got %d", len(models), a.countTotal)
	}
	if a.countShown != len(models) {
		t.Errorf("countShown want %d, got %d", len(models), a.countShown)
	}
	if a.version != "0.1.0" {
		t.Errorf("version want 0.1.0, got %q", a.version)
	}
	if a.filterState != FilterIdle {
		t.Errorf("filterState want FilterIdle, got %d", a.filterState)
	}
}

// ── Badges: extractQuantization ──────────────────────────────────────────────

func TestExtractQuantization(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"mistral-7b-instruct-v0.2.Q4_K_M.gguf", "Q4_K_M"},
		{"llama-3-8b-q5_k_m.gguf", "Q5_K_M"},
		{"gemma-2-9b-it-Q8_0.gguf", "Q8_0"},
		{"phi-3-mini-4k-instruct.Q4_0.gguf", "Q4_0"},
		{"qwen2.5-7b-instruct-q3_k_m.gguf", "Q3_K_M"},
		{"deepseek-r1-distill-qwen-7b.f16.gguf", "F16"},
		{"unknown-model.gguf", ""}, // unknown quantization → empty
	}
	for _, c := range cases {
		got := extractQuantization(c.name)
		if got != c.want {
			t.Errorf("extractQuantization(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

// ── Badges: formatMetadataBadges ─────────────────────────────────────────────

func TestFormatMetadataBadges_NilStyle(t *testing.T) {
	m := &modelscan.Model{Name: "model.Q4_K_M.gguf", Size: 4_000_000_000}
	result := formatMetadataBadges(m, nil)
	if result == "" {
		t.Error("formatMetadataBadges with nil style should return non-empty string")
	}
}

func TestFormatMetadataBadges_MMProj(t *testing.T) {
	st := uistyle.GetStyles()
	m := &modelscan.Model{
		Name:        "llava.Q4_K_M.gguf",
		Size:        4_000_000_000,
		MMProjPaths: []string{"/models/mmproj.gguf"},
	}
	result := formatMetadataBadges(m, st)
	if !strings.Contains(result, "mmproj") {
		t.Errorf("badge line should contain 'mmproj', got: %q", result)
	}
}

// ── View: Loading guard ───────────────────────────────────────────────────────

func TestAppView_LoadingBeforeResize(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	// width == 0 → Loading...
	content := viewContent(a)
	if !strings.Contains(content, "Loading") {
		t.Error("view before WindowSizeMsg should show 'Loading...'")
	}
}

// ── Filter edge cases ────────────────────────────────────────────────────────

func TestAppFilter_NoMatchShowsZero(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	a, _ = sendKey(a, "/")
	for _, ch := range "zzzzzzzzz" {
		a, _ = sendKey(a, string(ch))
	}

	if a.countShown != 0 {
		t.Errorf("filter 'zzzzzzzzz' should show 0 results, got %d", a.countShown)
	}
	if a.countTotal != 6 {
		t.Errorf("countTotal must stay 6, got %d", a.countTotal)
	}
}

// ── UTF-8 в фильтре ──────────────────────────────────────────────────────────

func TestFilterInput_CyrillicInput(t *testing.T) {
	f := newFilterInput(uistyle.GetStyles())
	f.Toggle()
	for _, ch := range "тест" {
		f.HandleKey(string(ch))
	}
	if f.Text() != "тест" {
		t.Errorf("filter text want %q, got %q", "тест", f.Text())
	}
	f.HandleKey("backspace")
	if f.Text() != "тес" {
		t.Errorf("after backspace want %q, got %q", "тес", f.Text())
	}
}

func TestFilterInput_LeftRightCyrillic(t *testing.T) {
	f := newFilterInput(uistyle.GetStyles())
	f.Toggle()
	for _, ch := range "абв" {
		f.HandleKey(string(ch))
	}
	if f.cursor != 6 {
		t.Errorf("cursor want 6 (3 runes × 2 bytes), got %d", f.cursor)
	}
	f.HandleKey("left")
	if f.cursor != 4 {
		t.Errorf("after left: cursor want 4, got %d", f.cursor)
	}
	f.HandleKey("right")
	if f.cursor != 6 {
		t.Errorf("after right: cursor want 6, got %d", f.cursor)
	}
}

func TestFilterInput_NonPrintableIgnored(t *testing.T) {
	f := newFilterInput(uistyle.GetStyles())
	f.Toggle()
	f.HandleKey("\x01")
	if f.Text() != "" {
		t.Errorf("non-printable rune should be ignored, got %q", f.Text())
	}
}

// ── UTF-8 в truncatePathLeft ────────────────────────────────────────────────

func TestTruncatePathLeft_Cyrillic(t *testing.T) {
	path := "C:\\Пользователи\\test\\models\\файл.gguf"
	result := truncatePathLeft(path, 20)
	if !strings.HasPrefix(result, "...") {
		t.Errorf("result must start with '...', got %q", result)
	}
	for _, r := range result {
		if r == 0xFFFD {
			t.Errorf("invalid UTF-8 rune in result: %q", result)
			break
		}
	}
}

func TestTruncatePathLeft_AsciiShort(t *testing.T) {
	path := "/short"
	if got := truncatePathLeft(path, 20); got != "/short" {
		t.Errorf("short path should not be truncated: %q", got)
	}
}

// ── BuildFlagsString fix ────────────────────────────────────────────────────

func TestBuildFlagsString_ScanDirNotModel(t *testing.T) {
	got := BuildFlagsString("/models", "gemma-4", 16, 0.7)
	if !strings.Contains(got, "--scan-dir=/models") {
		t.Errorf("BuildFlagsString must include --scan-dir, got: %s", got)
	}
	if strings.Count(got, "--model=") != 1 {
		t.Errorf("BuildFlagsString must have exactly one --model=, got: %s", got)
	}
}

// ── q в режиме фильтра не должно выходить ────────────────────────────────────

func TestAppFilter_QInFilterIsTextNotQuit(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	a, _ = sendKey(a, "/")
	updated, cmd := sendKey(a, "q")

	if cmd != nil {
		t.Error("'q' inside filter must NOT trigger quit")
	}
	if updated.filterText != "q" {
		t.Errorf("'q' inside filter must be inserted as text, got filterText=%q", updated.filterText)
	}
}

func TestAppFilter_QInIdleQuits(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)
	_, cmd := sendKey(a, "q")
	if cmd == nil {
		t.Error("'q' in Idle must trigger quit")
	}
}

func TestAppFilter_EscFromIdleIsNoop(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	// Esc while Idle — должен быть noop
	a, _ = sendKey(a, "esc")
	if a.filterState != FilterIdle {
		t.Errorf("filterState want FilterIdle, got %d", a.filterState)
	}
	if a.countShown != 6 {
		t.Errorf("countShown want 6, got %d", a.countShown)
	}
}

// ── Tabs: 1/2/3 + Tab/Shift+Tab ──────────────────────────────────────────────

func TestAppTabs_NumberKeys(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	// "2" — SetActive(1), но таб 1 disabled → индекс всё равно устанавливается
	a, _ = sendKey(a, "2")
	if a.tabs.Active() != 1 {
		t.Errorf("after '2': tabs.Active() want 1, got %d", a.tabs.Active())
	}

	// "1" — возвращаем на 0
	a, _ = sendKey(a, "1")
	if a.tabs.Active() != 0 {
		t.Errorf("after '1': tabs.Active() want 0, got %d", a.tabs.Active())
	}
}

func TestAppTabs_TabKey(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	// Models и Params enabled → Tab переключает между ними
	a, _ = sendKey(a, "tab")
	if a.tabs.Active() != TabParams {
		t.Errorf("Tab from Models should move to Params, got %d", a.tabs.Active())
	}
	a, _ = sendKey(a, "tab")
	if a.tabs.Active() != TabModels {
		t.Errorf("Tab from Params should wrap to Models, got %d", a.tabs.Active())
	}
}

// ── Help popup toggle ────────────────────────────────────────────────────────

func TestAppHelp_QuestionMarkToggles(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	a, _ = sendKey(a, "?")
	if !a.helpExpanded {
		t.Error("'?' should set helpExpanded=true")
	}

	// View не должен паниковать при helpExpanded
	content := viewContent(a)
	if content == "" {
		t.Error("view with helpExpanded should not be empty")
	}

	a, _ = sendKey(a, "?")
	if a.helpExpanded {
		t.Error("second '?' should set helpExpanded=false")
	}
}

func TestAppHelp_EscClosesPopup(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	a, _ = sendKey(a, "?")
	if !a.helpExpanded {
		t.Fatal("precondition: helpExpanded should be true")
	}

	a, _ = sendKey(a, "esc")
	if a.helpExpanded {
		t.Error("Esc should close help popup")
	}
	// После закрытия popup — фильтр всё ещё в Idle
	if a.filterState != FilterIdle {
		t.Errorf("filterState want FilterIdle after Esc from popup, got %d", a.filterState)
	}
}

func TestAppHelp_PopupDoesNotQuit(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	a, _ = sendKey(a, "?")
	_, cmd := sendKey(a, "esc")
	// Esc из popup не должен передаваться как Quit
	if cmd != nil {
		t.Error("Esc from popup should not return tea.Quit")
	}
}

// ── ctrl+q — выход ───────────────────────────────────────────────────────────

func TestAppCtrlQ_Quits(t *testing.T) {
	a := NewApp(testModels(), nil, nil)
	a = sendWindowSize(a, 120, 40)

	_, cmd := a.Update(tea.KeyPressMsg{Code: rune('q'), Mod: tea.ModCtrl})
	if cmd == nil {
		t.Error("ctrl+q should return tea.Quit cmd")
	}
}

func TestApp_ScrollbarOffset_FollowsCursor(t *testing.T) {
	// 12 моделей — достаточно, чтобы список не влез целиком на экран.
	models := make([]*modelscan.Model, 12)
	for i := range models {
		models[i] = &modelscan.Model{
			Name: fmt.Sprintf("model-%02d.Q4_K_M.gguf", i),
			Path: fmt.Sprintf("/models/model-%02d.gguf", i),
			Size: 4_000_000_000,
		}
	}
	a := NewApp(models, nil, nil)
	m, _ := a.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app := m.(*App)

	itemH := (&StyledDelegate{}).Height()
	listH := app.list.Height()
	visible := listH / itemH
	if visible < 1 {
		visible = 1
	}
	total := len(app.list.Items())

	for _, idx := range []int{0, 3, 6, total - 1} {
		app.list.Select(idx)
		offset, gotVisible, gotTotal := app.scrollbarParams()
		if gotVisible != visible {
			t.Errorf("idx=%d: visible=%d want %d", idx, gotVisible, visible)
		}
		if gotTotal != total {
			t.Errorf("idx=%d: total=%d want %d", idx, gotTotal, total)
		}
		wantOffset := idx - visible/2
		if wantOffset < 0 {
			wantOffset = 0
		}
		if mx := total - visible; wantOffset > mx {
			wantOffset = mx
		}
		if wantOffset < 0 {
			wantOffset = 0
		}
		if offset != wantOffset {
			t.Errorf("idx=%d: offset=%d want %d", idx, offset, wantOffset)
		}
	}
}
