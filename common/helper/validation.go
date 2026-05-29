package helper

import (
	"sort"
	"strings"
)

// BuildTextOnlyModelImageInputValidationMessage builds a consistent validation
// message for requests that include image input for text-only models.
// Parameters: modelName is the requested model identifier; contentTypes contains
// detected image-related content types; reason explains why routing cannot proceed.
// Returns: a user-facing validation message with actionable next steps.
func BuildTextOnlyModelImageInputValidationMessage(modelName string, contentTypes []string, reason string) string {
	trimmedModel := strings.TrimSpace(modelName)
	if trimmedModel == "" {
		trimmedModel = "<unknown-model>"
	}

	filteredTypes := make([]string, 0, len(contentTypes))
	seen := make(map[string]struct{}, len(contentTypes))
	for idx := range contentTypes {
		typeName := strings.TrimSpace(strings.ToLower(contentTypes[idx]))
		if typeName == "" {
			continue
		}
		if _, exists := seen[typeName]; exists {
			continue
		}
		seen[typeName] = struct{}{}
		filteredTypes = append(filteredTypes, typeName)
	}
	sort.Strings(filteredTypes)

	typeSummary := "image"
	if len(filteredTypes) > 0 {
		typeSummary = strings.Join(filteredTypes, ",")
	}

	msg := "validation failed: model \"" + trimmedModel + "\" only supports text content, but image input was detected (content types: " + typeSummary + ")"
	if strings.TrimSpace(reason) != "" {
		msg += "; " + strings.TrimSpace(reason)
	}

	msg += ". Remove image input and retry with text, or switch to a vision-capable model"
	return msg
}