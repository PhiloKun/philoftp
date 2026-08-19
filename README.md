# PhiloFTP — 内网 FTP 服务器（带 Web 管理端）

一个用 Go 编写的高性能内网 FTP 服务器，内置网页管理端，支持**多用户 + 密码 + 目录权限控制**，单文件二进制，跨平台运行（macOS / Windows / Linux）。

## 特性

- 🚀 **高性能**：Go 编写，并发处理多客户端，资源占用低
- 👥 **多用户 + 密码**：每个用户独立账号密码，存储于 SQLite 数据库（`configs/users.db`），密码以 bcrypt 加盐哈希落盘
- 🔒 **权限控制（RBAC 分级）**：两级角色 `admin`（全部权限）与 `user`（仅文件操作），各自独立根目录（chroot 隔离，互不可见）
- 🌐 **Web 管理端**：浏览器即可管理用户、浏览/上传/下载文件，无需命令行
- 📦 **单文件分发**：编译为单个可执行文件，双击即可运行，无需安装运行时
- 🔒 **单实例运行**：Windows 使用「文件锁 + 命名互斥量」双保险，其他平台使用 `flock`，确保同一时刻仅一个进程；重复启动时弹出中文提示而**不会**再生成一个实例
- 🔌 **被动模式**：内置 PASV 端口范围，适配内网/NAT 环境
- 🖥 **Windows 系统托盘**（GUI）：常驻后台，托盘菜单支持查看状态/启动/停止/打开 Web/打开日志/退出；启动成功后弹出**确认提示框**，明确告知程序已运行及 Web 管理地址（GUI 模式无控制台，避免"启动无反馈"的困惑）
- 📋 **日志文件**：运行日志实时写入 `~/.philoftp/logs/philoftp.log`，便于排查与追溯
- 📡 **局域网访问**：自动注册 mDNS `philoftp.local`，登录页展示本机 IP / mDNS 主机名 / 端口与**访问地址二维码**，其他设备扫码即可访问，无需记 IP
- 🗂 **Web 文件管理增强**：支持**重命名 / 批量移动（剪切）**、**递归文件名搜索**（结果可一键定位到所在目录）、**ZIP 打包下载**（多文件/目录服务端打包）与**在线解压**（上传或已有的 .zip 一键解压，同名自动重命名，防 zip 穿越攻击），全部复用越权防护，仅可写用户可操作
- 📱 **移动端适配**：导航栏在窄屏自动转为顶部横向滚动；局域网访问卡片/二维码在窄屏居中堆叠；表单、按钮、二维码等触控目标适配 ≥44px，避免横向滚动与内容溢出

## 快速开始

### 方式一：直接运行（已编译好二进制）

1. 下载对应平台的 `philoftp` 可执行文件（见 `dist/macos`、`dist/windows`、`dist/linux` 目录）
2. 双击运行，或终端执行：
   ```bash
   ./philoftp
   ```
3. 打开浏览器访问 `http://<本机IP>:8080` 进入管理端

### 方式二（Windows）：安装包（推荐）

Windows 用户可直接使用 `dist/windows/PhiloFTP-Setup.exe` 一键安装：

- **标准 Windows 安装流程**：选择安装目录 → 创建桌面/开始菜单快捷方式 → 完成
- 默认安装到**当前用户目录** `%LocalAppData%\PhiloFTP`（无需管理员权限，避免 Program Files 写权限导致启动失败）
- 安装时自动**放行 Windows 防火墙**（FTP 控制端口 2121 + Web 管理端口 8080）
- 安装完成后勾选「立即启动」，程序以**系统托盘图标**方式常驻后台，**不会闪退/关控制台**
- **托盘右键菜单**：查看运行状态 / 打开 Web 管理页 / 启动服务 / 停止服务 / 打开日志目录 / 退出
- 通过「开始菜单 → 卸载 PhiloFTP」或「设置 → 应用」可完整卸载
- 数据与日志默认在 `%USERPROFILE%\.philoftp\`（配置/数据库、data、logs 日志文件）

> 兼容 Windows 10 / 11（64 位）。

### 方式三：从源码运行（需 Go 1.26+）

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

## 命令行参数 / 环境变量

```bash
./philoftp \
  -config config.json \   # 配置文件路径
  -db     users.db \      # SQLite 用户数据库路径
  -ftp-port 2121 \        # FTP 控制端口（覆盖配置）
  -web-port 8080 \        # Web 管理端口（覆盖配置）
  -data   ./data          # 数据根目录（覆盖配置）
  -headless               # Windows 下强制控制台模式（无托盘）
