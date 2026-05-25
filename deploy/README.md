# deploy 目录说明

本目录用于集中存放项目部署相关文件，避免与源码根目录混杂。

## 文件清单

- `deploy.sh`：本地执行的远程升级入口脚本。
- `deploy_remote.sh`：上传到目标服务器后执行的远端构建与替换脚本。
- `setup_auto.sh`：Ubuntu/Debian 环境自动化安装与构建辅助脚本。
- `uniapi.service`：systemd 服务示例配置文件。
- `Dockerfile`：容器镜像构建文件。

## 使用建议

- 远程升级：在仓库根目录执行 `bash deploy/deploy.sh`。
- systemd 部署参考：使用 `deploy/uniapi.service` 作为模板并按实际环境调整。
- 容器构建：在仓库根目录执行 `docker build -f deploy/Dockerfile .`。

## 迁移说明

本目录由根目录部署文件归档而来，目的是统一部署资产位置并降低根目录复杂度。
