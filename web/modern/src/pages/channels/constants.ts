export interface ChannelType {
  key: number;
  text: string;
  value: number;
  color?: string;
  tip?: string;
  description?: string;
}

export interface Model {
  id: string;
  name: string;
}

export const CHANNEL_TYPES: ChannelType[] = [
  {
    key: 20,
    text: 'OpenRouter',
    value: 20,
    color: 'black',
    description: 'OpenRouter aggregated model marketplace.',
  },
  {
    key: 36,
    text: 'DeepSeek',
    value: 36,
    color: 'blue',
    description: 'DeepSeek native API (deepseek-v4-pro, deepseek-v4-flash).',
  },
  {
    key: 23,
    text: 'Tencent Hunyuan (腾讯混元)',
    value: 23,
    color: 'blue',
    description: 'Tencent Hunyuan native API.',
  },
  {
    key: 16,
    text: 'Zhipu GLM',
    value: 16,
    color: 'violet',
    description: 'Zhipu GLM (智谱) native API — glm-5.1.',
  },
  {
    key: 25,
    text: 'Moonshot AI (KIMI)',
    value: 25,
    color: 'black',
    description: 'Moonshot/Kimi native API — kimi-k2.6.',
  },
  {
    key: 27,
    text: 'MiniMax',
    value: 27,
    color: 'red',
    description: 'MiniMax M2 series (MiniMax-M2.7, MiniMax-M2.5).',
  },
  {
    key: 17,
    text: 'Alibaba Tongyi Qianwen (Qwen)',
    value: 17,
    color: 'orange',
    description: 'DashScope Qwen API — qwen3.6-plus, qwen3.6-flash.',
  },
  {
    key: 50,
    text: 'OpenAI Compatible',
    value: 50,
    color: 'olive',
    description: 'Custom api_base; OpenAI-style API with ChatCompletion or Response API payloads.',
  },
];

export const CHANNEL_TYPES_WITH_DEDICATED_BASE_URL = new Set<number>([50]);
export const CHANNEL_TYPES_WITH_CUSTOM_KEY_FIELD = new Set<number>();

// Mainstream model whitelist per channel type (5-8 models each).
// Used to filter the model dropdown to show only actively-used models.
// The "Fill All" button still exposes every model from the backend catalog.
export const MAINSTREAM_MODELS: Record<number, string[]> = {
  // OpenRouter - curated 10 models
  20: [
    'anthropic/claude-opus-4.7',
    'anthropic/claude-opus-4.6',
    'anthropic/claude-sonnet-4.6',
    'openai/gpt-5.4',
    'openai/gpt-4',
    'openai/gpt-4o',
    'openai/gpt-4o-mini',
    'openai/gpt-5.3-codex',
    'xiaomi/mimo-v2-pro',
    'x-ai/grok-4.1-fast',
  ],
  // Tencent Hunyuan (腾讯混元)
  23: ['hunyuan-lite', 'hunyuan-pro', 'hunyuan-turbo'],
  // DeepSeek
  36: ['deepseek-v4-pro', 'deepseek-v4-flash'],
  // Zhipu GLM
  16: ['glm-5.1'],
  // Moonshot / Kimi
  25: ['kimi-k2.6'],
  // MiniMax
  27: ['MiniMax-M2.7', 'MiniMax-M2.7-highspeed', 'MiniMax-M2.5', 'MiniMax-M2.5-highspeed'],
  // Alibaba Qwen
  17: ['qwen3.6-plus', 'qwen3.6-flash'],
  // OpenAI Compatible (no whitelist — user defines their own models)
  50: [],
};

export const OPENAI_COMPATIBLE_API_FORMAT_OPTIONS = [
  { value: 'chat_completion', label: 'ChatCompletion (default)' },
  { value: 'response', label: 'Response' },
];

export const COZE_AUTH_OPTIONS = [
  {
    key: 'personal_access_token',
    text: 'Personal Access Token',
    value: 'personal_access_token',
  },
  { key: 'oauth_jwt', text: 'OAuth JWT', value: 'oauth_jwt' },
];

export const MODEL_MAPPING_EXAMPLE = {
  'gpt-3.5-turbo-0301': 'gpt-3.5-turbo',
  'gpt-4-0314': 'gpt-4',
  'gpt-4-32k-0314': 'gpt-4-32k',
};

export const MODEL_CONFIGS_EXAMPLE = {
  'gpt-3.5-turbo-0301': {
    ratio: 0.0015,
    completion_ratio: 2.0,
    max_tokens: 65536,
  },
  'gpt-4': {
    ratio: 0.03,
    completion_ratio: 2.0,
    max_tokens: 128000,
  },
} satisfies Record<string, Record<string, unknown>>;

export const TOOLING_CONFIG_EXAMPLE = {
  whitelist: ['web_search'],
  pricing: {
    web_search: {
      usd_per_call: 0.025,
    },
  },
} satisfies Record<string, unknown>;

export const OAUTH_JWT_CONFIG_EXAMPLE = {
  client_type: 'jwt',
  client_id: '123456789',
  coze_www_base: 'https://www.coze.cn',
  coze_api_base: 'https://api.coze.cn',
  private_key: '-----BEGIN PRIVATE KEY-----\n***\n-----END PRIVATE KEY-----',
  public_key_id: '***********************************************************',
};

export const INFERENCE_PROFILE_ARN_MAP_EXAMPLE = {
  'anthropic.claude-3-5-sonnet-20240620-v1:0':
    'arn:aws:bedrock:us-east-1:123456789012:inference-profile/us.anthropic.claude-3-5-sonnet-20240620-v1:0',
};
