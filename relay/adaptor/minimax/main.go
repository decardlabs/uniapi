package minimax

import (
	"github.com/Laisky/errors/v2"

	"github.com/songquanpeng/one-api/relay/adaptor/openai_compatible"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.Mode {
	case relaymode.ChatCompletions, relaymode.ClaudeMessages, relaymode.ResponseAPI:
		return openai_compatible.GetFullRequestURL(meta.BaseURL, "/v1/chat/completions", 0), nil
	}
	return "", errors.Errorf("unsupported relay mode %d for minimax", meta.Mode)
}
