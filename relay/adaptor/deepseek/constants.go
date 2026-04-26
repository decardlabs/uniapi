package deepseek

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on official DeepSeek pricing: https://api-docs.deepseek.com/quick_start/pricing
var ModelRatios = map[string]adaptor.ModelConfig{
	// DeepSeek V4 (2026-04)
	// Based on https://api-docs.deepseek.com/quick_start/pricing
	"deepseek-v4-pro": {
		Ratio:            1.2 * ratio.MilliTokensUsd,
		CachedInputRatio: 0.12 * ratio.MilliTokensUsd,
		CompletionRatio:  2.0,
	},
	"deepseek-v4-flash": {
		Ratio:            0.3 * ratio.MilliTokensUsd,
		CachedInputRatio: 0.03 * ratio.MilliTokensUsd,
		CompletionRatio:  2.0,
	},
}

// DeepseekToolingDefaults documents that DeepSeek does not publish built-in tool pricing (retrieved 2025-11-12).
// Source: https://r.jina.ai/https://api-docs.deepseek.com/quick_start/pricing
var DeepseekToolingDefaults = adaptor.ChannelToolConfig{}
