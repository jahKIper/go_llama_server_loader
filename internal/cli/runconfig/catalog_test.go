package runconfig

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"llama-server-loader/internal/config"
)

// ── ParamKey ──────────────────────────────────────────────────────────────────

func TestParamKey(t *testing.T) {
	cases := []struct {
		short, long string
		want        string
	}{
		{"", "--ctx-size N", "ctx_size"},
		{"", "--temp", "temp"},
		{"", "--n-predict N", "n_predict"},
		// Только ShortFlag (--n-predict вместо long_flag)
		{"--n-predict", "", "n_predict"},
		// ShortFlag однобуквенный
		{"-c", "", "c"},
		// Long без пробела
		{"", "--no-warmup", "no_warmup"},
		// Long с несколькими словами через пробел (берём до первого)
		{"", "--spec-replace TARGET DRAFT", "spec_replace"},
	}

	for _, tc := range cases {
		meta := &config.ParamMeta{ShortFlag: tc.short, LongFlag: tc.long}
		got := ParamKey(meta)
		if got != tc.want {
			t.Errorf("ParamKey({Short:%q Long:%q}) = %q, want %q", tc.short, tc.long, got, tc.want)
		}
	}
}

// ── ResolveParamsFile ─────────────────────────────────────────────────────────

func TestResolveParamsFile_Explicit(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "params_ru.json")
	if err := os.WriteFile(f, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveParamsFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != f {
		t.Errorf("got %q, want %q", got, f)
	}
}

func TestResolveParamsFile_ExplicitMissing(t *testing.T) {
	// explicit задан, но не существует — должен пробовать fallback-и
	_, err := ResolveParamsFile("/nonexistent/path/params_ru.json")
	// Ожидаем либо ErrNotExist (если fallback тоже нет), либо путь к реальному файлу
	// В тестовой среде скорее всего ни один fallback не сработает
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Errorf("unexpected error type: %v", err)
	}
}

func TestResolveParamsFile_NotFound(t *testing.T) {
	// Переопределяем CWD во временный пустой каталог, чтобы ./params_ru.json не нашёлся
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	_, err = ResolveParamsFile("")
	if !errors.Is(err, os.ErrNotExist) {
		// Может найти файл рядом с бинарником (go test бинарник) — это тоже ok
		if err == nil {
			return
		}
		t.Errorf("expected ErrNotExist, got %v", err)
	}
}

// ── FlattenCatalog ────────────────────────────────────────────────────────────

func TestFlattenCatalog_Nil(t *testing.T) {
	entries := FlattenCatalog(nil)
	if len(entries) != 0 {
		t.Errorf("expected empty slice for nil input, got %d", len(entries))
	}
}

func TestFlattenCatalog(t *testing.T) {
	pf := &config.ParamFile{
		Categories: []config.ParamCategory{
			{
				Name: "Кат1",
				Params: []config.ParamMeta{
					{LongFlag: "--foo", ShortFlag: "-f"},
					{LongFlag: "--bar", ShortFlag: "-b"},
				},
			},
			{
				Name: "Кат2",
				Params: []config.ParamMeta{
					{LongFlag: "--baz", ShortFlag: ""},
				},
			},
		},
	}

	entries := FlattenCatalog(pf)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Category != "Кат1" || entries[0].Meta.LongFlag != "--foo" {
		t.Errorf("unexpected entry[0]: %+v", entries[0])
	}
	if entries[2].Category != "Кат2" || entries[2].Meta.LongFlag != "--baz" {
		t.Errorf("unexpected entry[2]: %+v", entries[2])
	}
}
