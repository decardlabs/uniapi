# UniAPI

[![Go Version](https://img.shields.io/badge/Go-1.25-blue)]()
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react)]()
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?logo=typescript)]()

Open‑source version of OpenRouter — a unified AI API gateway that aggregates 40+ LLM providers behind a single OpenAI-compatible API.

- Aggregates chat, image, speech, TTS, embeddings, rerank and more
- 40+ providers: OpenAI, Anthropic, Google, DeepSeek, Azure, AWS Bedrock, etc.
- Automatic format conversion between Chat Completions, Response API, and Claude Messages
- Multi‑tenant management with per‑tenant quotas and permissions
- Sub‑API Keys with per‑key model binding and quota limits

---

## Table of Contents

- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Why UniAPI](#why-uniapi)
- [Build Modes](#build-modes)
- [Features](#features)
- [API Compatibility Matrix](#api-compatibility-matrix)
- [Development](#development)
- [Deployment](#deployment)
- [Extending Channels](#extending-channels)
- [Documentation](#documentation)
- [Changelog](#changelog)
- [Contributors](#contributors)

---

## Quick Start

```sh
# 1. Backend
cp .env.example .env    # SQLite by default, edit for MySQL/Redis
go run .                # starts on http://localhost:3000

# 2. Frontend (development mode with hot-reload)
make dev                # starts on http://localhost:3001

# 3. Open http://localhost:3001, register, and log in.
#    The first registered user becomes admin. Default: root / 123456
```

> 💡 Use `go run .` (not `go run main.go`) — the project uses multiple files in the `main` package including embedded frontend assets.

---

## Architecture

```
Client → Gin Router → Middleware (auth, rate-limit, distribute) → Controller → Relay → Adaptor → Upstream
```

### Key Packages

| Package | Purpose |
|---|---|
| [middleware/distributor.go](middleware/distributor.go) | Channel selection, sticky sessions, rate-limit precheck, retry logic |
| [relay/adaptor/](relay/adaptor/) | Per-provider adaptors implementing `Adaptor` interface |
| [relay/adaptor/openai/](relay/adaptor/openai/) | Reference adaptor — model new ones after this |
| [relay/adaptor/openai_compatible/](relay/adaptor/openai_compatible/) | Base adaptor for OpenAI-compatible providers |
| [relay/meta/](relay/meta/) | Per-request metadata (channel, model, pricing, tokens) |
| [relay/billing/](relay/billing/) | Token billing ratio calculations |
| [relay/pricing/](relay/pricing/) | Global pricing manager, model price resolution |
| [relay/channeltype/](relay/channeltype/) | Channel type constants (50+ provider types) |
| [common/config/](common/config/) | All configuration via environment variables |
| [model/](model/) | GORM models and data access |

### Billing Pipeline (4-layer)

```
Channel overrides → Adaptor defaults → Global fallback → Safe default (ratio=1)
```

See [docs/arch/billing.md](docs/arch/billing.md) for details.

---

## Why UniAPI?

- **统一** — 一个网关聚合所有主流 AI 模型服务商
- **稳定** — 基于 One API 多年生产实践，主线功能可靠
- **现代** — Modern UI，响应式设计、深色模式

---

## Build Modes

| Command | Size | Description |
|---|---|---|
| `make build` | ~87 MB | Standard build with debug symbols, rebuilds frontend |
| `make build-release` | ~59 MB | Stripped (`-s -w -trimpath`), rebuilds frontend |
| `make build-release-no-frontend` | ~59 MB | Stripped, skips frontend rebuild |
| `make build-release-external-static` | ~55 MB | Loads frontend from disk at runtime |

For API-only deployments (frontend served via reverse proxy), use `build-release-no-frontend`.

---

## Features

### Universal

| Feature | Description |
|---|---|
| **I18n** | English, Chinese, French, Spanish, Japanese |
| **Billing** | 4-tier pricing pipeline, cached prompt buckets, per-second media metering |
| **OpenTelemetry** | Trace export via OTLP |
| **Prompt Caching** | Anthropic-style prompt caching support |
| **Thinking/Reasoning** | URL params `thinking` & `reasoning_format` (reasoning_content / reasoning / thinking) |
| **MCP Aggregator** | MCP server aggregation as built-in tools for any model |
| **Cached Input** | Reduced costs via cached prompt tokens |

### Provider Features

| Provider | Supported APIs |
|---|---|
| **OpenAI** | Chat, Response, Images, Audio, Whisper, Sora, GPT-5/4.1/o3/o4 family, Codex CLI |
| **Anthropic** | Claude Messages API, Claude Code, Claude 4.x, Thinking |
| **Google** | Gemini 2.0/2.5/3, Vertex Imagen3/4, multimodal output |
| **DeepSeek** | reasoning_content, cache billing |
| **AWS Bedrock** | Claude, cross-region inference, inference profiles |
| **Azure** | GPT-5 nano and all OpenAI models |
| **OpenRouter** | reasoning content passthrough |
| **Zhipu (GLM)** | GLM-4/5, CogView, GLM OCR, GLM-4V vision |
| **MiniMax** | Chat, Response API fallback |
| **Moonshot (Kimi)** | K2 model family |
| **Replicate** | Flux, Flux Remix, chat models |
| **Cohere** | Command R, Rerank |
| **Coze** | OAuth authentication |
| **Doubao** | ByteDance volcanic engine |
| **XAI/Grok** | Text & image models |
| **Black Forest Labs** | Flux Kontext Pro |

### Multi-format Support

UniAPI transparently converts between these request formats:

- **OpenAI Chat Completions** (`/v1/chat/completions`)
- **OpenAI Response API** (`/v1/responses`) with WebSocket support
- **Claude Messages** (`/v1/messages`)
- **OpenAI Images** (`/v1/images/generations`, `/v1/images/edits`)
- **OpenAI Audio** (`/v1/audio/transcriptions`)
- **OpenAI Video** (`/v1/videos`) — Sora integration
- **Rerank** (`/v1/rerank`)

### Request Cost Tracking

Every response includes an `X-Oneapi-Request-Id` header. Retrieve cost breakdown at `GET /api/cost/request/:request_id`.

### External Billing

Update user quotas via API key:

```sh
curl -X POST https://your-host/api/token/consume \
  -H "Authorization: Bearer <TOKEN_API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"add_reason": "async-transcode", "add_used_quota": 150}'
```

See [docs/manuals/external_billing.md](docs/manuals/external_billing.md).

---

## API Compatibility Matrix

Tested against 8 models across 21 request formats (2025-12-12):

```
Request Format                         gpt-4o-mini  gpt-5-mini   claude-haiku-4-5  gemini-2.5-flash  openai/gpt-oss-20b  deepseek-chat  grok-4-1-fast-non-reasoning  azure-gpt-5-nano
Chat (stream=false)                    PASS 11.21s  PASS 13.10s  PASS 8.52s        PASS 4.64s        PASS 9.52s          PASS 7.08s     PASS 3.08s                 PASS 14.68s
Chat (stream=true)                     PASS 13.23s  PASS 13.37s  PASS 2.31s        PASS 6.02s        PASS 4.56s          PASS 10.92s    PASS 9.72s                 PASS 15.30s
Chat Tools (stream=false)              PASS 5.60s   PASS 12.94s  PASS 7.69s        PASS 7.11s        PASS 3.14s          PASS 8.71s     PASS 5.48s                 PASS* 35.02s
Chat Tools (stream=true)               PASS 14.51s  PASS 18.90s  PASS 7.60s        PASS 4.36s        PASS 8.87s          PASS 7.56s     PASS 7.45s                 PASS 13.13s
Chat Tools History (stream=false)      PASS 9.09s   PASS 14.28s  PASS 12.04s       PASS 7.45s        PASS 10.40s         PASS 9.52s     PASS 6.26s                 PASS 13.61s
Chat Tools History (stream=true)       PASS 14.80s  PASS 25.49s  PASS 3.08s        PASS 11.24s       PASS 5.22s          PASS 4.97s     PASS 5.14s                 PASS 15.56s
Chat Structured (stream=false)         PASS 10.51s  PASS 15.71s  PASS 12.66s       PASS 13.68s       PASS 8.24s          PASS 6.95s     PASS 13.42s                PASS 13.80s
Chat Structured (stream=true)          PASS 11.26s  PASS 14.50s  PASS 6.07s        PASS 4.84s        PASS 6.97s          PASS 6.86s     PASS 4.51s                 PASS 14.04s
Response (stream=false)                PASS 14.65s  PASS 15.31s  PASS 10.51s       PASS 3.03s        PASS 3.98s          PASS 12.83s    PASS 11.29s                PASS 15.70s
Response (stream=true)                 PASS 8.91s   PASS 17.54s  PASS 6.51s        PASS 5.81s        PASS 5.26s          PASS 7.56s     PASS 9.51s                 PASS 15.66s
Response Vision (stream=false)         PASS 12.32s  PASS 14.49s  PASS 14.12s       PASS 8.82s        SKIP                SKIP           PASS 8.74s                 PASS 16.59s
Response Vision (stream=true)          PASS 11.04s  PASS 9.50s   PASS 10.75s       PASS 13.60s       SKIP                SKIP           PASS 9.05s                 PASS 11.51s
Response Tools (stream=false)          PASS 11.02s  PASS 11.71s  PASS 7.68s        PASS 10.55s       PASS 4.04s          PASS 10.30s    PASS 10.15s                PASS 12.93s
Response Tools (stream=true)           PASS 8.64s   PASS 14.40s  PASS 10.73s       PASS 13.20s       PASS 6.81s          PASS 7.62s     PASS 13.42s                PASS 12.03s
Response Tools History (stream=false)  PASS 8.04s   PASS 14.45s  PASS 9.63s        PASS 5.54s        PASS 5.88s          PASS 9.30s     PASS 5.22s                 PASS 11.11s
Response Tools History (stream=true)   PASS 9.89s   PASS 12.22s  PASS 6.58s        PASS 5.18s        PASS 7.40s          PASS 5.84s     PASS 4.50s                 PASS 16.86s
Response Structured (stream=false)     PASS 14.35s  PASS 15.40s  PASS 12.74s       PASS 12.78s       PASS 7.59s          PASS 5.99s     PASS 12.10s                PASS 13.18s
Response Structured (stream=true)      PASS 15.04s  PASS 12.68s  PASS 12.52s       PASS 7.83s        PASS 7.85s          PASS 3.81s     PASS 8.35s                 PASS 11.01s
Claude (stream=false)                  PASS 4.78s   PASS 11.79s  PASS 12.18s       PASS 10.58s       PASS 8.75s          PASS 12.46s    PASS 9.66s                 PASS 14.93s
Claude (stream=true)                   PASS 4.46s   PASS 9.82s   PASS 6.43s        PASS 14.37s       PASS 9.22s          PASS 12.17s    PASS 3.13s                 PASS 20.63s
Claude Tools (stream=false)            PASS 9.20s   PASS 11.08s  PASS 11.79s       PASS 3.55s        PASS 7.39s          PASS 6.32s     PASS 12.71s                PASS 14.85s
Claude Tools (stream=true)             PASS 3.01s   PASS 6.56s   PASS 14.15s       PASS 8.11s        PASS 9.11s          PASS 8.37s     PASS 4.16s                 PASS 12.80s
Claude Tools History (stream=false)    PASS 9.67s   PASS 15.07s  PASS 7.45s        PASS 6.70s        PASS 8.47s          PASS 9.25s     PASS 13.92s                PASS 15.36s
Claude Tools History (stream=true)     PASS 11.15s  PASS 19.37s  PASS 13.52s       PASS 8.90s        PASS 7.20s          PASS 8.89s     PASS 5.81s                 PASS 9.87s
Claude Structured (stream=false)       PASS 5.39s   SKIP         PASS 7.89s        PASS 11.51s       PASS 13.30s         PASS 8.31s     PASS 6.16s                 SKIP
Claude Structured (stream=true)        PASS 6.43s   SKIP         PASS 11.05s       PASS 9.62s        PASS 3.05s          PASS 4.64s     PASS 4.69s                 SKIP

Totals  | Requests: 208 | Passed: 200 | Failed: 0 | Skipped: 8
```

---

## Development

```sh
# Lint
make lint

# Test
go test -race ./...

# Live channel probing
go run ./cmd/test live

# Database migrations
go run ./cmd/migrate
```

### VS Code

| Action | Key / Task |
|---|---|
| Build | `Cmd+Shift+B` — select `build-frontend-modern` or `build` |
| Debug | `F5` — selects `Debug Main`, builds frontend then starts Go debugger |
| Frontend dev | Run `make dev` or use the `Dev Frontend` launch config |

### Environment Variables

All configuration via environment variables. See [common/config/config.go](common/config/config.go) for the full list. Key ones:

| Variable | Default | Description |
|---|---|---|
| `PORT` | `3000` | Server port |
| `GIN_MODE` | `debug` | Gin framework mode |
| `SESSION_SECRET` | auto | Session encryption key |
| `SQLITE_PATH` | `./data/uniapi.db` | SQLite database path |
| `SQL_DSN` | — | MySQL DSN (takes precedence over SQLite) |
| `REDIS_CONN_STRING` | — | Redis connection string |
| `RELAY_TIMEOUT` | `300` | Upstream request timeout (seconds) |

---

## Deployment

### Docker Compose

```yaml
services:
  uniapi:
    image: ppcelery/one-api:latest
    restart: unless-stopped
    volumes:
      - /var/lib/oneapi:/data
    ports:
      - 3000:3000
```

Images: `ppcelery/one-api:latest`, `ppcelery/one-api:arm64-latest`

### Kubernetes

See [docs/manuals/k8s.md](docs/manuals/k8s.md).

### Remote Server Upgrade

```sh
bash deploy/deploy.sh
```

This builds the frontend locally, uploads to the server, builds Linux AMD64 binary, and hot-replaces the service with rollback on failure. See [deploy/deploy.sh](deploy/deploy.sh).

---

## Extending Channels

Channel types and their parameter schemas are defined server-side and served via `GET /api/channel/types`. The frontend dynamically renders forms and validation from this template — adding a new parameter requires **zero frontend changes**.

To add a new channel type:

1. Register the type and its parameter template in the backend registry
2. Add i18n keys for labels/descriptions
3. (Optional) Extend the frontend renderer for custom input widgets

**Key files:**
- [relay/channeltype/helper.go](relay/channeltype/helper.go) — channel type registry
- [web/modern/src/components/channel/ChannelDynamicParams.tsx](web/modern/src/components/channel/ChannelDynamicParams.tsx) — dynamic form renderer

---

## Documentation

| Document | Description |
|---|---|
| [CHANGELOG.md](CHANGELOG.md) | Full version history |
| [docs/arch/billing.md](docs/arch/billing.md) | Billing pipeline internals |
| [docs/manuals/billing.md](docs/manuals/billing.md) | Billing operations guide |
| [docs/manuals/mcp_aggregator.md](docs/manuals/mcp_aggregator.md) | MCP aggregator setup |
| [docs/manuals/k8s.md](docs/manuals/k8s.md) | Kubernetes deployment |
| [docs/manuals/closed-loop-testing.md](docs/manuals/closed-loop-testing.md) | Testing & CI |
| [docs/manuals/external_billing.md](docs/manuals/external_billing.md) | External billing API |

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for the complete version history.

---

## Contributors

<a href="https://github.com/Laisky/one-api/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=Laisky/one-api" />
</a>