```

端口/路径同样可通过**环境变量**配置（命令行参数优先级更高）：

| 环境变量 | 作用 |
|---|---|
| `PHILOFTP_WEB_PORT` | Web 管理端口（如 `9090`） |
| `PHILOFTP_FTP_PORT` | FTP 控制端口（如 `2121`） |
| `PHILOFTP_DATA_DIR` | 数据根目录 |
| `PHILOFTP_CONFIG` | 配置文件路径 |
| `PHILOFTP_DB` | 用户数据库路径 |

> 端口优先级：命令行 `-web-port`/`-ftp-port` > 环境变量 `PHILOFTP_*` > 配置文件 `config.json`。Web 管理端口修改后**需重启进程**生效（FTP 端口经热重载立即生效）。

## Web 管理端功能

后台采用**左侧菜单栏 + 右侧视图**的单页（SPA）布局，响应式适配桌面与移动端，菜单项均可点击跳转：

- **概览**：仪表盘式概览页（深空控制台玻璃拟态风格），一次加载完整统计快照：`/api/overview` 提供 4+3 块 KPI 卡片（用户总数/活跃会话/文件目录数/存储使用率 + 管理员数/数据目录/Go 运行时），三列均衡布局：左列用户分布条形图（管理员/普通/启用/禁用/活跃会话 5 项，列高时自动均匀铺开）、中列服务器状态（含数据目录完整路径，等宽字体换行显示）+ 活跃会话列表 + 最近文件更新时间、右列存储使用率**环形图**（按用量自适应红/黄/青三色）+ 文件类型分布（按大小降序，隐藏文件归"其它"）+ Top 5 大文件；**卡片完全可自定义**：所有卡片均支持拖拽摆放（自由调整 KPI 行内顺序、在三列间移动、跨列重排），位置自动保存到 localStorage 并在下次访问时恢复；顶部「⚙ 卡片设置」统一管理每张卡的**可拖动开关**（关闭后该卡固定不可拖）与**显示开关**（隐藏/显示该卡），自定义卡片可删除；顶部「＋ 添加卡片」可创建**自定义备注卡片**（标题 + 多行文本，纯前端 localStorage 持久化），放入指定列；拖动时启用**智能布局保护**（三列不出现空列，空列自动从相邻列移入卡片）；「重置布局」按钮一键恢复默认排布；KPI 卡片可直接跳转对应视图
- **登录鉴权**：Web 端基于会话（HTTP-Only Cookie）登录，未登录自动跳转登录页；所有管理与数据接口均受保护
- **自助注册**：默认开放注册（可在 `config.json` 关闭 `allow_register`），注册即创建独立文件空间的可写用户；表单含用户名/密码格式校验与实时密码强度提示
- **用户管理**：新增（弹层表单，含实时密码强度提示）/ 删除用户，设置主目录、角色（管理员/普通用户）、启用状态；删除需经**自定义确认弹窗**二次确认，提示不可恢复
- **文件管理**：按用户浏览文件树、上传、批量上传、新建目录、下载；目录内支持**面包屑导航 + 返回上一级 + 双击目录行进入**，文件行支持**双击预览**，新建目录后停留当前目录、可打开或返回；新建目录经**自定义输入弹窗**（青色图标 + 居中输入框 + 实时名称校验，禁用 `/\:*?"<>|` 等非法字符）；常见文件（图片、音视频、文本/代码、PDF）支持**在线预览**，无需下载即可查看（文本/代码与 PDF 通过浏览器内嵌渲染，`text/html` 强制下载以防 XSS）；下载带实时进度条与速度显示；**批量删除 + 回收站**：文件列表带**勾选列**，工具栏支持**全选 / 反选 / 取消选择**，批量操作栏显示已选数并提供「删除选中」按钮，删除前经**自定义确认弹窗**二次确认（含选中项数量与名称预览）；删除操作（单删/批量）统一**移入回收站**而非物理删除，回收站视图支持**恢复**与**清空**；单删成功 toast 带**「撤销」按钮**一键恢复；后端 `DELETE /api/files` 与 `POST /api/files/batch-delete` 均经**权限校验**（仅可写用户），防护越权与主目录根路径误删，`GET /api/trash` / `POST /api/trash/restore` / `DELETE /api/trash` 管理回收站；**上传同名冲突处理**：上传前自动检测目标目录是否已有同名文件，若有则弹出**冲突确认弹窗**（黄色警示列出冲突文件），提供「覆盖原文件 / 自动重命名 / 取消上传」三种处理方式——覆盖直接替换、自动重命名在文件名后追加时间戳序号（如 `file_123456.txt`）、取消则不写入；后端 `POST /api/upload` 支持 `mode` 参数（`rename`/`overwrite`/`cancel`），`cancel` 模式遇冲突返回 409 + 冲突清单
- **重命名 / 批量移动**：文件行「重命名」按钮（同级改名，禁用路径分隔符）；批量勾选后工具栏「移动到…」可将多个文件/目录剪切到目标目录（输入相对根路径如 `/docs`，同名自动重命名 `name_1`）；后端 `POST /api/rename`、`POST /api/move` 均经越权校验与 `safeJoin` 防护
- **递归搜索**：工具栏搜索框输入关键词（文件名子串、大小写不敏感），结果以模态列出「名称 / 所在目录 / 大小」，点击「定位」跳转回所在目录并高亮；当前目录递归遍历，结果上限 500 条（超出提示收窄关键词）以防卡顿；后端 `GET /api/search?q=&path=`
- **ZIP 打包下载**：批量勾选后「打包下载」或单文件/目录直接下载，多文件服务端打包为 `.zip`（以公共目录为基准，条目不含 `../` 前缀）；后端 `GET /api/download/zip?paths=`
- **在线解压**：`.zip` 文件行「解压」按钮，将压缩包解压到同名目录（默认）或指定目录，支持 `rename`（默认，同名自动重命名）/ `overwrite` / `cancel` 三种冲突策略；内置 zip 穿越（zip slip）防护，仅可写用户可操作；后端 `POST /api/unzip`
- **系统设置**：由原「基础设置 + 系统配置」合并而来，分「服务端口 / 存储与访问」两个分组统一管理：FTP/Web 端口、PASV 范围、数据目录、自助注册开关；保存即写入 `config.json` 并**即时生效**（FTP 服务端口/PASV 自动热重载、数据目录与注册开关实时读取），仅 Web 管理端口需重启进程；PASV 范围以 `min-max` 形式输入，前端自动组合/拆分为端口字段

