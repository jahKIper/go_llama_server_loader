package modelscan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchModelAndMMproj(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "matcher-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a model directory with an mmproj file
	modelDir := filepath.Join(tmpDir, "llama")
	require.NoError(t, os.MkdirAll(modelDir, 0755))

	// Create a mmproj file
	mmprojFile := filepath.Join(modelDir, "mmproj.gguf")
	_, err = os.Create(mmprojFile)
	require.NoError(t, err)

	// Create a model file
	modelFile := filepath.Join(modelDir, "llama.gguf")
	_, err = os.Create(modelFile)
	require.NoError(t, err)

	model := &Model{
		Path: modelFile,
		Name: "llama",
	}

	matches := MatchModelAndMMproj(model)
	require.NotEmpty(t, matches)
	// FindMMprojByDirAndName returns paths normalized with filepath.ToSlash()
	expectedPath := filepath.ToSlash(mmprojFile)
	assert.Equal(t, expectedPath, matches[0])
}

func TestMatchModelAndMMproj_NoMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "matcher-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create directory without mmproj files
	require.NoError(t, os.MkdirAll(tmpDir, 0755))

	model := &Model{
		Path: tmpDir,
		Name: "test",
	}

	matches := MatchModelAndMMproj(model)
	assert.Empty(t, matches)
}

func TestMatchModelAndMMproj_MultipleMatches(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "matcher-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a model directory with multiple mmproj files
	modelDir := filepath.Join(tmpDir, "model")
	require.NoError(t, os.MkdirAll(modelDir, 0755))

	// Create two mmproj files with different names
	_, err = os.Create(filepath.Join(modelDir, "mmproj-1"))
	require.NoError(t, err)
	_, err = os.Create(filepath.Join(modelDir, "mmproj-2"))
	require.NoError(t, err)

	model := &Model{
		Path: filepath.Join(modelDir, "dummy"),
		Name: "test",
	}

	matches := MatchModelAndMMproj(model)
	assert.Len(t, matches, 2)
}

func TestFindMMprojByDirAndName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "matcher-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create multiple mmproj files with unique names
	for i := 0; i < 3; i++ {
		path := filepath.Join(tmpDir, "mmproj-test"+string(rune('a'+i)))
		_, err = os.Create(path)
		require.NoError(t, err)
	}

	matches, err := FindMMprojByDirAndName(tmpDir)
	require.NoError(t, err)
	assert.Len(t, matches, 3)
}

func TestFindMMprojByDirAndName_EmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "matcher-test")
	require.NoError(t, err)

	matches, err := FindMMprojByDirAndName(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, matches)
}

func TestIsMMproj(t *testing.T) {
	tests := []struct {
		name   string
		want   bool
	}{
		{"llama-mmproj", true},
		{"test-mmproj", true},
		{"mmproj-quant", true},
		{"model", false},
		{"quant.gguf", false},
	}

	for _, tt := range tests {
		if got := IsMMproj(tt.name); got != tt.want {
			t.Errorf("IsMMproj(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestBaseContainsMMProj(t *testing.T) {
	tests := []struct {
		name   string
		want   bool
	}{
		{"llama-mmproj", true},
		{"test-mmproj", true},
		{"mmproj-quant", true},
		{"mmproj", true},
		{"model", false},
		{"quant.gguf", false},
		{"gemma-7b", false},
	}

	for _, tt := range tests {
		if got := BaseContainsMMProj(tt.name); got != tt.want {
			t.Errorf("BaseContainsMMProj(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestMatchMMProj(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "matcher-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	modelDir := filepath.Join(tmpDir, "test-model")
	require.NoError(t, os.MkdirAll(modelDir, 0755))

	// Create model files
	_, err = os.Create(filepath.Join(modelDir, "model-a.gguf"))
	require.NoError(t, err)
	_, err = os.Create(filepath.Join(modelDir, "model-b.gguf"))
	require.NoError(t, err)

	// Create mmproj file
	_, err = os.Create(filepath.Join(modelDir, "mmproj.gguf"))
	require.NoError(t, err)

	models := []*Model{
		{Name: "model-a", Path: filepath.Join(modelDir, "model-a.gguf")},
		{Name: "model-b", Path: filepath.Join(modelDir, "model-b.gguf")},
	}

	result, err := MatchMMProj(models)
	require.NoError(t, err)
	require.Len(t, result, 2)

	// Each model should have mmproj_paths populated
	for _, m := range result {
		require.NotEmpty(t, m.MMProjPaths)
	}
}

func TestMatchMMProj_EmptyModels(t *testing.T) {
	result, err := MatchMMProj([]*Model{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestMatchMMProj_MultipleDirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "matcher-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create two directories with different mmproj files
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	require.NoError(t, os.MkdirAll(dir1, 0755))
	require.NoError(t, os.MkdirAll(dir2, 0755))

	// dir1 has mmproj-1
	_, err = os.Create(filepath.Join(dir1, "mmproj-1"))
	require.NoError(t, err)
	// dir2 has mmproj-2
	_, err = os.Create(filepath.Join(dir2, "mmproj-2"))
	require.NoError(t, err)

	// Create model files
	_, err = os.Create(filepath.Join(dir1, "model.gguf"))
	require.NoError(t, err)
	_, err = os.Create(filepath.Join(dir2, "model.gguf"))
	require.NoError(t, err)

	models := []*Model{
		{Name: "model1", Path: filepath.Join(dir1, "model.gguf")},
		{Name: "model2", Path: filepath.Join(dir2, "model.gguf")},
	}

	result, err := MatchMMProj(models)
	require.NoError(t, err)
	require.Len(t, result, 2)

	// Each model should have exactly one mmproj path from its directory
	for _, m := range result {
		require.Len(t, m.MMProjPaths, 1)
	}
}
