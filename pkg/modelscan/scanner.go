package modelscan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

const (
	// maxConcurrency определяет максимальное количество горутин для сканирования.
	maxConcurrency = 4
)

// Model представляет собой обнаруженную модель или mmproj файл.
type Model struct {
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	IsMMProj   bool    `json:"is_mmproj"`
	Size       int64   `json:"size"`
	MMProjPaths []string `json:"mmproj_paths,omitempty"`
}

// ScanResult содержит результаты сканирования директории.
type ScanResult struct {
	Models   []*Model `json:"models"`
	MMModels []*Model `json:"mm_models"`
	Errors   []error  `json:"-"`
	mu       sync.Mutex
}

// Add добавляет модель в результат потокобезопасно.
func (r *ScanResult) Add(m *Model) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m.IsMMProj {
		r.MMModels = append(r.MMModels, m)
	} else {
		r.Models = append(r.Models, m)
	}
}

// AddError добавляет ошибку в результат потокобезопасно.
func (r *ScanResult) AddError(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err != nil {
		r.Errors = append(r.Errors, err)
	}
}

// IsMMProj проверяет, является ли файл mmproj по имени.
func IsMMProj(name string) bool {
	return strings.Contains(strings.ToLower(name), "mmproj")
}

// isGGUF проверяет, является ли файл gguf по расширению.
func isGGUF(name string) bool {
	return strings.EqualFold(filepath.Ext(name), ".gguf")
}

// ScanDir выполняет рекурсивное сканирование директории на наличие .gguf файлов
// с использованием до 4 горутин для параллельного обхода.
func ScanDir(root string) (*ScanResult, error) {
	result := &ScanResult{
		Models:   make([]*Model, 0),
		MMModels: make([]*Model, 0),
		Errors:   make([]error, 0),
	}

	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("cannot access root directory %q: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root path is not a directory: %q", root)
	}

	// Собираем все файлы сначала синхронно для предсказуемости
	var allFiles []string
	var walkErr error

	walkErr = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.AddError(fmt.Errorf("error walking path %q: %w", path, err))
			return nil // Продолжаем даже при ошибках
		}
		if !info.IsDir() && isGGUF(info.Name()) {
			allFiles = append(allFiles, path)
		}
		return nil
	})

	if walkErr != nil {
		result.AddError(fmt.Errorf("error walking directory tree: %w", walkErr))
	}

	// Если файлов нет — возвращаем результат сразу
	if len(allFiles) == 0 {
		return result, nil
	}

	// Параллельно обрабатываем найденные файлы с ограничением 4 горутин
	var wg errgroup.Group
	wg.SetLimit(maxConcurrency)

	for _, filePath := range allFiles {
		fp := filePath // для замыкания в горутине
		wg.Go(func() error {
			info, statErr := os.Stat(fp)
			if statErr != nil {
				result.AddError(fmt.Errorf("cannot stat file %q: %w", fp, statErr))
				return nil
			}

			model := &Model{
				Name:     info.Name(),
				Path:     filepath.ToSlash(fp),
				IsMMProj: IsMMProj(info.Name()),
				Size:     info.Size(),
			}

			result.Add(model)
			return nil
		})
	}

	waitErr := wg.Wait()
	if waitErr != nil {
		result.AddError(waitErr)
	}

	return result, nil
}
