set -euo pipefail
cd /Users/macairm5/Documents/uniapi
TS=$(date +%Y%m%d%H%M%S)
ARCHIVE=/tmp/uniapi_src_${TS}.tgz

# 1) Package source
tar --exclude='.git' --exclude='logs' --exclude='uniapi_linux_amd64' --exclude='uniapi' --exclude='web/modern/node_modules' --exclude='web/modern/.yarn/cache' -czf "$ARCHIVE" .
ls -lh "$ARCHIVE"

# 2) Upload source archive
scp -o ConnectTimeout=10 "$ARCHIVE" root@43.165.186.252:/tmp/

# 3) Remote build + deploy
ssh -o ConnectTimeout=10 root@43.165.186.252 "bash -s" << 'REMOTE'
set -euo pipefail
TS=\$(date +%Y%m%d%H%M%S)
BUILD_DIR=/tmp/uniapi-build-\$TS
BIN_NEW=/tmp/uniapi-new-\$TS
TARGET_BIN=/opt/uniapi/uniapi
BACKUP_BIN=/opt/uniapi/uniapi.bak.\$TS

# install build deps
if ! command -v gcc >/dev/null 2>&1; then
  apt-get update && apt-get install -y build-essential curl ca-certificates
fi

# Go 1.23.6
if ! /usr/local/go/bin/go version 2>/dev/null | grep -q 'go1.23.6'; then
  rm -rf /usr/local/go
  curl -fsSL https://go.dev/dl/go1.23.6.linux-amd64.tar.gz -o /tmp/go1.23.6.linux-amd64.tar.gz
  tar -C /usr/local -xzf /tmp/go1.23.6.linux-amd64.tar.gz
fi
export PATH=/usr/local/go/bin:\$PATH

mkdir -p "\$BUILD_DIR"
LATEST_SRC=\$(ls -t /tmp/uniapi_src_*.tgz | head -n1)
tar -xzf "\$LATEST_SRC" -C "\$BUILD_DIR"
cd "\$BUILD_DIR"

CGO_ENABLED=1 GOOS=linux GOARCH=amd64 /usr/local/go/bin/go build -o "\$BIN_NEW" .

mkdir -p /opt/uniapi
if [ -f "\$TARGET_BIN" ]; then
  cp -a "\$TARGET_BIN" "\$BACKUP_BIN"
fi
cp -a "\$BIN_NEW" "\$TARGET_BIN"
chmod +x "\$TARGET_BIN"

if id -u uniapi >/dev/null 2>&1; then
  chown uniapi:uniapi "\$TARGET_BIN"
fi

if systemctl list-unit-files | grep -q uniapi.service; then
  systemctl daemon-reload || true
  systemctl restart uniapi
  sleep 2
  if ! systemctl is-active --quiet uniapi; then
    echo 'Rollback...'
    [ -f "\$BACKUP_BIN" ] && cp -a "\$BACKUP_BIN" "\$TARGET_BIN" && systemctl restart uniapi
    exit 1
  fi
  systemctl status uniapi --no-pager | head -n 20
fi

ss -lntp | grep ':10780' || true
echo "DEPLOY_OK backup=\$BACKUP_BIN"
REMOTE
