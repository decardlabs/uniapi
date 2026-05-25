set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "$REPO_ROOT"
TS=$(date +%Y%m%d%H%M%S)
ARCHIVE=/tmp/uniapi_src_${TS}.tgz
REMOTE_HOST=43.165.186.252
REMOTE_USER=root
REMOTE_SCRIPT=/tmp/uniapi_deploy_remote_${TS}.sh
GIT_TAG=$(git describe --tags --always 2>/dev/null || echo dev)

# 1) Build modern frontend assets before packaging.
make build-frontend-modern

# 2) Package source
tar --exclude='.git' --exclude='logs' --exclude='uniapi_linux_amd64' --exclude='uniapi' --exclude='web/modern/node_modules' --exclude='web/modern/.yarn/cache' -czf "$ARCHIVE" .
ls -lh "$ARCHIVE"

# 3) Upload source archive
scp -o ConnectTimeout=10 "$ARCHIVE" "${REMOTE_USER}@${REMOTE_HOST}:/tmp/"

# 4) Upload unified remote deploy script
scp -o ConnectTimeout=10 "${SCRIPT_DIR}/deploy_remote.sh" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_SCRIPT}"

# 5) Remote build + upgrade
ssh -o ConnectTimeout=10 "${REMOTE_USER}@${REMOTE_HOST}" "chmod +x ${REMOTE_SCRIPT} && GIT_TAG='${GIT_TAG}' bash ${REMOTE_SCRIPT} && rm -f ${REMOTE_SCRIPT}"
