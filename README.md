# OpenVPN 管理系统（linkvpn）


![login](/docs/images/login.png)

基于 Web 的 OpenVPN 服务端管理系统：通过浏览器完成 PKI 初始化、用户与证书管理、服务端参数配置、连接记录与日志查看；支持账号密码认证（SQLite + bcrypt），支持 Docker 一键部署与可选客户端镜像。

---

## 一、为什么做这个项目

- **统一管理**：在服务器上改 OpenVPN 配置、建用户、发证书通常要 SSH + 手敲命令，本面板用 Web 界面完成 PKI、用户、CCD、服务端参数和日志查看，减少人工操作和出错。
- **开箱即用**：Docker 部署时，挂载空目录即可自动初始化 PKI 和默认 `server.conf`，避免「首次启动缺证书」等报错。
- **账号密码认证**：除客户端证书外，支持用户名/密码认证（`auth-user-pass-verify`），密码存 SQLite 并做 bcrypt 哈希，便于在面板里增删改用户。
- **客户端友好**：下载的 `.ovpn` 已包含 `auth-user-pass`，桌面客户端会弹窗输入账号密码；另提供通用 OpenVPN 客户端 Docker 镜像（`linkvpnclient`），挂载 `.ovpn` 与可选密码文件即可使用。

适合自建 VPN、内网穿透、远程办公等场景，单机部署即可使用。

---

## 二、开发框架

| 层级     | 技术栈 | 说明 |
|----------|--------|------|
| 后端     | Go 1.21+ | 主语言 |
| Web 框架 | Gin | 路由、中间件、静态文件 |
| 数据库   | SQLite (GORM) | 用户、配置、连接记录、OpenVPN 参数 |
| 认证     | JWT + bcrypt | 面板登录 JWT；VPN 用户密码 bcrypt |
| 前端     | Vue 3 + Vite | Composition API、单页应用 |
| UI       | Element Plus | 表格、表单、日期、消息提示等 |
| 状态/路由 | Pinia + Vue Router | 登录态、前端路由 |
| 请求     | Axios | 调用后端 API |
| 打包     | GoReleaser（可选） | 多平台二进制（含 server、authcheck） |

**后端目录结构概览：**

- `cmd/server`：HTTP 服务入口；`cmd/authcheck`：OpenVPN 账号密码校验可执行文件（via-file 调用）。
- `internal/config`：配置加载（环境变量）。
- `internal/router`：Gin 路由与鉴权。
- `internal/handler`：接口实现（用户、配置、日志、VPN 初始化与状态等）。
- `internal/model`：GORM 模型（用户、服务配置、连接记录、OpenVPN 参数等）。
- `internal/vpn`：OpenVPN 相关逻辑（PKI/easy-rsa、server.conf 生成、客户端 .ovpn 生成）。

**前端目录结构概览：**

- `web/src/views`：页面（首页、用户列表、系统配置、参数配置、在线统计、连接记录、服务日志、连接日志等）。
- `web/src/router`：前端路由与鉴权。
- `web/src/stores`：Pinia 状态（如登录 token）。
- `web/src/api`：Axios 封装与 API 调用。

---

## 三、功能说明

- **首页**：PKI 是否初始化、在线用户数、最近登录、服务状态等概览。
- **用户管理**：增删改 VPN 用户；设置类型（用户/客户端）、昵称、到期时间、启用状态；为用户生成证书并下载 `.ovpn`（内含 `auth-user-pass`）；CCD（client-config-dir）：静态 IP、iroute、push route、redirect-gateway、route-nopull 等。
- **系统配置**：客户端下载所用「服务器地址」与「端口」；首次进入可强制弹窗填写。
- **OpenVPN 参数配置**：端口、协议、网段、密码学参数等，界面化编辑后写入 `server.conf` 并支持重启 OpenVPN。
- **在线统计**：当前在线用户列表（来自 OpenVPN status）。
- **连接记录**：按日期查看 VPN 连接记录（成功/未成功及原因说明）。
- **服务日志 / 连接日志**：查看 OpenVPN 主日志与 status 日志。
- **账号密码认证**：服务端使用 `auth-user-pass-verify` 调用 `authcheck`，对数据库中的用户做校验（存在、启用、未过期、密码 bcrypt 匹配）。


