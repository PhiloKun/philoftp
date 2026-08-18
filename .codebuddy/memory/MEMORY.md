# 项目长期记忆：philoftp

## 项目概况
PhiloFTP —— 内网 FTP 服务器，带 Gin Web 管理端（用户管理、文件浏览/上传/下载/搜索）。单二进制 Go 程序。

## 目录分层（2026-08-15 重构）
源码按职责拆分为内部包，根目录 `main.go` 组装启动：
- `model/`      → `User` 模型 + `ResolveHome`
- `config/`     → `Config`、`LoadConfig/SaveConfig`、`DataDirOf`、`ToAPI`、`StartTime`
- `repository/` → `DBStore`（`repository/dbstore.go`，**基于 SQLite** 的用户存储；纯 Go 驱动 `modernc.org/sqlite` 无 cgo；bcrypt 哈希存储密码；首次启动种子管理员 `admin/admin123`；自动迁移旧 `configs/users.json`；`guardLastAdmin` 防止删除/降级最后一个管理员）
- `service/`    → `StartFTP`（FTP 文件系统驱动 goftp/server；支持**热重载**：`ftpController` 注册 `cfg.RegisterFTPReloader`，配置变更时 `Shutdown()`+重建，无需重启进程）
- `config/`     → `Config` 加 `ftpReloader` 回调 + `RegisterFTPReloader`；`Save()` 改为写锁内 `saveLocked()`（修复旧版 RLock→写文件锁降级竞态）；`UpdateFromMap` 检测 FTP 字段变更后触发热重载
- `handler/`    → `StartWeb`（Gin API）+ `DashboardHTML`/`LoginHTML`/`RegisterHTML`（内置左侧菜单 SPA；`StartWeb` 会在绑定失败时返回错误，避免 main 死锁）
  - **前端实现约束**：采用**纯内联 CSS**（零外部依赖、离线可用、单二进制分发）。早期曾用 Tailwind CDN，因「内网服务器可能无外网」已改为内联 CSS。设计语言为「深空控制台风」：玻璃拟态 + 青色(#22d3ee)辉光 + JetBrains Mono 等宽字体 + 入场动画。
  - ⚠️ 因 Go 原始字符串（反引号）不能含反引号，JS 模板字面量必须改用字符串拼接。

## 数据存储
- 用户：**SQLite 数据库** `configs/users.db`（`repository.DBStore`，`modernc.org/sqlite` 纯 Go 驱动，无 cgo，单二进制可分发）。密码以 bcrypt 加盐哈希存储。**不再使用 JSON 用户文件**（旧 `configs/users.json` 会在首次启动自动迁移为哈希记录并改名 `.migrated`）。
- 配置：`configs/config.json`（`config.Config.Save()` 运行时热修改+持久化）
  - **配置即时生效规则**：PUT /api/config 保存后，FTP 端口/PASV/FTPS 经 `ftpReloader` 热重载 FTP server 立即生效；data_dir/allow_register/enable_ftps 实时读取生效；**仅 Web 管理端口需重启进程**（Gin 绑定后无法换端口）。前端保存成功后会重新 `loadBasic()`+`loadOverview()` 刷新界面并提示。
- `configs/`    → 默认 `config.json` / `users.db`（运行时自此读取）
- `dist/`       → 各平台构建产物（见下方"不入库"约定）
- 构建：`Makefile`（`make build` / `make run` / `make dist`）

## 权限模型（RBAC，2026-08-17 引入）
- 角色两级：`admin`（全部权限：用户管理/系统配置/文件读写/系统信息）与 `user`（普通用户，仅文件上传/下载/删除/新建目录/浏览）。
- 服务端通过 `handler/auth.go` 的 `RequireRole("admin")` 中间件对 `/api/users`、`/api/config`、`/api/system` 强制鉴权（非管理员返回 403）；文件操作接口对所有已登录用户开放。
- 前端在 `/api/me` 返回 `is_admin` 后隐藏「用户管理/基础设置/系统配置」导航与「新增用户」按钮，并阻止非管理员跳转管理视图。
- 普通用户与管理员均可写文件，仅 `enabled=false`（禁用）账户被拒绝写操作。

## 重要约定
- **dist/ 不纳入 git**：二进制/zip 由 `make dist` 重新生成。`.gitignore` 用 `/dist/` 整体忽略。
- 默认 config 路径：`configs/config.json`，user DB：`configs/users.db`（命令行 `-config`/`-db` 可覆盖，已弃用 `-users`）。
- `data/` 运行时生成（各用户 home），仅保留 `.gitkeep`。

## Git 远端
- GitHub：已创建 `https://github.com/PhiloKun/philoftp`（gh CLI，`gh repo create` 成功）。
  - 推送需注意：dist 二进制过大导致 GitHub 408，必须先 `git rm --cached dist/` 再提交。
- Gitee：已创建并推送 `https://gitee.com/PhiloKun/philoftp`（remote 名 `gitee`）。
  - Gitee token 存于 macOS Keychain（security find-internet-password -a PhiloKun -s gitee.com -w），git credential 自动复用，无需手动 token。

## 工作流强制约束（来自用户 rules）
对 philoftp 任何代码改动后流程：
1) 先同步所有相关 md 文档（README.md、docs/）与代码实际参数一致；
2) 再 commit 并推送到 gitee + github 两个远端。
务必先同步文档再上传。