> 界面采用「深空控制台风」设计（玻璃拟态 + 青色辉光 + 等宽技术字体 + 入场动画），样式与脚本位于 `web/assets/`。Google Fonts 异步加载并配系统字体强回退，无网时自动降级为系统字体，保证内网/离线环境仍可正常使用。前端登录/注册输入框提供聚焦提示、格式示例与动态校验反馈；删除类危险操作使用**自定义玻璃拟态确认弹窗**（替换浏览器原生 `confirm()`），带脉冲图标、红色警示与快捷键（Esc 取消 / Enter 确认），交互与整体风格完全统一；新建目录等需要输入的交互使用**自定义玻璃拟态输入弹窗**（替换浏览器原生 `prompt()`），青色脉冲图标、居中等宽输入框、实时名称校验（非法字符 / `.`/`..` / 长度限制）、Esc 取消 / Enter 提交。

## 配置文件 `config.json`

```json
{
  "ftp_port": 2121,
  "pasv_min_port": 21100,
  "pasv_max_port": 21110,
  "web_port": 8080,
  "data_dir": "data",
  "allow_register": true
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
| `admin`（管理员） | 全部权限：用户管理、系统设置、文件上传/下载/删除/建目录/浏览 |
| `user`（普通用户） | 仅文件操作：上传、下载、删除、新建目录、浏览自身目录；**无法访问**用户管理、系统设置等管理接口（调用将返回 403） |

- 管理类接口（`/api/users`、`/api/config`）服务端通过 `RequireRole("admin")` 中间件强制鉴权，前端亦对非管理员隐藏对应入口。
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

## 构建各平台二进制

```bash
# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o dist/macos/philoftp-darwin-arm64 .

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o dist/macos/philoftp-darwin-amd64 .

# Windows（系统托盘版，需 CGO + mingw-w64）
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
  go build -ldflags "-H=windowsgui" -o dist/windows/philoftp-windows-amd64.exe .

