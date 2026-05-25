#!/usr/bin/env bash
set -euo pipefail

# eval-codex-support.sh runs a practical compatibility check for UniAPI against
# OpenAI Responses API behavior and codex-like model naming.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

API_BASE="${API_BASE:-http://localhost:3000}"
CODEX_MODEL="${CODEX_MODEL:-codex-mini}"
TIMEOUT_SECONDS="${TIMEOUT_SECONDS:-90}"

mask_secret() {
  local value="$1"
  local len=${#value}
  if (( len <= 8 )); then
    printf '****'
    return
  fi

  printf '%s****%s' "${value:0:4}" "${value:len-4:4}"
}

pick_token() {
  if [[ -n "${API_TOKEN:-}" ]]; then
    printf '%s' "$API_TOKEN"
    return
  fi

  if [[ -f "uniapi.db" ]]; then
    local db_token
    db_token="$(sqlite3 uniapi.db "select key from tokens where status=1 limit 1;" 2>/dev/null || true)"
    if [[ -n "$db_token" ]]; then
      printf '%s' "$db_token"
      return
    fi
  fi

  printf ''
}

pick_baseline_model() {
  if [[ -n "${BASELINE_MODEL:-}" ]]; then
    printf '%s' "$BASELINE_MODEL"
    return
  fi

  if [[ -f "uniapi.db" ]]; then
    local db_model
    db_model="$(sqlite3 uniapi.db "select trim(substr(models,1, instr(models||',',',')-1)) from channels where status=1 and length(models)>0 limit 1;" 2>/dev/null || true)"
    db_model="$(echo "$db_model" | tr -d '\r' | xargs 2>/dev/null || true)"
    if [[ -n "$db_model" ]]; then
      printf '%s' "$db_model"
      return
    fi
  fi

  printf ''
}

post_response() {
  local token="$1"
  local model="$2"
  local input_text="$3"

  curl -sS --max-time "$TIMEOUT_SECONDS" -X POST "$API_BASE/v1/responses" \
    -H "Authorization: Bearer $token" \
    -H "Content-Type: application/json" \
    -d "{\"model\":\"$model\",\"input\":\"$input_text\"}" \
    -w '\nHTTP_STATUS:%{http_code}\n'
}

summarize_result() {
  local label="$1"
  local response="$2"

  local status
  status="$(echo "$response" | sed -n 's/^HTTP_STATUS://p' | tail -n1)"

  local body
  body="$(echo "$response" | sed '/^HTTP_STATUS:/d')"

  local verdict="FAIL"
  local reason="unknown"

  if [[ "$status" =~ ^2 ]]; then
    verdict="PASS"
    reason="protocol_ok"
  elif echo "$body" | grep -qiE 'model_not_found|does not have permission to use the model|invalid model|model.*not.*found'; then
    reason="model_not_configured_or_not_authorized"
  elif echo "$body" | grep -qiE 'No token provided|authorization|invalid api key|unauthorized'; then
    reason="auth_issue"
  else
    reason="protocol_or_upstream_error"
  fi

  printf '%s|%s|%s|%s\n' "$label" "$verdict" "$status" "$reason"
  printf -- '--- %s raw body (truncated) ---\n' "$label"
  echo "$body" | head -c 500
  printf '\n\n'
}

main() {
  echo "[codex-eval] root: $ROOT_DIR"
  echo "[codex-eval] API_BASE: $API_BASE"
  echo "[codex-eval] CODEX_MODEL: $CODEX_MODEL"

  if ! curl -sS --max-time 5 "$API_BASE/" >/dev/null; then
    echo "[codex-eval] ERROR: backend is not reachable at $API_BASE"
    exit 1
  fi

  local token
  token="$(pick_token)"
  if [[ -z "$token" ]]; then
    echo "[codex-eval] ERROR: no API token found. Set API_TOKEN or prepare active token in uniapi.db"
    exit 1
  fi

  local baseline_model
  baseline_model="$(pick_baseline_model)"
  if [[ -z "$baseline_model" ]]; then
    echo "[codex-eval] ERROR: no baseline model found. Set BASELINE_MODEL or configure channel models in DB"
    exit 1
  fi

  echo "[codex-eval] API_TOKEN(masked): $(mask_secret "$token")"
  echo "[codex-eval] BASELINE_MODEL: $baseline_model"

  echo "[codex-eval] Running live probe (rounds=1)..."
  local live_output
  if live_output="$(API_TOKEN="$token" API_BASE="$API_BASE" ONEAPI_TEST_MODELS="$baseline_model" go run ./cmd/test live --rounds 1 --concurrency 1 --timeout "${TIMEOUT_SECONDS}s" 2>&1)"; then
    echo "[codex-eval] live_probe: PASS"
  else
    echo "[codex-eval] live_probe: FAIL"
    echo "$live_output" | tail -n 40
  fi

  echo "[codex-eval] Probing Responses API with baseline model..."
  local base_resp
  base_resp="$(post_response "$token" "$baseline_model" "baseline ping")"

  echo "[codex-eval] Probing Responses API with codex model..."
  local codex_resp
  codex_resp="$(post_response "$token" "$CODEX_MODEL" "codex ping")"

  echo
  echo "label|verdict|http_status|reason"
  summarize_result "baseline_model" "$base_resp"
  summarize_result "codex_model" "$codex_resp"

  local base_line codex_line
  base_line="$(summarize_result "baseline_model_summary" "$base_resp" | head -n1)"
  codex_line="$(summarize_result "codex_model_summary" "$codex_resp" | head -n1)"

  local base_verdict codex_reason
  base_verdict="$(echo "$base_line" | cut -d'|' -f2)"
  codex_reason="$(echo "$codex_line" | cut -d'|' -f4)"

  echo "[codex-eval] Final judgment:"
  if [[ "$base_verdict" == "PASS" && "$codex_reason" == "model_not_configured_or_not_authorized" ]]; then
    echo "[codex-eval] Protocol is compatible; codex failure is caused by model/channel configuration."
    exit 0
  fi

  if [[ "$base_verdict" == "PASS" ]]; then
    echo "[codex-eval] Baseline passes, codex result requires manual inspection (possible upstream or policy mismatch)."
    exit 0
  fi

  echo "[codex-eval] Baseline failed; investigate gateway/protocol/credential before codex conclusion."
  exit 1
}

main "$@"
