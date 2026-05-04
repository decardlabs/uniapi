package novita

import (
	"fmt"

	"github.com/Laisky/errors/v2"

	"github.com/decardlabs/uniapi/relay/meta"
	"github.com/decardlabs/uniapi/relay/relaymode"
)

func GetRequestURL(meta *meta.Meta) (string, error) {
	if meta.Mode == relaymode.ChatCompletions {
		return fmt.Sprintf("%s/chat/completions", meta.BaseURL), nil
	}
	return "", errors.Errorf("unsupported relay mode %d for novita", meta.Mode)
}
