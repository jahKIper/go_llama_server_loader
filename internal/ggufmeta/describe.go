package ggufmeta

import (
	"strings"
)

// FileTypeName — отображение general.file_type в имя кванта (llama_ftype).
var FileTypeName = map[uint64]string{
	0: "F32", 1: "F16",
	2: "Q4_0", 3: "Q4_1",
	7: "Q8_0", 8: "Q5_0", 9: "Q5_1",
	10: "Q2_K", 11: "Q3_K_XS", 12: "Q3_K_S", 13: "Q3_K_M", 14: "Q3_K_L",
	15: "Q4_K_M", 16: "Q5_K_S", 17: "Q5_K_M", 18: "Q6_K",
	19: "IQ2_XXS", 20: "IQ2_XS", 21: "Q2_K_S",
	22: "IQ3_XS", 23: "IQ3_XXS", 24: "IQ1_S", 25: "IQ4_NL",
	26: "IQ3_S", 27: "IQ3_M", 28: "IQ2_S", 29: "IQ2_M",
	30: "IQ4_XS", 31: "IQ1_M", 32: "BF16",
	33: "TQ1_0", 34: "TQ2_0",
	1000: "GUESSED",
}

// staticDesc — описания для ключей без архитектурного префикса.
var staticDesc = map[string]string{
	// general.*
	"general.architecture":         `Идентификатор архитектуры (llama, qwen3, gemma3, gemma4, qwen35 и т.п.) — определяет LLM_ARCH в llama.cpp`,
	"general.basename":             `Базовое имя модели без суффиксов квантизации/размера (например, "Gemma-4-E4B-it")`,
	"general.size_label":           `Метка размера модели ("4B", "9B", "70B"), используется в имени файла`,
	"general.type":                 `Тип артефакта: "model", "adapter" (LoRA) и т.п.`,
	"general.name":                 `Человекочитаемое имя модели`,
	"general.author":               `Автор модели`,
	"general.version":              `Версия модели`,
	"general.organization":         `Организация-разработчик`,
	"general.license":              `Лицензия модели`,
	"general.license.name":         `Название лицензии`,
	"general.license.link":         `Ссылка на текст лицензии`,
	"general.url":                  `URL модели`,
	"general.description":          `Описание модели`,
	"general.repo_url":             `URL репозитория`,
	"general.source.url":           `URL исходной модели`,
	"general.source.huggingface.repository": `Hugging Face репозиторий источника`,
	"general.file_type":            `Числовой код общей квантизации файла (enum LLAMA_FTYPE_*: 0=F32, 1=F16, 2=Q4_0, 7=Q8_0, 15=Q4_K_M, 18=Q6_K и т.д.)`,
	"general.quantization_version": `Версия формата квантизации GGUF (текущая — 2)`,
	"general.alignment":            `Выравнивание тензорных данных в файле, байт`,
	"general.languages":            `Поддерживаемые языки`,
	"general.tags":                 `Теги модели`,
	"general.datasets":             `Использованные датасеты`,

	// Рекомендуемые семплер-дефолты (новые ключи llama.cpp)
	"general.sampling.temp":  `Рекомендуемая авторами модели температура семплирования по умолчанию`,
	"general.sampling.top_k": `Рекомендуемый top-k по умолчанию`,
	"general.sampling.top_p": `Рекомендуемый top-p (nucleus) по умолчанию`,

	// tokenizer.*
	"tokenizer.chat_template":            `Jinja-шаблон чат-промпта (форматирование ролей user/assistant/system)`,
	"tokenizer.chat_template.tool_use":   `Jinja-шаблон чата с поддержкой инструментов (tool/function calling)`,
	"tokenizer.ggml.model":               `Тип токенизатора: "gpt2" (BPE), "llama" (SentencePiece), "bert" (WordPiece), "no_vocab" и т.д.`,
	"tokenizer.ggml.pre":                 `Имя pre-tokenizer'а для BPE-моделей ("llama-bpe", "qwen2", "gemma" и т.д.) — задаёт regex-разбиение текста перед BPE`,
	"tokenizer.ggml.tokens":              `Массив всех токенов словаря (строки)`,
	"tokenizer.ggml.scores":              `Массив log-prob/score для каждого токена (используется SentencePiece-моделями)`,
	"tokenizer.ggml.token_type":          `Массив типов токенов (1=normal, 2=unknown, 3=control, 4=user-defined, 5=unused, 6=byte)`,
	"tokenizer.ggml.merges":              `Массив BPE-merge правил ("a b" → "ab") для BPE-токенизаторов`,
	"tokenizer.ggml.bos_token_id":        `ID токена начала последовательности (BOS)`,
	"tokenizer.ggml.eos_token_id":        `ID токена конца последовательности (EOS)`,
	"tokenizer.ggml.eot_token_id":        `ID токена конца хода (EOT, end-of-turn)`,
	"tokenizer.ggml.eom_token_id":        `ID токена конца сообщения (EOM, end-of-message)`,
	"tokenizer.ggml.unknown_token_id":    `ID unknown-токена (UNK)`,
	"tokenizer.ggml.separator_token_id":  `ID токена-разделителя (SEP)`,
	"tokenizer.ggml.padding_token_id":    `ID padding-токена`,
	"tokenizer.ggml.cls_token_id":        `ID токена CLS (BERT)`,
	"tokenizer.ggml.mask_token_id":       `ID mask-токена (BERT-подобные модели для MLM)`,
	"tokenizer.ggml.add_bos_token":       `Добавлять ли BOS автоматически при токенизации`,
	"tokenizer.ggml.add_eos_token":       `Добавлять ли EOS автоматически при токенизации`,
	"tokenizer.ggml.add_space_prefix":    `Добавлять ли пробел в начало текста перед токенизацией (особенность SentencePiece у LLaMA)`,
	"tokenizer.ggml.remove_extra_whitespaces": `Удалять лишние пробелы перед токенизацией`,
	"tokenizer.ggml.precompiled_charsmap":     `Скомпилированная карта символов (Unicode-нормализация)`,

	// quantize.imatrix.*
	"quantize.imatrix.file":          `Файл importance-matrix, использованный при квантизации`,
	"quantize.imatrix.dataset":       `Датасет для importance-matrix`,
	"quantize.imatrix.entries_count": `Число записей в importance-matrix`,
	"quantize.imatrix.chunks_count":  `Число чанков в importance-matrix`,

	// split.*
	"split.no":            `Номер сплита (для разбитых на несколько файлов моделей)`,
	"split.count":         `Общее число сплитов`,
	"split.tensors.count": `Число тензоров в сплите`,
}

