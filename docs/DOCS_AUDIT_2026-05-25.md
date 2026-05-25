# Documentation Consistency Audit (2026-05-25)

## Scope

This audit focused on active documentation and excluded archival planning/reporting paths:

- Included: root docs, docs/arch, docs/manuals, docs/refs, package READMEs
- Excluded from mandatory rename: docs/legacy, docs/reviews, docs/superpowers, tool memory dirs

## Method

1. Enumerated all Markdown files.
2. Scanned for naming and API mismatches (`one-api`, service name, endpoint paths, env var names, startup commands).
3. Verified endpoint and env-key facts against code:
   - `router/relay.go`
   - `router/api.go`
   - `common/config/config.go`
4. Applied fixes only where statements describe current behavior (not historical links or external upstream names).

## Facts Verified Against Code

- Response API endpoint is `/v1/responses`.
- Status endpoint is `/api/status`.
- Current binary name in docs should be `uniapi`.
- Systemd service in this repo should be referenced as `uniapi.service`.
- DB env key should be `SQL_DSN` (plus `SQLITE_PATH` fallback).

## Files Updated in This Audit

- `README.md`
  - Unified active naming to `UniAPI` in non-historical prose.
  - Fixed stale TOC anchor section.
  - Kept historical PR links and Docker image references unchanged.
- `DEV_SETUP_GUIDE.md`
  - Updated startup binary, service name, frontend build command, and SSH host.
- `CONFIG_GUIDE.md`
  - Updated startup binary and DB env key guidance.
- `docs/API接口文档-完整版.md`
  - Corrected Response API path (`/v1/responses`).
  - Corrected env key (`SQL_DSN`) and startup binary (`./uniapi`).
- `docs/DOCS_INDEX.md`
  - Added documentation taxonomy and consistency rules.

## Intentional Non-Changes

The following references were retained because they represent historical or external compatibility context:

- GitHub links to upstream repositories/PRs (`songquanpeng/one-api`, `Laisky/one-api`).
- Docker image names currently published as `ppcelery/one-api`.
- Reference/demo documents that quote upstream naming conventions.

## Remaining High-Volume Mentions

Large counts remain mainly in:

- `docs/refs/openai_websearch_demo.md`
- `docs/manuals/k8s.md`

These are mostly compatibility/demo context and should be handled in a separate pass only if you want full branding normalization (which may require changing deployment object names and examples).

## Recommendation

- Treat this audit as baseline for active docs.
- In future doc PRs, enforce the rules in `docs/DOCS_INDEX.md` section "Consistency Rules".
