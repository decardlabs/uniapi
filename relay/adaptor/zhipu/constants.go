package zhipu

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
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