// suffixDesc — описания по суффиксу после первой точки в ключе ({arch}.<suffix>).
// Применяются ко всем архитектурам.
var suffixDesc = map[string]string{
	// LLM-топология
	"context_length":                    `Максимальная длина контекста в токенах, на которую модель обучена (training-time, без YARN/rope-scaling)`,
	"embedding_length":                  `Размерность скрытого состояния (hidden size, d_model)`,
	"block_count":                       `Количество трансформерных блоков (слоёв)`,
	"feed_forward_length":               `Размерность промежуточного слоя FFN (intermediate size)`,
	"vocab_size":                        `Размер словаря токенизатора`,
	"embedding_length_per_layer_input":  `Per-layer input embedding (PLE) — специфика Gemma 3n/E4B: для каждого слоя добавляется отдельный вектор поверх общего эмбеддинга (обычно d=256)`,
	"final_logit_softcapping":           `Коэффициент soft-cap для выходных логитов (logits = cap·tanh(logits/cap); типично 30 у Gemma)`,
	"attn_logit_softcapping":            `Soft-cap для attention-логитов (применяется до softmax; типично 50 у Gemma)`,
	"full_attention_interval":           `Интервал full-attention слоёв в гибридной SSM/Attn архитектуре (Qwen3-Next/3.5): full-attn только на каждом N-м слое (обычно 4), остальные — SSM`,
	"expert_feed_forward_length":        `FFN intermediate size для экспертов MoE`,
	"expert_shared_feed_forward_length": `FFN intermediate size для shared-эксперта`,
	"use_parallel_residual":             `Использовать parallel residual (Falcon/MPT-style)`,
	"tensor_data_layout":                `Layout тензорных данных`,
	"pooling_type":                      `Тип пулинга эмбеддингов (для embedding-моделей)`,
	"logit_scale":                       `Множитель выходных логитов`,
	"decoder_start_token_id":            `Стартовый токен декодера (encoder-decoder модели)`,
	"causal":                            `Causal LM (true для декодеров)`,

	// Attention
	"attention.head_count":              `Число голов внимания (Q)`,
	"attention.head_count_kv":           `Число KV-голов (для GQA/MQA; равно head_count при обычном MHA)`,
	"attention.key_length":              `Размерность одной K-головы (head_dim для ключей) в обычных (global/full) слоях`,
	"attention.value_length":            `Размерность одной V-головы в обычных (global/full) слоях`,
	"attention.key_length_swa":          `Размерность K-головы в слоях со sliding-window attention (Gemma 3+: у SWA-слоёв может отличаться от global)`,
	"attention.value_length_swa":        `Размерность V-головы в слоях со sliding-window attention`,
	"attention.layer_norm_epsilon":      `Epsilon для LayerNorm (численная стабильность)`,
	"attention.layer_norm_rms_epsilon":  `Epsilon в RMSNorm перед attention/FFN (типично 1e-5…1e-6)`,
	"attention.sliding_window":          `Размер окна для sliding-window attention в токенах (Gemma 3 — 1024, Gemma 3n — 512)`,
	"attention.sliding_window_pattern":  `Шаблон чередования local(SWA)/global слоёв: у Gemma 3 целое N (каждый N-й — global, типично 6); у Gemma 4 — bool-массив длиной block_count с пометкой SWA-слоёв`,
	"attention.shared_kv_layers":        `Количество последних слоёв, переиспользующих KV-проекции более ранних слоёв (оптимизация Gemma 3+ для экономии KV-cache)`,
	"attention.max_alibi_bias":          `Максимальный ALiBi bias`,
	"attention.clamp_kqv":               `Clamp для K/Q/V`,
	"attention.causal":                  `Каузальная маска внимания (true для декодеров)`,
	"attention.q_lora_rank":             `Q LoRA rank (DeepSeek MLA)`,
	"attention.kv_lora_rank":            `KV LoRA rank (DeepSeek MLA)`,
	"attention.relative_buckets_count":  `Число relative-position бакетов (T5-style)`,
	"attention.scale":                   `Множитель скалирования attention-логитов (обычно 1/sqrt(head_dim))`,

	// RoPE
	"rope.dimension_count":                 `Число измерений head_dim, на которые применяется RoPE (обычно равно key_length)`,
	"rope.dimension_count_swa":             `Число RoPE-измерений в SWA-слоях (Gemma 3+, может отличаться от global)`,
	"rope.dimension_sections":              `Разбиение RoPE-измерений на секции для multi-axis M-RoPE (Qwen3-VL/3.5: длины секций time/height/width)`,
	"rope.freq_base":                       `Базовая частота θ для RoPE (обычно 10000; 1e6 у global-слоёв Gemma 3+ и моделей с расширенным контекстом)`,
	"rope.freq_base_swa":                   `Базовая частота θ для RoPE в SWA-слоях (у Gemma 3+ обычно 10000, тогда как global — 1e6)`,
	"rope.freq_scale":                      `Масштаб частоты RoPE`,
	"rope.scaling.type":                    `Тип RoPE-масштабирования (none/linear/yarn/longrope)`,
	"rope.scaling.factor":                  `Коэффициент масштабирования RoPE для расширения контекста`,
	"rope.scaling.original_context_length": `Исходный context length до скейлинга`,
	"rope.scaling.attn_factor":             `Attention factor для YARN`,
	"rope.scaling.finetuned":               `Был ли RoPE fine-tuned`,

	// SSM (Mamba / Qwen3-Next)
	"ssm.conv_kernel":    `Размер 1D-свёрточного ядра в SSM-блоке (Mamba-2; типично 4)`,
	"ssm.group_count":    `Число групп в multi-group SSM (аналог GQA для state-space; разделяет B/C-проекции между головами)`,
	"ssm.inner_size":     `Внутренняя размерность SSM-блока (d_inner; обычно 2·embedding_length)`,
	"ssm.state_size":     `Размер скрытого состояния SSM (d_state; типично 64/128/256)`,
	"ssm.time_step_rank": `Ранг низкоранговой параметризации шага Δt в SSM (dt_rank; обычно ceil(d_model/16))`,

	// MoE
	"expert_count":              `Количество MoE-экспертов`,
	"expert_used_count":         `Сколько экспертов активируется на токен (top-k)`,
	"expert_shared_count":       `Количество shared-экспертов (DeepSeekMoE)`,
	"expert_weights_scale":      `Множитель весов экспертов`,
	"expert_weights_norm":       `Нормализация весов экспертов`,
	"expert_gating_func":        `Функция гейтинга экспертов`,
	"leading_dense_block_count": `Число dense-блоков перед MoE`,
}

// Describe возвращает русское описание ключа метаданных.
func Describe(key string) string {
	if d, ok := staticDesc[key]; ok {
		return d
	}
	if i := strings.Index(key, "."); i >= 0 {
		suffix := key[i+1:]
		if d, ok := suffixDesc[suffix]; ok {
			return d
		}
	}
	return ""
}
