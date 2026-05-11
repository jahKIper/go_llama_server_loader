package runconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"

	"llama-server-loader/pkg/modelscan"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func sendAppKey(a *RunConfigApp, key string) (*RunConfigApp, tea.Cmd) {
	m, cmd := a.Update(keyMsg(key))
	return m.(*RunConfigApp), cmd
}

func sendAppWindowSize(a *RunConfigApp, w, h int) *RunConfigApp {
	m, _ := a.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return m.(*RunConfigApp)
}

func testModel() *modelscan.Model {
	return &modelscan.Model{
		Name: "llama-3-8b-q4_k_m.gguf",
		Path: "/models/llama-3-8b-q4_k_m.gguf",
		Size: 4_661_248_000,
	}
}

// writeTempCatalog записывает минимальный params_ru.json во временную директорию
// и возвращает путь к файлу. Каталог содержит два параметра:
//   - --ctx-size N  (ключ ctx_size)
//   - --temp N      (ключ temp)
func writeTempCatalog(t *testing.T, dir string) string {
	t.Helper()
	type paramMeta struct {
		ShortFlag string `json:"short_flag"`
		LongFlag  string `json:"long_flag"`
		DescRU    string `json:"description_ru"`
	}
	type category struct {
		Name   string      `json:"name"`
		Params []paramMeta `json:"params"`
	}
	type paramFile struct {
		Version          string     `json:"version"`
		TotalParamsCount int        `json:"total_params_count"`
		Categories       []category `json:"categories"`
	}
	pf := paramFile{
		Version:          "1.0",
		TotalParamsCount: 2,
		Categories: []category{
			{
				Name: "Основные",
				Params: []paramMeta{
					{
						ShortFlag: "-c N",
						LongFlag:  "--ctx-size N",
						DescRU:    "Размер контекста модели",
					},
					{
						ShortFlag: "-t N",
						LongFlag:  "--temp N",
						DescRU:    "Температура сэмплирования",
					},
				},
			},
		},
	}
	raw, err := json.Marshal(pf)
	if err != nil {
		t.Fatalf("writeTempCatalog: marshal: %v", err)
	}
	path := filepath.Join(dir, "params_ru.json")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("writeTempCatalog: write: %v", err)
	}
	return path
}

// ── Сценарий A: полный happy-path ────────────────────────────────────────────
//
// Tab (нет, не нужен) → / → ctx → Enter (закрыть фильтр) → Enter (добавить)
// → Tab (FocusLeft) → e (редактировать) → 4096 → Enter (подтвердить) → r (Run)
// Проверяем: Result.Action == ActionRun, Rows[0].Long == "--ctx-size", Value == "4096".

func TestRunConfigApp_ScenarioA_AddEditRun(t *testing.T) {
	dir := t.TempDir()
	catalogPath := writeTempCatalog(t, dir)

	a := NewApp(testModel(), catalogPath, "")
	a = sendAppWindowSize(a, 120, 40)

	// Начальный фокус — правая панель.
	if a.focus != FocusRight {
		t.Fatalf("initial focus want FocusRight, got %v", a.focus)
	}

	// Активировать фильтр.
	a, _ = sendAppKey(a, "/")
	if !a.right.IsFilterActive() {
		t.Fatal("filter should be active after '/'")
	}

	// Набрать "ctx" → только --ctx-size остаётся в отфильтрованном списке.
	for _, ch := range "ctx" {
		a, _ = sendAppKey(a, string(ch))
	}
	if len(a.right.filtered) != 1 {
		t.Fatalf("expected 1 filtered entry after 'ctx', got %d", len(a.right.filtered))
	}
	if a.right.filtered[0].Meta.LongFlag != "--ctx-size N" {
		t.Errorf("unexpected filtered entry: %q", a.right.filtered[0].Meta.LongFlag)
	}

	// Enter 1 — закрыть фильтр (filterActive → false).
	a, _ = sendAppKey(a, "enter")
	if a.right.IsFilterActive() {
		t.Fatal("filter should be closed after Enter")
	}

	// Enter 2 — добавить выбранный параметр в левую панель.
	a, _ = sendAppKey(a, "enter")
	rows := a.left.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row in left panel after Enter, got %d", len(rows))
	}
	if rows[0].Long != "--ctx-size" {
		t.Errorf("expected Long='--ctx-size', got %q", rows[0].Long)
	}
	if rows[0].Key != "ctx_size" {
		t.Errorf("expected Key='ctx_size', got %q", rows[0].Key)
	}

	// Tab — переключить фокус в левую панель.
	a, _ = sendAppKey(a, "tab")
	if a.focus != FocusLeft {
		t.Fatalf("focus want FocusLeft after Tab, got %v", a.focus)
	}

	// Enter — начать редактирование значения.
	a, _ = sendAppKey(a, "enter")
	if !a.left.IsEditing() {
		t.Fatal("left panel should be in edit mode after Enter")
	}

	// Ввести "4096" в textinput.
	for _, ch := range "4096" {
		a, _ = sendAppKey(a, string(ch))
	}

	// Enter — подтвердить редактирование.
	a, _ = sendAppKey(a, "enter")
	if a.left.IsEditing() {
		t.Fatal("editing should be done after Enter")
	}
	if a.left.Rows()[0].Value != "4096" {
		t.Errorf("expected value '4096', got %q", a.left.Rows()[0].Value)
	}

	// r — Save+Run.
	a2, cmd := sendAppKey(a, "r")
	if cmd == nil {
		t.Error("'r' should return tea.Quit cmd")
	}

	res := a2.Result()
	if res.Action != ActionRun {
		t.Errorf("expected ActionRun (%d), got %d", ActionRun, res.Action)
	}
	if len(res.Rows) != 1 {
		t.Fatalf("expected 1 row in result, got %d", len(res.Rows))
	}
	if res.Rows[0].Long != "--ctx-size" {
		t.Errorf("result: Long want '--ctx-size', got %q", res.Rows[0].Long)
	}
	if res.Rows[0].Value != "4096" {
		t.Errorf("result: Value want '4096', got %q", res.Rows[0].Value)
	}
	if res.Model == nil {
		t.Error("result: Model must not be nil")
	}
}

