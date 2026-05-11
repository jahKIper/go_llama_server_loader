package runconfig

import (
	"log"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/modelparams"
	"llama-server-loader/internal/cli/uistyle"
	"llama-server-loader/pkg/modelscan"
	"llama-server-loader/pkg/servercmd"
)

// narrowModeThreshold — ширина, ниже которой включается режим одной панели.
const narrowModeThreshold = 80

// minSupportedWidth — минимальная поддерживаемая ширина терминала.
const minSupportedWidth = 60

// FocusZone определяет активную панель экрана.
type FocusZone int

const (
	FocusRight FocusZone = iota // правая панель (каталог) — по умолчанию
	FocusLeft                   // левая панель (выбранные параметры)
)

// RunConfigApp — второй tea.Model, открывается после выбора модели.
type RunConfigApp struct {
	model          *modelscan.Model
	rows           []ParamRow
	width          int
	height         int
	action         RunConfigAction
	styles         *uistyle.StyleConfig
	catalog        []CatalogEntry
	catalogErr     error
	paramsFilePath string

	focus    FocusZone
	right    *RightPanel
	left     *LeftPanel
	showDesc bool
	descText string

	// GGUF-параметры выбранной модели (загружаются из models.json).
	params  *modelparams.Lookup
	curated modelparams.Curated

	// Узкий режим: одна панель за раз
	narrowMode bool

	// Конфликты флагов
	showConflicts    bool
	conflictWarnings []string

	// Комментарий пользователя к модели
	comment      string
	showComment  bool
	commentInput textinput.Model
}

// NewApp создаёт RunConfigApp для заданной модели.
// paramsFilePath — явный путь к params_ru.json (пустая строка = поиск по умолчанию).
// modelsCfgPath — путь к models.json для подгрузки сохранённых параметров модели
// (пустая строка = не подгружать).
func NewApp(m *modelscan.Model, paramsFilePath string, modelsCfgPath string) *RunConfigApp {
	a := &RunConfigApp{
		model:          m,
		rows:           []ParamRow{},
		action:         ActionCancel,
		styles:         uistyle.GetStyles(),
		paramsFilePath: paramsFilePath,
		focus:          FocusRight,
	}

	// GGUF-параметры выбранной модели: читаем из models.json по имени.
	a.params = modelparams.LoadFromFile(modelsCfgPath)
	if m != nil {
		a.curated = a.params.ForPathCurated(m.Path)
	}

	resolved, err := ResolveParamsFile(paramsFilePath)
	if err != nil {
		a.catalogErr = err
		log.Printf("runconfig: params file not found: %v", err)
		a.right = NewRightPanel(nil, a.styles, 40, 20)
		a.left = NewLeftPanel(a.styles, 60, 20)
		if m != nil {
			a.left.SetParams(a.params, m.Path)
		}
		return a
	}
	a.paramsFilePath = resolved

	pf, err := LoadCatalog(resolved)
	if err != nil {
		a.catalogErr = err
		log.Printf("runconfig: failed to load catalog from %s: %v", resolved, err)
		a.right = NewRightPanel(nil, a.styles, 40, 20)
		a.left = NewLeftPanel(a.styles, 60, 20)
		if m != nil {
			a.left.SetParams(a.params, m.Path)
		}
		return a
	}
	a.catalog = FlattenCatalog(pf)
	a.right = NewRightPanel(a.catalog, a.styles, 40, 20)
	if m != nil {
		a.right.SetGGUFParams(a.params.ForPath(m.Path))
	}
	a.left = NewLeftPanel(a.styles, 60, 20)
	if m != nil {
		a.left.SetParams(a.params, m.Path)
	}

	a.comment = LoadCommentForModel(modelsCfgPath, m)

	rows := PrefilledRowsForModel(a.catalog, m)
	if saved, ok := LoadSavedFlagsForModel(modelsCfgPath, m); ok {
		rows = MergeWithSavedFlags(a.catalog, rows, saved)
	}
	// Автодефолты добавляются всегда; LeftPanel.Seed дедуплицирует по Long,
	// поэтому уже сохранённые пользователем флаги не перезаписываются — а
	// недостающие подмешиваются с бейджем [default].
	rows = append(rows, ComputeModelDefaults(a.catalog, m, a.params)...)
	a.left.Seed(rows)
	if got := a.left.Rows(); len(got) > 0 {
		a.rows = got
	}
	return a
}