# Linux
GOOS=linux GOARCH=amd64 go build -o dist/linux/philoftp-linux-amd64 .
```

### 版本号管理（单一真源）

版本号统一写在仓库根目录 **`version.txt`**（如 `1.0.1`）。构建时由 Makefile 读取并：

- 注入二进制（`-ldflags -X main.version` / `-X github.com/philoftp/handler.version`），体现在**启动日志**与**关于页**；
- 拼接到**所有产物文件名**（如 `philoftp-1.0.1-windows-setup.exe`、`philoftp-1.0.1-macos.zip`）；
- Windows 安装包内部版本信息（`VIProductVersion` / `ProductVersion`）同步为对应 `x.y.z.0`。

发布新版本流程（推荐一键脚本）：

```bash
# 1) 确保已安装构建依赖
brew install nsis mingw-w64 && pip3 install Pillow

# 2) 运行 release.sh：自动递增版本号、构建全平台产物、打 tag、推双远端、发布 GitHub + Gitee Release
./release.sh            # 默认 patch 递增（1.0.1 -> 1.0.2）
./release.sh minor      # 次版本递增
./release.sh 1.2.3      # 直接指定版本号
./release.sh -n 1.2.3   # 仅改 version.txt，不构建/发布（干跑）
```

### 发布说明（Changelog）约定

每个正式版本须在仓库根目录 **`CHANGELOG.md`** 中维护结构化发布说明，`release.sh` 发布时会**自动读取当前版本条目**作为 GitHub / Gitee Release 的说明文本。

- **维护时机**：每次代码变更后，在 `CHANGELOG.md` 顶部的「未发布（Unreleased）」小节记录变更；发布时把内容迁移到对应 `## [X.Y.Z] - 日期` 条目并标注发布日期。
- **格式要求**：按类别分组（`🆕 新功能` / `✨ 改进` / `🐛 Bug 修复`），每个版本标注**版本号 + 发布日期**。
- **Bug 修复条目**须包含：`描述`（问题现象）、`影响范围`（哪些功能/场景）、`修复方式`（如何修复）。
- **若未在 CHANGELOG 中找到当前版本条目**，`release.sh` 会使用通用说明并在注释中提醒补充，建议发布前先补全。

示例条目：

```markdown
## [1.1.0] - 2026-08-30

### 🆕 新功能
- 描述新增能力。

### ✨ 改进
- 描述既有功能优化。

### 🐛 Bug 修复
- **描述**：问题现象
- **影响范围**：受影响的功能/场景
- **修复方式**：如何修复
```


脚本会执行以下步骤：

1. 读取 `version.txt` 并按规则递增，写回单一真源；
2. `make dist` + `make installer-windows` 构建 macOS / Windows / Linux 全部平台产物（文件名均带版本号）；
3. 创建 `vX.Y.Z` tag 并推送到 `origin`（GitHub）和 `gitee`；
4. `gh release create` 发布 GitHub Release 并上传附件；
5. 通过 Gitee API 发布 Gitee Release 并上传附件（大文件直连，避免代理 HTTP:100 失败）。

如需手动发布，仍可按以下步骤操作：

```bash
make dist
make installer-windows
make tag-version
# 再分别在 GitHub / Gitee 创建 Release 并上传 dist/ 下对应产物
```

> Windows 二进制使用 `-H=windowsgui` 构建为 GUI 程序（无控制台窗口），配合系统托盘常驻后台。

## 项目结构

```
philoftp/
├── main.go            # 程序入口、命令行参数解析、组装各层并启动
├── app.go             # 服务管理器（FTP/Web 统一启停 + 状态）
├── logging.go         # 日志初始化（文件 + 控制台双写）
├── tray_windows.go    # Windows 系统托盘实现（状态/启动/停止/打开Web/退出 + 启动提示）
├── tray_other.go      # 非 Windows 控制台模式
├── singleinstance.go  # 单实例锁（跨平台分发，重复启动提示并退出）
├── singleinstance_windows.go / singleinstance_other.go  # 平台具体文件锁实现
├── notify_windows.go / notify_other.go  # 平台提示框（Windows MessageBox / 其他空实现）
├── icon.go            # //go:embed assets/icon.png 托盘图标
├── webstatic.go       # //go:embed web —— 将前端静态资源嵌入单二进制
├── assets/icon.png    # 应用图标
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
├── build/windows/     # Windows 安装包构建脚本（NSIS 模板 + 图标生成 + 说明）
├── dist/              # 各平台构建产物
│   ├── macos/           # macOS 二进制 / zip / PhiloFTP.app
│   ├── windows/         # Windows exe / zip / PhiloFTP-Setup.exe
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