---

## 四、使用说明

### 4.1 环境要求

- 若 Docker 部署：Docker、Docker Compose；宿主机需开放 UDP 1194（OpenVPN）与 HTTP 8789（面板）。
- 若本地运行：Go 1.21+、Node 18+（前端开发）、OpenVPN、easy-rsa（或由 Docker 提供）。

### 4.2 Docker 部署（推荐）

1. **使用预构建镜像（若已有）：**

   ```bash
   # 项目目录
   cd linkvpn
   
   docker run -d --name linkvpn \
        --network host \
        -e ADMIN_PASS="${ADMIN_PASS:-admin}" \
        -e JWT_SECRET="${JWT_SECRET:-admin}" \
        -v "$(pwd)/openvpn:/etc/openvpn" \
        --device /dev/net/tun \
        --cap-add=NET_ADMIN \
        --restart unless-stopped \
        liguanglong1234/linkvpn:1.0
   ```

2. **首次启动**：  
   将数据卷挂载到 `/etc/openvpn`。若挂载目录为空，容器内 `start.sh` 会自动初始化 easy-rsa、生成 CA 与服务端证书、创建默认 `server.conf`（含账号密码认证配置）。

3. **访问面板**：  
   `http://<服务器IP>:8789`。默认管理员账号/密码见环境变量 `ADMIN_USER` / `ADMIN_PASS`（如 `admin` / `admin`），生产环境请修改并设置 `JWT_SECRET`。

4. **首次使用**：  
   进入系统后按提示完成「客户端下载服务器地址与端口」配置。

### 4.3 OpenVPN 客户端

- **桌面/手机**：使用系统下载的 `.ovpn`，连接时按提示输入 VPN 用户名与密码。
- **Docker 客户端**：使用本项目提供的 `linkvpnclient` 镜像，挂载 `.ovpn` 与 `auth.txt`（用户名/密码文件）, `auth.txt` 内容格式为第一行账号第二行密码。

```shell
    # 项目目录
    cd linkvpnclient
    docker run -d --name linkvpnclient \
          --network host \
          --restart=always \
          --cap-add=NET_ADMIN \
          --device /dev/net/tun:/dev/net/tun \
          -v "$(pwd)/xxxx.ovpn:/config/client.ovpn" \
          -v "$(pwd)/auth.txt:/config/auth.txt" \
          liguanglong1234/linkvpnclient:1.1-amd
```

## 五、配置与端口

| 环境变量        | 含义           | 默认值 |
|-----------------|----------------|--------|
| `VPN_BASE_PATH` | OpenVPN 工作目录 | `/etc/openvpn` |
| `HTTP_PORT`     | 面板 HTTP 端口 | `8789` |
| `STATIC_DIR`    | 前端静态目录   | `/app/web` |
| `ADMIN_USER` / `ADMIN_PASS` | 面板管理员账号密码 | `admin` / `admin` |
| `JWT_SECRET`    | JWT 签名密钥   | 建议生产环境设置 |
| `DATABASE_PATH` | SQLite 库路径  | `$VPN_BASE_PATH/data/panel.db` |

OpenVPN 默认：UDP 1194；可通过「OpenVPN 参数配置」或直接改 `server.conf` 修改。

---

## 六、系统截图
![login](/docs/images/login.png)
![首页](/docs/images/home.png)
![user](/docs/images/user.png)
![log](/docs/images/log.png)
![log2](/docs/images/log2.png)
![log3](/docs/images/log3.png)
![log4](/docs/images/log4.png)
## 六、许可证与免责

本项目仅供学习与自建使用。使用 OpenVPN 时请遵守当地法律法规与网络政策。生产环境请务必修改默认管理员密码与 `JWT_SECRET`，并做好备份与权限控制。

