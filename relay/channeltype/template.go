package channeltype

// ChannelTypeTemplateField describes a parameter for a channel type.
type ChannelTypeTemplateField struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"` // string, number, bool, select
	Required    bool          `json:"required"`
	Default     interface{}   `json:"default,omitempty"`
	Description string        `json:"desc,omitempty"`
	Options     []interface{} `json:"options,omitempty"` // for select
	Pattern     string        `json:"pattern,omitempty"` // for validation
}

// ChannelTypeTemplate is a slice of fields.
type ChannelTypeTemplate []ChannelTypeTemplateField

// ChannelTypeInfoV2 is the new structure for API, with template and category.
type ChannelTypeInfoV2 struct {
	ID          int                 `json:"id"`
	Name        string              `json:"name"`
	Label       string              `json:"label"`
	Category    string              `json:"category"`
	Description string              `json:"description,omitempty"`
	Template    ChannelTypeTemplate `json:"template"`
}
