#!/usr/bin/env bash
set -e

# ======================
# VPN NAT / 转发配置脚本
# 功能：
# 1. 开启内核转发 net.ipv4.ip_forward=1（含持久化）
# 2. 检测 ping 10.8.8.1 是否通
# 3. 配置 iptables NAT + FORWARD 规则
# 4. 持久化 iptables 规则
#
# 用法：
#   sudo bash setup_vpn_nat.sh                # 自动检测外网网卡
#   sudo bash setup_vpn_nat.sh eth0          # 手动指定外网网卡
# ======================

VPN_NET="10.8.8.0/24"
VPN_GW="10.8.8.1"
VPN_TUN="tun0"

# 优先使用参数指定外网网卡，否则根据默认路由自动检测
WAN_IF="$1"
if [ -z "$WAN_IF" ]; then
  # 通过默认路由检测外网网卡
  if command -v ip >/dev/null 2>&1; then
    WAN_IF=$(ip route get 1.1.1.1 2>/dev/null | awk '/dev/ {for(i=1;i<=NF;i++) if($i=="dev") print $(i+1); exit}')
  fi
fi

if [ -z "$WAN_IF" ]; then
  echo "无法自动检测外网网卡，请手动指定，例如："
  echo "  sudo bash setup_vpn_nat.sh eth0"
  exit 1
fi

echo "外网网卡检测/使用为：$WAN_IF"
echo "VPN 网段：$VPN_NET，VPN 设备：$VPN_TUN，VPN 网关：$VPN_GW"

# 1. 开启 IP 转发（含持久化）
echo "开启 net.ipv4.ip_forward = 1 ..."
sysctl -w net.ipv4.ip_forward=1 >/dev/null

# 持久化到 /etc/sysctl.conf
if grep -q '^net.ipv4.ip_forward' /etc/sysctl.conf 2>/dev/null; then
  sed -i 's/^net.ipv4.ip_forward.*/net.ipv4.ip_forward = 1/' /etc/sysctl.conf
else
  echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.conf
fi

# 2. 检测 ping 10.8.8.1
echo "检测是否能 ping 到 VPN 网关 $VPN_GW ..."
if ping -c 2 -W 1 "$VPN_GW" >/dev/null 2>&1; then
  echo "ping $VPN_GW 正常。"
else
  echo "警告：暂时无法 ping 通 $VPN_GW（可能是 OpenVPN 未启动或路由未就绪），继续配置 iptables。"
fi

# 3. 配置 iptables 规则
echo "配置 iptables NAT / FORWARD 规则 ..."

# NAT：VPN 网段经过外网网卡做源地址伪装
iptables -t nat -C POSTROUTING -s "$VPN_NET" -o "$WAN_IF" -j MASQUERADE 2>/dev/null || \
iptables -t nat -A POSTROUTING -s "$VPN_NET" -o "$WAN_IF" -j MASQUERADE

# 允许 VPN 流量从 tun0 转发到外网
iptables -C FORWARD -i "$VPN_TUN" -o "$WAN_IF" -j ACCEPT 2>/dev/null || \
iptables -A FORWARD -i "$VPN_TUN" -o "$WAN_IF" -j ACCEPT

# 允许外网返回流量回到 tun0（状态跟踪）
iptables -C FORWARD -i "$WAN_IF" -o "$VPN_TUN" -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || \
iptables -A FORWARD -i "$WAN_IF" -o "$VPN_TUN" -m state --state RELATED,ESTABLISHED -j ACCEPT

echo "iptables 规则已配置："
iptables -t nat -L POSTROUTING -n -v | grep "$VPN_NET" || true
iptables -L FORWARD -n -v | grep -E "$VPN_TUN|$WAN_IF" || true

# 4. 持久化 iptables 规则
echo "尝试持久化 iptables 规则 ..."

if command -v netfilter-persistent >/dev/null 2>&1; then
  netfilter-persistent save
  echo "已通过 netfilter-persistent 保存规则。"
elif command -v iptables-save >/dev/null 2>&1; then
  # Debian/Ubuntu 常用：/etc/iptables/rules.v4
  mkdir -p /etc/iptables
  iptables-save > /etc/iptables/rules.v4
  echo "已将规则保存到 /etc/iptables/rules.v4，重启时可通过自定义服务加载。"
else
  echo "未找到 netfilter-persistent 或 iptables-save，请手动持久化当前规则。"
fi

echo "全部完成。"