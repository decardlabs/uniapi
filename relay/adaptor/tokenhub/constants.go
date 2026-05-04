package tokenhub

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// ModelList is the list of models available on Tencent TokenHub aggregator platform.
// TokenHub is an OpenRouter-like aggregator that provides access to Hunyuan models
// and third-party models (DeepSeek, GLM, Qwen, etc.) through an OpenAI-compatible API.
var ModelList = []string{
	// Hunyuan models
	"hunyuan-turbos-latest",
	"hunyuan-turbos-20250116",
	"hunyuan-t1-latest",
	"hunyuan-t1-20250321",
	"hunyuan-standard",
	"hunyuan-standard-256k",
	"hunyuan-lite",
	// DeepSeek via TokenHub
	"deepseek-v4-flash",
	"deepseek-v3",
	"deepseek-r1",
	// GLM via TokenHub
	"glm-4-plus",
	"glm-4-flash",
	"glm-z1-flash",
	// Qwen via TokenHub
	"qwen-turbo",
	"qwen-plus",
	"qwen-max",
	// MiniMax via TokenHub
	"MiniMax-M2.7",
	"MiniMax-Text-01",
}

// ModelRatio provides pricing for TokenHub models (RMB per 1M tokens, using median pricing).
// TokenHub uses RMB pricing; we use MilliTokensRmb as base unit.
// 1 ratio unit = 0.5 milli-USD per 1K tokens = MilliTokensUsd
// For RMB: 1 RMB/1M tokens = ratio.MilliTokensRmb
var ModelRatio = map[string]float64{
	// Hunyuan turbos: ¥0.8/1M input, ¥2.0/1M output (median ~¥1.4/1M)
	"hunyuan-turbos-latest":   0.8 * ratio.MilliTokensRmb,
	"hunyuan-turbos-20250116": 0.8 * ratio.MilliTokensRmb,
	// Hunyuan T1: ¥3/1M input, ¥15/1M output (median ~¥5/1M)
	"hunyuan-t1-latest":   3 * ratio.MilliTokensRmb,
	"hunyuan-t1-20250321": 3 * ratio.MilliTokensRmb,
	// Hunyuan standard: ¥4.5/1M
	"hunyuan-standard":      4.5 * ratio.MilliTokensRmb,
	"hunyuan-standard-256k": 15 * ratio.MilliTokensRmb,
	// Hunyuan lite: ¥0.8/1M
	"hunyuan-lite": 0.8 * ratio.MilliTokensRmb,
	// DeepSeek via TokenHub (median market price)
	"deepseek-v4-flash": 0.3 * ratio.MilliTokensRmb,
	"deepseek-v3":       1 * ratio.MilliTokensRmb,
	"deepseek-r1":       4 * ratio.MilliTokensRmb,
	// GLM via TokenHub (median)
	"glm-4-plus":   7 * ratio.MilliTokensRmb,
	"glm-4-flash":  0.1 * ratio.MilliTokensRmb,
	"glm-z1-flash": 0.1 * ratio.MilliTokensRmb,
	// Qwen via TokenHub (median)
	"qwen-turbo": 0.3 * ratio.MilliTokensRmb,
	"qwen-plus":  0.8 * ratio.MilliTokensRmb,
	"qwen-max":   2.4 * ratio.MilliTokensRmb,
	// MiniMax via TokenHub
	"MiniMax-M2.7":    2.1 * ratio.MilliTokensRmb,
	"MiniMax-Text-01": 1 * ratio.MilliTokensRmb,
}
