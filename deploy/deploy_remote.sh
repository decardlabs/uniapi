set -euo pipefail
TS=$(date +%Y%m%d%H%M%S)
BUILD_DIR="/tmp/uniapi-build-$TS"
BIN_NEW="/tmp/uniapi-new-$TS"
TARGET_BIN="/opt/uniapi/uniapi"
BACKUP_BIN="/opt/uniapi/uniapi.bak.$TS"
SERVICE_NAME="uniapi"
LISTEN_PORT="10780"
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
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 "$GO_BIN" build -trimpath -ldflags "$LDFLAGS" -o "$BIN_NEW" .

mkdir -p /opt/uniapi

# Upgrade only: restart existing systemd service in place.
if systemctl list-unit-files | grep -q "${SERVICE_NAME}.service"; then
  systemctl stop "${SERVICE_NAME}" || true
fi

if [ -f "$TARGET_BIN" ]; then
  cp -a "$TARGET_BIN" "$BACKUP_BIN"
fi

# Use 'cp' with follow-links or just remove the old file first to avoid busy errors
rm -f "$TARGET_BIN"
cp -a "$BIN_NEW" "$TARGET_BIN"
chmod +x "$TARGET_BIN"

if id -u uniapi >/dev/null 2>&1; then
  chown uniapi:uniapi "$TARGET_BIN"
fi

if systemctl list-unit-files | grep -q "${SERVICE_NAME}.service"; then
  systemctl daemon-reload || true
  systemctl start "${SERVICE_NAME}"
  sleep 2
  if ! systemctl is-active --quiet "${SERVICE_NAME}"; then
    echo "Rollback..."
    rm -f "$TARGET_BIN"
    [ -f "$BACKUP_BIN" ] && cp -a "$BACKUP_BIN" "$TARGET_BIN" && systemctl restart "${SERVICE_NAME}"
    exit 1
  fi
  systemctl status "${SERVICE_NAME}" --no-pager | head -n 20
fi

ss -lntp | grep ":${LISTEN_PORT}" || true
echo "DEPLOY_OK backup=$BACKUP_BIN"
