# UniAPI Documentation Index

This index classifies project documentation by purpose and marks which files are the current source of truth.

## 1) Core Entry Docs

- Product overview and quick start: [README.md](../README.md)
- Local development setup: [DEV_SETUP_GUIDE.md](../DEV_SETUP_GUIDE.md)
- Configuration and channel templates: [CONFIG_GUIDE.md](../CONFIG_GUIDE.md)
- Design guidelines: [DESIGN.md](../DESIGN.md)

## 2) Architecture Docs

Path: [docs/arch](./arch)

- API conversion and protocol adaptation
- Billing and pricing model
- Dashboard and frontend architecture
- Logging, tracing, graceful restart, MCP aggregation

## 3) Operations Manuals

Path: [docs/manuals](./manuals)

- Kubernetes deployment: [docs/manuals/k8s.md](./manuals/k8s.md)
- OpenTelemetry and Prometheus operations
- Channel and billing operations
- Closed-loop testing and regression manuals

## 4) Reference Notes

Path: [docs/refs](./refs)

- Provider-specific capabilities and protocol notes (OpenAI, Claude, Gemini, AWS, MCP, etc.)
- Integration references and behavior notes used during implementation

## 5) Package-level Docs

- Frontend package: [web/README.md](../web/README.md)
- Migration tool: [cmd/migrate/README.md](../cmd/migrate/README.md)
- Regression tool: [cmd/test/README.md](../cmd/test/README.md)

## 6) Historical / Archive Docs

- Legacy plans and upgrade notes: [docs/legacy](./legacy)
- Point-in-time review reports: [docs/reviews](./reviews)
- Generated planning artifacts: [docs/superpowers/specs](./superpowers/specs)

## 7) Consistency Rules

When updating docs, keep these code-aligned facts consistent:

- Binary name: `uniapi` (default build output in [Makefile](../Makefile))
- Systemd service name: `uniapi.service`
- Response API endpoint: `/v1/responses` (router definition in [router/relay.go](../router/relay.go))
- Status endpoint: `/api/status` (router definition in [router/api.go](../router/api.go))
- Database env key: `SQL_DSN` (plus `SQLITE_PATH`) from [common/config/config.go](../common/config/config.go)
- Frontend production build command: `make build-frontend-modern`

## 8) Audit Scope (2026-05-25)

Checked and aligned in this pass:

- [README.md](../README.md)
- [DEV_SETUP_GUIDE.md](../DEV_SETUP_GUIDE.md)
- [CONFIG_GUIDE.md](../CONFIG_GUIDE.md)
- [docs/API接口文档-完整版.md](./API接口文档-完整版.md)

Audit report:

- [docs/DOCS_AUDIT_2026-05-25.md](./DOCS_AUDIT_2026-05-25.md)

Recommended follow-up audits:

- Sweep docs for legacy naming where context is not historical reference.
- Keep deployment section and scripts synchronized after each release process change.
