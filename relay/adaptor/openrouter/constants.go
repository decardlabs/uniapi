package openrouter

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios defines the comprehensive pricing configuration for all OpenRouter models.
//
// Note (H0llyW00dzZ): This price may need modified, due these price are outdated, as for me I can't modified them because it's too much for me,
// even previously used OpenRouter with oneapi I only used for specific model that I need
var ModelRatios = map[string]adaptor.ModelConfig{
	// OpenRouter Mainstream Models (2026-04)
	// Based on https://openrouter.ai/models
	// Anthropic Claude
	"anthropic/claude-opus-4.7":   {Ratio: 40 * ratio.MilliTokensUsd, CompletionRatio: 3},
	"anthropic/claude-opus-4.6":   {Ratio: 38 * ratio.MilliTokensUsd, CompletionRatio: 3},
	"anthropic/claude-sonnet-4.6": {Ratio: 8 * ratio.MilliTokensUsd, CompletionRatio: 3},
	// OpenAI
	"openai/gpt-5.4":       {Ratio: 2.5 * ratio.MilliTokensUsd, CompletionRatio: 5},
	"openai/gpt-4":         {Ratio: 30 * ratio.MilliTokensUsd, CompletionRatio: 2},
	"openai/gpt-4o":        {Ratio: 2.5 * ratio.MilliTokensUsd, CompletionRatio: 4},
	"openai/gpt-4o-mini":   {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 4},
	"openai/gpt-5.3-codex": {Ratio: 1.75 * ratio.MilliTokensUsd, CompletionRatio: 8},
	// Xiaomi
	"xiaomi/mimo-v2-pro": {Ratio: 0.5 * ratio.MilliTokensUsd, CompletionRatio: 2},
	// xAI
	"x-ai/grok-4.1-fast": {Ratio: 0.6 * ratio.MilliTokensUsd, CompletionRatio: 3},
}

// OpenRouterToolingDefaults enumerates OpenRouter's published web tooling prices (retrieved 2025-11-12).
// Source: https://r.jina.ai/https://openrouter.ai/docs/features/web-search
var OpenRouterToolingDefaults = adaptor.ChannelToolConfig{
	Pricing: map[string]adaptor.ToolPricingConfig{
		// Exa plugin defaults to 5 results per request (5 * $0.004 = $0.02)
		"web_plugin_exa":                      {UsdPerCall: 0.02},
		"openai_native_search_low":            {UsdPerCall: 0.03},
		"openai_native_search_medium":         {UsdPerCall: 0.035},
		"openai_native_search_high":           {UsdPerCall: 0.05},
		"openai_mini_native_search_low":       {UsdPerCall: 0.025},
		"openai_mini_native_search_medium":    {UsdPerCall: 0.0275},
		"openai_mini_native_search_high":      {UsdPerCall: 0.03},
		"perplexity_native_search_low":        {UsdPerCall: 0.005},
		"perplexity_native_search_medium":     {UsdPerCall: 0.008},
		"perplexity_native_search_high":       {UsdPerCall: 0.012},
		"perplexity_pro_native_search_low":    {UsdPerCall: 0.006},
		"perplexity_pro_native_search_medium": {UsdPerCall: 0.01},
		"perplexity_pro_native_search_high":   {UsdPerCall: 0.014},
	},
}
