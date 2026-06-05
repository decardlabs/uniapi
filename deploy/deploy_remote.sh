set -euo pipefail
TS=$(date +%Y%m%d%H%M%S)
BUILD_DIR="/tmp/uniapi-build-$TS"
BIN_NEW="/tmp/uniapi-new-$TS"
BIN_STAGING="/tmp/uniapi-staging-$TS"
TARGET_BIN="/opt/uniapi/uniapi"
BACKUP_BIN="/opt/uniapi/uniapi.bak.$TS"
SERVICE_NAME="uniapi"
LISTEN_PORT="10780"
STAGING_PORT="17980"
GO_VERSION="1.25.0"
GIT_TAG="${GIT_TAG:-dev}"

if ! command -v gcc >/dev/null 2>&1; then
  apt-get update && apt-get install -y build-essential curl ca-certificates
fi

GO_BIN="/usr/local/go/bin/go"
if ! "$GO_BIN" version 2>/dev/null | grep -q 'go1.25'; then
  rm -rf /usr/local/go
  mkdir -p /tmp/go
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o "/tmp/go/go${GO_VERSION}.linux-amd64.tar.gz"
  tar -C /usr/local -xzf "/tmp/go/go${GO_VERSION}.linux-amd64.tar.gz"
fi

mkdir -p "$BUILD_DIR"
LATEST_SRC=$(ls -t /tmp/uniapi_src_*.tgz | head -n1)
tar --warning=no-unknown-keyword -xzf "$LATEST_SRC" -C "$BUILD_DIR"
cd "$BUILD_DIR"

if [ ! -f "web/build/modern/index.html" ]; then
  echo "frontend build artifacts missing: web/build/modern/index.html"
  exit 1
fi

MODULE_PATH=$(sed -n 's/^module[[:space:]]\+//p' go.mod | head -n1)
if [ -z "$MODULE_PATH" ]; then
  echo "failed to parse module path from go.mod"
  exit 1
fi

LDFLAGS="-X ${MODULE_PATH}/common.Version=${GIT_TAG}"
echo "Building uniapi (linux/amd64)..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 "$GO_BIN" build -trimpath -ldflags "$LDFLAGS" -o "$BIN_NEW" .

# === Pre-flight: verify new binary can start and respond ===
echo "Starting pre-flight validation on port ${STAGING_PORT}..."
"$BIN_NEW" --port "$STAGING_PORT" --log-dir "/tmp/uniapi-prefight-$TS" &
NEW_PID=$!
sleep 3
if ! curl -fsS "http://127.0.0.1:${STAGING_PORT}/api/status" > /dev/null 2>&1; then
  echo "Pre-flight FAILED: new binary is broken or unresponsive"
  kill "$NEW_PID" 2>/dev/null || true
  wait "$NEW_PID" 2>/dev/null || true
  rm -f "$BIN_NEW"
  exit 1
fi
echo "Pre-flight passed, new binary is healthy"
kill "$NEW_PID"
wait "$NEW_PID" 2>/dev/null || true

# === Deploy: replace binary and restart service ===
mkdir -p /opt/uniapi
if [ -f "$TARGET_BIN" ]; then
  cp -a "$TARGET_BIN" "$BACKUP_BIN"
  echo "Backed up current binary to $BACKUP_BIN"
fi

rm -f "$TARGET_BIN"
cp -a "$BIN_NEW" "$TARGET_BIN"
chmod +x "$TARGET_BIN"

if id -u uniapi >/dev/null 2>&1; then
  chown uniapi:uniapi "$TARGET_BIN"
fi

if systemctl list-unit-files | grep -q "${SERVICE_NAME}.service"; then
  systemctl restart "${SERVICE_NAME}"
  sleep 3

  if ! systemctl is-active --quiet "${SERVICE_NAME}"; then
    echo "Rollback: new binary failed to start..."
    rm -f "$TARGET_BIN"
    if [ -f "$BACKUP_BIN" ]; then
      cp -a "$BACKUP_BIN" "$TARGET_BIN"
      chmod +x "$TARGET_BIN"
      systemctl restart "${SERVICE_NAME}"
      echo "Rollback complete"
    fi
    exit 1
  fi

  systemctl status "${SERVICE_NAME}" --no-pager | head -n 20
fi

ss -lntp | grep ":${LISTEN_PORT}" || true
echo "DEPLOY_OK backup=$BACKUP_BIN"
