#!/bin/bash
# uniapi 自动化环境搭建脚本
# 适用于 Ubuntu/Debian 系统
# 请以 root 或 sudo 权限运行

set -e

# 1. 安装依赖
apt update
apt install -y git curl build-essential mysql-client redis-server

# 安装 Go（如已安装可跳过）
if ! command -v go &> /dev/null; then
  curl -LO https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
  tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
  echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
  export PATH=$PATH:/usr/local/go/bin
fi

echo "Go 版本: $(go version)"

# 安装 Node.js 和 yarn（如已安装可跳过）
if ! command -v node &> /dev/null; then
  curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
  apt install -y nodejs
fi
if ! command -v yarn &> /dev/null; then
  npm install -g yarn
fi

echo "Node 版本: $(node -v)"
echo "Yarn 版本: $(yarn -v)"

# 2. 拉取代码（如已拉取可跳过）
# git clone git@github.com:你的仓库/one-api.git
# cd one-api

# 3. 安装 Go 依赖
cd ~/Documents/uniapi
export PATH=$PATH:/usr/local/go/bin
go mod tidy

# 4. 构建后端
make build

# 5. 安装前端依赖并构建
cd web/modern
yarn install
yarn build
cd ../../

# 6. 数据库和 Redis 检查
systemctl enable redis-server --now

# 7. 可选：安装 Delve 调试工具
go install github.com/go-delve/delve/cmd/dlv@latest

# 8. 启动后端
./uniapi &

# 9. 启动前端开发环境（可选）
# cd web/modern && yarn dev &

# 10. 提示
cat <<EOF

环境搭建完成！
- 后端已启动，前端已构建。
- 如需生产部署，请配置 systemd 服务（见 deploy/uniapi.service）。
- 如需调试，推荐使用 Delve (dlv)。
- 如需自定义环境变量，请编辑 .env 文件。

EOF
