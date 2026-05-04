package alibailian

import (
	"github.com/decardlabs/uniapi/relay/adaptor"
	"github.com/decardlabs/uniapi/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on Alibaba Bailian pricing: https://help.aliyun.com/zh/model-studio/billing
var ModelRatios = map[string]adaptor.ModelConfig{
	// Legacy aliases kept for backward compatibility
	"qwen-turbo":              {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 2},
	"qwen-plus":               {Ratio: 0.8 * ratio.MilliTokensRmb, CompletionRatio: 2.5},
	"qwen-long":               {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 4},
	"qwen-max":                {Ratio: 2.4 * ratio.MilliTokensRmb, CompletionRatio: 4},
	"qwen-coder-plus":         {Ratio: 3.5 * ratio.MilliTokensRmb, CompletionRatio: 2},
	"qwen-coder-plus-latest":  {Ratio: 3.5 * ratio.MilliTokensRmb, CompletionRatio: 2},
	"qwen-coder-turbo":        {Ratio: 2.0 * ratio.MilliTokensRmb, CompletionRatio: 3},
	"qwen-coder-turbo-latest": {Ratio: 2.0 * ratio.MilliTokensRmb, CompletionRatio: 3},
	"qwen-mt-plus":            {Ratio: 1.8 * ratio.MilliTokensRmb, CompletionRatio: 3},
	"qwen-mt-turbo":           {Ratio: 0.7 * ratio.MilliTokensRmb, CompletionRatio: 2.8},
	"qwq-32b-preview":         {Ratio: 2.0 * ratio.MilliTokensRmb, CompletionRatio: 3},
	"deepseek-r1":             {Ratio: 4.0 * ratio.MilliTokensRmb, CompletionRatio: 4},
	"deepseek-v3":             {Ratio: 2.0 * ratio.MilliTokensRmb, CompletionRatio: 4},

	// ── Official versioned IDs ────────────────────────────────────────────────
	// Prices sourced from Alibaba Bailian billing page (standard tier):
	// https://help.aliyun.com/zh/model-studio/billing

	// qwen3-max series: ¥2.5/M input, ¥10/M output (0<Token≤32K)
	"qwen3-max": {
		Ratio:           2.5 * ratio.MilliTokensRmb,
		CompletionRatio: 4,
	},
	"qwen3-max-preview": {
		Ratio:           6.0 * ratio.MilliTokensRmb,
		CompletionRatio: 4,
	},
	"qwen3-max-2025-09-23": {
		Ratio:           6.0 * ratio.MilliTokensRmb,
		CompletionRatio: 4,
	},

	// qwen3.6-plus series: ¥2/M input, ¥12/M output (0<Token≤256K, non-thinking)
	"qwen3.6-plus": {
		Ratio:           2.0 * ratio.MilliTokensRmb,
		CompletionRatio: 6,
	},
	"qwen3.6-plus-2026-04-02": {
		Ratio:           2.0 * ratio.MilliTokensRmb,
		CompletionRatio: 6,
	},

	// qwen3.5-plus series: ¥0.8/M input, ¥4.8/M output (0<Token≤128K, non-thinking)
	"qwen3.5-plus": {
		Ratio:           0.8 * ratio.MilliTokensRmb,
		CompletionRatio: 6,
	},
	"qwen3.5-plus-2026-02-15": {
		Ratio:           0.8 * ratio.MilliTokensRmb,
		CompletionRatio: 6,
	},

	// qwen-plus series: ¥0.8/M input, ¥2/M output (non-thinking, 0<Token≤128K)
	"qwen-plus-latest": {
		Ratio:           0.8 * ratio.MilliTokensRmb,
		CompletionRatio: 2.5,
	},
	"qwen-plus-2024-12-20": {
		Ratio:           0.8 * ratio.MilliTokensRmb,
		CompletionRatio: 2.5,
	},

	// qwen3.5-flash series: ¥0.2/M input, ¥2/M output (0<Token≤128K)
	"qwen3.5-flash": {
		Ratio:           0.2 * ratio.MilliTokensRmb,
		CompletionRatio: 10,
	},
	"qwen3.5-flash-2026-02-23": {
		Ratio:           0.2 * ratio.MilliTokensRmb,
		CompletionRatio: 10,
	},

	// qwen-flash series: ¥0.15/M input, ¥1.5/M output (0<Token≤128K)
	"qwen-flash": {
		Ratio:           0.15 * ratio.MilliTokensRmb,
		CompletionRatio: 10,
	},
	"qwen-flash-2025-07-28": {
		Ratio:           0.15 * ratio.MilliTokensRmb,
		CompletionRatio: 10,
	},

	// Open-source qwen3.5 MoE series: prices per ¥/M token (0<Token≤128K)
	// qwen3.5-397b-a17b: ¥1.2/M input, ¥7.2/M output
	"qwen3.5-397b-a17b": {
		Ratio:           1.2 * ratio.MilliTokensRmb,
		CompletionRatio: 6,
	},
	// qwen3.5-122b-a10b: ¥0.8/M input, ¥6.4/M output
	"qwen3.5-122b-a10b": {
		Ratio:           0.8 * ratio.MilliTokensRmb,
		CompletionRatio: 8,
	},
	// qwen3.5-27b: ¥0.6/M input, ¥4.8/M output
	"qwen3.5-27b": {
		Ratio:           0.6 * ratio.MilliTokensRmb,
		CompletionRatio: 8,
	},
	// qwen3.5-35b-a3b: ¥0.4/M input, ¥3.2/M output
	"qwen3.5-35b-a3b": {
		Ratio:           0.4 * ratio.MilliTokensRmb,
		CompletionRatio: 8,
	},

	// Open-source qwen3 series
	// qwen3-next-80b-a3b: ¥1/M input (thinking ¥10/M output, non-thinking ¥4/M output)
	"qwen3-next-80b-a3b-thinking": {
		Ratio:           1.0 * ratio.MilliTokensRmb,
		CompletionRatio: 10,
	},
	"qwen3-next-80b-a3b-instruct": {
		Ratio:           1.0 * ratio.MilliTokensRmb,
		CompletionRatio: 4,
	},
	// qwen3-235b-a22b: ¥2/M input, ¥8/M output (non-thinking) / ¥20/M (thinking)
	"qwen3-235b-a22b-thinking-2507": {
		Ratio:           2.0 * ratio.MilliTokensRmb,
		CompletionRatio: 10,
	},
	"qwen3-235b-a22b-instruct-2507": {
		Ratio:           2.0 * ratio.MilliTokensRmb,
		CompletionRatio: 4,
	},
	// qwen3-30b-a3b: ¥0.75/M input, ¥3/M output (non-thinking) / ¥7.5/M (thinking)
	"qwen3-30b-a3b-thinking-2507": {
		Ratio:           0.75 * ratio.MilliTokensRmb,
		CompletionRatio: 10,
	},
	"qwen3-30b-a3b-instruct-2507": {
		Ratio:           0.75 * ratio.MilliTokensRmb,
		CompletionRatio: 4,
	},
	// qwen3-235b-a22b (non-versioned): ¥2/M input, ¥8/M non-thinking, ¥20/M thinking
	"qwen3-235b-a22b": {
		Ratio:           2.0 * ratio.MilliTokensRmb,
		CompletionRatio: 4,
	},
	// qwen3-32b: ¥2/M input, ¥8/M output (non-thinking) / ¥20/M (thinking)
	"qwen3-32b": {
		Ratio:           2.0 * ratio.MilliTokensRmb,
		CompletionRatio: 4,
	},
	// qwen3-30b-a3b: ¥0.75/M input, ¥3/M output non-thinking
	"qwen3-30b-a3b": {
		Ratio:           0.75 * ratio.MilliTokensRmb,
		CompletionRatio: 4,
	},
	// qwen3-14b: ¥1/M input, ¥4/M output (non-thinking) / ¥10/M (thinking)
	"qwen3-14b": {
		Ratio:           1.0 * ratio.MilliTokensRmb,
		CompletionRatio: 4,
	},
	// qwen3-8b: ¥0.5/M input, ¥2/M output (non-thinking) / ¥5/M (thinking)
	"qwen3-8b": {
		Ratio:           0.5 * ratio.MilliTokensRmb,
		CompletionRatio: 4,
	},
}

// ModelList enumerates the model IDs that Alibaba explicitly lists on its
// OpenAI-compatible chat-completions documentation page.
var ModelList = []string{
	"qwen3-max",
	"qwen3-max-preview",
	"qwen3-max-2025-09-23",
	"qwen3.6-plus",
	"qwen3.6-plus-2026-04-02",
	"qwen3.5-plus",
	"qwen3.5-plus-2026-02-15",
	"qwen-plus",
	"qwen-plus-latest",
	"qwen-plus-2024-12-20",
	"qwen3.5-flash",
	"qwen3.5-flash-2026-02-23",
	"qwen-flash",
	"qwen-flash-2025-07-28",
	"qwen3.5-397b-a17b",
	"qwen3.5-122b-a10b",
	"qwen3.5-27b",
	"qwen3.5-35b-a3b",
	"qwen3-next-80b-a3b-thinking",
	"qwen3-next-80b-a3b-instruct",
	"qwen3-235b-a22b-thinking-2507",
	"qwen3-235b-a22b-instruct-2507",
	"qwen3-30b-a3b-thinking-2507",
	"qwen3-30b-a3b-instruct-2507",
	"qwen3-235b-a22b",
	"qwen3-32b",
	"qwen3-30b-a3b",
	"qwen3-14b",
	"qwen3-8b",
}

// AlibailianToolingDefaults reflects that Bailian's public docs do not disclose per-tool pricing (retrieved 2025-11-12).
// Source: https://r.jina.ai/https://www.alibabacloud.com/help/en/model-studio/latest/billing (returns 404 for unauthenticated access)
var AlibailianToolingDefaults = adaptor.ChannelToolConfig{}
