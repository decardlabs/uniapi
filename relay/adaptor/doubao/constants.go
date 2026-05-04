package doubao

import (
	"github.com/decardlabs/uniapi/relay/adaptor"
	"github.com/decardlabs/uniapi/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on Volcengine Ark pricing: https://www.volcengine.com/docs/82379/1099320
// Prices use the standard [0, 32k] input-length tier (RMB per million tokens)
var ModelRatios = map[string]adaptor.ModelConfig{
	// doubao-seed-2.0 series (2026)
	// doubao-seed-2.0-pro: ¥3.2/M input, ¥16.0/M output, ¥0.64/M cached
	"doubao-seed-2.0-pro": {
		Ratio:            3.2 * ratio.MilliTokensRmb,
		CompletionRatio:  5,
		CachedInputRatio: 0.64 * ratio.MilliTokensRmb,
	},
	// doubao-seed-2.0-lite: ¥0.6/M input, ¥3.6/M output, ¥0.12/M cached
	"doubao-seed-2.0-lite": {
		Ratio:            0.6 * ratio.MilliTokensRmb,
		CompletionRatio:  6,
		CachedInputRatio: 0.12 * ratio.MilliTokensRmb,
	},
	// doubao-seed-2.0-mini: ¥0.2/M input, ¥2.0/M output, ¥0.04/M cached
	"doubao-seed-2.0-mini": {
		Ratio:            0.2 * ratio.MilliTokensRmb,
		CompletionRatio:  10,
		CachedInputRatio: 0.04 * ratio.MilliTokensRmb,
	},
	// doubao-seed-2.0-code: ¥3.2/M input, ¥16.0/M output, ¥0.64/M cached
	"doubao-seed-2.0-code": {
		Ratio:            3.2 * ratio.MilliTokensRmb,
		CompletionRatio:  5,
		CachedInputRatio: 0.64 * ratio.MilliTokensRmb,
	},
	// doubao-seed-1.8 series
	// doubao-seed-1.8: ¥0.80/M input, ¥8.00/M output (long), ¥0.16/M cached
	"doubao-seed-1.8": {
		Ratio:            0.8 * ratio.MilliTokensRmb,
		CompletionRatio:  10,
		CachedInputRatio: 0.16 * ratio.MilliTokensRmb,
	},
	// doubao-seed-character: ¥0.80/M input, ¥6.00/M output, ¥0.16/M cached
	"doubao-seed-character": {
		Ratio:            0.8 * ratio.MilliTokensRmb,
		CompletionRatio:  7,
		CachedInputRatio: 0.16 * ratio.MilliTokensRmb,
	},
	// doubao-seed-code: ¥1.20/M input, ¥8.00/M output, ¥0.24/M cached
	"doubao-seed-code": {
		Ratio:            1.2 * ratio.MilliTokensRmb,
		CompletionRatio:  7,
		CachedInputRatio: 0.24 * ratio.MilliTokensRmb,
	},
	// doubao-seed-1.6 series
	// doubao-seed-1.6: ¥0.80/M input, ¥8.00/M output (long), ¥0.16/M cached
	"doubao-seed-1.6": {
		Ratio:            0.8 * ratio.MilliTokensRmb,
		CompletionRatio:  10,
		CachedInputRatio: 0.16 * ratio.MilliTokensRmb,
	},
	// doubao-seed-1.6-lite: ¥0.30/M input, ¥2.40/M output (long), ¥0.06/M cached
	"doubao-seed-1.6-lite": {
		Ratio:            0.3 * ratio.MilliTokensRmb,
		CompletionRatio:  8,
		CachedInputRatio: 0.06 * ratio.MilliTokensRmb,
	},
	// doubao-seed-1.6-flash: ¥0.15/M input, ¥1.50/M output, ¥0.03/M cached
	"doubao-seed-1.6-flash": {
		Ratio:            0.15 * ratio.MilliTokensRmb,
		CompletionRatio:  10,
		CachedInputRatio: 0.03 * ratio.MilliTokensRmb,
	},
	// doubao-seed-1.6-vision: ¥0.80/M input, ¥8.00/M output, ¥0.16/M cached
	"doubao-seed-1.6-vision": {
		Ratio:            0.8 * ratio.MilliTokensRmb,
		CompletionRatio:  10,
		CachedInputRatio: 0.16 * ratio.MilliTokensRmb,
	},
	// doubao-seed-translation: ¥1.20/M input, ¥3.60/M output, no cache
	"doubao-seed-translation": {Ratio: 1.2 * ratio.MilliTokensRmb, CompletionRatio: 3},
	// doubao-1.5 series
	// doubao-1.5-pro-32k: ¥0.80/M input, ¥2.00/M output, ¥0.16/M cached
	"doubao-1.5-pro-32k": {
		Ratio:            0.8 * ratio.MilliTokensRmb,
		CompletionRatio:  2.5,
		CachedInputRatio: 0.16 * ratio.MilliTokensRmb,
	},
	// doubao-1.5-lite-32k: ¥0.30/M input, ¥0.60/M output, ¥0.06/M cached
	"doubao-1.5-lite-32k": {
		Ratio:            0.3 * ratio.MilliTokensRmb,
		CompletionRatio:  2,
		CachedInputRatio: 0.06 * ratio.MilliTokensRmb,
	},
	// doubao-1.5-vision-pro: ¥3.00/M input, ¥9.00/M output
	"doubao-1.5-vision-pro": {Ratio: 3.0 * ratio.MilliTokensRmb, CompletionRatio: 3},
	// Legacy doubao-pro/lite models (kept for backward compatibility)
	// doubao-pro-32k: ¥0.80/M input, ¥2.00/M output
	"doubao-pro-32k": {Ratio: 0.8 * ratio.MilliTokensRmb, CompletionRatio: 2.5},
	// doubao-pro-128k: ¥0.005/M input, ¥0.009/M output (legacy pricing, using old ratio)
	"doubao-pro-128k": {Ratio: 0.005 * ratio.MilliTokensRmb, CompletionRatio: 1},
	// doubao-pro-4k: ¥0.0008/M input
	"doubao-pro-4k": {Ratio: 0.0008 * ratio.MilliTokensRmb, CompletionRatio: 1},
	// doubao-lite-128k, 32k, 4k (legacy)
	"doubao-lite-128k": {Ratio: 0.0008 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"doubao-lite-32k":  {Ratio: 0.0006 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"doubao-lite-4k":   {Ratio: 0.0003 * ratio.MilliTokensRmb, CompletionRatio: 1},
	// Embedding Models
	"Doubao-embedding": {Ratio: 0.0002 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// ── Official versioned IDs ────────────────────────────────────────────────
	// Prices sourced from Volcengine Ark pricing page (standard [0,32k] tier):
	// https://www.volcengine.com/docs/82379/1099320

	// doubao-seed-2.0-pro series: ¥3.2/M input, ¥16.0/M output, ¥0.64/M cached
	"doubao-seed-2-0-pro-260215": {
		Ratio:            3.2 * ratio.MilliTokensRmb,
		CompletionRatio:  5,
		CachedInputRatio: 0.64 * ratio.MilliTokensRmb,
	},
	// doubao-seed-2.0-lite series: ¥0.6/M input, ¥3.6/M output, ¥0.12/M cached
	"doubao-seed-2-0-lite-250820": {
		Ratio:            0.6 * ratio.MilliTokensRmb,
		CompletionRatio:  6,
		CachedInputRatio: 0.12 * ratio.MilliTokensRmb,
	},
	// doubao-seed-2.0-mini series: ¥0.2/M input, ¥2.0/M output, ¥0.04/M cached
	"doubao-seed-2-0-mini-250820": {
		Ratio:            0.2 * ratio.MilliTokensRmb,
		CompletionRatio:  10,
		CachedInputRatio: 0.04 * ratio.MilliTokensRmb,
	},
	// doubao-seed-2.0-code series: ¥3.2/M input, ¥16.0/M output, ¥0.64/M cached
	"doubao-seed-2-0-code-250923": {
		Ratio:            3.2 * ratio.MilliTokensRmb,
		CompletionRatio:  5,
		CachedInputRatio: 0.64 * ratio.MilliTokensRmb,
	},
	// doubao-seed-1.8: ¥0.80/M input, ¥8.00/M output (long), ¥0.16/M cached
	"doubao-seed-1-8-251228": {
		Ratio:            0.8 * ratio.MilliTokensRmb,
		CompletionRatio:  10,
		CachedInputRatio: 0.16 * ratio.MilliTokensRmb,
	},
	// doubao-seed-character: ¥0.80/M input, ¥6.00/M output, ¥0.16/M cached
	"doubao-seed-character-251128": {
		Ratio:            0.8 * ratio.MilliTokensRmb,
		CompletionRatio:  7.5,
		CachedInputRatio: 0.16 * ratio.MilliTokensRmb,
	},
	// doubao-seed-code: ¥1.20/M input, ¥8.00/M output, ¥0.24/M cached
	"doubao-seed-code-preview-251028": {
		Ratio:            1.2 * ratio.MilliTokensRmb,
		CompletionRatio:  7,
		CachedInputRatio: 0.24 * ratio.MilliTokensRmb,
	},
	// doubao-seed-1.6 series: ¥0.80/M input, ¥8.00/M output (long), ¥0.16/M cached
	"doubao-seed-1-6-251015": {
		Ratio:            0.8 * ratio.MilliTokensRmb,
		CompletionRatio:  10,
		CachedInputRatio: 0.16 * ratio.MilliTokensRmb,
	},
	"doubao-seed-1-6-250615": {
		Ratio:            0.8 * ratio.MilliTokensRmb,
		CompletionRatio:  10,
		CachedInputRatio: 0.16 * ratio.MilliTokensRmb,
	},
	// doubao-seed-1.6-lite: ¥0.30/M input, ¥2.40/M output (long), ¥0.06/M cached
	"doubao-seed-1-6-lite-251015": {
		Ratio:            0.3 * ratio.MilliTokensRmb,
		CompletionRatio:  8,
		CachedInputRatio: 0.06 * ratio.MilliTokensRmb,
	},
	// doubao-seed-1.6-flash: ¥0.15/M input, ¥1.50/M output, ¥0.03/M cached
	"doubao-seed-1-6-flash-250828": {
		Ratio:            0.15 * ratio.MilliTokensRmb,
		CompletionRatio:  10,
		CachedInputRatio: 0.03 * ratio.MilliTokensRmb,
	},
	"doubao-seed-1-6-flash-250615": {
		Ratio:            0.15 * ratio.MilliTokensRmb,
		CompletionRatio:  10,
		CachedInputRatio: 0.03 * ratio.MilliTokensRmb,
	},
	// doubao-seed-1.6-vision: ¥0.80/M input, ¥8.00/M output, ¥0.16/M cached
	"doubao-seed-1-6-vision-250815": {
		Ratio:            0.8 * ratio.MilliTokensRmb,
		CompletionRatio:  10,
		CachedInputRatio: 0.16 * ratio.MilliTokensRmb,
	},
	// doubao-seed-translation: ¥1.20/M input, ¥3.60/M output, no cache
	"doubao-seed-translation-250915": {Ratio: 1.2 * ratio.MilliTokensRmb, CompletionRatio: 3},
	// doubao-1.5-pro-32k series: ¥0.80/M input, ¥2.00/M output, ¥0.16/M cached
	"doubao-1-5-pro-32k-250115": {
		Ratio:            0.8 * ratio.MilliTokensRmb,
		CompletionRatio:  2.5,
		CachedInputRatio: 0.16 * ratio.MilliTokensRmb,
	},
	"doubao-1-5-pro-32k-character-250715": {
		Ratio:            0.8 * ratio.MilliTokensRmb,
		CompletionRatio:  2.5,
		CachedInputRatio: 0.16 * ratio.MilliTokensRmb,
	},
	"doubao-1-5-pro-32k-character-250228": {
		Ratio:            0.8 * ratio.MilliTokensRmb,
		CompletionRatio:  2.5,
		CachedInputRatio: 0.16 * ratio.MilliTokensRmb,
	},
	// doubao-1.5-lite-32k: ¥0.30/M input, ¥0.60/M output, ¥0.06/M cached
	"doubao-1-5-lite-32k-250115": {
		Ratio:            0.3 * ratio.MilliTokensRmb,
		CompletionRatio:  2,
		CachedInputRatio: 0.06 * ratio.MilliTokensRmb,
	},
	// doubao-1.5-vision-pro: ¥3.00/M input, ¥9.00/M output, no cache
	"doubao-1-5-vision-pro-32k-250115": {Ratio: 3.0 * ratio.MilliTokensRmb, CompletionRatio: 3},
	// glm-4.7: ¥2.0/M input (short), ¥8.0/M output (short), ¥0.4/M cached
	"glm-4-7-251222": {
		Ratio:            2.0 * ratio.MilliTokensRmb,
		CompletionRatio:  4,
		CachedInputRatio: 0.4 * ratio.MilliTokensRmb,
	},
	// deepseek-v3.2: ¥2.0/M input (short), ¥3.0/M output, ¥0.4/M cached
	"deepseek-v3-2-251201": {
		Ratio:            2.0 * ratio.MilliTokensRmb,
		CompletionRatio:  1.5,
		CachedInputRatio: 0.4 * ratio.MilliTokensRmb,
	},
	// deepseek-v3.1: ¥4.0/M input, ¥12.0/M output, ¥0.8/M cached
	"deepseek-v3-1-terminus": {
		Ratio:            4.0 * ratio.MilliTokensRmb,
		CompletionRatio:  3,
		CachedInputRatio: 0.8 * ratio.MilliTokensRmb,
	},
	// deepseek-v3: ¥2.0/M input, ¥8.0/M output, ¥0.4/M cached
	"deepseek-v3-250324": {
		Ratio:            2.0 * ratio.MilliTokensRmb,
		CompletionRatio:  4,
		CachedInputRatio: 0.4 * ratio.MilliTokensRmb,
	},
	// deepseek-r1: ¥4.0/M input, ¥16.0/M output, ¥0.8/M cached
	"deepseek-r1-250528": {
		Ratio:            4.0 * ratio.MilliTokensRmb,
		CompletionRatio:  4,
		CachedInputRatio: 0.8 * ratio.MilliTokensRmb,
	},
	// doubao-embedding-vision: ¥0.70/M text input
	"doubao-embedding-vision-251215": {Ratio: 0.7 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"doubao-embedding-vision-250615": {Ratio: 0.7 * ratio.MilliTokensRmb, CompletionRatio: 1},
}

// ModelList enumerates the public Ark text-generation and embedding model IDs.
// It intentionally uses the official versioned IDs from Volcengine's public model list
// instead of the pricing-map aliases that are kept for backward compatibility.
var ModelList = []string{
	"doubao-seed-2-0-pro-260215",
	"doubao-seed-2-0-lite-250820",
	"doubao-seed-2-0-mini-250820",
	"doubao-seed-2-0-code-250923",
	"doubao-seed-1-8-251228",
	"doubao-seed-character-251128",
	"doubao-seed-code-preview-251028",
	"doubao-seed-1-6-251015",
	"doubao-seed-1-6-250615",
	"doubao-seed-1-6-lite-251015",
	"doubao-seed-1-6-flash-250828",
	"doubao-seed-1-6-flash-250615",
	"doubao-seed-1-6-vision-250815",
	"doubao-seed-translation-250915",
	"doubao-1-5-pro-32k-250115",
	"doubao-1-5-pro-32k-character-250715",
	"doubao-1-5-pro-32k-character-250228",
	"doubao-1-5-lite-32k-250115",
	"doubao-1-5-vision-pro-32k-250115",
	"glm-4-7-251222",
	"glm-4-5-air",
	"deepseek-v3-2-251201",
	"deepseek-v3-1-terminus",
	"deepseek-v3-250324",
	"deepseek-r1-250528",
	"qwen3-32b-20250429",
	"qwen3-14b-20250429",
	"qwen3-8b-20250429",
	"qwen3-0-6b-20250429",
	"qwen2-5-72b-20240919",
	"doubao-embedding-vision-251215",
	"doubao-embedding-vision-250615",
}

// DoubaoToolingDefaults documents that Bytedance's Doubao cloud pricing does not list per-tool fees publicly (retrieved 2025-11-12).
// Source: https://r.jina.ai/https://www.volcengine.com/docs/82379/1099320
var DoubaoToolingDefaults = adaptor.ChannelToolConfig{}
