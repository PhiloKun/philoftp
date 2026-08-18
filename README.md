# PhiloFTP — 内网 FTP 服务器（带 Web 管理端）

一个用 Go 编写的高性能内网 FTP 服务器，内置网页管理端，支持**多用户 + 密码 + 目录权限控制**，单文件二进制，跨平台运行（macOS / Windows / Linux）。

## 特性

- 🚀 **高性能**：Go 编写，并发处理多客户端，资源占用低
- 👥 **多用户 + 密码**：每个用户独立账号密码，存储于 SQLite 数据库（`configs/users.db`），密码以 bcrypt 加盐哈希落盘
- 🔒 **权限控制（RBAC 分级）**：两级角色 `admin`（全部权限）与 `user`（仅文件操作），各自独立根目录（chroot 隔离，互不可见）
- 🌐 **Web 管理端**：浏览器即可管理用户、浏览/上传/下载文件，无需命令行
- 📦 **单文件分发**：编译为单个可执行文件，双击即可运行，无需安装运行时
- 🔌 **被动模式**：内置 PASV 端口范围，适配内网/NAT 环境
- 🔐 **可选 FTPS**：配置 TLS 证书后可启用加密传输

## 快速开始

### 方式一：直接运行（已编译好二进制）

1. 下载对应平台的 `philoftp` 可执行文件（见 `dist/macos`、`dist/windows`、`dist/linux` 目录）
2. 双击运行，或终端执行：
   ```bash
   ./philoftp
   ```
3. 打开浏览器访问 `http://<本机IP>:8080` 进入管理端

### 方式二：从源码运行（需 Go 1.26+）

```bash
git clone <repo> && cd philoftp
make build        # 编译到 dist/macos/philoftp
# 或： go build -o dist/macos/philoftp .
./dist/macos/philoftp
```

## 默认账户

| 用户 | 密码 | 权限 | 根目录 |
|------|------|------|--------|
| `admin` | `admin123` | 管理员（全部权限） | `data/admin` |

> ⚠️ 首次部署请务必修改默认密码，或直接删除默认用户在 Web 端新建。

## 命令行参数

```bash
./philoftp \
  -config config.json \   # 配置文件路径
  -db     users.db \      # SQLite 用户数据库路径
  -ftp-port 2121 \        # FTP 控制端口（覆盖配置）
  -web-port 8080 \        # Web 管理端口（覆盖配置）
  -data   ./data          # 数据根目录（覆盖配置）
```

## Web 管理端功能

后台采用**左侧菜单栏 + 右侧视图**的单页（SPA）布局，响应式适配桌面与移动端，菜单项均可点击跳转：

- **概览**：状态卡片一览（FTP/Web 端口、用户数、PASV 范围、FTPS 状态、运行时长）
- **登录鉴权**：Web 端基于会话（HTTP-Only Cookie）登录，未登录自动跳转登录页；所有管理与数据接口均受保护
- **自助注册**：默认开放注册（可在 `config.json` 关闭 `allow_register`），注册即创建独立文件空间的可写用户；表单含用户名/密码格式校验与实时密码强度提示
- **用户管理**：新增（弹层表单，含实时密码强度提示）/ 删除用户，设置主目录、角色（管理员/普通用户）、启用状态
- **文件管理**：按用户浏览文件树、上传、批量上传、新建目录、下载；下载带实时进度条与速度显示
- **基础设置**：在 Web 端查看并修改配置（FTP/Web 端口、PASV 范围、数据目录、FTPS 开关与证书、自助注册开关），保存即写入 `config.json` 并**即时生效**：FTP 服务（端口/PASV/FTPS）自动热重载、数据目录与注册开关实时读取；仅 Web 管理端口需重启服务进程后切换监听
- **系统配置**：运行时信息（Go 版本、协程数、运行时长、数据/配置路径等）与关于

> 界面采用「深空控制台风」设计（玻璃拟态 + 青色辉光 + 等宽技术字体 + 入场动画），样式与脚本位于 `web/assets/`。Google Fonts 异步加载并配系统字体强回退，无网时自动降级为系统字体，保证内网/离线环境仍可正常使用。前端登录/注册输入框提供聚焦提示、格式示例与动态校验反馈。

## 配置文件 `config.json`

```json
{
  "ftp_port": 2121,
  "pasv_min_port": 21100,
  "pasv_max_port": 21110,
  "web_port": 8080,
  "data_dir": "data",
  "enable_ftps": false,
  "allow_register": true,
  "tls_cert": "",
  "tls_key": ""
}
```

### 防火墙 / 端口放行

内网使用需放行：
- FTP 控制端口（默认 `2121`）
- 被动端口范围（默认 `21100-21110`）

例如在 Linux 上：
```bash
sudo ufw allow 2121/tcp
sudo ufw allow 21100:21110/tcp
sudo ufw allow 8080/tcp
```

## 用户与权限（SQLite + 角色分级）

用户数据存储在 SQLite 数据库 `configs/users.db`（纯 Go 驱动，无 cgo，单二进制可分发），密码以 **bcrypt 加盐哈希**存储，绝不明文落盘。首次启动会自动创建默认管理员 `admin / admin123`；若同目录存在旧版 `configs/users.json`，将自动迁移为哈希记录（明文密码 → bcrypt）。

