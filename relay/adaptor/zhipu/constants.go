package zhipu

import (
	"github.com/decardlabs/uniapi/relay/adaptor"
	"github.com/decardlabs/uniapi/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Default: input $0.30/M tokens, output $1.50/M tokens (USD)
var ModelRatios = map[string]adaptor.ModelConfig{
	// GLM-5.1 (2026-04)
	// Based on https://open.bigmodel.cn/pricing
	"glm-5.1": {
		Ratio:            0.15 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
	},
	// GLM-5 series models
	// Based on https://open.bigmodel.cn/pricing
	"glm-5-turbo": {
		Ratio:            0.15 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
	},
	"glm-5": {
		Ratio:            0.15 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
	},
	// GLM-4.7 series
	"glm-4.7": {
		Ratio:            0.1 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.02 * ratio.MilliTokensUsd,
	},
	"glm-4.7-flash": {
		Ratio:            0.05 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.01 * ratio.MilliTokensUsd,
	},
	"glm-4.7-flashx": {
		Ratio:            0.05 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.01 * ratio.MilliTokensUsd,
	},
	// GLM-4.6 series
	"glm-4.6": {
		Ratio:            0.1 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.02 * ratio.MilliTokensUsd,
	},
	// GLM-4.5 series
	"glm-4.5-air": {
		Ratio:            0.05 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.01 * ratio.MilliTokensUsd,
	},
	"glm-4.5-airx": {
		Ratio:            0.05 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.01 * ratio.MilliTokensUsd,
	},
	"glm-4.5-flash": {
		Ratio:            0.05 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.01 * ratio.MilliTokensUsd,
	},
	// Legacy GLM-4 flash models
	"glm-4-flash-250414": {
		Ratio:            0.05 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.01 * ratio.MilliTokensUsd,
	},
	"glm-4-flashx-250414": {
		Ratio:            0.05 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.01 * ratio.MilliTokensUsd,
	},
	// GLM multimodal models (keep sane defaults; channel-level pricing can override)
	"glm-5v-turbo": {
		Ratio:            0.15 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
	},
	"glm-4.6v": {
		Ratio:            0.15 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
	},
	"glm-4.6v-flashx": {
		Ratio:            0.15 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
	},
	"glm-4.5v": {
		Ratio:            0.15 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
	},
	"glm-4.6v-flash": {
		Ratio:            0.15 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
	},
	"glm-4v-flash": {
		Ratio:            0.15 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
	},
	// Additional vision models
	"autoglm-phone": {
		Ratio:            0.15 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
	},
	"glm-4.1v-thinking-flashx": {
		Ratio:            0.15 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
	},
	"glm-4.1v-thinking-flash": {
		Ratio:            0.15 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
	},
	// OCR model
	"glm-ocr": {
		Ratio:            0.15 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
	},
}

// ZhipuToolingDefaults captures Open BigModel's published search-tool pricing tiers (retrieved 2025-11-12).
// Source: https://r.jina.ai/https://open.bigmodel.cn/pricing
var ZhipuToolingDefaults = adaptor.ChannelToolConfig{
	Pricing: map[string]adaptor.ToolPricingConfig{
		"search_std":       {UsdPerCall: 0.01},
		"search_pro":       {UsdPerCall: 0.03},
		"search_pro_sogou": {UsdPerCall: 0.05},
		"search_pro_quark": {UsdPerCall: 0.05},
	},
}
