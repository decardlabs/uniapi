package moonshot

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Default: input $0.30/M tokens, output $1.50/M tokens (USD)
var ModelRatios = map[string]adaptor.ModelConfig{
	// Kimi K2.6 (2026-04)
	// Based on https://platform.moonshot.cn/docs/pricing
	"kimi-k2.6": {
		Ratio:            0.15 * ratio.MilliTokensUsd,
		CompletionRatio:  5,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
	},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)

// MoonshotToolingDefaults notes that Moonshot's pricing page lists model fees only; no tool metering is published (retrieved 2025-11-12).
// Source: https://r.jina.ai/https://platform.moonshot.cn/docs/pricing
var MoonshotToolingDefaults = adaptor.ChannelToolConfig{}
