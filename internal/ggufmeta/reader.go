// Package ggufmeta — обёртка над github.com/gpustack/gguf-parser-go,
// которая отдаёт KV-метаданные в виде «дружелюбного» среза для сериализации в JSON.
package ggufmeta

import (
	"fmt"

	gguf "github.com/gpustack/gguf-parser-go"
)

// KV — одна запись метаданных, готовая к сериализации.
type KV struct {
	Key   string
	Type  string
	Value any
}

// File — упрощённый view на GGUFFile.
type File struct {
	Path        string
	Size        int64
	Version     uint32
	TensorCount uint64
	KV          []KV
	index       map[string]int
}

func (f *File) Get(key string) (any, bool) {
	if i, ok := f.index[key]; ok {
		return f.KV[i].Value, true
	}
	return nil, false
}

// ReadFile парсит GGUF и возвращает плоский список KV.
// Большие массивы (tokenizer.ggml.tokens и т.д.) сворачиваются в placeholder.
func ReadFile(path string) (*File, error) {
	gf, err := gguf.ParseGGUFFile(path, gguf.SkipCache())
	if err != nil {
		return nil, err
	}

	f := &File{
		Path:        path,
		Size:        int64(gf.Size),
		Version:     uint32(gf.Header.Version),
		TensorCount: gf.Header.TensorCount,
		index:       map[string]int{},
	}

	for _, kv := range gf.Header.MetadataKV {
		f.index[kv.Key] = len(f.KV)
		f.KV = append(f.KV, KV{
			Key:   kv.Key,
			Type:  kv.ValueType.String(),
			Value: normalizeValue(kv),
		})
	}
	return f, nil
}

// normalizeValue приводит значение к удобному для JSON виду.
func normalizeValue(kv gguf.GGUFMetadataKV) any {
	switch kv.ValueType {
	case gguf.GGUFMetadataValueTypeArray:
		av := kv.ValueArray()
		// массивы строк/чисел могут быть огромными — отдаём placeholder
		if av.Len > 1024 {
			return fmt.Sprintf("<array len=%d type=%s>", av.Len, av.Type.String())
		}
		return av.Array
	default:
		return kv.Value
	}
}
