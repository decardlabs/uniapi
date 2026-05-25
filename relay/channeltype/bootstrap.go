package channeltype

// registerBaseTemplate returns common base-url and key fields.
func registerBaseTemplate(defaultBaseURL string) ChannelTypeTemplate {
	return ChannelTypeTemplate{
		{Name: "api_base", Type: "string", Required: true, Default: defaultBaseURL, Description: "API base URL"},
		{Name: "key", Type: "string", Required: true, Description: "API key or token"},
	}
}

// registerOpenAICompatibleTemplate returns template fields for OpenAI-compatible providers.
func registerOpenAICompatibleTemplate(defaultBaseURL string) ChannelTypeTemplate {
	return ChannelTypeTemplate{
		{Name: "api_base", Type: "string", Required: true, Default: defaultBaseURL, Description: "API base URL"},
		{Name: "key", Type: "string", Required: true, Description: "API key"},
		{Name: "api_format", Type: "select", Required: false, Default: "chat_completion", Description: "OpenAI-compatible API format", Options: []interface{}{"chat_completion", "response"}},
	}
}

// registerSignedAuthTemplate returns template fields for AK/SK style providers.
func registerSignedAuthTemplate(defaultBaseURL string) ChannelTypeTemplate {
	return ChannelTypeTemplate{
		{Name: "api_base", Type: "string", Required: true, Default: defaultBaseURL, Description: "API base URL"},
		{Name: "ak", Type: "string", Required: true, Description: "Access key"},
		{Name: "sk", Type: "string", Required: true, Description: "Secret key"},
		{Name: "region", Type: "string", Required: false, Description: "Service region"},
	}
}

// registerVertexAITemplate returns template fields for Google Vertex AI style providers.
func registerVertexAITemplate() ChannelTypeTemplate {
	return ChannelTypeTemplate{
		{Name: "vertex_ai_project_id", Type: "string", Required: true, Description: "Google Cloud project ID"},
		{Name: "vertex_ai_adc", Type: "string", Required: true, Description: "Google Application Default Credentials JSON"},
		{Name: "region", Type: "string", Required: true, Default: "us-central1", Description: "Vertex AI region"},
	}
}

func init() {
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          OpenRouter,
		Name:        "openrouter",
		Label:       "OpenRouter",
		Category:    "official",
		Description: "OpenRouter unified routing endpoint",
		Template:    registerBaseTemplate("https://openrouter.ai/api/v1"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          DeepSeek,
		Name:        "deepseek",
		Label:       "DeepSeek",
		Category:    "official",
		Description: "DeepSeek official endpoint",
		Template:    registerBaseTemplate("https://api.deepseek.com/v1"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          GLM,
		Name:        "glm",
		Label:       "GLM (智谱)",
		Category:    "official",
		Description: "Zhipu GLM official endpoint",
		Template:    registerBaseTemplate("https://open.bigmodel.cn/api/paas/v4"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Moonshot,
		Name:        "moonshot",
		Label:       "Kimi (Moonshot)",
		Category:    "official",
		Description: "Moonshot Kimi official endpoint",
		Template:    registerBaseTemplate("https://api.moonshot.cn/v1"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Minimax,
		Name:        "minimax",
		Label:       "MiniMax",
		Category:    "official",
		Description: "MiniMax（海螺AI）official endpoint，支持 MiniMax-Text-01、MiniMax-M1 等主要模型",
		Template:    registerBaseTemplate("https://api.minimaxi.com/v1"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          SiliconFlow,
		Name:        "siliconflow",
		Label:       "SiliconFlow",
		Category:    "official",
		Description: "SiliconFlow OpenAI-compatible endpoint",
		Template:    registerBaseTemplate("https://api.siliconflow.cn/v1"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Doubao,
		Name:        "ark",
		Label:       "Doubao (Volcengine Ark)",
		Category:    "official",
		Description: "Volcengine Ark / 字节豆包 OpenAI-compatible endpoint",
		Template:    registerOpenAICompatibleTemplate("https://ark.cn-beijing.volces.com/api/v3"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Ali,
		Name:        "qwen",
		Label:       "Alibaba Qwen",
		Category:    "official",
		Description: "Alibaba DashScope / Qwen compatible endpoint",
		Template:    registerBaseTemplate("https://dashscope.aliyuncs.com/compatible-mode/v1"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          TencentTokenHub,
		Name:        "tokenhub",
		Label:       "Tencent TokenHub",
		Category:    "official",
		Description: "Tencent TokenHub OpenAI-compatible endpoint",
		Template:    registerOpenAICompatibleTemplate("https://api.lkeap.cloud.tencent.com/TokenHub/v3"),
	})
}
