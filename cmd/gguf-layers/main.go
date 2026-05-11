// gguf-layers печатает по-слоевую разбивку тензоров GGUF-файла:
// для каждого блока (blk.N.*) — список тензоров, их типы (квантизация),
// размерности и итоговый размер в байтах. Плюс — non-block тензоры
// (token_embd, output_norm, output, …).
//
// Использование:
//   go run ./cmd/gguf-layers <path-to.gguf> [-v]
//   go run ./cmd/gguf-layers <path-to.gguf> -summary
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	gguf "github.com/gpustack/gguf-parser-go"
)

var blkRe = regexp.MustCompile(`^blk\.(\d+)\.(.+)$`)

type tensor struct {
	name  string
	suf   string
	typ   string
	dims  []uint64
	bytes uint64
}

func main() {
	verbose := flag.Bool("v", false, "печатать каждый тензор слоя")
	summary := flag.Bool("summary", false, "только сводка по слоям без деталей")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "использование: gguf-layers <file.gguf> [-v] [-summary]")
		os.Exit(2)
	}

	f, err := gguf.ParseGGUFFile(flag.Arg(0), gguf.SkipCache())
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	byBlock := map[int][]tensor{}
	var nonBlock []tensor
	var totalBytes uint64

	for _, ti := range f.TensorInfos {
		t := tensor{
			name:  ti.Name,
			typ:   ti.Type.String(),
			dims:  ti.Dimensions,
			bytes: ti.Bytes(),
		}
		totalBytes += t.bytes
		if m := blkRe.FindStringSubmatch(ti.Name); m != nil {
			idx, _ := strconv.Atoi(m[1])
			t.suf = m[2]
			byBlock[idx] = append(byBlock[idx], t)
		} else {
			nonBlock = append(nonBlock, t)
		}
	}

	// --- Non-block тензоры ---
	if !*summary {
		fmt.Printf("== Non-block tensors (%d) ==\n", len(nonBlock))
		sort.Slice(nonBlock, func(i, j int) bool { return nonBlock[i].name < nonBlock[j].name })
		for _, t := range nonBlock {
			fmt.Printf("  %-30s %-8s %-20s %s\n", t.name, t.typ, dimStr(t.dims), humanBytes(t.bytes))
		}
		fmt.Println()
	}

	// --- Per-block ---
	blocks := make([]int, 0, len(byBlock))
	for k := range byBlock {
		blocks = append(blocks, k)
	}
	sort.Ints(blocks)

	fmt.Printf("== Blocks (%d) ==\n", len(blocks))
	// header сводки
	fmt.Printf("%-5s %-7s %-20s %s\n", "blk", "tnsrs", "size", "quant types")
	for _, idx := range blocks {
		ts := byBlock[idx]
		var bsum uint64
		typeCount := map[string]int{}
		for _, t := range ts {
			bsum += t.bytes
			typeCount[t.typ]++
		}
		fmt.Printf("%-5d %-7d %-20s %s\n", idx, len(ts), humanBytes(bsum), typesStr(typeCount))
		if *verbose {
			sort.Slice(ts, func(i, j int) bool { return ts[i].suf < ts[j].suf })
			for _, t := range ts {
				fmt.Printf("        %-30s %-8s %-20s %s\n", t.suf, t.typ, dimStr(t.dims), humanBytes(t.bytes))
			}
		}
	}

	fmt.Println()
	fmt.Printf("== Total ==\n")
	fmt.Printf("  tensors:        %d\n", len(f.TensorInfos))
	fmt.Printf("  blocks:         %d\n", len(blocks))
	fmt.Printf("  non-block:      %d\n", len(nonBlock))
	fmt.Printf("  data bytes:     %s\n", humanBytes(totalBytes))
	fmt.Printf("  file bytes:     %s\n", humanBytes(uint64(f.Size)))
}

func dimStr(d []uint64) string {
	s := make([]string, len(d))
	for i, x := range d {
		s[i] = strconv.FormatUint(x, 10)
	}
	return "[" + strings.Join(s, "×") + "]"
}

func typesStr(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%s×%d", k, m[k])
	}
	return strings.Join(parts, " ")
}

func humanBytes(b uint64) string {
	const (
		KiB = 1024
		MiB = 1024 * KiB
		GiB = 1024 * MiB
	)
	switch {
	case b >= GiB:
		return fmt.Sprintf("%.2f GiB", float64(b)/GiB)
	case b >= MiB:
		return fmt.Sprintf("%.2f MiB", float64(b)/MiB)
	case b >= KiB:
		return fmt.Sprintf("%.2f KiB", float64(b)/KiB)
	default:
		return fmt.Sprintf("%d B", b)
	}
}