// ── Сценарий B: Esc не закрывает приложение ──────────────────────────────────

func TestRunConfigApp_ScenarioB_EscDoesNotQuit(t *testing.T) {
	dir := t.TempDir()
	catalogPath := writeTempCatalog(t, dir)

	a := NewApp(testModel(), catalogPath, "")
	a = sendAppWindowSize(a, 120, 40)

	// Esc на правой панели без активного фильтра — не должен завершать программу.
	_, cmd := sendAppKey(a, "esc")
	if cmd != nil {
		t.Error("Esc should NOT return tea.Quit cmd — only 'q' closes the app")
	}
}

func TestRunConfigApp_ScenarioB_QCancel(t *testing.T) {
	dir := t.TempDir()
	catalogPath := writeTempCatalog(t, dir)

	a := NewApp(testModel(), catalogPath, "")
	a = sendAppWindowSize(a, 120, 40)

	a2, cmd := sendAppKey(a, "q")
	if cmd == nil {
		t.Error("'q' should return tea.Quit cmd")
	}
	res := a2.Result()
	if res.Action != ActionCancel {
		t.Errorf("expected ActionCancel, got %d", res.Action)
	}
}

func TestRunConfigApp_ScenarioB_CtrlCCancel(t *testing.T) {
	a := NewApp(testModel(), "", "")
	a2, cmd := a.Update(tea.KeyPressMsg{Code: rune('c'), Mod: tea.ModCtrl})
	_ = a2
	if cmd == nil {
		t.Error("ctrl+c should return tea.Quit cmd")
	}
}

// ── Сценарий C: пустой каталог (файл не найден) ──────────────────────────────

func TestRunConfigApp_ScenarioC_EmptyCatalog_NoPanic(t *testing.T) {
	// Передаём несуществующий путь — каталог не загрузится.
	a := NewApp(testModel(), "/nonexistent/path/params_ru.json", "")

	// catalogErr должен быть установлен.
	if a.catalogErr == nil {
		// Файл мог найтись рядом с бинарником (go test) — тогда это нормально.
		t.Log("note: params file was found via fallback — skipping 'not found' check")
		return
	}

	// View не должен паниковать.
	a = sendAppWindowSize(a, 120, 40)
	v := a.View()
	if v.Content == "" {
		t.Error("View should return non-empty content even with empty catalog")
	}

	// Правая панель должна рендериться без паники и содержать сообщение о пустом каталоге.
	rendered := a.right.Render(true)
	if rendered == "" {
		t.Error("right panel Render should return non-empty string for empty catalog")
	}
}

func TestRunConfigApp_ScenarioC_EmptyCatalog_ShowsNotFound(t *testing.T) {
	// Меняем CWD во временную пустую директорию, чтобы исключить fallback на ./params_ru.json.
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	a := NewApp(testModel(), "", "")
	if a.catalogErr == nil {
		// Нашёлся файл рядом с бинарником — не можем проверить "not found".
		t.Skip("params_ru.json found via binary-dir fallback, skipping empty-catalog test")
	}

	a = sendAppWindowSize(a, 120, 40)

	// Панель не должна паниковать.
	rendered := a.right.Render(true)
	if rendered == "" {
		t.Error("right panel Render must not be empty")
	}
	// right.all должен быть nil/пустым.
	if len(a.right.all) != 0 {
		t.Errorf("right panel should have 0 entries, got %d", len(a.right.all))
	}
}

// ── Дополнительные smoke-тесты ────────────────────────────────────────────────

