package openai

import (
	"github.com/decardlabs/uniapi/relay/adaptor/ai360"
	"github.com/decardlabs/uniapi/relay/adaptor/alibailian"
	"github.com/decardlabs/uniapi/relay/adaptor/baichuan"
	"github.com/decardlabs/uniapi/relay/adaptor/baiduv2"
	"github.com/decardlabs/uniapi/relay/adaptor/doubao"
	"github.com/decardlabs/uniapi/relay/adaptor/geminiOpenaiCompatible"
	"github.com/decardlabs/uniapi/relay/adaptor/groq"
	"github.com/decardlabs/uniapi/relay/adaptor/lingyiwanwu"
	"github.com/decardlabs/uniapi/relay/adaptor/minimax"
	"github.com/decardlabs/uniapi/relay/adaptor/mistral"
	"github.com/decardlabs/uniapi/relay/adaptor/moonshot"
	"github.com/decardlabs/uniapi/relay/adaptor/novita"
	"github.com/decardlabs/uniapi/relay/adaptor/openrouter"
	"github.com/decardlabs/uniapi/relay/adaptor/siliconflow"
	"github.com/decardlabs/uniapi/relay/adaptor/stepfun"
	"github.com/decardlabs/uniapi/relay/adaptor/togetherai"
	"github.com/decardlabs/uniapi/relay/adaptor/tokenhub"
	"github.com/decardlabs/uniapi/relay/adaptor/xai"
	"github.com/decardlabs/uniapi/relay/adaptor/xunfeiv2"
	"github.com/decardlabs/uniapi/relay/channeltype"
)

var CompatibleChannels = []int{
	channeltype.Azure,
	channeltype.AI360,
	channeltype.Moonshot,
	channeltype.Baichuan,
	channeltype.Minimax,
	channeltype.Doubao,
	channeltype.Mistral,
	channeltype.Groq,
	channeltype.LingYiWanWu,
	channeltype.StepFun,
	channeltype.DeepSeek,
	channeltype.TogetherAI,
	channeltype.Novita,
	channeltype.SiliconFlow,
	channeltype.XAI,
	channeltype.BaiduV2,
	channeltype.XunfeiV2,
	channeltype.AliBailian,
	channeltype.TencentTokenHub,
}

func GetCompatibleChannelMeta(channelType int) (string, []string) {
	switch channelType {
	case channeltype.Azure:
		return "azure", ModelList
	case channeltype.AI360:
		return "360", ai360.ModelList
	case channeltype.Moonshot:
		return "moonshot", moonshot.ModelList
	case channeltype.Baichuan:
		return "baichuan", baichuan.ModelList
	case channeltype.Minimax:
		return "minimax", minimax.ModelList
	case channeltype.Mistral:
		return "mistralai", mistral.ModelList
	case channeltype.Groq:
		return "groq", groq.ModelList
	case channeltype.LingYiWanWu:
		return "lingyiwanwu", lingyiwanwu.ModelList
	case channeltype.StepFun:
		return "stepfun", stepfun.ModelList
	case channeltype.DeepSeek:
		return "deepseek", []string{"deepseek-chat", "deepseek-reasoner"}
	case channeltype.TogetherAI:
		return "together.ai", togetherai.ModelList
	case channeltype.Doubao:
		return "doubao", doubao.ModelList
	case channeltype.Novita:
		return "novita", novita.ModelList
	case channeltype.SiliconFlow:
		return "siliconflow", siliconflow.ModelList
	case channeltype.XAI:
		return "xai", xai.ModelList
	case channeltype.BaiduV2:
		return "baiduv2", baiduv2.ModelList
	case channeltype.XunfeiV2:
		return "xunfeiv2", xunfeiv2.ModelList
	case channeltype.OpenRouter:
		adaptor := &openrouter.Adaptor{}
		return "openrouter", adaptor.GetModelList()
	case channeltype.AliBailian:
		return "alibailian", alibailian.ModelList
	case channeltype.TencentTokenHub:
		return "tencenttokenhub", tokenhub.ModelList
	case channeltype.GeminiOpenAICompatible:
		return "geminiv2", geminiOpenaiCompatible.ModelList
	default:
		return "openai", ModelList
	}
}
