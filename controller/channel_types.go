package controller

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/channeltype"
)

// ChannelTypeTemplateOptionResponse defines a selectable option in a channel template field.
type ChannelTypeTemplateOptionResponse struct {
	Value interface{} `json:"value"`
	Label string      `json:"label"`
}

// ChannelTypeTemplateFieldResponse defines one dynamic parameter field for frontend rendering.
type ChannelTypeTemplateFieldResponse struct {
	Key      string                              `json:"key"`
	Label    string                              `json:"label"`
	Type     string                              `json:"type"`
	Required bool                                `json:"required"`
	Help     string                              `json:"help,omitempty"`
	Default  interface{}                         `json:"default,omitempty"`
	Options  []ChannelTypeTemplateOptionResponse `json:"options,omitempty"`
}

// ChannelTypeTemplateResponse defines the template container for dynamic parameters.
type ChannelTypeTemplateResponse struct {
	Fields []ChannelTypeTemplateFieldResponse `json:"fields"`
}

// ChannelTypeResponse defines one channel type option returned to the Modern frontend.
type ChannelTypeResponse struct {
	Key         int                         `json:"key"`
	Value       int                         `json:"value"`
	Text        string                      `json:"text"`
	Description string                      `json:"description,omitempty"`
	Category    string                      `json:"category,omitempty"`
	Template    ChannelTypeTemplateResponse `json:"template"`
}

var fieldLabelOverrides = map[string]string{
	"api_base":             "API Base URL",
	"key":                  "API Key",
	"api_format":           "API Format",
	"api_version":          "API Version",
	"region":               "Region",
	"ak":                   "Access Key",
	"sk":                   "Secret Key",
	"auth_type":            "Auth Type",
	"vertex_ai_project_id": "Vertex AI Project ID",
	"vertex_ai_adc":        "Vertex AI ADC JSON",
}

// GetChannelTypes returns all registered channel types in frontend-compatible shape.
func GetChannelTypes(c *gin.Context) {
	types := channeltype.AllChannelTypes()
	result := make([]ChannelTypeResponse, 0, len(types))

	for _, ct := range types {
		fields := make([]ChannelTypeTemplateFieldResponse, 0, len(ct.Template))
		for _, f := range ct.Template {
			if isCoreTemplateField(f.Name) {
				continue
			}

			options := make([]ChannelTypeTemplateOptionResponse, 0, len(f.Options))
			for _, opt := range f.Options {
				options = append(options, ChannelTypeTemplateOptionResponse{
					Value: opt,
					Label: toOptionLabel(opt),
				})
			}

			label := toFieldLabel(f.Name)
			help := strings.TrimSpace(f.Description)
			if help == "" {
				help = fmt.Sprintf("Configure %s for this channel.", strings.ToLower(label))
			}

			fields = append(fields, ChannelTypeTemplateFieldResponse{
				Key:      f.Name,
				Label:    label,
				Type:     f.Type,
				Required: f.Required,
				Help:     help,
				Default:  f.Default,
				Options:  options,
			})
		}

		text := strings.TrimSpace(ct.Label)
		if text == "" {
			text = strings.TrimSpace(ct.Name)
		}

		result = append(result, ChannelTypeResponse{
			Key:         ct.ID,
			Value:       ct.ID,
			Text:        text,
			Description: strings.TrimSpace(ct.Description),
			Category:    strings.TrimSpace(ct.Category),
			Template: ChannelTypeTemplateResponse{
				Fields: fields,
			},
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Text == result[j].Text {
			return result[i].Value < result[j].Value
		}
		return result[i].Text < result[j].Text
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// isCoreTemplateField reports whether the field maps to a first-class form control.
func isCoreTemplateField(fieldName string) bool {
	switch fieldName {
	case "api_base", "key":
		return true
	default:
		return false
	}
}

// toFieldLabel converts an internal field key to a frontend display label.
func toFieldLabel(fieldName string) string {
	if label, ok := fieldLabelOverrides[fieldName]; ok {
		return label
	}
	parts := strings.Split(fieldName, "_")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, " ")
}

// toOptionLabel converts an option value to display text.
func toOptionLabel(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}
