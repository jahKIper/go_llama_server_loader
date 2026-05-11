package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"

	"llama-server-loader/internal/cli/runconfig"
	"llama-server-loader/internal/cli/uistyle"
	"llama-server-loader/pkg/modelscan"
)

// CLI represents the command-line interface application.
type CLI struct {
	scanDir          string
	modelName        string
	threads          int
	temperature      float64
	startWebUI       bool
	webPort          int
	saveConfig       string
	generateParams   bool
	output           string
	selectedModel    *modelscan.Model // Selected model after UI interaction
	modelsConfigPath string           // путь к models.json (--models-config)
	paramsFile       string           // путь к params_ru.json (--params-file)
}

// Flags holds all CLI flag values.
type Flags struct {
	ScanDir        string  `long:"scan-dir" description:"Directory to scan for models"`
	Model          string  `long:"model" description:"Model name to run"`
	Threads        int     `long:"threads" description:"Number of threads"`
	Temperature    float64 `long:"temperature" description:"Temperature for inference"`
	StartWebUI     bool    `long:"start-webui" description:"Start Web UI server"`
	WebPort        int     `long:"port" description:"Web UI port (default 8080)"`
	SaveConfig     string  `long:"save-config" description:"Save configuration file"`
	GenerateParams bool    `long:"generate-params" description:"Generate parameters"`
	Output         string  `long:"output" description:"Output file for generated params"`
	ParamsFile     string  `long:"params-file"   description:"Path to params_ru.json"`
	ModelsConfig   string  `long:"models-config" description:"Path to models.json (default: ./models.json)"`
}

// NewCLI creates a new CLI instance with the given flags.
func NewCLI(flags *Flags) *CLI {
	c := &CLI{
		scanDir:          flags.ScanDir,
		modelName:        flags.Model,
		threads:          flags.Threads,
		temperature:      flags.Temperature,
		startWebUI:       flags.StartWebUI,
		webPort:          flags.WebPort,
		saveConfig:       flags.SaveConfig,
		generateParams:   flags.GenerateParams,
		output:           flags.Output,
		paramsFile:       flags.ParamsFile,
		modelsConfigPath: flags.ModelsConfig,
	}
	if c.webPort == 0 {
		c.webPort = 8080
	}
	return c
}

// ParseFlags parses command-line flags and returns a Flags struct.
func ParseFlags(args []string) (*Flags, error) {
	var flags Flags
	i := 0
	for i < len(args) {
		arg := args[i]
		switch arg {
		case "--scan-dir":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--scan-dir requires a value")
			}
			flags.ScanDir = args[i+1]
			i += 2
		case "--model":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--model requires a value")
			}
			flags.Model = args[i+1]
			i += 2
		case "--threads":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--threads requires a value")
			}
			val, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid --threads value: %s", args[i+1])
			}
			flags.Threads = val
			i += 2
		case "--temperature":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--temperature requires a value")
			}
			val, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid --temperature value: %s", args[i+1])
			}
			flags.Temperature = val
			i += 2
		case "--start-webui":
			flags.StartWebUI = true
			i++
		case "--port":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--port requires a value")
			}
			val, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid --port value: %s", args[i+1])
			}
			flags.WebPort = val
			i += 2
		case "--save-config":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--save-config requires a value")
			}
			flags.SaveConfig = args[i+1]
			i += 2
		case "--generate-params":
			flags.GenerateParams = true
			i++
		case "--output":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--output requires a value")
			}
			flags.Output = args[i+1]
			i += 2
		case "--params-file":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--params-file requires a value")
			}
			flags.ParamsFile = args[i+1]
			i += 2
		case "--models-config":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--models-config requires a value")
			}
			flags.ModelsConfig = args[i+1]
			i += 2
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		default:
			// Check for --flag=value format
			if eqIdx := strings.Index(arg, "="); eqIdx > 0 {
				flagName := arg[:eqIdx]
				value := arg[eqIdx+1:]
				switch flagName {
				case "--scan-dir":
					flags.ScanDir = value
				case "--model":
					flags.Model = value
				case "--threads":
					val, err := strconv.Atoi(value)
					if err != nil {
						return nil, fmt.Errorf("invalid --threads=value: %s", value)
					}
					flags.Threads = val
				case "--temperature":
					val, err := strconv.ParseFloat(value, 64)
					if err != nil {
						return nil, fmt.Errorf("invalid --temperature=value: %s", value)
					}
					flags.Temperature = val
				case "--start-webui":
					flags.StartWebUI = true
				case "--port":
					val, err := strconv.Atoi(value)
					if err != nil {
						return nil, fmt.Errorf("invalid --port=value: %s", value)
					}
					flags.WebPort = val
				case "--save-config":
					flags.SaveConfig = value
				case "--generate-params":
					flags.GenerateParams = true
				case "--output":
					flags.Output = value
				case "--params-file":
					flags.ParamsFile = value
				case "--models-config":
					flags.ModelsConfig = value
				default:
					return nil, fmt.Errorf("unknown flag: %s", arg)
				}
			} else {
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
		i++
		}
	}
	return &flags, nil
}

