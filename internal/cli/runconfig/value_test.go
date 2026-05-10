package runconfig

import (
	"testing"
)

func TestParseFlagValue(t *testing.T) {
	tests := []struct {
		raw  string
		want any
	}{
		{"", ""},
		{"  ", ""},
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"false", false},
		{"False", false},
		{"FALSE", false},
		{"42", 42},
		{"-1", -1},
		{"0", 0},
		{"3.14", 3.14},
		{"-0.5", -0.5},
		{"1e2", 100.0},
		{"hello", "hello"},
		{"--flag", "--flag"},
		{"some string with spaces", "some string with spaces"},
	}

	for _, tc := range tests {
		got := ParseFlagValue(tc.raw)
		if got != tc.want {
			t.Errorf("ParseFlagValue(%q) = %v (%T), want %v (%T)", tc.raw, got, got, tc.want, tc.want)
		}
	}
}

func TestBuildFlagsMap(t *testing.T) {
	t.Run("adds model path when not present", func(t *testing.T) {
		rows := []ParamRow{
			{Key: "ctx_size", Value: "4096"},
		}
		m := BuildFlagsMap(rows, "/models/my.gguf")
		if m["model"] != "/models/my.gguf" {
			t.Errorf("expected model=/models/my.gguf, got %v", m["model"])
		}
		if m["ctx_size"] != 4096 {
			t.Errorf("expected ctx_size=4096 (int), got %v (%T)", m["ctx_size"], m["ctx_size"])
		}
	})

	t.Run("does not override explicit model key", func(t *testing.T) {
		rows := []ParamRow{
			{Key: "model", Value: "/custom/path.gguf"},
		}
		m := BuildFlagsMap(rows, "/default/path.gguf")
		if m["model"] != "/custom/path.gguf" {
			t.Errorf("expected model=/custom/path.gguf, got %v", m["model"])
		}
	})

	t.Run("skips rows with empty key", func(t *testing.T) {
		rows := []ParamRow{
			{Key: "", Value: "ignored"},
			{Key: "temp", Value: "0.7"},
		}
		m := BuildFlagsMap(rows, "/m.gguf")
		if len(m) != 2 { // temp + model
			t.Errorf("expected 2 entries, got %d: %v", len(m), m)
		}
	})

	t.Run("bool value", func(t *testing.T) {
		rows := []ParamRow{
			{Key: "flash_attn", Value: "true"},
		}
		m := BuildFlagsMap(rows, "/m.gguf")
		if m["flash_attn"] != true {
			t.Errorf("expected flash_attn=true, got %v", m["flash_attn"])
		}
	})

	t.Run("float value", func(t *testing.T) {
		rows := []ParamRow{
			{Key: "temp", Value: "0.8"},
		}
		m := BuildFlagsMap(rows, "/m.gguf")
		if m["temp"] != 0.8 {
			t.Errorf("expected temp=0.8, got %v", m["temp"])
		}
	})

	t.Run("string value", func(t *testing.T) {
		rows := []ParamRow{
			{Key: "host", Value: "0.0.0.0"},
		}
		m := BuildFlagsMap(rows, "/m.gguf")
		if m["host"] != "0.0.0.0" {
			t.Errorf("expected host=0.0.0.0, got %v", m["host"])
		}
	})

	t.Run("empty value becomes empty string", func(t *testing.T) {
		rows := []ParamRow{
			{Key: "grammar", Value: ""},
		}
		m := BuildFlagsMap(rows, "/m.gguf")
		if m["grammar"] != "" {
			t.Errorf("expected grammar=\"\", got %v", m["grammar"])
		}
	})
}