// Init implements tea.Model.
func (a *RunConfigApp) Init() tea.Cmd {
	return nil
}

// tryRun проверяет конфликты флагов и либо запрашивает подтверждение, либо сразу завершает.
// Возвращает (model, cmd) — если конфликтов нет, cmd = tea.Quit и action=Run.
func (a *RunConfigApp) tryRun() (tea.Model, tea.Cmd) {
	var rows []ParamRow
	if a.left != nil {
		rows = a.left.Rows()
	}
	flagsMap := BuildFlagsMap(rows, a.model.Path)
	conflicts := servercmd.DetectFlagConflicts(flagsMap)
	if len(conflicts) > 0 {
		a.showConflicts = true
		a.conflictWarnings = conflicts
		log.Printf("runconfig: flag conflicts detected: %v", conflicts)
		return a, nil
	}
	a.action = ActionRun
	return a, tea.Quit
}

// Update implements tea.Model.
func (a *RunConfigApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.narrowMode = a.width < narrowModeThreshold
		rw, rh := a.rightPanelSize()
		if a.right != nil {
			a.right.SetSize(rw, rh)
		}
		if a.left != nil {
			lw, lh := a.leftPanelSize(rw, rh)
			a.left.SetSize(lw, lh)
		}
		return a, nil

	case tea.KeyPressMsg:
		key := msg.String()
		filterActive := a.right != nil && a.right.IsFilterActive()
		editing := a.left != nil && a.left.IsEditing()

		// ctrl+c всегда
		if key == "ctrl+c" {
			a.action = ActionCancel
			return a, tea.Quit
		}

		// Конфликтный modal — перехватывает все клавиши
		if a.showConflicts {
			switch key {
			case "r":
				// Подтверждение запуска несмотря на конфликты
				a.action = ActionRun
				return a, tea.Quit
			case "esc", "q":
				a.showConflicts = false
			}
			return a, nil
		}

		// Закрыть popup описания по Esc или ?
		if a.showDesc {
			if key == "esc" || key == "?" {
				a.showDesc = false
			}
			return a, nil
		}

		// Попап комментария — перехватывает все клавиши
		if a.showComment {
			switch key {
			case "enter":
				a.comment = a.commentInput.Value()
				a.showComment = false
			case "esc":
				a.showComment = false
			default:
				var cmd tea.Cmd
				a.commentInput, cmd = a.commentInput.Update(msg)
				return a, cmd
			}
			return a, nil
		}

		// В режиме редактирования — перехватываем Enter/Esc, остальное форвардим в input
		if editing {
			switch key {
			case "enter":
				a.left.ConfirmEdit()
				return a, nil
			case "esc":
				a.left.CancelEdit()
				return a, nil
			default:
				cmd := a.left.UpdateInput(msg)
				return a, cmd
			}
		}

		// ? — показать описание для текущего параметра
		if key == "?" {
			a.showDesc = true
			a.descText = a.currentDescription()
			return a, nil
		}

		// c — редактировать комментарий к модели
		if key == "c" && !filterActive && !editing {
			ti := textinput.New()
			ti.SetValue(a.comment)
			ti.SetWidth(60)
			ti.Focus()
			a.commentInput = ti
			a.showComment = true
			return a, nil
		}

		// Backspace — вернуться на первый экран (выбор моделей).
		// Перехватываем только когда нет ввода в фильтре/редакторе.
		if key == "backspace" && !filterActive && !editing {
			a.action = ActionBack
			return a, tea.Quit
		}

		// tab всегда переключает фокус
		if key == "tab" {
			if a.focus == FocusRight {
				a.focus = FocusLeft
			} else {
				a.focus = FocusRight
			}
			return a, nil
		}

		// Форвардим в правую панель, если она в фокусе
		if a.focus == FocusRight && a.right != nil {
			prevFilterActive := a.right.IsFilterActive()

			// 'g' — переключение вкладки CLI ↔ GGUF (вне режима фильтра).
			if key == "g" && !prevFilterActive {
				a.right.ToggleTab()
				return a, nil
			}

			// Enter на правой — добавить параметр в левую (только в CLI-табе).
			if key == "enter" && !prevFilterActive {
				if a.right.Tab() == RightTabCLI {
					if sel := a.right.Selected(); sel != nil && a.left != nil {
						a.left.Add(sel.Meta)
					}
				}
				return a, nil
			}

			cmd := a.right.Update(msg, true)

			// q/r: выход и запуск — только вне режима ввода фильтра.
			// Esc намеренно не закрывает приложение (правая панель сама очищает
			// активный фильтр / текст фильтра при esc).
			if !filterActive {
				switch key {
				case "q":
					a.action = ActionCancel
					return a, tea.Quit
				case "r":
					return a.tryRun()
				}
			}
			return a, cmd
		}

		// Левая панель в фокусе
		if a.focus == FocusLeft && a.left != nil {
			switch key {
			case "up":
				a.left.MoveUp()
				return a, nil
			case "down":
				a.left.MoveDown()
				return a, nil
			case "d", "delete":
				a.left.Remove(a.left.CursorIndex())
				return a, nil
			case "m":
				// Apply-from-model: подставить GGUF-значение в текущий CLI-флаг.
				a.left.ApplyModelValue()
				return a, nil
			case "enter":
				cmd := a.left.StartEdit()
				return a, cmd
			case "q":
				a.action = ActionCancel
				return a, tea.Quit
			case "r":
				return a.tryRun()
			}
			return a, nil
		}

		// Глобальные биндинги (фокус не определён или панели нет)
		switch key {
		case "q":
			a.action = ActionCancel
			return a, tea.Quit
		case "r":
			return a.tryRun()
		}
	}
	return a, nil
}