系统采用**两级角色**的 RBAC 权限模型，严格分级：

| 角色 | 权限范围 |
| --- | --- |
| `admin`（管理员） | 全部权限：用户管理、系统配置、文件上传/下载/删除/建目录/浏览、系统信息与状态查看 |
| `user`（普通用户） | 仅文件操作：上传、下载、删除、新建目录、浏览自身目录；**无法访问**用户管理、基础设置、系统配置等管理接口（调用将返回 403） |

- 管理类接口（`/api/users`、`/api/config`、`/api/system`）服务端通过 `RequireRole("admin")` 中间件强制鉴权，前端亦对非管理员隐藏对应入口。
- 至少保留一个管理员账户：将最后一个管理员降级或删除会被拒绝（"至少需保留一个管理员账户"）。
- `enabled`：`false` 时该用户无法登录，且禁用账户不可执行任何写操作。

> 安全提示：部署后请尽快在「用户管理」中修改 `admin` 默认密码。

## 各平台客户端连接示例

```bash
# 命令行（lftp / ftp）
ftp <服务器IP> 2121
# 输入用户名密码

# 文件管理器（Windows 资源管理器 / macOS Finder）
ftp://<服务器IP>:2121
```

## 启用 FTPS（加密）

1. 准备证书 `server.crt` 与私钥 `server.key`
2. 修改 `config.json`：
   ```json
   { "enable_ftps": true, "tls_cert": "server.crt", "tls_key": "server.key" }
   ```
3. 在 Web 端「基础设置」开启 FTPS 并保存即可热重载生效（或在 `config.json` 修改后重启服务），客户端使用 `FTPS（显式）` 模式连接

## 构建各平台二进制

```bash
# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o dist/macos/philoftp-darwin-arm64 .

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o dist/macos/philoftp-darwin-amd64 .

# Windows
GOOS=windows GOARCH=amd64 go build -o dist/windows/philoftp-windows-amd64.exe .

# Linux
GOOS=linux GOARCH=amd64 go build -o dist/linux/philoftp-linux-amd64 .
```

## 项目结构

```
philoftp/
├── main.go            # 程序入口、命令行参数解析、组装各层并启动
├── webstatic.go       # //go:embed web —— 将前端静态资源嵌入单二进制
├── go.mod / go.sum
├── model/             # 领域模型
│   └── user.go          # User 模型 + ResolveHome 路径辅助
├── config/            # 配置层
│   └── config.go        # Config 结构体、加载/持久化、DataDirOf、ToAPI
├── repository/        # 数据访问层（持久化）
│   └── dbstore.go       # DBStore：基于 SQLite 的用户存储（bcrypt 哈希 + 迁移 + 最后管理员保护）
├── service/           # 业务/服务层
│   └── ftpserver.go     # FTP 服务器启动与文件系统驱动（goftp/server）
├── handler/           # 接入层（HTTP API + 静态资源托管，不含界面代码）
│   ├── web.go           # Web 管理端 API（Gin）
│   └── web_ui.go        # 前端静态资源托管（页面路由 + /assets 挂载）
├── web/               # 【前端】纯静态资源，与后端 Go 代码完全分离
│   ├── index.html       # 登录页
│   ├── register.html    # 注册页
│   ├── app.html         # 控制台单页（SPA）
│   └── assets/
│       ├── style.css      # 共享样式（深空控制台风：玻璃拟态 + 青色辉光）
│       ├── auth.js        # 登录 / 注册逻辑
│       └── app.js         # 控制台交互逻辑
├── configs/           # 默认配置模板（运行时从此处读取）
│   ├── config.json       # 服务器配置（端口、PASV 范围、FTPS 等）
│   └── users.db          # SQLite 用户数据库（默认 admin/admin123，密码 bcrypt 哈希）
├── data/              # 文件存储根目录（运行时生成，含各用户 home）
├── dist/              # 各平台构建产物
│   ├── macos/           # macOS 二进制 / zip / PhiloFTP.app
│   ├── windows/         # Windows exe / zip
│   └── linux/           # Linux 二进制 / zip
├── Makefile           # 构建 / 运行 / 打包脚本
└── .gitignore
```

> **前后端职责边界**
> - **后端（Go）**：专注业务逻辑与接口。`handler/` 仅提供 RESTful API（`/api/*`）与静态资源托管，**不内联任何 HTML/JS/CSS**；前端文件全部位于 `web/`，经 `//go:embed` 嵌入后由 `handler/web_ui.go` 托管，保持单二进制分发。
> - **前端（静态资源）**：专注界面与交互，纯 HTML/CSS/JS，不依赖后端编译，可独立开发/调试（如本地用任意静态服务器打开 `web/`）。
>
> 源码按职责拆分为内部包（model / config / repository / service / handler），
> 由根目录 `main.go`（package main）负责组装与启动。`go build .` 即可编译单二进制。

### 常用命令（Makefile）

```bash
make build      # 编译到 dist/macos/philoftp
make run        # 直接运行（读取 configs/config.json）
make dist       # 构建 macOS / Windows / Linux 全部平台产物
make clean      # 清理本地编译产物
```
