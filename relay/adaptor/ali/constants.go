package ali

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios.
// The model list is derived from the keys of this map.
// Default: input $0.30/M tokens, output $1.50/M tokens (USD)
//
// https://help.aliyun.com/zh/model-studio/models
var ModelRatios = map[string]adaptor.ModelConfig{
	// Qwen 3.6 (2026-04)
	// Based on https://help.aliyun.com/zh/model-studio/models
	"qwen3.6-plus":  {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 5},
	"qwen3.6-flash": {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 5},
}

// AliToolingDefaults notes that Alibaba Model Studio does not expose public built-in tool pricing (retrieved 2025-11-12).
// Source: https://r.jina.ai/https://help.aliyun.com/en/model-studio/developer-reference/tools-reference (requires authentication)
var AliToolingDefaults = adaptor.ChannelToolConfig{}
