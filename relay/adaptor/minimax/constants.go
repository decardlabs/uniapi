package minimax

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Source: https://platform.minimaxi.com/document/price
var ModelRatios = map[string]adaptor.ModelConfig{
	// MiniMax-Text-01: ¥0.6/M input, ¥3/M output
	"MiniMax-Text-01": {Ratio: 0.6 * ratio.MilliTokensRmb, CompletionRatio: 5},
	// MiniMax-M1 (reasoning model): ¥2.5/M input, ¥10/M output
	"MiniMax-M1": {Ratio: 2.5 * ratio.MilliTokensRmb, CompletionRatio: 4},
	// MiniMax M2 series (2026-04) — pricing TBD, placeholder values
	"MiniMax-M2.7":           {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 5},
	"MiniMax-M2.7-highspeed": {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 5},
	"MiniMax-M2.5":           {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 5},
	"MiniMax-M2.5-highspeed": {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 5},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)

// MinimaxToolingDefaults notes that MiniMax's pricing reference lists model rates only (no tool pricing) as of 2025-11-12.
// Source: https://r.jina.ai/https://api.minimax.chat/document/price
var MinimaxToolingDefaults = adaptor.ChannelToolConfig{}
