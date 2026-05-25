package openai_compatible

import (
	"strings"

	"github.com/decardlabs/uniapi/relay/model"
)

// BackfillToolMessageNamesFromToolCalls backfills role=tool message names from prior assistant
// tool_calls when tool_call_id can be mapped to a function name.
// Some providers (MiniMax, DeepSeek, Moonshot) require the name field on role=tool messages.
func BackfillToolMessageNamesFromToolCalls(request *model.GeneralOpenAIRequest) {
	if request == nil || len(request.Messages) == 0 {
		return
	}

	toolCallNames := make(map[string]string)
	for i := range request.Messages {
		message := &request.Messages[i]
		for _, toolCall := range message.ToolCalls {
			if toolCall.Id == "" || toolCall.Function == nil || toolCall.Function.Name == "" {
				continue
			}
			for _, key := range toolCallIDVariants(toolCall.Id) {
				if key == "" {
					continue
				}
				toolCallNames[key] = toolCall.Function.Name
			}
		}
	}

	for i := range request.Messages {
		message := &request.Messages[i]
		if message.Role != "tool" || message.Name != nil || message.ToolCallId == "" {
			continue
		}

		if name, ok := toolCallNames[message.ToolCallId]; ok && name != "" {
			nameCopy := name
			message.Name = &nameCopy
		}
	}
}

// toolCallIDVariants returns normalized key variants for matching call IDs from
// different protocol representations (call_*, fc_*, or bare suffix).
func toolCallIDVariants(id string) []string {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return nil
	}

	variants := []string{trimmed}
	if strings.HasPrefix(trimmed, "call_") {
		suffix := strings.TrimPrefix(trimmed, "call_")
		variants = append(variants, "fc_"+suffix, suffix)
	} else if strings.HasPrefix(trimmed, "fc_") {
		suffix := strings.TrimPrefix(trimmed, "fc_")
		variants = append(variants, "call_"+suffix, suffix)
	} else {
		variants = append(variants, "call_"+trimmed, "fc_"+trimmed)
	}

	return variants
}
