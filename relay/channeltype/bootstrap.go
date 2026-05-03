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
		ID:          OpenAI,
		Name:        "openai",
		Label:       "OpenAI",
		Category:    "official",
		Description: "OpenAI official endpoint",
		Template:    registerBaseTemplate("https://api.openai.com/v1"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          OpenAICompatible,
		Name:        "openai_compatible",
		Label:       "OpenAI Compatible",
		Category:    "generic",
		Description: "Any OpenAI-compatible provider endpoint",
		Template:    registerOpenAICompatibleTemplate("https://api.openai.com/v1"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Azure,
		Name:        "azure_openai",
		Label:       "Azure OpenAI",
		Category:    "cloud",
		Description: "Azure OpenAI endpoint",
		Template: ChannelTypeTemplate{
			{Name: "api_base", Type: "string", Required: true, Description: "Azure endpoint, for example https://<resource>.openai.azure.com"},
			{Name: "key", Type: "string", Required: true, Description: "Azure API key"},
			{Name: "api_version", Type: "string", Required: false, Default: "2024-10-21", Description: "Azure API version"},
		},
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          OpenRouter,
		Name:        "openrouter",
		Label:       "OpenRouter",
		Category:    "official",
		Description: "OpenRouter unified routing endpoint",
		Template:    registerBaseTemplate("https://openrouter.ai/api"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Anthropic,
		Name:        "anthropic",
		Label:       "Anthropic Claude",
		Category:    "official",
		Description: "Anthropic Claude official endpoint",
		Template:    registerBaseTemplate("https://api.anthropic.com/v1"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          ClaudeCompatible,
		Name:        "claude_compatible",
		Label:       "Claude Compatible",
		Category:    "generic",
		Description: "Anthropic-compatible provider endpoint",
		Template: ChannelTypeTemplate{
			{Name: "api_base", Type: "string", Required: true, Description: "Anthropic-compatible API base URL"},
			{Name: "key", Type: "string", Required: true, Description: "API key"},
			{Name: "api_format", Type: "select", Required: false, Default: "chat_completion", Description: "Transport format", Options: []interface{}{"chat_completion", "response"}},
		},
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Gemini,
		Name:        "gemini",
		Label:       "Google Gemini",
		Category:    "official",
		Description: "Gemini official endpoint",
		Template:    registerBaseTemplate("https://generativelanguage.googleapis.com/v1beta"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          GeminiOpenAICompatible,
		Name:        "gemini_openai_compatible",
		Label:       "Gemini OpenAI Compatible",
		Category:    "official",
		Description: "Gemini endpoint via OpenAI-compatible API",
		Template:    registerOpenAICompatibleTemplate("https://generativelanguage.googleapis.com/v1beta/openai"),
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
		ID:          Moonshot,
		Name:        "moonshot",
		Label:       "Moonshot",
		Category:    "official",
		Description: "Moonshot official endpoint",
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
		ID:          GLM,
		Name:        "glm",
		Label:       "GLM",
		Category:    "official",
		Description: "Zhipu GLM official endpoint",
		Template:    registerBaseTemplate("https://open.bigmodel.cn/api/paas/v4"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Ali,
		Name:        "qwen",
		Label:       "Alibaba Qwen",
		Category:    "official",
		Description: "Alibaba DashScope compatible endpoint",
		Template:    registerBaseTemplate("https://dashscope.aliyuncs.com/compatible-mode/v1"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Cohere,
		Name:        "cohere",
		Label:       "Cohere",
		Category:    "official",
		Description: "Cohere official endpoint",
		Template:    registerBaseTemplate("https://api.cohere.ai/v1"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Groq,
		Name:        "groq",
		Label:       "Groq",
		Category:    "official",
		Description: "Groq OpenAI-compatible endpoint",
		Template:    registerBaseTemplate("https://api.groq.com/openai/v1"),
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
		ID:          XAI,
		Name:        "xai",
		Label:       "xAI",
		Category:    "official",
		Description: "xAI endpoint",
		Template:    registerBaseTemplate("https://api.x.ai/v1"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Ollama,
		Name:        "ollama",
		Label:       "Ollama",
		Category:    "self_hosted",
		Description: "Local or self-hosted Ollama endpoint",
		Template:    registerBaseTemplate("http://localhost:11434/v1"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Coze,
		Name:        "coze",
		Label:       "Coze",
		Category:    "official",
		Description: "Coze endpoint with PAT or OAuth JWT",
		Template: ChannelTypeTemplate{
			{Name: "api_base", Type: "string", Required: true, Default: "https://api.coze.com", Description: "Coze API base URL"},
			{Name: "auth_type", Type: "select", Required: false, Default: "personal_access_token", Description: "Authentication type", Options: []interface{}{"personal_access_token", "oauth_jwt"}},
			{Name: "key", Type: "textarea", Required: true, Description: "Personal access token or OAuth JWT JSON config"},
		},
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Baidu,
		Name:        "baidu",
		Label:       "Baidu Wenxin",
		Category:    "official",
		Description: "Baidu Wenxin endpoint",
		Template: ChannelTypeTemplate{
			{Name: "api_base", Type: "string", Required: true, Default: "https://aip.baidubce.com", Description: "Baidu API base URL"},
			{Name: "key", Type: "string", Required: true, Description: "AK|SK, for example API_KEY|SECRET_KEY"},
		},
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Xunfei,
		Name:        "xunfei",
		Label:       "Xunfei Spark",
		Category:    "official",
		Description: "Xunfei Spark endpoint",
		Template:    registerSignedAuthTemplate("https://spark-api-open.xf-yun.com"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          Tencent,
		Name:        "tencent_hunyuan",
		Label:       "Tencent Hunyuan",
		Category:    "official",
		Description: "Tencent Hunyuan endpoint",
		Template:    registerSignedAuthTemplate("https://hunyuan.tencentcloudapi.com"),
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          AwsClaude,
		Name:        "aws_bedrock_claude",
		Label:       "AWS Bedrock Claude",
		Category:    "cloud",
		Description: "AWS Bedrock Anthropic Claude",
		Template: ChannelTypeTemplate{
			{Name: "ak", Type: "string", Required: true, Description: "AWS access key ID"},
			{Name: "sk", Type: "string", Required: true, Description: "AWS secret access key"},
			{Name: "region", Type: "string", Required: true, Default: "us-east-1", Description: "AWS region"},
		},
	})
	RegisterChannelType(ChannelTypeInfoV2{
		ID:          VertextAI,
		Name:        "vertex_ai",
		Label:       "Google Vertex AI",
		Category:    "cloud",
		Description: "Google Vertex AI endpoint",
		Template:    registerVertexAITemplate(),
	})
}
