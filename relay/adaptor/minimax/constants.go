package minimax

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Source: https://platform.minimaxi.com/docs/guides/pricing-paygo
var ModelRatios = map[string]adaptor.ModelConfig{
	// MiniMax-Text-01: ¥0.6/M input, ¥3/M output
	"MiniMax-Text-01": {Ratio: 0.6 * ratio.MilliTokensRmb, CompletionRatio: 5},
	// MiniMax-M1 (reasoning model): ¥2.5/M input, ¥10/M output
	"MiniMax-M1": {Ratio: 2.5 * ratio.MilliTokensRmb, CompletionRatio: 4},
	// MiniMax M2 series (2026-04)
	// Source: https://platform.minimaxi.com/docs/guides/pricing-paygo
	// M2.7: ¥2.1/M input, ¥8.4/M output, ¥0.42/M cached input
	"MiniMax-M2.7": {
		Ratio:            2.1 * ratio.MilliTokensRmb,
		CompletionRatio:  4,
		CachedInputRatio: 0.42 * ratio.MilliTokensRmb,
	},
	// M2.7-highspeed: ¥4.2/M input, ¥16.8/M output, ¥0.42/M cached input
	"MiniMax-M2.7-highspeed": {
		Ratio:            4.2 * ratio.MilliTokensRmb,
		CompletionRatio:  4,
		CachedInputRatio: 0.42 * ratio.MilliTokensRmb,
	},
	// M2.5: ¥2.1/M input, ¥8.4/M output, ¥0.21/M cached input
	"MiniMax-M2.5": {
		Ratio:            2.1 * ratio.MilliTokensRmb,
		CompletionRatio:  4,
		CachedInputRatio: 0.21 * ratio.MilliTokensRmb,
	},
	// M2.5-highspeed: ¥4.2/M input, ¥16.8/M output, ¥0.21/M cached input
	"MiniMax-M2.5-highspeed": {
		Ratio:            4.2 * ratio.MilliTokensRmb,
		CompletionRatio:  4,
		CachedInputRatio: 0.21 * ratio.MilliTokensRmb,
	},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)

// MinimaxToolingDefaults notes that MiniMax's pricing reference lists model rates only (no tool pricing) as of 2025-11-12.
// Source: https://r.jina.ai/https://api.minimax.chat/document/price
var MinimaxToolingDefaults = adaptor.ChannelToolConfig{}