// printHelp prints the help message.
func printHelp() {
	help := `llama-server-loader - A terminal UI for managing and running llama.cpp servers

Usage:
  llama-server-loader [options]

Options:
  --scan-dir <path>        Directory to scan for .gguf models
  --model <name>           Model name to run (from scanned list)
  --threads <count>        Number of CPU threads (default: auto-detect)
  --temperature <float>    Sampling temperature (default: 0.8)
  --start-webui            Start the embedded Web UI server
  --port <number>          Web UI port (default: 8080)
  --save-config <file>     Save configuration to file
  --generate-params        Generate parameter configuration
  --output <file>          Output file for generated params
  --params-file <file>     Path to params_ru.json (parameter catalog)
  --models-config <file>   Path to models.json (default: ./models.json)
  -h, --help               Show this help message

Examples:
  llama-server-loader --scan-dir=./models
  llama-server-loader --scan-dir=/models --model=gemma-4
  llama-server-loader --start-webui --port=8080
  llama-server-loader --scan-dir=./models --threads=16 --temperature=0.9
  llama-server-loader --scan-dir=./models --params-file=./params_ru.json --models-config=./models.json
`
	fmt.Println(help)
}

// Run starts the CLI application.
func (c *CLI) Run() error {
	// Validate flags
	if err := c.validate(); err != nil {
		return fmt.Errorf("invalid flags: %w", err)
	}

	// Handle --generate-params
	if c.generateParams {
		return c.generateParameters()
	}

	// Handle --start-webui without scan
	if c.startWebUI && c.scanDir == "" {
		fmt.Println("Starting Web UI server on port", c.webPort)
		return nil // TODO: Start web UI
	}

	// Interactive mode: scan and show model list
	if c.scanDir != "" {
		return c.runInteractive()
	}

	// Default: show help
	printHelp()
	return nil
}

