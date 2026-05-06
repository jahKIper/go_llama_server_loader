package config

import (
	"fmt"
)

// ValidateParams validates a loaded params file for basic correctness.
func ValidateParams(pf *ParamFile) error {
	if pf == nil {
		return fmt.Errorf("config: params file is nil")
	}
	if len(pf.Categories) == 0 {
		return fmt.Errorf("config: no categories found")
	}
	totalCount := 0
	for _, cat := range pf.Categories {
		for _, p := range cat.Params {
			if p.ShortFlag == "" && p.LongFlag == "" {
				return fmt.Errorf("config: param has empty short_flag and long_flag")
			}
			if p.DescRU == "" {
				return fmt.Errorf("config: param has empty description")
			}
			totalCount++
		}
	}
	if totalCount == 0 {
		return fmt.Errorf("config: no params found")
	}
	return nil
}
