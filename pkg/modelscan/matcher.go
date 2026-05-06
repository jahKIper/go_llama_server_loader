package modelscan

import (
	"path/filepath"
	"strings"
)

// MatchModelAndMMproj matches gguf files with mmproj files found in the same directory
// where the filename contains 'mmproj'.
func MatchModelAndMMproj(gguf *Model) []string {
	dir := filepath.Dir(gguf.Path)
	matches, _ := FindMMprojByDirAndName(dir)
	return matches
}

// FindMMprojByDirAndName searches a given directory for files whose name contains "mmproj"
// and returns them.
func FindMMprojByDirAndName(dir string) ([]string, error) {
	entries, err := filepath.Glob(filepath.Join(dir, "*mmproj*"))
	if err != nil {
		return nil, err
	}
	var results []string
	for _, e := range entries {
		results = append(results, filepath.ToSlash(e))
	}
	return results, nil
}

// IsMMproj checks if a filename contains "mmproj"
func IsMMproj(name string) bool {
	return strings.Contains(name, "mmproj")
}

// BaseContainsMMProj checks if a base path contains "mmproj" in its basename or directory name.
func BaseContainsMMProj(base string) bool {
	base = filepath.Base(base)
	return strings.Contains(strings.ToLower(base), "mmproj")
}

// MatchMMProj matches gguf models with mmproj files based on directory and name matching.
// It groups models by common prefixes and returns enriched models with mmproj paths.
func MatchMMProj(models []*Model) ([]*Model, error) {
	// Build a map of directory -> mmproj files
	dirMMProjMap := make(map[string][]string)

	// Collect mmproj files from all directories
	for _, model := range models {
		dir := filepath.Dir(model.Path)
		if _, exists := dirMMProjMap[dir]; !exists {
			matches, err := FindMMprojByDirAndName(dir)
			if err != nil {
				return nil, err
			}
			dirMMProjMap[dir] = matches
		}
	}

	// Match models with mmproj files
	var result []*Model
	for _, model := range models {
		dir := filepath.Dir(model.Path)
		mmprojPaths := dirMMProjMap[dir]

		enriched := *model
		enriched.MMProjPaths = mmprojPaths
		result = append(result, &enriched)
	}

	return result, nil
}
