package alibailian

import (
	"fmt"

	"github.com/Laisky/errors/v2"

	"github.com/decardlabs/uniapi/relay/meta"
	"github.com/decardlabs/uniapi/relay/relaymode"
)

func GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.Mode {
	case relaymode.ChatCompletions:
		return fmt.Sprintf("%s/compatible-mode/v1/chat/completions", meta.BaseURL), nil
	case relaymode.Embeddings:
		return fmt.Sprintf("%s/compatible-mode/v1/embeddings", meta.BaseURL), nil
	default:
	}
	return "", errors.Errorf("unsupported relay mode %d for ali bailian", meta.Mode)
}