// validate checks flag values for correctness.
func (c *CLI) validate() error {
	if c.scanDir != "" {
		info, err := os.Stat(c.scanDir)
		if err != nil {
			return fmt.Errorf("cannot access scan directory %s: %w", c.scanDir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", c.scanDir)
		}
	}

	if c.threads == 0 && c.modelName != "" {
		return fmt.Errorf("threads must be >= 1 when model is specified")
	}

	if c.temperature < 0 || c.temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	return nil
}

// runInteractive starts the terminal UI with a model list.
func (c *CLI) runInteractive() error {
	// Scan directory using pkg/modelscan
	scanResult, err := modelscan.ScanDir(c.scanDir)
	if err != nil {
		return fmt.Errorf("error scanning directory: %w", err)
	}

	log.Printf("Found %d models and %d mmproj files", len(scanResult.Models), len(scanResult.MMModels))

	// Match models with mmproj files
	enrichedModels, err := modelscan.MatchMMProj(scanResult.Models)
	if err != nil {
		return fmt.Errorf("error matching mmproj: %w", err)
	}

	if len(enrichedModels) == 0 {
		return fmt.Errorf("no models found in directory: %s", c.scanDir)
	}

	// Set terminal background color to dark (#0a0f18 from StyleConfig)
	// This ensures the terminal itself has a solid background, not transparent
	SetTerminalBackground("#0a0f18")

	// Ensure background is reset on exit (defer cleanup)
	defer ResetTerminalBackground()

	// Create and run TUI
	app := NewApp(enrichedModels)
	p := tea.NewProgram(app)
	_, err = p.Run()
	if err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	// Store selected model
	c.selectedModel = app.selected

	// If no model was selected (e.g., user pressed Esc), return without error
	if c.selectedModel == nil {
		log.Println("No model selected, returning to model list")
		fmt.Println("\nМодель не выбрана")
		return nil
	}

	log.Printf("Selected model: %+v", c.selectedModel)

	// Запускаем второй экран — параметры запуска модели
	paramsFilePath, _ := runconfig.ResolveParamsFile(c.paramsFile)
	rcApp := runconfig.NewApp(c.selectedModel, paramsFilePath)
	p2 := tea.NewProgram(rcApp)
	if _, err = p2.Run(); err != nil {
		return fmt.Errorf("error running run config TUI: %w", err)
	}

	res := rcApp.Result()
	log.Printf("runconfig: action=%v rows=%d", res.Action, len(res.Rows))

	if res.Action == runconfig.ActionRun {
		modelsCfgPath := c.modelsConfigPath
		if modelsCfgPath == "" {
			modelsCfgPath = "models.json"
		}
		ResetTerminalBackground()
		return runconfig.SaveAndRun(modelsCfgPath, res.Model, res.Rows)
	}

	return nil
}

// generateParameters creates a parameter configuration file.
func (c *CLI) generateParameters() error {
	output := c.output
	if output == "" {
		output = "generated_params.json"
	}

	dir := filepath.Dir(output)
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create directory %s: %w", dir, err)
		}
	}

	threads := c.threads
	if threads == 0 {
		threads = runtime.NumCPU()
	}
	temperature := c.temperature
	if temperature == 0 {
		temperature = 0.8
	}

	params := map[string]any{
		"output_file":         output,
		"generated_at":        time.Now().UTC().Format(time.RFC3339),
		"default_threads":     threads,
		"default_temperature": temperature,
	}

	data, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal params: %w", err)
	}

	if err := os.WriteFile(output, data, 0644); err != nil {
		return fmt.Errorf("cannot write params file: %w", err)
	}

	fmt.Println("Parameters generated:", output)
	return nil
}

// FormatModelName formats a model name for display.
func FormatModelName(name string, path string, size int64) string {
	sizeStr := formatSize(size)
	if name == "" {
		name = filepath.Base(path)
	}
	return fmt.Sprintf("%s (%s)", name, sizeStr)
}

