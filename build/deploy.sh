#!/usr/bin/env bash
# 使用 Dockerfile.prebuilt 构建并重新运行 linkvpn 容器
# 使用方式：在项目根目录（含 build/、Dockerfile.prebuilt）执行 ./scripts/deploy.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$SCRIPT_DIR"
# 若从 scripts/deploy.sh 执行，则构建根目录在上一级
if [ ! -f "$SCRIPT_DIR/Dockerfile.prebuilt" ] && [ -f "$SCRIPT_DIR/../Dockerfile.prebuilt" ]; then
  ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
fi
cd "$ROOT_DIR" || exit 1

echo ">>> 工作目录: $ROOT_DIR"
echo ">>> 构建镜像..."
docker build -f Dockerfile.prebuilt -t linkvpn:latest .

echo ">>> 停止并删除旧容器..."
docker rm -f linkvpn 2>/dev/null || true

echo ">>> 启动新容器 (host 网络模式)..."
docker run -d --name linkvpn \
  --network host \
  -e ADMIN_PASS="${ADMIN_PASS:-admin}" \
  -e JWT_SECRET="${JWT_SECRET:-admin}" \
  -v /root/openvpn:/etc/openvpn \
  --device /dev/net/tun \
  --cap-add=NET_ADMIN \
  --restart unless-stopped \
  linkvpn:latest

echo ">>> 完成。面板: http://<服务器IP>:8789  OpenVPN: UDP 1194"
docker ps --filter name=linkvpn

echo ">>> 删除部署目录下的构建产物（server、dist）..."
rm -rf "server" "dist"
echo ">>> 已删除 server、dist"
