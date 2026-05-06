package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"llama-server-loader/internal/cli"
	"llama-server-loader/internal/config"
	"llama-server-loader/internal/webui"
	"llama-server-loader/pkg/modelscan"
	"llama-server-loader/pkg/servercmd"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetPrefix("main: ")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	flags, err := cli.ParseFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Ошибка парсинга флагов: %v", err)
	}

	c := cli.NewCLI(flags)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-sigChan:
			log.Println("Получен сигнал завершения...")
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	if flags.StartWebUI && c.ScanDir() == "" {
		log.Printf("Запуск Web UI сервера на %s", fmt.Sprintf(":%d", flags.WebPort))
		ws := webui.NewServer(fmt.Sprintf(":%d", flags.WebPort))
		if err := ws.Start(ctx); err != nil {
			log.Fatalf("Ошибка запуска Web UI: %v", err)
		}
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		ws.Shutdown(shutdownCtx)
		return
	}

	if c.GenerateParams() {
		if err := c.Run(); err != nil {
			log.Fatalf("Ошибка генерации параметров: %v", err)
		}
		return
	}

	if c.ScanDir() != "" {
		if err := runInteractive(ctx, c); err != nil {
			log.Fatalf("Ошибка интерактивного режима: %v", err)
		}
		return
	}

	printUsage()
}

func printUsage() {
	fmt.Println(`llama-server-loader - Терминальный UI для управления и запуска llama-server

Использование:
  llama-server-loader [опции]

Опции:
  --scan-dir <path>        Каталог для сканирования .gguf моделей
  --model <name>           Имя модели для запуска
  --threads <count>        Количество CPU потоков (по умолчанию: авто)
  --temperature <float>    Температура сэмплинга (по умолчанию: 0.8)
  --start-webui            Запустить встроенный Web UI сервер
  --port <number>          Порт Web UI (по умолчанию: 8080)
  --save-config <file>     Сохранить конфигурацию в файл
  --generate-params        Сгенерировать конфигурацию параметров
  --output <file>          Файл вывода для сгенерированных параметров
  -h, --help               Показать справку

Примеры:
  llama-server-loader --scan-dir=./models
  llama-server-loader --scan-dir=/models --model=gemma-4
  llama-server-loader --start-webui --port=8080
  llama-server-loader --scan-dir=./models --threads=16 --temperature=0.9`)
}

func runInteractive(ctx context.Context, c *cli.CLI) error {
	log.Printf("Сканирование каталога: %s", c.ScanDir())

	// Используем CLI.Run() который уже содержит логику сканирования и UI
	if err := c.Run(); err != nil {
		return fmt.Errorf("ошибка интерактивного режима: %w", err)
	}

	// После выбора модели в UI, если модель выбрана - сохраняем конфиг
	selectedModel := c.SelectedModel()
	if selectedModel != nil && c.SaveConfig() != "" {
		cfg := &config.Config{
			Version:    "1.0",
			ScanPaths:  []string{c.ScanDir()},
			Models:     buildModelConfigs(selectedModel),
		}
		if err := config.SaveConfig(cfg, c.SaveConfig()); err != nil {
			return fmt.Errorf("ошибка сохранения конфига: %w", err)
		}
		log.Printf("Конфигурация сохранена в: %s", c.SaveConfig())
	}

	// Запускаем llama-server через servercmd если выбрана модель
	if selectedModel != nil && (c.ModelName() != "" || c.SelectedModelName() != "") {
		return startLlamaServer(ctx, selectedModel, c)
	}

	return nil
}

// startLlamaServer запускает llama-server через pkg/servercmd.
func startLlamaServer(ctx context.Context, model *modelscan.Model, c *cli.CLI) error {
	log.Printf("Запуск llama-server для модели: %s", model.Path)

	cfg := &servercmd.Config{
		Version: "1.0",
		Models: []servercmd.ModelConfig{
			{
				Name:       strings.TrimSuffix(filepath.Base(model.Path), ".gguf"),
				ModelPath:  model.Path,
				MMProjPath: "",
				MMProjOn:   false,
				Size:       model.Size,
				Flags: map[string]any{
					"threads":     c.Threads(),
					"temperature": c.Temperature(),
					"port":        8080,
				},
			},
		},
	}

	if len(model.MMProjPaths) > 0 {
		cfg.Models[0].MMProjPath = model.MMProjPaths[0]
		cfg.Models[0].MMProjOn = true
	}

	cmd, err := servercmd.BuildCommand(cfg)
	if err != nil {
		return fmt.Errorf("ошибка сборки команды: %w", err)
	}

	log.Println("Запуск llama-server...")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ошибка запуска сервера: %w", err)
	}

	// Ждем сигнал завершения
	<-ctx.Done()
	log.Println("llama-server остановлен")

	return nil
}

// buildModelConfigs создает []config.ModelConfig from selected model.
func buildModelConfigs(m *modelscan.Model) []config.ModelConfig {
	mc := config.ModelConfig{
		Name:       strings.TrimSuffix(filepath.Base(m.Path), ".gguf"),
		ModelPath:  m.Path,
		Size:       m.Size,
		MMProjPath: "",
		MMProjOn:   false,
	}

	if len(m.MMProjPaths) > 0 {
		mc.MMProjPath = m.MMProjPaths[0]
		mc.MMProjOn = true
	}

	return []config.ModelConfig{mc}
}
