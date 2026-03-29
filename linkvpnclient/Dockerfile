FROM alpine:3.19

# 安装 openvpn 和基础工具
RUN apk add --no-cache \
    openvpn \
    ca-certificates \
    bash \
    && update-ca-certificates

# 设置容器时区为 Asia/Shanghai，保证 OpenVPN 与面板日志时间与国内时间一致
ENV TZ=Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

# 默认配置 / 密码文件路径（可被环境变量覆盖）
ENV OVPN_FILE=/config/client.ovpn \
    AUTH_FILE=/config/auth.txt

WORKDIR /config

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
