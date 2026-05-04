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
```

**Frontend**: Use `yarn` (not npm). Frontend source is in `web/modern/`. Language files in `web/modern/src/i18n/locales/`.

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

| Package | Purpose |
|---------|---------|
| [common/config/](common/config/) | All configuration via environment variables (single package, sectioned by group) |
| [common/logger/](common/logger/) | Zap-based structured logging with rotation and retention |
| [model/](model/) | GORM models and data access. Use GORM for writes, raw SQL for complex reads. |
| [relay/meta/](relay/meta/) | Per-request metadata (channel, model, pricing, token info) threaded through the relay |
| [relay/channeltype/](relay/channeltype/) | Channel type constants (50+ provider types) |
| [relay/relaymode/](relay/relaymode/) | Relay mode constants (ChatCompletions, Embeddings, ResponseAPI, ClaudeMessages, etc.) |
| [relay/billing/](relay/billing/) | Token billing ratio calculations |
| [relay/pricing/](relay/pricing/) | Global pricing manager and model price resolution |
| [relay/streaming/](relay/streaming/) | SSE streaming helpers |
| [relay/mcp/](relay/mcp/) | MCP (Model Context Protocol) proxy and aggregation |
| [dto/](dto/) | Data transfer objects for API responses |
| [monitor/](monitor/) | Prometheus + OpenTelemetry monitoring setup |

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