// leftPanelSize вычисляет ширину и высоту левой панели.
func (a *RunConfigApp) leftPanelSize(rw, rh int) (w, h int) {
	padH := a.effectivePadH()
	innerW := a.width - 2*padH
	if innerW < 1 {
		innerW = 1
	}
	if a.narrowMode {
		return innerW, rh
	}
	lw := innerW - rw - uistyle.BlockGap
	if lw < 10 {
		lw = 10
	}
	return lw, rh
}

// rightPanelSize вычисляет ширину и высоту правой панели.
func (a *RunConfigApp) rightPanelSize() (w, h int) {
	padH := a.effectivePadH()
	innerW := a.width - 2*padH
	if innerW < 1 {
		innerW = 1
	}

	var rw int
	if a.narrowMode {
		// В узком режиме — полная ширина
		rw = innerW
	} else {
		rw = innerW * 40 / 100
		if rw > 56 {
			rw = 56
		}
		if rw < 28 {
			rw = 28
		}
	}

	// высота content-блока: учитываем top-bar (1 строка) + model header
	headerStr := RenderHeader(a.model, a.styles, innerW, a.comment, a.curated)
	modelHeaderH := strings.Count(headerStr, "\n") + 1
	topBarH := 1
	// В узком режиме добавляем строку-таббар
	tabBarH := 0
	if a.narrowMode {
		tabBarH = 1 + uistyle.BlockGap
	}
	footerH := 1
	rh := a.height - 2*uistyle.OuterPadV - topBarH - modelHeaderH - footerH - 3*uistyle.BlockGap - tabBarH
	if rh < 5 {
		rh = 5
	}
	return rw, rh
}

func (a *RunConfigApp) effectivePadH() int {
	if a.width < minSupportedWidth {
		return 0
	}
	return uistyle.OuterPadH
}

