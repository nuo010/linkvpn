#!/usr/bin/env bash
set -eu

IMAGE_NAME="liguanglong1234/linkvpnclient"

VERSION="${1:-}"
if [ -z "${VERSION}" ]; then
  echo "用法: $0 <版本号>"
  echo "例如: $0 1.0   将构建并推送:"
  echo "  ${IMAGE_NAME}:1.0-aarch64"
  echo "  ${IMAGE_NAME}:1.0-x86_64"
  echo "  ${IMAGE_NAME}:1.0        (多架构 manifest，包含上述两种架构)"
  exit 1
fi

echo "[linkvpn] 将构建并推送版本: ${VERSION}"

if ! command -v docker >/dev/null 2>&1; then
  echo "[linkvpn] 未找到 docker 命令，请先安装 Docker。" >&2
  exit 1
fi

if ! docker buildx version >/dev/null 2>&1; then
  echo "[linkvpn] 当前 Docker 不支持 buildx，请确认 Docker 版本并启用 buildx。" >&2
  exit 1
fi

BUILDER_NAME="linkvpnclient-builder"

if ! docker buildx inspect "${BUILDER_NAME}" >/dev/null 2>&1; then
  docker buildx create --name "${BUILDER_NAME}" --use >/dev/null
else
  docker buildx use "${BUILDER_NAME}" >/dev/null
fi

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t "${IMAGE_NAME}:${VERSION}-x86_64" \
  -t "${IMAGE_NAME}:${VERSION}-aarch64" \
  -t "${IMAGE_NAME}:${VERSION}" \
  -f Dockerfile \
  . \
  --push

echo "[linkvpn] 已推送镜像:"
echo "  ${IMAGE_NAME}:${VERSION}-x86_64"
echo "  ${IMAGE_NAME}:${VERSION}-aarch64"
echo "  ${IMAGE_NAME}:${VERSION}        (多架构 manifest)"
