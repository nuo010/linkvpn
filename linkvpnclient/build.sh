#!/bin/sh
set -eu

# 本地镜像名
LOCAL_IMAGE_NAME="xxgl"
# Docker Hub 仓库名
REMOTE_IMAGE_NAME="xxgl/linkvpnclient"

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  echo "用法: $0 <版本号>"
  echo "示例: $0 1.0"
  echo "效果:"
  echo "  构建本地镜像: ${LOCAL_IMAGE_NAME}:1.0"
  echo "  打 tag 并推送: ${REMOTE_IMAGE_NAME}:1.0"
  echo "  保存镜像 tar 包: ${LOCAL_IMAGE_NAME}-1.0.tar"
  exit 1
fi

echo "[linkvpnclient] 构建本地镜像: ${LOCAL_IMAGE_NAME}:${VERSION}"
docker build -t "${LOCAL_IMAGE_NAME}:${VERSION}" .

TAR_NAME="${LOCAL_IMAGE_NAME}-${VERSION}.tar"
echo "[linkvpnclient] 保存镜像到 tar 包: ${TAR_NAME}"
docker save -o "${TAR_NAME}" "${LOCAL_IMAGE_NAME}:${VERSION}"

echo "[linkvpnclient] 完成，已生成: ${TAR_NAME}"


echo "[linkvpnclient] 为 Docker Hub 打标签: ${REMOTE_IMAGE_NAME}:${VERSION}"
docker tag "${LOCAL_IMAGE_NAME}:${VERSION}" "${REMOTE_IMAGE_NAME}:${VERSION}"

echo "[linkvpnclient] 推送到 Docker Hub: ${REMOTE_IMAGE_NAME}:${VERSION}"
docker push "${REMOTE_IMAGE_NAME}:${VERSION}"

