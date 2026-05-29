# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build (includes frontend build)
make build

# Build release binary (stripped + compressed)
make build-release

# Build frontend only
make build-frontend-modern

# Start frontend dev server
make dev

# Run all tests with race detector
go test -race ./...

# Run a single package's tests
go test -race ./controller/...

# Run a single test function
go test -race -run TestGetChannelTypes ./controller/...

# Lint (goimports, gofmt, go vet, golangci-lint, govulncheck)
make lint

# Install dev tooling (golangci-lint, goimports, govulncheck)
make install

# Live channel probing (tests real upstream connectivity)
go run ./cmd/test live

# Run database migrations
go run ./cmd/migrate
```

**Frontend**: Use `yarn` (not npm). Frontend source is in `web/modern/`. Language files in `web/modern/src/i18n/locales/`.

**Local dev setup**: Copy `.env.example` to `.env`. The current template defaults to SQLite via `SQLITE_PATH` for local startup, and you can switch to MySQL with `SQL_DSN` if needed. Redis is optional in local development and uses `REDIS_CONN_STRING` when enabled. Default port is `3000`, default admin credentials are `root` / `123456`. VS Code: `F5` to debug, `Cmd+Shift+B` to build.

**Module path**: `github.com/decardlabs/uniapi` (Go 1.25).

## Architecture

UniAPI is an AI API gateway that aggregates 40+ LLM providers behind a unified OpenAI-compatible API. Users can send requests in **Chat Completion**, **Response API**, or **Claude Messages** formats — the gateway transparently converts between them to match each upstream provider's native format.

### Request Flow

```
Client → Gin Router → Middleware (auth, rate-limit, distribute) → Controller → Relay → Adaptor → Upstream
```

1. **Router** ([router/](router/)) defines three route groups: `/api` (management), `/v1` (relay), and `/dashboard` (web UI).
2. **Middleware** ([middleware/](middleware/)) applies auth (token/user/admin/root), rate limiting, CORS, channel distribution, and API format auto-detection.
3. **Controller** ([controller/](controller/)) contains HTTP handlers for both relay endpoints (text, image, audio, video, response API, Claude messages, realtime) and management endpoints (channels, tokens, users, billing, MCP).
4. **Relay engine** ([relay/controller/](relay/controller/)) handles request processing: validation, model mapping, billing calculation, upstream dispatch, and response conversion.
5. **Adaptor** ([relay/adaptor/](relay/adaptor/)) — each provider has an adaptor implementing the `Adaptor` interface ([relay/adaptor/interface.go](relay/adaptor/interface.go#L313)):
   - `GetRequestURL` — construct the upstream URL
   - `SetupRequestHeader` — set auth/content-type headers
   - `ConvertRequest` / `ConvertImageRequest` / `ConvertClaudeRequest` — translate from OpenAI format to provider-native format
   - `DoRequest` — execute the HTTP call (typically via `DoRequestHelper`)
   - `DoResponse` — parse upstream response, extract usage/billing data
   - Pricing: `GetDefaultModelPricing`, `GetModelRatio`, `GetCompletionRatio`

### Key Packages

| Package                                  | Purpose                                                                               |
| ---------------------------------------- | ------------------------------------------------------------------------------------- |
| [common/config/](common/config/)         | All configuration via environment variables (single package, sectioned by group)      |
| [common/logger/](common/logger/)         | Zap-based structured logging with rotation and retention                              |
| [model/](model/)                         | GORM models and data access. Use GORM for writes, raw SQL for complex reads.          |
| [relay/meta/](relay/meta/)               | Per-request metadata (channel, model, pricing, token info) threaded through the relay |
| [relay/channeltype/](relay/channeltype/) | Channel type constants (50+ provider types)                                           |
| [relay/relaymode/](relay/relaymode/)     | Relay mode constants (ChatCompletions, Embeddings, ResponseAPI, ClaudeMessages, etc.) |
| [relay/billing/](relay/billing/)         | Token billing ratio calculations                                                      |
| [relay/pricing/](relay/pricing/)         | Global pricing manager and model price resolution                                     |
| [relay/streaming/](relay/streaming/)     | SSE streaming helpers                                                                 |
| [relay/mcp/](relay/mcp/)                 | MCP (Model Context Protocol) proxy and aggregation                                    |
| [dto/](dto/)                             | Data transfer objects for API responses                                               |
| [monitor/](monitor/)                     | Prometheus + OpenTelemetry monitoring setup                                           |

### Billing Pipeline

Four-layer pricing resolution ([docs/arch/billing.md](docs/arch/billing.md)):

1. **Channel overrides** — per-model pricing set on a specific channel
2. **Adaptor defaults** — provider-supplied default pricing via `GetDefaultModelPricing`
3. **Global fallback** — configured global model price list
4. **Safe default** — ratio=1 with no completion multiplier

### Channel Type Template System

Channel types and their parameter schemas are defined server-side and served via `GET /api/channel/types`. The Modern frontend dynamically renders channel forms and validation from this template — adding a new channel type or parameter in the backend requires no frontend changes. See [README.md](README.md#channel-parameter-template-mechanism--extension) for details.

### Key Files

- [middleware/distributor.go](middleware/distributor.go) — channel selection, sticky sessions, rate-limit precheck, and retry logic. This is the core routing brain of the gateway.
- [relay/adaptor/openai/](relay/adaptor/openai/) — the **reference adaptor**. Model new provider adaptors after this one.
- [relay/adaptor/openai_compatible/](relay/adaptor/openai_compatible/) — base adaptor for providers with OpenAI-compatible APIs. Many providers embed this.

### Conventions

- **Go 1.25**, modern syntax. Files should not exceed 600 lines for Go, 800 for other code.
- **Error handling**: wrap errors with `github.com/Laisky/errors/v2` (`errors.Wrap`, `errors.Wrapf`, `errors.WithStack`). Never return bare errors. Each error must be either returned or logged — never both.
- **Logging**: use `gmw.GetLogger(c)` once per function and store locally. Use `zap.Error(err)` for errors. Use structured Zap logging, avoid `fmt.Sprintf`.
- **Testing**: use `github.com/stretchr/testify/require` for assertions. Tests run with `-race` flag. Test files use `_test.go` suffix.
- **Time**: always use UTC. Date-range queries should end just before 00:00 of the next day to include the full final day.
- **Context**: thread `context.Context` through call chains. Business logic pulls the context-aware logger from `context.Context`.
- **DB**: `gorm.io/gorm` — use ORM for writes, explicit SQL for complex reads. Minimize DB pressure.
- **I18n**: all UI strings must be in locale files. Language files: `web/modern/src/i18n/locales/`.
- **CSS**: avoid `!important` and inline styles; use CSS classes.
- **Security**: constant-time comparisons for sensitive values. OWASP-compliant password hashing (≥10,000 iterations). Always validate and sanitize user input.
- **Comments**: every exported function and interface must have a comment starting with the function/interface name.