// View implements tea.Model.
func (a *RunConfigApp) View() tea.View {
	if a.width == 0 {
		v := tea.NewView("Загрузка...")
		v.AltScreen = true
		return v
	}

	st := a.styles
	padH := a.effectivePadH()
	innerW := a.width - 2*padH
	if innerW < 1 {
		innerW = 1
	}

	// ── Block 0: TopBar (version + title + tabs, такая же как на первом экране)
	topBar := RenderTopBar("llama-server-loader - Параметры запуска", "0.1.0", st, innerW)

	// ── Block 1: Header модели ────────────────────────────────────────────────
	header := RenderHeader(a.model, st, innerW, a.comment, a.curated)

	// ── Block 2: Content ──────────────────────────────────────────────────────
	rw, rh := a.rightPanelSize()
	lw, lh := a.leftPanelSize(rw, rh)

	if a.right != nil {
		a.right.SetSize(rw, rh)
	}
	if a.left != nil {
		a.left.SetSize(lw, lh)
	}

	bgGap := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Width(innerW).
		Render("")

	var contentRow string
	if a.narrowMode {
		contentRow = a.renderNarrowContent(innerW, rw, rh)
	} else {
		contentRow = a.renderWideContent(lw, rw, st)
	}

	// ── Block 3: Footer ───────────────────────────────────────────────────────
	footerLine := RenderFooter(st, innerW)

	// ── Сборка стека ──────────────────────────────────────────────────────────
	stackParts := []string{topBar}
	for i := 0; i < uistyle.BlockGap; i++ {
		stackParts = append(stackParts, bgGap)
	}
	stackParts = append(stackParts, header)
	for i := 0; i < uistyle.BlockGap; i++ {
		stackParts = append(stackParts, bgGap)
	}
	stackParts = append(stackParts, contentRow)
	for i := 0; i < uistyle.BlockGap; i++ {
		stackParts = append(stackParts, bgGap)
	}
	stackParts = append(stackParts, footerLine)

	// Каждую строку из всех stackParts оборачиваем в Width(innerW) + Background(DarkBg),
	// затем добавляем горизонтальные padding-стрипы того же фона слева и справа.
	// Это надёжнее, чем outer Padding+Width — lipgloss не всегда красит свободные
	// колонки в недостающих по ширине строках многострочного контента.
	rowFill := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Width(innerW)
	padCell := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Width(padH).
		Render("")
	fullRowStrip := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Width(a.width).
		Render("")

	var lines []string
	// Верхний vertical padding.
	for i := 0; i < uistyle.OuterPadV; i++ {
		lines = append(lines, fullRowStrip)
	}
	for _, part := range stackParts {
		for _, ln := range strings.Split(part, "\n") {
			painted := rowFill.Render(ln)
			lines = append(lines, padCell+painted+padCell)
		}
	}
	// Нижний vertical padding.
	for i := 0; i < uistyle.OuterPadV; i++ {
		lines = append(lines, fullRowStrip)
	}
	screen := strings.Join(lines, "\n")

	// ── Overlays ─────────────────────────────────────────────────────────────
	if a.showConflicts {
		overlay := a.renderConflictModal(a.width, a.height)
		screen = lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, overlay)
	} else if a.showDesc {
		overlay := a.renderDescPopup(a.width, a.height)
		screen = lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, overlay)
	} else if a.showComment {
		overlay := a.renderCommentPopup(a.width, a.height)
		screen = lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, overlay)
	}

	v := tea.NewView(screen)
	v.AltScreen = true
	v.BackgroundColor = lipgloss.Color(st.DarkBg)
	return v
}