// formatSize formats a file size in bytes to human-readable form.
func formatSize(size int64) string {
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

// BuildFlagsString builds a command-line flags string from CLI config.
func BuildFlagsString(scanDir, modelName string, threads int, temperature float64) string {
	var parts []string
	if scanDir != "" {
		parts = append(parts, fmt.Sprintf("--scan-dir=%s", scanDir))
	}
	if modelName != "" {
		parts = append(parts, fmt.Sprintf("--model=%s", modelName))
	}
	if threads > 0 {
		parts = append(parts, fmt.Sprintf("--threads=%d", threads))
	}
	if temperature > 0 {
		parts = append(parts, fmt.Sprintf("--temperature=%.2f", temperature))
	}
	return strings.Join(parts, " ")
}

// DetectCPUCores returns the number of CPU cores on the system.
func DetectCPUCores() int {
	return runtime.NumCPU()
}

// ScanDir returns the scan directory.
func (c *CLI) ScanDir() string {
	return c.scanDir
}

// ModelName returns the model name.
func (c *CLI) ModelName() string {
	return c.modelName
}

// Threads returns the number of threads.
func (c *CLI) Threads() int {
	return c.threads
}

// Temperature returns the temperature value.
func (c *CLI) Temperature() float64 {
	return c.temperature
}

// SaveConfig returns the config file path.
func (c *CLI) SaveConfig() string {
	return c.saveConfig
}

// GenerateParams returns whether to generate parameters.
func (c *CLI) GenerateParams() bool {
	return c.generateParams
}

// SelectedModel returns the selected model after UI interaction.
func (c *CLI) SelectedModel() *modelscan.Model {
	return c.selectedModel
}

// SelectedModelName returns the name of the selected model after UI interaction.
func (c *CLI) SelectedModelName() string {
	if c.selectedModel != nil {
		return strings.TrimSuffix(filepath.Base(c.selectedModel.Path), ".gguf")
	}
	return ""
}

// ============================================================================
// Terminal UI components using charm.land/bubbles/v2/list and bubbletea/v2
// ============================================================================

// Layout-константы — алиасы из uistyle для совместимости с пакетом cli.
const (
	OuterPadH = uistyle.OuterPadH
	OuterPadV = uistyle.OuterPadV
	BlockGap  = uistyle.BlockGap
)

// App represents the main TUI application.
type App struct {
	list         list.Model
	selected     *modelscan.Model
	filterState  FilterState
	filterInput  *filterInput
	filterText   string
	allModels    []*modelscan.Model
	countShown   int
	countTotal   int
	title        string
	version      string
	width        int
	height       int
	styles       *uistyle.StyleConfig
	dims         *Dimensions
	footer       *Footer
	tabs         *TabBar
	helpPopup    *HelpPopup
	helpExpanded bool // toggled by "?"
}

// NewApp creates a new App with model list.
func NewApp(models []*modelscan.Model) *App {
	st := uistyle.GetStyles()
	dims := DefaultDimensions()

	items := make([]list.Item, len(models))
	for i, m := range models {
		items[i] = NewListItem(m)
	}

	delegate := &StyledDelegate{base: list.NewDefaultDelegate(), styles: st}
	l := list.New(items, delegate, 60, 20)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)       // встроенный help отключён — используем кастомный Footer
	l.SetShowPagination(false) // pagination dots не нужны — список со скроллом
	l.SetShowTitle(false)      // встроенный title тоже не нужен (есть свой Header)
	// «No items.» (пустой результат фильтра) — на фоне BgPanel, чтобы
	// не светил терминальный bg.
	l.Styles.NoItems = lipgloss.NewStyle().
		Background(lipgloss.Color(st.BgPanel)).
		Foreground(lipgloss.Color(st.TextMuted)).
		Padding(0, 0, 0, 2)

	return &App{
		list:        l,
		allModels:   models,
		countShown:  len(models),
		countTotal:  len(models),
		title:       "llama-server-loader - Model Selector",
		version:     "0.1.0",
		styles:      st,
		dims:        &dims,
		footer:      NewFooter(st),
		filterInput: newFilterInput(st),
		filterState: FilterIdle,
		tabs:        NewTabBar(st),
		helpPopup:   NewHelpPopup(st),
	}
}

// Init implements tea.Model interface.
func (a *App) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model interface.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.footer.SetWidth(msg.Width)
		a.recomputeListSize()
		return a, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit

		case "ctrl+q":
			return a, tea.Quit

		case "q":
			// q — выход только вне режима фильтра. В фильтре "q" — обычный символ.
			if a.filterState == FilterIdle {
				return a, tea.Quit
			}
			a.filterInput.HandleKey("q")
			a.filterText = a.filterInput.Text()
			a.applyFilter()
			return a, nil

		case "enter":
			if item, ok := a.list.SelectedItem().(*ListItem); ok {
				a.selected = item.Model()
				return a, tea.Quit
			}

		case "esc":
			// Esc-иерархия: popup → filter → no-op
			if a.helpExpanded {
				a.helpExpanded = false
				return a, nil
			}
			if a.filterState != FilterIdle {
				a.filterInput.Toggle()
				a.filterInput.Clear()
				a.filterState = FilterIdle
				a.filterText = ""
				setListItems(a, a.allModels)
				a.countShown = a.countTotal
			}
			return a, nil

		case "/":
			if a.filterState == FilterIdle {
				a.filterInput.Toggle()
				a.filterState = FilterActive
			} else {
				a.filterInput.HandleKey("/")
				a.filterText = a.filterInput.Text()
				a.applyFilter()
			}
			return a, nil

		case "?":
			if a.filterState == FilterIdle {
				a.helpExpanded = !a.helpExpanded
				return a, nil
			}
			// В режиме фильтра "?" — обычный символ
			a.filterInput.HandleKey("?")
			a.filterText = a.filterInput.Text()
			a.applyFilter()
			return a, nil

		case "1", "2", "3":
			if a.filterState == FilterIdle {
				idx := int(msg.String()[0]-'1')
				a.tabs.SetActive(idx)
				return a, nil
			}
			a.filterInput.HandleKey(msg.String())
			a.filterText = a.filterInput.Text()
			a.applyFilter()
			return a, nil

		case "tab":
			if a.filterState == FilterIdle {
				a.tabs.Next()
				return a, nil
			}

		case "shift+tab":
			if a.filterState == FilterIdle {
				a.tabs.Prev()
				return a, nil
			}

		default:
			keyStr := msg.String()
			if a.filterState == FilterActive || a.filterState == Filtering {
				if keyStr == "up" || keyStr == "down" {
					break
				}
				a.filterInput.HandleKey(keyStr)
				a.filterText = a.filterInput.Text()
				a.applyFilter()
				return a, nil
			}
		}
	}

	newList, cmd := a.list.Update(msg)
	a.list = newList
	return a, cmd
}

