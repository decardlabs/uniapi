package relay

import (
	"github.com/decardlabs/uniapi/relay/adaptor"
	"github.com/decardlabs/uniapi/relay/adaptor/ai360"
	"github.com/decardlabs/uniapi/relay/adaptor/aiproxy"
	"github.com/decardlabs/uniapi/relay/adaptor/ali"
	"github.com/decardlabs/uniapi/relay/adaptor/anthropic"
	"github.com/decardlabs/uniapi/relay/adaptor/aws"
	"github.com/decardlabs/uniapi/relay/adaptor/baidu"
	"github.com/decardlabs/uniapi/relay/adaptor/cloudflare"
	"github.com/decardlabs/uniapi/relay/adaptor/cohere"
	"github.com/decardlabs/uniapi/relay/adaptor/copilot"
	"github.com/decardlabs/uniapi/relay/adaptor/coze"
	"github.com/decardlabs/uniapi/relay/adaptor/deepl"
	"github.com/decardlabs/uniapi/relay/adaptor/deepseek"
	"github.com/decardlabs/uniapi/relay/adaptor/gemini"
	"github.com/decardlabs/uniapi/relay/adaptor/groq"
	"github.com/decardlabs/uniapi/relay/adaptor/minimax"
	"github.com/decardlabs/uniapi/relay/adaptor/mistral"
	"github.com/decardlabs/uniapi/relay/adaptor/moonshot"
	"github.com/decardlabs/uniapi/relay/adaptor/ollama"
	"github.com/decardlabs/uniapi/relay/adaptor/openai"
	"github.com/decardlabs/uniapi/relay/adaptor/openrouter"
	"github.com/decardlabs/uniapi/relay/adaptor/palm"
	"github.com/decardlabs/uniapi/relay/adaptor/proxy"
	"github.com/decardlabs/uniapi/relay/adaptor/replicate"
	"github.com/decardlabs/uniapi/relay/adaptor/tencent"
	"github.com/decardlabs/uniapi/relay/adaptor/vertexai"
	"github.com/decardlabs/uniapi/relay/adaptor/xai"
	"github.com/decardlabs/uniapi/relay/adaptor/xunfei"
	"github.com/decardlabs/uniapi/relay/adaptor/zhipu"
	"github.com/decardlabs/uniapi/relay/apitype"
	"github.com/decardlabs/uniapi/relay/pricing"
)

func GetAdaptor(apiType int) adaptor.Adaptor {
	switch apiType {
	case apitype.OpenAI:
		return &openai.Adaptor{}
	case apitype.Anthropic:
		return &anthropic.Adaptor{}
	case apitype.PaLM:
		return &palm.Adaptor{}
	case apitype.Baidu:
		return &baidu.Adaptor{}
	case apitype.Zhipu:
		return &zhipu.Adaptor{}
	case apitype.Ali:
		return &ali.Adaptor{}
	case apitype.Xunfei:
		return &xunfei.Adaptor{}
	case apitype.AIProxyLibrary:
		return &aiproxy.Adaptor{}
	case apitype.Tencent:
		return &tencent.Adaptor{}
	case apitype.Gemini:
		return &gemini.Adaptor{}
	case apitype.Ollama:
		return &ollama.Adaptor{}
	case apitype.AwsClaude:
		return &aws.Adaptor{}
	case apitype.Coze:
		return &coze.Adaptor{}
	case apitype.Cohere:
		return &cohere.Adaptor{}
	case apitype.Cloudflare:
		return &cloudflare.Adaptor{}
	case apitype.DeepL:
		return &deepl.Adaptor{}
	case apitype.VertexAI:
		return &vertexai.Adaptor{}
	case apitype.Proxy:
		return &proxy.Adaptor{}
	case apitype.Replicate:
		return &replicate.Adaptor{}
	case apitype.DeepSeek:
		return &deepseek.Adaptor{}
	case apitype.Groq:
		return &groq.Adaptor{}
	case apitype.Mistral:
		return &mistral.Adaptor{}
	case apitype.Moonshot:
		return &moonshot.Adaptor{}
	case apitype.XAI:
		return &xai.Adaptor{}
	case apitype.OpenRouter:
		return &openrouter.Adaptor{}
	case apitype.Copilot:
		return &copilot.Adaptor{}
	case apitype.Minimax:
		return &minimax.Adaptor{}
	}

	return nil
}

// getAI360Adaptor returns AI360 adaptor (used for pricing/channel testing only)
// AI360 maps to OpenAI apitype so it goes through openai adaptor by default,
// but for pricing purposes we expose the specific adaptor.
func getAI360Adaptor() adaptor.Adaptor {
	return &ai360.Adaptor{}
}

// InitializeGlobalPricing initializes the global pricing manager with the GetAdaptor function
func InitializeGlobalPricing() {
	pricing.InitializeGlobalPricingManager(GetAdaptor)
}
