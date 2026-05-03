# UniAPI Configuration Guide

This guide describes the configuration process after installing UniAPI, including environment setup, channel template management, and mainstream model template examples.

---

## 1. Installation Recap

- Backend: Build and run with Go 1.25+
  ```sh
  go build -o one-api main.go
  ./one-api
  ```
- Frontend (Modern):
  ```sh
  cd web/modern
  yarn install && yarn build
  cd ../../web/build/modern
  python3 -m http.server 8080
  ```
- Access the UI at http://localhost:8080

---

## 2. Environment Configuration

- All configuration can be set via environment variables or config files (see `common/config/config.go`).
- Common variables:
  - `PORT`: Backend listening port (default: 3000)
  - `DB_DSN`: Database connection string
  - `REDIS_URL`: Redis connection string (if used)
  - `OTEL_*`: OpenTelemetry tracing (optional)
- For production, use Docker Compose or Kubernetes (see README and docs/manuals/k8s.md).

---

## 3. Channel Type & Parameter Template System

UniAPI uses a backend-driven registry for all channel types (model providers). Each channel type is described by a template, which defines the required parameters for connecting to that provider.

- Templates are registered in Go code (see `relay/channeltype/bootstrap.go`).
- The `/api/channel/types` endpoint returns all available types and their parameter templates for the frontend to render.
- Each template field includes:
  - `name`: Parameter key (e.g., `api_base`, `key`)
  - `type`: Data type (`string`, `number`, `bool`, `select`)
  - `required`: Whether the field is required
  - `default`: Default value (if any)
  - `description`: Field description (for UI/help)
  - `options`: For select fields, allowed values
  - `pattern`: Regex for validation (optional)

**Example template field:**
```go
{Name: "api_base", Type: "string", Required: true, Default: "https://api.openai.com/v1", Description: "API Base URL"}
```

---

## 4. Mainstream Model Template Examples

Below are sample template definitions for mainstream providers. These are registered in `relay/channeltype/bootstrap.go` and exposed via the API for frontend use.

### OpenAI
```go
RegisterChannelType(ChannelTypeInfoV2{
  ID: OpenAI,
  Name: "openai",
  Label: "OpenAI 兼容",
  Category: "official",
  Description: "OpenAI 官方及兼容协议（如 Azure、API2D、OhMyGPT 等）",
  Template: ChannelTypeTemplate{
    {Name: "api_base", Type: "string", Required: true, Default: "https://api.openai.com/v1", Description: "API Base URL"},
    {Name: "key", Type: "string", Required: true, Description: "API Key"},
  },
})
```

### Anthropic Claude
```go
RegisterChannelType(ChannelTypeInfoV2{
  ID: Anthropic,
  Name: "anthropic",
  Label: "Anthropic Claude",
  Category: "official",
  Description: "Anthropic Claude 官方及兼容协议",
  Template: ChannelTypeTemplate{
    {Name: "api_base", Type: "string", Required: true, Default: "https://api.anthropic.com/v1", Description: "API Base URL"},
    {Name: "key", Type: "string", Required: true, Description: "API Key"},
  },
})
```

### Google Gemini
```go
RegisterChannelType(ChannelTypeInfoV2{
  ID: Gemini,
  Name: "gemini",
  Label: "Google Gemini",
  Category: "official",
  Description: "Google Gemini 官方及兼容协议",
  Template: ChannelTypeTemplate{
    {Name: "api_base", Type: "string", Required: true, Default: "https://generativelanguage.googleapis.com/v1beta", Description: "API Base URL"},
    {Name: "key", Type: "string", Required: true, Description: "API Key"},
  },
})
```

### DeepSeek
```go
RegisterChannelType(ChannelTypeInfoV2{
  ID: DeepSeek,
  Name: "deepseek",
  Label: "DeepSeek",
  Category: "official",
  Description: "DeepSeek 官方及兼容协议",
  Template: ChannelTypeTemplate{
    {Name: "api_base", Type: "string", Required: true, Default: "https://api.deepseek.com/v1", Description: "API Base URL"},
    {Name: "key", Type: "string", Required: true, Description: "API Key"},
  },
})
```

### Replicate
```go
RegisterChannelType(ChannelTypeInfoV2{
  ID: Replicate,
  Name: "replicate",
  Label: "Replicate",
  Category: "official",
  Description: "Replicate 官方及兼容协议",
  Template: ChannelTypeTemplate{
    {Name: "api_base", Type: "string", Required: true, Default: "https://api.replicate.com/v1", Description: "API Base URL"},
    {Name: "key", Type: "string", Required: true, Description: "API Key"},
  },
})
```

### Groq
```go
RegisterChannelType(ChannelTypeInfoV2{
  ID: Groq,
  Name: "groq",
  Label: "Grok",
  Category: "official",
  Description: "Grok 官方及兼容协议",
  Template: ChannelTypeTemplate{
    {Name: "api_base", Type: "string", Required: true, Default: "https://api.groq.com/v1", Description: "API Base URL"},
    {Name: "key", Type: "string", Required: true, Description: "API Key"},
  },
})
```

### 智谱 GLM
```go
RegisterChannelType(ChannelTypeInfoV2{
  ID: GLM,
  Name: "glm",
  Label: "智谱 GLM",
  Category: "official",
  Description: "智谱 GLM 官方及兼容协议",
  Template: ChannelTypeTemplate{
    {Name: "api_base", Type: "string", Required: true, Default: "https://open.bigmodel.cn/api/paas/v4", Description: "API Base URL"},
    {Name: "key", Type: "string", Required: true, Description: "API Key"},
  },
})
```

---

## 5. Adding or Modifying Channel Templates

- To add a new provider, add a new `RegisterChannelType` block in `relay/channeltype/bootstrap.go`.
- Restart the backend to reload templates.
- All templates are exposed via `/api/channel/types` and rendered dynamically in the frontend.
- For advanced options (grouping, dependencies, select fields), see the README and code comments in `template.go`.

---

## 6. Troubleshooting

- If the channel list is empty, ensure the backend is running and `bootstrap.go` is loaded without syntax errors.
- Check logs for Go build/runtime errors.
- For frontend issues, ensure the Modern build is up to date and the API endpoint is reachable.

---

For more details, see:
- `README.md`
- `relay/channeltype/bootstrap.go`
- `relay/channeltype/template.go`
- `controller/channel_types.go`
- `web/modern/src/constants.ts`