// recomputeListSize пересчитывает размер списка по текущим width/height.
// Учитывает outer-padding и BlockGap, чтобы list не вылезал за «воздушный» layout.
func (a *App) recomputeListSize() {
	if a.width == 0 || a.height == 0 {
		return
	}

	innerW, padH := a.outerInnerWidth()
	_ = padH

	// header(1) + footer(1) — фиксированные. Зазоры — 2*BlockGap.
	headerH := 1
	footerH := 1
	contentH := a.height - 2*OuterPadV - headerH - 2*BlockGap - footerH
	if contentH < 7 {
		contentH = 7
	}

	contentStyle := a.styles.ContentBlockStyle()
	// listW = inner area content-блока минус scrollbar
	listW := innerW - contentStyle.GetHorizontalFrameSize() - ScrollbarWidth
	// filterRow (3 строки с рамкой) + verticalFrame content-блока
	listH := contentH - contentStyle.GetVerticalFrameSize() - 3

	if listW < 10 {
		listW = 10
	}
	if listH < 7 {
		listH = 7
	}
	a.list.SetSize(listW, listH)
}

// scrollbarParams вычисляет (offset, visible, total) для RenderScrollbar.
// offset аппроксимируется по индексу курсора, а не по пагинатору, поэтому
// thumb следует за курсором плавно, а не прыгает страницами.
func (a *App) scrollbarParams() (offset, visible, total int) {
	itemH := (&StyledDelegate{}).Height()
	if itemH < 1 {
		itemH = 1
	}
	listH := a.list.Height()
	total = len(a.list.Items())
	visible = listH / itemH
	if visible < 1 {
		visible = 1
	}
	idx := a.list.Index()
	offset = idx - visible/2
	if offset < 0 {
		offset = 0
	}
	if mx := total - visible; offset > mx {
		offset = mx
	}
	if offset < 0 {
		offset = 0
	}
	return offset, visible, total
}

// outerInnerWidth возвращает ширину «inner column» (a.width минус 2*OuterPadH)
// и фактический padH. При очень узком терминале (<20 col) padding отключается,
// чтобы UI не ломался.
func (a *App) outerInnerWidth() (innerW, padH int) {
	padH = OuterPadH
	if a.width < 20 {
		padH = 0
	}
	innerW = a.width - 2*padH
	if innerW < 1 {
		innerW = 1
	}
	return innerW, padH
}

// applyFilter применяет filterText к списку и синхронизирует filterState и countShown.
func (a *App) applyFilter() {
	if a.filterText == "" {
		a.filterState = FilterActive
		setListItems(a, a.allModels)
		a.countShown = a.countTotal
	} else {
		a.filterState = Filtering
		filtered := FilterModels(a.allModels, a.filterText)
		setListItems(a, filtered)
		a.countShown = len(filtered)
	}
}

