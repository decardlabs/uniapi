set -euo pipefail
TS=$(date +%Y%m%d%H%M%S)
BUILD_DIR="/tmp/uniapi-build-$TS"
BIN_NEW="/tmp/uniapi-new-$TS"
TARGET_BIN="/opt/uniapi/uniapi"
BACKUP_BIN="/opt/uniapi/uniapi.bak.$TS"

if ! command -v gcc >/dev/null 2>&1; then
  apt-get update && apt-get install -y build-essential curl ca-certificates
fi

GO_BIN="/usr/local/go/bin/go"
if ! "$GO_BIN" version 2>/dev/null | grep -q 'go1.23.6'; then
  rm -rf /usr/local/go
  mkdir -p /tmp/go
  curl -fsSL https://go.dev/dl/go1.23.6.linux-amd64.tar.gz -o /tmp/go/go1.23.6.linux-amd64.tar.gz
  tar -C /usr/local -xzf /tmp/go/go1.23.6.linux-amd64.tar.gz
fi

mkdir -p "$BUILD_DIR"
LATEST_SRC=$(ls -t /tmp/uniapi_src_*.tgz | head -n1)
tar --warning=no-unknown-keyword -xzf "$LATEST_SRC" -C "$BUILD_DIR"
cd "$BUILD_DIR"

CGO_ENABLED=1 GOOS=linux GOARCH=amd64 "$GO_BIN" build -o "$BIN_NEW" .

mkdir -p /opt/uniapi

# Stop service if running to avoid "Text file busy"
if systemctl list-unit-files | grep -q uniapi.service; then
    systemctl stop uniapi || true
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

if systemctl list-unit-files | grep -q uniapi.service; then
  systemctl daemon-reload || true
  systemctl start uniapi
  sleep 2
  if ! systemctl is-active --quiet uniapi; then
    echo "Rollback..."
    rm -f "$TARGET_BIN"
    [ -f "$BACKUP_BIN" ] && cp -a "$BACKUP_BIN" "$TARGET_BIN" && systemctl restart uniapi
    exit 1
  fi
  systemctl status uniapi --no-pager | head -n 20
fi

ss -lntp | grep ':10780' || true
echo "DEPLOY_OK backup=$BACKUP_BIN"
