package runconfig

import (
	"testing"

	"llama-server-loader/internal/cli/modelparams"
	"llama-server-loader/internal/config"
	"llama-server-loader/pkg/modelscan"
)

// fakeCatalog собирает минимальный каталог для тестов — без чтения params_ru.json.
func fakeCatalog() []CatalogEntry {
	metas := []config.ParamMeta{
		{LongFlag: "--ctx-size N", ShortFlag: "-c"},
		{LongFlag: "--gpu-layers N"},
		{LongFlag: "--n-cpu-moe N"},
		{LongFlag: "--cache-type-k TYPE", ShortFlag: "-ctk"},
		{LongFlag: "--cache-type-v TYPE", ShortFlag: "-ctv"},
		{LongFlag: "--threads N", ShortFlag: "-t"},
	}
	out := make([]CatalogEntry, len(metas))
	for i := range metas {
		out[i] = CatalogEntry{Category: "test", Meta: &metas[i]}
	}
	return out
}

// makeLookupWithParams строит Lookup на основе фейковой записи в config.
func makeLookupWithParams(name string, kv []config.ModelParam) *modelparams.Lookup {
	cfg := &config.Config{
		Version: "1.0",
		Models:  []config.ModelConfig{{Name: name, ModelPath: "/m/" + name + ".gguf", Params: kv}},
	}
	return modelparams.NewLookup(cfg)
}

func rowByLong(rows []ParamRow, long string) *ParamRow {
	for i := range rows {
		if rows[i].Long == long {
			return &rows[i]
		}
	}
	return nil
}

func TestComputeModelDefaults_FromLookup(t *testing.T) {
	m := &modelscan.Model{Path: "/m/qwen3.gguf"}
	lookup := makeLookupWithParams("qwen3", []config.ModelParam{
		{Key: "general.architecture", Value: "qwen3"},
		{Key: "qwen3.context_length", Value: int64(160000)},
		{Key: "qwen3.block_count", Value: int64(40)},
		{Key: "qwen3.expert_count", Value: int64(64)},
	})

	rows := ComputeModelDefaults(fakeCatalog(), m, lookup)
	if len(rows) == 0 {
		t.Fatal("expected non-empty defaults")
	}
	for _, r := range rows {
		if !r.IsDefault {
			t.Errorf("row %s: expected IsDefault=true", r.Long)
		}
	}

	if r := rowByLong(rows, "--ctx-size"); r == nil || r.Value != "80000" {
		t.Errorf("ctx-size: got %+v, want value=80000", r)
	}
	if r := rowByLong(rows, "--gpu-layers"); r == nil || r.Value != "40" {
		t.Errorf("gpu-layers: got %+v, want value=40", r)
	}
	if r := rowByLong(rows, "--n-cpu-moe"); r == nil || r.Value != "2" {
		t.Errorf("n-cpu-moe: got %+v, want value=2 (MoE detected)", r)
	}
	if r := rowByLong(rows, "--cache-type-k"); r == nil || r.Value != "q4_0" {
		t.Errorf("cache-type-k: got %+v, want q4_0", r)
	}
	if r := rowByLong(rows, "--cache-type-v"); r == nil || r.Value != "q4_0" {
		t.Errorf("cache-type-v: got %+v, want q4_0", r)
	}
	if r := rowByLong(rows, "--threads"); r == nil || r.Value != "10" {
		t.Errorf("threads: got %+v, want 10", r)
	}
}

func TestComputeModelDefaults_NonMoE_OmitsNcmoe(t *testing.T) {
	m := &modelscan.Model{Path: "/m/llama.gguf"}
	lookup := makeLookupWithParams("llama", []config.ModelParam{
		{Key: "general.architecture", Value: "llama"},
		{Key: "llama.context_length", Value: int64(8192)},
		{Key: "llama.block_count", Value: int64(32)},
	})
	rows := ComputeModelDefaults(fakeCatalog(), m, lookup)
	if r := rowByLong(rows, "--n-cpu-moe"); r != nil {
		t.Errorf("non-MoE model should not include --n-cpu-moe, got %+v", r)
	}
}

func TestComputeModelDefaults_NoBlockCount_FallbackNGL(t *testing.T) {
	m := &modelscan.Model{Path: "/m/x.gguf"}
	lookup := makeLookupWithParams("x", []config.ModelParam{
		{Key: "general.architecture", Value: "x"},
		// context_length и block_count отсутствуют
	})
	rows := ComputeModelDefaults(fakeCatalog(), m, lookup)
	if r := rowByLong(rows, "--gpu-layers"); r == nil || r.Value != "999" {
		t.Errorf("gpu-layers fallback: got %+v, want 999", r)
	}
	if r := rowByLong(rows, "--ctx-size"); r != nil {
		t.Errorf("ctx-size should be omitted when context_length is missing, got %+v", r)
	}
}

func TestComputeModelDefaults_NilModel(t *testing.T) {
	if got := ComputeModelDefaults(fakeCatalog(), nil, nil); got != nil {
		t.Errorf("expected nil for nil model, got %+v", got)
	}
}
