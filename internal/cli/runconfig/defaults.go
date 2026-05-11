package runconfig

import (
	"log"
	"strconv"

	"llama-server-loader/internal/cli/modelparams"
	"llama-server-loader/internal/config"
	"llama-server-loader/internal/ggufmeta"
	"llama-server-loader/pkg/modelscan"
)

// ngpuFallback — значение для --gpu-layers, когда block_count в GGUF недоступен.
// По договорённости: 999 → llama-server сам срежет до фактического числа слоёв.
const ngpuFallback = 999

// ncmoeDefault — дефолтное значение -ncmoe / --n-cpu-moe для MoE-моделей.
const ncmoeDefault = 2

// threadsDefault — дефолтное значение для --threads.
const threadsDefault = 10

// ComputeModelDefaults возвращает строки с автодефолтами для запуска модели.
// Применяется только при первом открытии экрана (когда в models.json ещё нет
// записи). Все возвращённые строки помечены IsDefault=true.
//
// Источники значений:
//   - GGUF-зависимые: --ctx-size = context_length/2, --gpu-layers = block_count,
//     --n-cpu-moe = 2 (если expert_count > 0).
//   - Фиксированные: --cache-type-k = q4_0, --cache-type-v = q4_0,
//     --threads = 10.
//
// Параметры:
//   - catalog: плоский каталог из params_ru.json — нужен для резолва Meta по
//     long-флагу. Если флаг не найден в каталоге, соответствующая строка
//     пропускается (без Meta мы не сможем построить корректный Key).
//   - m: просканированная модель (для отсева, если nil).
//   - lookup: GGUF-метаданные, может быть nil — тогда GGUF-зависимые дефолты
//     пропускаются.
func ComputeModelDefaults(catalog []CatalogEntry, m *modelscan.Model, lookup *modelparams.Lookup) []ParamRow {
	if m == nil {
		return nil
	}
	var curated modelparams.Curated
	if lookup != nil {
		curated = lookup.ForPathCurated(m.Path)
	}
	// Если в lookup-е модели нет (первое открытие — запись ещё не сохранена),
	// читаем GGUF напрямую с диска.
	if curated.TotalCount == 0 {
		if ps, err := ggufmeta.ExtractParams(m.Path); err == nil {
			mps := make([]config.ModelParam, len(ps))
			for i, p := range ps {
				mps[i] = config.ModelParam{Key: p.Key, Value: p.Value, DescriptionRU: p.DescriptionRU}
			}
			curated = modelparams.ExtractCurated(mps)
		} else {
			log.Printf("runconfig: defaults: read GGUF %s: %v", m.Path, err)
		}
	}

	var rows []ParamRow
	add := func(longFlag, value string) {
		meta := FindMetaByLong(catalog, longFlag)
		if meta == nil {
			return
		}
		long := stripFlagArg(meta.LongFlag)
		if long == "" {
			return
		}
		rows = append(rows, ParamRow{
			Long:      long,
			Short:     stripFlagArg(meta.ShortFlag),
			Key:       ParamKey(meta),
			Value:     value,
			Meta:      meta,
			IsDefault: true,
		})
	}

	// --ctx-size = context_length / 2
	if curated.ContextLength > 0 {
		add("--ctx-size", strconv.FormatInt(curated.ContextLength/2, 10))
	}

	// --gpu-layers = block_count (fallback 999)
	ngl := curated.BlockCount
	if ngl <= 0 {
		ngl = ngpuFallback
	}
	add("--gpu-layers", strconv.FormatInt(ngl, 10))

	// --n-cpu-moe для MoE
	if curated.ExpertCount > 0 {
		add("--n-cpu-moe", strconv.Itoa(ncmoeDefault))
	}

	// Фиксированные дефолты
	add("--cache-type-k", "q4_0")
	add("--cache-type-v", "q4_0")
	add("--threads", strconv.Itoa(threadsDefault))

	return rows
}
