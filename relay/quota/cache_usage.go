package quota

import relaymodel "github.com/decardlabs/uniapi/relay/model"

// ApplyDeepSeekCacheUsage maps DeepSeek top-level cache-hit usage fields into
// PromptTokensDetails so shared quota computation can apply cached input pricing.
func ApplyDeepSeekCacheUsage(usage *relaymodel.Usage) {
	if usage == nil || usage.PromptCacheHitTokens <= 0 {
		return
	}
	if usage.PromptTokensDetails == nil {
		usage.PromptTokensDetails = &relaymodel.UsagePromptTokensDetails{}
	}
	if usage.PromptTokensDetails.CachedTokens == 0 {
		usage.PromptTokensDetails.CachedTokens = usage.PromptCacheHitTokens
	}
}
