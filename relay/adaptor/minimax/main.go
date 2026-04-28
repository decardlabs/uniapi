package minimax

import (
	"fmt"
	"strings"

	"github.com/Laisky/errors/v2"

	"github.com/songquanpeng/one-api/relay/adaptor/openai_compatible"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func GetRequestURL(meta *meta.Meta) (string, error) {
	requestPath := meta.RequestURLPath
	if idx := strings.Index(requestPath, "?"); idx >= 0 {
		requestPath = requestPath[:idx]
	}
	if requestPath == "/v1/messages" {
		return openai_compatible.GetFullRequestURL(meta.BaseURL, "/v1/chat/completions", 0), nil
	}
	if meta.Mode == relaymode.ChatCompletions {
		return fmt.Sprintf("%s/v1/chat/completions", meta.BaseURL), nil
	}
	return "", errors.Errorf("unsupported relay mode %d for minimax", meta.Mode)
}
