package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tidwall/gjson"
)

// ParamMeta describes a single parameter from params_ru.json
type ParamMeta struct {
	ShortFlag  string `json:"short_flag"`
	LongFlag   string `json:"long_flag"`
	DescRU     string `json:"description_ru"`
}

// ParamCategory groups related parameters
type ParamCategory struct {
	Name string     `json:"name"`
	Params []ParamMeta `json:"params"`
}

// ParamFile is the top-level structure for params_ru.json
type ParamFile struct {
	Version         string           `json:"version"`
	Categories      []ParamCategory  `json:"categories"`
	TotalParamsCount int64           `json:"total_params_count"`
	SourceDocs       []string         `json:"source_docs"`
}

// LoadParams reads and parses params_ru.json, returning the parsed structure.
func LoadParams(path string) (*ParamFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: cannot read file %s: %w", path, err)
	}

	result := gjson.ParseBytes(raw)
	if !result.IsObject() {
		return nil, fmt.Errorf("config: expected JSON object, got %s", result.Type)
	}

	var pf ParamFile
	if err := json.Unmarshal(raw, &pf); err != nil {
		return nil, fmt.Errorf("config: cannot unmarshal params: %w", err)
	}
	return &pf, nil
}
