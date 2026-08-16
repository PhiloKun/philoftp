# 项目长期记忆：philoftp

## 项目概况
PhiloFTP —— 内网 FTP 服务器，带 Gin Web 管理端（用户管理、文件浏览/上传/下载/搜索）。单二进制 Go 程序。

## 目录分层（2026-08-15 重构）
源码按职责拆分为内部包，根目录 `main.go` 组装启动：
- `model/`      → `User` 模型 + `ResolveHome`
- `config/`     → `Config`、`LoadConfig/SaveConfig`、`DataDirOf`、`ToAPI`、`StartTime`
- `repository/` → `UserStore`（用户 JSON 持久化，读写锁；`Upsert`/`Delete` 使用 `saveLocked()` 避免写锁内再次 RLock 死锁）
- `service/`    → `StartFTP` + FTP 文件系统驱动（goftp/server）
- `handler/`    → `StartWeb`（Gin API）+ `DashboardHTML`（内置左侧菜单 SPA；`StartWeb` 会在绑定失败时返回错误，避免 main 死锁）

## 数据存储
本项目不使用 SQLite；用户与配置均使用 JSON 文件：
- 用户：`configs/users.json`（由 `repository/userstore.go` 管理）
- 配置：`configs/config.json`（`config.Config.Save()` 支持运行时热修改并持久化）
- `configs/`    → 默认 `config.json` / `users.json`（运行时自此读取）
- `dist/`       → 各平台构建产物（见下方"不入库"约定）
- 构建：`Makefile`（`make build` / `make run` / `make dist`）

## 重要约定
- **dist/ 不纳入 git**：二进制/zip 由 `make dist` 重新生成。`.gitignore` 用 `/dist/` 整体忽略。
- 默认 config 路径：`configs/config.json`，user：`configs/users.json`（命令行 `-config`/`-users` 可覆盖）。
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
