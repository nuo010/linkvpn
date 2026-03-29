#!/usr/bin/env bash
set -e

OVPN_FILE="${OVPN_FILE:-/config/client.ovpn}"
AUTH_FILE="${AUTH_FILE:-/config/auth.txt}"

echo "[linkvpnclient] OpenVPN 版本:"
openvpn --version | head -n 1 || echo "[linkvpnclient] 无法获取 OpenVPN 版本"
echo "[linkvpnclient] 当前时间: $(date '+%Y-%m-%d %H:%M:%S %Z %z')"

if [ ! -f "$OVPN_FILE" ]; then
  echo "[linkvpnclient] ERROR: 未找到配置文件: $OVPN_FILE" >&2
  echo "请通过挂载方式提供 .ovpn，例如:" >&2
  echo "  -v /path/to/your.ovpn:/config/client.ovpn:ro" >&2
  exit 1
fi

CONFIG_TO_USE="$OVPN_FILE"

# 容器内无 TTY，OpenVPN 无法交互式询问用户名/密码。若提供了密码文件，则生成临时配置，
# 将 auth-user-pass 改为 auth-user-pass /path/to/auth.txt，避免 "can't ask for Enter Auth Username" 报错。
if [ -f "$AUTH_FILE" ]; then
  echo "[linkvpnclient] 使用密码文件: $AUTH_FILE"
  TMP_OVPN="/tmp/client-$$.ovpn"
  if grep -q '^auth-user-pass' "$OVPN_FILE"; then
    sed "s|^auth-user-pass.*|auth-user-pass $AUTH_FILE|" "$OVPN_FILE" > "$TMP_OVPN"
  else
    ( cat "$OVPN_FILE"; echo ""; echo "auth-user-pass $AUTH_FILE" ) > "$TMP_OVPN"
  fi
  CONFIG_TO_USE="$TMP_OVPN"
fi

echo "[linkvpnclient] 使用配置文件: $CONFIG_TO_USE"
exec openvpn --config "$CONFIG_TO_USE"
