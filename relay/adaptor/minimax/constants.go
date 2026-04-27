package minimax

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Default: input $0.30/M tokens, output $1.50/M tokens (USD)
var ModelRatios = map[string]adaptor.ModelConfig{
	// MiniMax M2 series (2026-04)
	// Based on https://api.minimax.chat/v1
	"MiniMax-M2.7":            {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 5},
	"MiniMax-M2.7-highspeed": {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 5},
	"MiniMax-M2.5":            {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 5},
	"MiniMax-M2.5-highspeed": {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 5},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)

// MinimaxToolingDefaults notes that MiniMax's pricing reference lists model rates only (no tool pricing) as of 2025-11-12.
// Source: https://r.jina.ai/https://api.minimax.chat/document/price
var MinimaxToolingDefaults = adaptor.ChannelToolConfig{}