// renderWideContent рендерит две панели рядом (обычный режим).
func (a *RunConfigApp) renderWideContent(lw, rw int, st *uistyle.StyleConfig) string {
	var rightBlock string
	if a.right != nil {
		rightBlock = a.right.Render(a.focus == FocusRight)
	} else {
		rightBlock = emptyPanel(st, rw)
	}

	var leftBlock string
	if a.left != nil {
		leftBlock = a.left.Render(a.focus == FocusLeft)
	} else {
		leftBlock = emptyPanel(st, lw)
	}

	// gap должен быть многострочным, иначе все строки кроме первой при JoinHorizontal
	// заполняются нестилизованными пробелами, через которые просвечивает фон терминала.
	gapH := strings.Count(leftBlock, "\n") + 1
	if rh := strings.Count(rightBlock, "\n") + 1; rh > gapH {
		gapH = rh
	}
	gapCell := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Width(uistyle.BlockGap).
		Render("")
	gapLines := make([]string, gapH)
	for i := 0; i < gapH; i++ {
		gapLines[i] = gapCell
	}
	gap := strings.Join(gapLines, "\n")

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, gap, rightBlock)
}

// renderNarrowContent рендерит одну панель с таббаром (узкий режим).
func (a *RunConfigApp) renderNarrowContent(innerW, panelW, panelH int) string {
	st := a.styles

	// Таббар
	tabBar := a.renderNarrowTabBar(innerW)

	// Активная панель
	var panel string
	if a.focus == FocusRight {
		if a.right != nil {
			a.right.SetSize(panelW, panelH)
			panel = a.right.Render(true)
		} else {
			panel = emptyPanel(st, panelW)
		}
	} else {
		if a.left != nil {
			a.left.SetSize(panelW, panelH)
			panel = a.left.Render(true)
		} else {
			panel = emptyPanel(st, panelW)
		}
	}

	bgGap := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Width(innerW).
		Render("")

	parts := []string{tabBar}
	for i := 0; i < uistyle.BlockGap; i++ {
		parts = append(parts, bgGap)
	}
	parts = append(parts, panel)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderNarrowTabBar рендерит строку-переключатель панелей для узкого режима.
func (a *RunConfigApp) renderNarrowTabBar(w int) string {
	st := a.styles

	makeTab := func(label string, active bool) string {
		fg := st.TextMuted
		bg := st.DarkBg
		if active {
			fg = st.NeonGreen
		}
		return lipgloss.NewStyle().
			Bold(active).
			Background(lipgloss.Color(bg)).
			Foreground(lipgloss.Color(fg)).
			Render(label)
	}

	leftTab := makeTab("[ Запуск ]", a.focus == FocusLeft)
	sep := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Foreground(lipgloss.Color(st.BorderIdle)).
		Render("  ")
	rightTab := makeTab("[ Параметры ]", a.focus == FocusRight)
	hint := lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Foreground(lipgloss.Color(st.TextMuted)).
		Render("  Tab — переключить")

	line := lipgloss.JoinHorizontal(lipgloss.Top, leftTab, sep, rightTab, hint)
	return lipgloss.NewStyle().
		Background(lipgloss.Color(st.DarkBg)).
		Width(w).
		Render(line)
}

// renderConflictModal рендерит модальное предупреждение о конфликтах флагов.
func (a *RunConfigApp) renderConflictModal(sw, sh int) string {
	st := a.styles

	popupW := sw * 60 / 100
	if popupW < 44 {
		popupW = 44
	}
	if popupW > 80 {
		popupW = 80
	}
	innerW := popupW - 4
	if innerW < 10 {
		innerW = 10
	}

	rowFill := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Width(innerW)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color("#fb923c")) // оранжевый — предупреждение

	warnStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextSecondary))

	closeStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextMuted))

	empty := rowFill.Render("")
	rows := []string{
		rowFill.Render(titleStyle.Render("⚠ Конфликты флагов")),
		empty,
	}
	for _, c := range a.conflictWarnings {
		wrapped := wordWrap("• "+c, innerW)
		for _, line := range strings.Split(wrapped, "\n") {
			rows = append(rows, rowFill.Render(warnStyle.Render(line)))
		}
	}
	rows = append(rows, empty,
		rowFill.Render(closeStyle.Render("r — запустить всё равно  ·  Esc — вернуться")),
	)

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color("#fb923c")).
		BorderBackground(lipgloss.Color(st.BgPanel)).
		Background(lipgloss.Color(st.BgPanel)).
		Padding(0, 1).
		Width(innerW).
		Render(body)
}