// setListItems заменяет элементы list.Model из среза моделей.
func setListItems(a *App, models []*modelscan.Model) {
	items := make([]list.Item, len(models))
	for i, m := range models {
		items[i] = NewListItem(m)
	}
	a.list.SetItems(items)
}

// View implements tea.Model interface — 3-блочная раскладка.
func (a *App) View() tea.View {
	// Guard: первый рендер до WindowSizeMsg
	if a.width == 0 {
		v := tea.NewView("Loading...")
		v.AltScreen = true
		return v
	}

	// ── Help popup (полноэкранный, early return) ──────────────────────────
	if a.helpExpanded {
		screen := a.helpPopup.Render(a.width, a.height)
		v := tea.NewView(screen)
		v.AltScreen = true
		v.BackgroundColor = lipgloss.Color(a.styles.DarkBg)
		return v
	}

	innerW, padH := a.outerInnerWidth()

	// ── Block 1: Header (version слева, title по центру, tabs справа) ─────
	headerStyle := a.styles.HeaderBlockStyle()
	headerInnerW := innerW - headerStyle.GetHorizontalFrameSize()
	if headerInnerW < 1 {
		headerInnerW = 1
	}
	header := headerStyle.
		Width(innerW).
		Render(RenderHeader(a.title, a.version, a.styles, headerInnerW, a.tabs.Render()))

	// ── Block 2: Content ──────────────────────────────────────────────────
	contentStyle := a.styles.ContentBlockStyle()
	contentInnerW := innerW - contentStyle.GetHorizontalFrameSize()

	filterRow := a.filterInput.Render(contentInnerW)

	// Scrollbar параметры (плавный offset по курсору)
	listW, listH := a.list.Width(), a.list.Height()
	sbOffset, sbVisible, sbTotal := a.scrollbarParams()
	scrollbar := RenderScrollbar(sbOffset, sbVisible, sbTotal, listH, a.styles)

	// Заливаем list-вьюху BgPanel: при пустом списке («No items.») и для
	// неиспользованного хвоста, чтобы терминальный фон не просвечивал.
	listView := lipgloss.NewStyle().
		Background(lipgloss.Color(a.styles.BgPanel)).
		Width(listW).
		Height(listH).
		Render(a.list.View())

	listWithScrollbar := lipgloss.JoinHorizontal(lipgloss.Top,
		listView,
		scrollbar,
	)

	innerContent := lipgloss.JoinVertical(lipgloss.Left,
		filterRow,
		listWithScrollbar,
	)

	rawContent := contentStyle.Width(innerW).Render(innerContent)

	counter := fmt.Sprintf("%d / %d", a.countShown, a.countTotal)
	content := InjectBorderTitle(rawContent, "Список моделей", counter)

	// ── Block 3: Footer ───────────────────────────────────────────────────
	if _, ok := a.list.SelectedItem().(*ListItem); ok {
		a.footer.SetPrimaryCTA(" Enter — Запустить выбранную модель ")
	} else {
		a.footer.SetPrimaryCTA("")
	}
	a.footer.SetWidth(innerW)
	footerLine := a.footer.Render()

	// ── Сборка стека с зазорами ───────────────────────────────────────────
	gapStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(a.styles.DarkBg)).
		Width(innerW)
	gap := gapStyle.Render("")
	stackParts := []string{header}
	for i := 0; i < BlockGap; i++ {
		stackParts = append(stackParts, gap)
	}
	stackParts = append(stackParts, content)
	for i := 0; i < BlockGap; i++ {
		stackParts = append(stackParts, gap)
	}
	stackParts = append(stackParts, footerLine)
	stack := lipgloss.JoinVertical(lipgloss.Left, stackParts...)

	// ── Outer padding: симметричный gutter + воздух сверху/снизу ──────────
	screen := lipgloss.NewStyle().
		Background(lipgloss.Color(a.styles.DarkBg)).
		Padding(OuterPadV, padH).
		Width(a.width).
		Render(stack)

	v := tea.NewView(screen)
	v.AltScreen = true
	v.BackgroundColor = lipgloss.Color(a.styles.DarkBg)
	return v
}

// Selected returns the currently selected model.
func (a *App) Selected() *modelscan.Model {
	return a.selected
}

