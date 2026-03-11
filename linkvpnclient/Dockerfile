FROM alpine:3.19

# 安装 openvpn 和基础工具
RUN apk add --no-cache openvpn ca-certificates bash

# 默认配置 / 密码文件路径（可被环境变量覆盖）
ENV OVPN_FILE=/config/client.ovpn \
    AUTH_FILE=/config/auth.txt

WORKDIR /config

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

CMD ["/entrypoint.sh"]