// renderDescPopup рендерит popup с описанием параметра по центру экрана.
func (a *RunConfigApp) renderDescPopup(sw, sh int) string {
	st := a.styles

	popupW := sw * 60 / 100
	if popupW < 40 {
		popupW = 40
	}
	if popupW > 80 {
		popupW = 80
	}
	innerW := popupW - 4
	if innerW < 10 {
		innerW = 10
	}

	desc := a.descText
	if desc == "" {
		desc = "(описание отсутствует)"
	}

	wrapped := wordWrap(desc, innerW)

	rowFill := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Width(innerW)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.GreenPrimary))

	bodyStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextSecondary))

	closeStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextMuted))

	empty := rowFill.Render("")
	rows := []string{
		rowFill.Render(titleStyle.Render("Описание параметра")),
		empty,
	}
	for _, line := range strings.Split(wrapped, "\n") {
		rows = append(rows, rowFill.Render(bodyStyle.Render(line)))
	}
	rows = append(rows, empty, rowFill.Render(closeStyle.Render("Esc или ? — закрыть")))

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(st.NeonGreen)).
		BorderBackground(lipgloss.Color(st.BgPanel)).
		Background(lipgloss.Color(st.BgPanel)).
		Padding(0, 1).
		Width(innerW).
		Render(body)
}

// currentDescription возвращает описание текущего параметра в активной панели.
func (a *RunConfigApp) currentDescription() string {
	if a.focus == FocusLeft && a.left != nil {
		rows := a.left.Rows()
		idx := a.left.CursorIndex()
		if idx >= 0 && idx < len(rows) && rows[idx].Meta != nil {
			return rows[idx].Meta.DescRU
		}
		return ""
	}
	if a.focus == FocusRight && a.right != nil {
		if sel := a.right.Selected(); sel != nil && sel.Meta != nil {
			return sel.Meta.DescRU
		}
	}
	return ""
}

// Result возвращает итог работы экрана.
func (a *RunConfigApp) Result() RunConfigResult {
	var rows []ParamRow
	if a.left != nil {
		rows = a.left.Rows()
	}
	return RunConfigResult{
		Action:  a.action,
		Rows:    rows,
		Model:   a.model,
		Comment: a.comment,
	}
}

// renderCommentPopup рендерит попап редактирования комментария к модели.
func (a *RunConfigApp) renderCommentPopup(sw, sh int) string {
	st := a.styles

	popupW := sw * 60 / 100
	if popupW < 44 {
		popupW = 44
	}
	if popupW > 80 {
		popupW = 80
	}
	innerW := popupW - 4
	if innerW < 10 {
		innerW = 10
	}

	rowFill := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Width(innerW)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.NeonGreen))

	hintStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextMuted))

	inputStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextSecondary))

	empty := rowFill.Render("")
	rows := []string{
		rowFill.Render(titleStyle.Render("Комментарий к модели")),
		empty,
		rowFill.Render(inputStyle.Render(a.commentInput.View())),
		empty,
		rowFill.Render(hintStyle.Render("Enter — сохранить  ·  Esc — отмена")),
	}

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(st.AccentPurple)).
		BorderBackground(lipgloss.Color(st.BgPanel)).
		Background(lipgloss.Color(st.BgPanel)).
		Padding(0, 1).
		Width(innerW).
		Render(body)
}

// emptyPanel рендерит пустую панель с рамкой (заглушка когда панель nil).
func emptyPanel(st *uistyle.StyleConfig, w int) string {
	inner := w - 2
	if inner < 0 {
		inner = 0
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(st.BorderIdle)).
		BorderBackground(lipgloss.Color(st.BgPanel)).
		Background(lipgloss.Color(st.BgPanel)).
		Width(inner).
		Render("")
}
