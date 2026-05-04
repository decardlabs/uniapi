package minimax

import (
	"github.com/Laisky/errors/v2"

	"github.com/decardlabs/uniapi/relay/adaptor/openai_compatible"
	"github.com/decardlabs/uniapi/relay/meta"
	"github.com/decardlabs/uniapi/relay/relaymode"
)

func GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.Mode {
	case relaymode.ChatCompletions, relaymode.ClaudeMessages, relaymode.ResponseAPI:
		return openai_compatible.GetFullRequestURL(meta.BaseURL, "/v1/chat/completions", 0), nil
	}
	return "", errors.Errorf("unsupported relay mode %d for minimax", meta.Mode)
}
