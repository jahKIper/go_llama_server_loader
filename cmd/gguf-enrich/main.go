// gguf-enrich читает models.json, для каждой модели парсит GGUF и заполняет
// поле "params" метаданными с русскими описаниями. Перезаписывает существующие.
//
// Использование:
//   go run ./cmd/gguf-enrich [-f models.json] [-o models.json]
package main

import (
	"flag"
	"fmt"
	"os"

	"llama-server-loader/internal/config"
	"llama-server-loader/internal/ggufmeta"
)

func main() {
	in := flag.String("f", "models.json", "путь к models.json")
	out := flag.String("o", "", "куда писать (по умолчанию = -f)")
	flag.Parse()
	if *out == "" {
		*out = *in
	}

	cfg, err := config.LoadConfig(*in)
	must(err)

	for i := range cfg.Models {
		m := &cfg.Models[i]
		ps, err := ggufmeta.ExtractParams(m.ModelPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", m.Name, err)
			continue
		}
		m.Params = make([]config.ModelParam, len(ps))
		for j, p := range ps {
			m.Params[j] = config.ModelParam{Key: p.Key, Value: p.Value, DescriptionRU: p.DescriptionRU}
		}
		fmt.Fprintf(os.Stderr, "ok  %s: %d params\n", m.Name, len(m.Params))
	}

	must(config.SaveConfig(cfg, *out))
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
