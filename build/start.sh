#!/bin/sh
set -e
# 挂载卷会覆盖 /etc/openvpn，导致镜像里拷贝的 easy-rsa 丢失。首次启动时从系统目录补全（失败也不影响面板启动）。
(
  if [ ! -f /etc/openvpn/easy-rsa/easyrsa ] && [ -d /usr/share/easy-rsa ]; then
    echo ">>> 检测到 /etc/openvpn 无 easy-rsa，从系统复制..."
    mkdir -p /etc/openvpn/easy-rsa
    cp -r /usr/share/easy-rsa/. /etc/openvpn/easy-rsa/
  fi
) || true
# 预先创建 status 与 log 文件，避免面板提示「status 文件不存在」（首次部署或 OpenVPN 未启动时也可正常进入系统）
LOG_FILE="/etc/openvpn/openvpn.log"
STATUS_FILE="/etc/openvpn/openvpn-status.log"
touch "$STATUS_FILE" "$LOG_FILE" 2>/dev/null || true

# 若挂载目录下没有 server.conf，则创建默认配置（与面板默认一致：UDP 1194，网段 10.8.8.0/24）
if [ ! -f /etc/openvpn/server.conf ]; then
  echo ">>> 未检测到 server.conf，创建默认配置..."
  cat > /etc/openvpn/server.conf << 'EOF'
# OpenVPN 服务端配置（可由管理面板编辑）
# 修改后需重启 OpenVPN 服务生效
# 本系统一用户一客户端，静态 IP 通过 ccd 目录实现

port 1194
proto udp
dev tun

# PKI 使用绝对路径，避免 OpenVPN 进程 CWD 非 /etc/openvpn 时找不到文件
ca /etc/openvpn/easy-rsa/pki/ca.crt
cert /etc/openvpn/easy-rsa/pki/issued/server.crt
key /etc/openvpn/easy-rsa/pki/private/server.key
dh /etc/openvpn/easy-rsa/pki/dh.pem

# 虚拟网段 10.8.8.0/24（服务端占 10.8.8.1，客户端从 10.8.8.2 起或由 ccd 指定静态 IP）
topology subnet
server 10.8.8.0 255.255.255.0
ifconfig-pool-persist /etc/openvpn/easy-rsa/pki/ipp.txt
client-config-dir /etc/openvpn/ccd
client-to-client

keepalive 10 120
cipher AES-256-GCM
auth SHA256
user nobody
group nogroup
persist-key
persist-tun
status /etc/openvpn/openvpn-status.log 10
log-append /etc/openvpn/openvpn.log
verb 3
explicit-exit-notify 1

# 账号密码认证（基于 SQLite）：使用 /usr/local/bin/authcheck 校验 VPN 用户名与密码
script-security 3
auth-user-pass-verify /usr/local/bin/authcheck via-file
username-as-common-name
verify-client-cert require
EOF
fi

# 挂载目录为空（无 PKI）时自动初始化 CA 与服务端证书，避免首次启动报 Options error
mkdir -p /etc/openvpn/ccd
PKI_CA="/etc/openvpn/easy-rsa/pki/ca.crt"
PKI_DH="/etc/openvpn/easy-rsa/pki/dh.pem"
if [ -f /etc/openvpn/server.conf ] && [ ! -f "$PKI_CA" ]; then
  echo "$(date '+%Y-%m-%d %H:%M:%S') 挂载目录为空，正在自动初始化 PKI（CA + 服务端证书 + DH）…" >> "$LOG_FILE" 2>/dev/null || true
  if [ -f /etc/openvpn/easy-rsa/easyrsa ]; then
    if ( export EASYRSA_BATCH=1
         cd /etc/openvpn/easy-rsa
         ./easyrsa init-pki
         ./easyrsa --batch build-ca nopass
         ./easyrsa --batch gen-dh
         ./easyrsa --batch build-server-full server nopass
       ) >> "$LOG_FILE" 2>&1; then
      if [ -f "$PKI_CA" ] && [ -f "$PKI_DH" ]; then
        echo "$(date '+%Y-%m-%d %H:%M:%S') PKI 初始化完成。" >> "$LOG_FILE" 2>/dev/null || true
      else
        echo "$(date '+%Y-%m-%d %H:%M:%S') PKI 初始化未成功（ca.crt/dh.pem 未生成），请查看上方 easy-rsa 输出。" >> "$LOG_FILE" 2>/dev/null || true
      fi
    else
      echo "$(date '+%Y-%m-%d %H:%M:%S') PKI 初始化命令执行失败，请查看上方 easy-rsa 输出。" >> "$LOG_FILE" 2>/dev/null || true
    fi
  else
    echo "$(date '+%Y-%m-%d %H:%M:%S') 未找到 easy-rsa 脚本（/etc/openvpn/easy-rsa/easyrsa），跳过自动初始化。" >> "$LOG_FILE" 2>/dev/null || true
  fi
fi

# 启动 OpenVPN（仅当 PKI 已存在时，避免刷 Options error；使用绝对路径配置保证工作目录正确）
if [ -f /etc/openvpn/server.conf ]; then
  if [ ! -f "$PKI_CA" ] || [ ! -f "$PKI_DH" ]; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') OpenVPN 未启动：PKI 初始化未完成。请检查 easy-rsa 或稍后在面板「配置」中点击「初始化 PKI」。" >> "$LOG_FILE" 2>/dev/null || true
  else
    echo "$(date '+%Y-%m-%d %H:%M:%S') 正在启动 OpenVPN..." >> "$LOG_FILE" 2>/dev/null || true
    ( openvpn --config /etc/openvpn/server.conf >> "$LOG_FILE" 2>&1 ) &
  fi
fi
exec /app/server
