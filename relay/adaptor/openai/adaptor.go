package openai

import "github.com/decardlabs/uniapi/relay/adaptor"

type Adaptor struct {
	adaptor.DefaultPricingMethods
	ChannelType int
}
