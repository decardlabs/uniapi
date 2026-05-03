package channeltype

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelTypeTemplateFieldJSON(t *testing.T) {
	field := ChannelTypeTemplateField{
		Name:        "api_base",
		Type:        "string",
		Required:    true,
		Default:     "https://api.openai.com/v1",
		Description: "API Base URL",
	}
	b, err := json.Marshal(field)
	require.NoError(t, err)
	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &out))
	require.Equal(t, "api_base", out["name"])
	require.Equal(t, "string", out["type"])
	require.Equal(t, true, out["required"])
	require.Equal(t, "https://api.openai.com/v1", out["default"])
	require.Equal(t, "API Base URL", out["desc"])
}

func TestChannelTypeInfoV2JSON(t *testing.T) {
	tmpl := ChannelTypeTemplate{
		{ Name: "api_base", Type: "string", Required: true, Default: "https://api.openai.com/v1", Description: "API Base URL" },
		{ Name: "key", Type: "string", Required: true, Description: "API Key" },
	}
	info := ChannelTypeInfoV2{
		ID: 50,
		Name: "openai",
		Label: "OpenAI 兼容",
		Category: "openai_compatible",
		Description: "OpenAI/兼容协议",
		Template: tmpl,
	}
	b, err := json.Marshal(info)
	require.NoError(t, err)
	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &out))
	require.Equal(t, float64(50), out["id"])
	require.Equal(t, "openai", out["name"])
	require.Equal(t, "OpenAI 兼容", out["label"])
	require.Equal(t, "openai_compatible", out["category"])
	require.Equal(t, "OpenAI/兼容协议", out["description"])
	tmplOut, ok := out["template"].([]interface{})
	require.True(t, ok)
	require.Len(t, tmplOut, 2)
}
