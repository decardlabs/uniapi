package channeltype

// ChannelTypeInfo holds id, name, label for frontend select.
type ChannelTypeInfo struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Label string `json:"label"`
}

var typeLabels = map[int]string{
	OpenAI:    "OpenAI 兼容",
	API2D:     "API2D",
	Azure:     "Azure OpenAI",
	Anthropic: "Anthropic Claude",
	Baidu:     "百度文心",
	Zhipu:     "智谱AI",
	Ali:       "阿里通义",
	Xunfei:    "讯飞星火",
	Gemini:    "Google Gemini",
	Cohere:    "Cohere",
	Groq:      "Groq",
	Ollama:    "Ollama",
	Custom:    "自定义/其他",
	// ...可补充其他类型
}

// AllTypesWithLabels returns all channel types for select.
func AllTypesWithLabels() []ChannelTypeInfo {
	var out []ChannelTypeInfo
	for id, label := range typeLabels {
		out = append(out, ChannelTypeInfo{
			ID:    id,
			Name:  TypeName(id),
			Label: label,
		})
	}
	return out
}

// TypeName returns the const name for a channel type id.
func TypeName(id int) string {
	switch id {
	case OpenAI:
		return "openai"
	case API2D:
		return "api2d"
	case Azure:
		return "azure"
	case Anthropic:
		return "anthropic"
	case Baidu:
		return "baidu"
	case Zhipu:
		return "zhipu"
	case Ali:
		return "ali"
	case Xunfei:
		return "xunfei"
	case Gemini:
		return "gemini"
	case Cohere:
		return "cohere"
	case Groq:
		return "groq"
	case Ollama:
		return "ollama"
	case Custom:
		return "custom"
	default:
		return "unknown"
	}
}
