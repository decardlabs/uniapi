# Closed Loop Testing For UniAPI

## Purpose

This document defines an end-to-end quality loop for UniAPI, covering backend, frontend, protocol conversion, real-channel probing, and production verification.

The loop is:

1. Code change enters pull request.
2. PR quality gates block regressions.
3. Main branch regression validates build and protocol compatibility.
4. Post-deploy probe verifies real traffic behavior.
5. Any incident must be converted into automated tests and re-enter the gate.

## Layers

### L0 Static And Unit Layer

- Backend: `go vet ./...`, `go test -race ./...`
- Frontend: `yarn type-check`, `yarn lint`, `yarn test --run`

### L1 Component And Contract Layer

- Frontend page and hook tests under `web/modern/src/**`.
- Backend tests in `*_test.go` files.
- Contract and format conversion behavior covered by relay/controller tests and cmd/test matrix.

### L2 E2E Smoke Layer

- Playwright smoke spec: `web/modern/e2e/smoke.spec.ts`
- Goal: verify service readiness and the login page's primary auth form rendering in CI.

### L3 Protocol Matrix Layer

- Command: `go run ./cmd/test run`
- Validates Chat Completion, Response API, and Claude Messages compatibility across selected models.

### L4 Live Probe Layer

- Command: `go run ./cmd/test live`
- Validates practical channel behavior:
  - `response_create`
  - `response_continue`
  - `chat_completion`
  - `chat_tool_invocation` (skipped for DeepSeek according to harness behavior)

## Workflows

### PR Gate

File: `.github/workflows/pr-gate.yml`

- Runs on pull requests to `main`.
- Contains three jobs:
  - backend vet and race tests
  - frontend typecheck/lint/unit test
  - chromium smoke e2e against a started local `uniapi` process

### Main Regression

File: `.github/workflows/main-regression.yml`

- Runs on push to `main`, nightly schedule, and manual trigger.
- Baseline job validates build + core test stability.
- Protocol matrix job runs `cmd/test run` when secret token is configured.
- Live probe job runs on schedule/manual to avoid slowing every push.

### Post Deploy Probe

File: `.github/workflows/post-deploy-probe.yml`

- Triggered manually or by `repository_dispatch`.
- Probes deployed environment with configurable base URL/model/rounds/timeout.
- Fails fast when live probe reports regressions.

## Required Secrets

- `ONEAPI_TEST_TOKEN`: token used by cmd/test regression jobs.
- `ONEAPI_TEST_BASE`: required base URL for matrix/live probe.
- `ONEAPI_TEST_MODELS`: optional model list for matrix job.
- `ONEAPI_TEST_PROBE_MODEL`: optional default model for scheduled live probe.

## Operational Rules

1. Do not merge PRs when PR Gate fails.
2. Treat scheduled live probe failures as release risks and investigate immediately.
3. For every production incident, add or update at least one automated test in the relevant layer.
4. Keep model list for matrix job small but representative to control runtime and cost.

## Rollout Plan

1. Enable PR Gate as required check.
2. Run Main Regression in non-blocking mode for one week and tune model set.
3. Enable Post Deploy Probe in deployment automation and treat failures as release blockers.
4. Review flaky tests weekly and stabilize before widening scope.