func TestRunConfigApp_WindowResize_NoPanic(t *testing.T) {
	dir := t.TempDir()
	catalogPath := writeTempCatalog(t, dir)
	a := NewApp(testModel(), catalogPath, "")

	for _, size := range [][2]int{{120, 40}, {80, 24}, {60, 20}, {40, 10}} {
		a = sendAppWindowSize(a, size[0], size[1])
		v := a.View()
		_ = v // не паникуем
	}
}

func TestRunConfigApp_TabSwitchesFocus(t *testing.T) {
	dir := t.TempDir()
	catalogPath := writeTempCatalog(t, dir)
	a := NewApp(testModel(), catalogPath, "")
	a = sendAppWindowSize(a, 120, 40)

	if a.focus != FocusRight {
		t.Fatal("initial focus should be FocusRight")
	}
	a, _ = sendAppKey(a, "tab")
	if a.focus != FocusLeft {
		t.Error("focus should be FocusLeft after first Tab")
	}
	a, _ = sendAppKey(a, "tab")
	if a.focus != FocusRight {
		t.Error("focus should be FocusRight after second Tab")
	}
}

func TestRunConfigApp_DescriptionPopup(t *testing.T) {
	dir := t.TempDir()
	catalogPath := writeTempCatalog(t, dir)
	a := NewApp(testModel(), catalogPath, "")
	a = sendAppWindowSize(a, 120, 40)

	// ? — открыть popup.
	a, _ = sendAppKey(a, "?")
	if !a.showDesc {
		t.Error("showDesc should be true after '?'")
	}
	// View не должен паниковать.
	v := a.View()
	if v.Content == "" {
		t.Error("View with popup should return non-empty content")
	}

	// Esc — закрыть popup (не выходить из программы).
	a2, cmd := sendAppKey(a, "esc")
	if cmd != nil {
		t.Error("Esc from popup should not trigger Quit")
	}
	if a2.showDesc {
		t.Error("showDesc should be false after Esc from popup")
	}
}

func TestRunConfigApp_EscFromFilter_NotQuit(t *testing.T) {
	dir := t.TempDir()
	catalogPath := writeTempCatalog(t, dir)
	a := NewApp(testModel(), catalogPath, "")
	a = sendAppWindowSize(a, 120, 40)

	// Активировать фильтр.
	a, _ = sendAppKey(a, "/")
	if !a.right.IsFilterActive() {
		t.Fatal("filter should be active")
	}

	// Esc — закрыть фильтр, но не выйти.
	a2, cmd := sendAppKey(a, "esc")
	if cmd != nil {
		t.Error("Esc from filter should not trigger Quit")
	}
	if a2.right.IsFilterActive() {
		t.Error("filter should be inactive after Esc")
	}
	if a2.action == ActionCancel {
		// action не должен был зафиксироваться как Cancel — quit не было вызвано
		// (action=Cancel — дефолт, но Quit не вернулся, значит программа продолжается)
	}
	_ = a2
}

func TestRunConfigApp_DeleteRow(t *testing.T) {
	dir := t.TempDir()
	catalogPath := writeTempCatalog(t, dir)
	a := NewApp(testModel(), catalogPath, "")
	a = sendAppWindowSize(a, 120, 40)

	// Добавить параметр через правую панель.
	a, _ = sendAppKey(a, "enter")
	if len(a.left.Rows()) != 1 {
		t.Fatalf("expected 1 row after Enter, got %d", len(a.left.Rows()))
	}

	// Перейти в левую панель и удалить строку.
	a, _ = sendAppKey(a, "tab")
	a, _ = sendAppKey(a, "d")
	if len(a.left.Rows()) != 0 {
		t.Errorf("expected 0 rows after 'd', got %d", len(a.left.Rows()))
	}
}

func TestRunConfigApp_NoDuplicateRows(t *testing.T) {
	dir := t.TempDir()
	catalogPath := writeTempCatalog(t, dir)
	a := NewApp(testModel(), catalogPath, "")
	a = sendAppWindowSize(a, 120, 40)

	// Добавить одну и ту же запись дважды.
	a, _ = sendAppKey(a, "enter")
	a, _ = sendAppKey(a, "enter")

	if len(a.left.Rows()) != 1 {
		t.Errorf("expected 1 row (no duplicates), got %d", len(a.left.Rows()))
	}
}

func TestRunConfigApp_ResultModel(t *testing.T) {
	dir := t.TempDir()
	catalogPath := writeTempCatalog(t, dir)
	m := testModel()
	a := NewApp(m, catalogPath, "")

	res := a.Result()
	if res.Model != m {
		t.Error("result Model should be the same pointer as passed to NewApp")
	}
}

func TestRunConfigApp_InitialStateBeforeResize(t *testing.T) {
	a := NewApp(testModel(), "", "")
	// width == 0 → "Загрузка..."
	v := a.View()
	if v.Content == "" {
		t.Error("View before WindowSizeMsg should return non-empty content")
	}
}
