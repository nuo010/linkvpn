# linkvpnclient — 通用 OpenVPN 客户端镜像

这个目录用于构建一个通用的 OpenVPN 客户端 Docker 镜像。

只需要挂载 .ovpn 与可选的密码文件，任何人都可以直接使用。

## 一、目录结构

```text
linkvpnclient/
├── Dockerfile       # 客户端基础镜像
├── entrypoint.sh    # 启动脚本
```

## 二、构建镜像

在项目根目录执行：

```bash
cd linkvpnclient
# 自己本地镜像名
docker build -t linkvpnclient:latest .

# 如需推送 Docker Hub:

# docker tag linkvpnclient:latest liguanglong1234/linkvpnclient:latest
# docker push liguanglong1234/linkvpnclient:latest
```

## 三、使用方式

### 1. 仅证书认证（不需要用户名密码）

1. 在面板中为某个用户下载 `.ovpn` 客户端配置。
2. 在要运行客户端的机器上：

```bash
mkdir -p ~/linkvpnclient-run
cp /path/to/downloaded.ovpn ~/linkvpnclient-run/client.ovpn

cd ~/linkvpnclient-run

docker run -d --name linkvpnclient \
  --cap-add=NET_ADMIN \
  --device /dev/net/tun:/dev/net/tun \
  -v "$(pwd)/client.ovpn:/config/client.ovpn:ro" \
  linkvpnclient:latest

# 查看日志
docker logs -f linkvpnclient
```

当日志中出现 `Initialization Sequence Completed` 时，即表示客户端已成功连接到 VPN 服务器。

### 2. 需要用户名/密码认证

1. 仍然准备好 `client.ovpn`。
2. 新建密码文件 `auth.txt`（第一行为用户名，第二行为密码）：

```bash
cd ~/linkvpnclient-run
echo "vpn_username" > auth.txt
echo "vpn_password" >> auth.txt
```

3. 启动容器：

```bash
docker run -d --name linkvpnclient \
  --cap-add=NET_ADMIN \
  --device /dev/net/tun:/dev/net/tun \
  -v "$(pwd)/client.ovpn:/config/client.ovpn:ro" \
  -v "$(pwd)/auth.txt:/config/auth.txt:ro" \
  linkvpnclient:latest
```

> 从面板下载的 `.ovpn` 已包含 `auth-user-pass`。在容器内无 TTY，OpenVPN 无法交互输入账号密码。挂载密码文件后，启动脚本会自动把配置改为使用该文件（`auth-user-pass /config/auth.txt`），无需手动编辑 .ovpn。
