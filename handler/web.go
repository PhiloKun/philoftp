package handler

import (
	"fmt"
	"io"
	"io/fs"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"

	"github.com/philoftp/config"
	"github.com/philoftp/model"
	"github.com/philoftp/repository"
)

// version / gitCommit 通过构建时 -ldflags 注入，用于"关于"页展示。
var (
	version   = "dev"
	gitCommit = ""
)

// WebServer 封装 Web 管理端（Gin），支持启动与优雅关闭。
type WebServer struct {
	auth *AuthManager
	srv  *http.Server
}

// Shutdown 优雅关闭 Web 服务
func (w *WebServer) Shutdown() error {
	if w == nil || w.srv == nil {
		return nil
	}
	return w.srv.Close()
}

// StartWeb 启动 Web 管理端（Gin）。webFS 为嵌入的前端静态资源（web/ 目录），返回可停止的句柄。
func StartWeb(cfg *config.Config, store *repository.DBStore, webFS fs.FS) (*WebServer, error) {
	auth := NewAuthManager(cfg, store)
	appAuth = auth
	r := gin.New()
	r.Use(gin.Recovery())

	// 公开 API
	r.POST("/api/login", auth.Login)
	r.POST("/api/register", registerHandler(cfg, store, auth))
	r.GET("/api/config/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"allow_register": cfg.AllowRegisterEnabled()})
	})
	// 访问信息（局域网 IP / mDNS 主机名）无需登录即可获取，供登录页展示"其他电脑如何访问"
	r.GET("/api/access", accessHandler(cfg))
	r.GET("/api/access/qr", accessQRHandler(cfg))

	// 受保护接口（所有已登录用户均可访问文件操作）
	authed := r.Group("")
	authed.Use(auth.RequireAuth())
	authed.GET("/api/status", statusHandler(cfg, store))
	authed.GET("/api/about", aboutHandler())
	authed.GET("/api/overview", overviewHandler(cfg, store, auth))
	authed.GET("/api/files", filesHandler(cfg, store))
	authed.DELETE("/api/files", deleteFileHandler(cfg, store))
	authed.POST("/api/files/batch-delete", batchDeleteHandler(cfg))
	authed.POST("/api/mkdir", mkdirHandler(cfg, store))
	authed.POST("/api/upload", uploadHandler(cfg, store))
	authed.GET("/api/download", downloadHandler(cfg, store))
	// 回收站
	authed.GET("/api/trash", trashHandler(cfg))
	authed.POST("/api/trash/restore", trashRestoreHandler(cfg))
	authed.DELETE("/api/trash", trashClearHandler(cfg))
	authed.POST("/api/logout", auth.Logout)
	authed.GET("/api/me", func(c *gin.Context) {
		u, _ := auth.CurrentUserOf(c)
		c.JSON(http.StatusOK, gin.H{
			"username": u.Username,
			"role":     u.Role,
			"is_admin": u.IsAdmin(),
		})
	})

	// 管理员专属接口（用户管理 / 系统配置）
	admin := r.Group("")
	admin.Use(auth.RequireAuth(), auth.RequireRole(model.RoleAdmin))
	admin.GET("/api/config", func(c *gin.Context) { configHandler(c, cfg) })
	admin.PUT("/api/config", func(c *gin.Context) { updateConfigHandler(c, cfg) })
	admin.GET("/api/users", usersHandler(store))
	admin.POST("/api/users", upsertUserHandler(store))
	admin.DELETE("/api/users/:username", deleteUserHandler(store))

	// 前端静态资源与页面（登录 / 注册 / 控制台）
	registerStatic(r, cfg, webFS, auth)

	addr := fmt.Sprintf(":%d", cfg.WebPort)
	srv := &http.Server{Addr: addr, Handler: r}
	ws := &WebServer{auth: auth, srv: srv}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	// 等待至多 1.5s 确认是否成功绑定；绑定失败（如端口被占用）立即返回错误
	select {
	case err := <-errCh:
		return nil, fmt.Errorf("Web 服务启动失败: %w", err)
	case <-time.After(1500 * time.Millisecond):
		return ws, nil
	}
}

// statusHandler 返回运行时状态
func statusHandler(cfg *config.Config, store *repository.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		pasv := fmt.Sprintf("%d-%d", cfg.PASVMinPort, cfg.PASVMaxPort)
		c.JSON(http.StatusOK, gin.H{
			"ftp_port":   cfg.FTPPort,
			"web_port":   cfg.WebPort,
			"pasv_ports": pasv,
			"ftps":       cfg.EnableFTPS,
			"uptime":     time.Since(config.StartTime).Round(time.Second).String(),
			"user_count": len(store.List()),
		})
	}
}

// aboutHandler 返回"关于"页面所需信息（版本/Go 版本/Git 提交/作者/版本历史/功能特性）。
func aboutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":    version,
			"go_version": runtime.Version(),
			"git_commit": gitCommit,
			"author":     "philokun",
			"repo_gitee": "https://gitee.com/PhiloKun/philoftp",
			"repo_github": "https://github.com/PhiloKun/philoftp",
			"history": []gin.H{
				{"ver": "v1.0.0", "date": "2026-08-18", "title": "稳定版：批量删除 + 回收站", "desc": "文件列表勾选/全选/反选；删除移入回收站可一键恢复或清空；回收站目录从文件列表中隐藏；数据目录路径完整显示。"},
				{"ver": "v0.9.0", "date": "2026-08-18", "title": "FTPS 配置入口移除", "desc": "简化系统设置，聚焦核心 FTP/管理功能。"},
				{"ver": "v0.8.0", "date": "2026-08-18", "title": "上传冲突与移动端", "desc": "上传同名文件支持覆盖/自动重命名/取消；概览页局域网访问卡含二维码+IP+mDNS；移动端响应式优化。"},
				{"ver": "v0.7.0", "date": "2026-08-18", "title": "mDNS 与局域网访问", "desc": "服务自动注册 philoftp.local；登录页与概览页展示访问地址与二维码，内网设备无需记 IP。"},
				{"ver": "v0.6.0", "date": "2026-08-18", "title": "概览卡片自定义", "desc": "概览页卡片支持拖拽排序与自定义添加；卡片设置统一管理可拖动/显示开关；环境变量配置端口。"},
				{"ver": "v0.5.0", "date": "2026-08-18", "title": "Windows 托盘 GUI 与安装包", "desc": "引入 systray 常驻后台；日志写入文件；NSIS 安装包；跨平台磁盘容量实现支持 Windows 编译。"},
				{"ver": "v0.4.0", "date": "2026-08-18", "title": "概览与权限", "desc": "概览页仪表盘快照（KPI+服务器/存储/活动）；RBAC 两级权限；自助注册开关；自定义输入弹窗。"},
				{"ver": "v0.3.0", "date": "2026-08-17", "title": "文件预览与导航", "desc": "在线预览常见文件（图片/音视频/文本/代码/PDF）；面包屑导航；自定义确认弹窗。"},
				{"ver": "v0.2.0", "date": "2026-08-16", "title": "深空控制台前端", "desc": "玻璃拟态 UI + 青色辉光 + JetBrains Mono 等宽字体；零外部依赖离线可用。"},
				{"ver": "v0.1.0", "date": "2026-08-15", "title": "初版", "desc": "内核 FTP 服务（goftp/server）+ Gin Web 管理端；SQLite 用户库（纯 Go 驱动，无 cgo）；配置热重载。"},
			},
			"features": []gin.H{
				{"icon": "📁", "name": "FTP 文件服务", "desc": "goftp/server 内核，支持 PASV 端口范围，适配内网/NAT 环境。"},
				{"icon": "🌐", "name": "Web 管理端", "desc": "Gin 提供 REST API，浏览器管理用户、文件、系统设置。"},
				{"icon": "👥", "name": "用户与权限", "desc": "SQLite + bcrypt 存储，admin/user 两级 RBAC；可关闭自助注册。"},
				{"icon": "📂", "name": "文件管理", "desc": "按用户隔离 home；上传/批量上传/新建目录/下载/在线预览；面包屑导航。"},
				{"icon": "🗑", "name": "回收站", "desc": "删除移入回收站可恢复；批量删除 + 全选/反选；一键撤销。"},
				{"icon": "🎨", "name": "深空控制台 UI", "desc": "玻璃拟态 + 青色辉光 + JetBrains Mono 等宽字体；零外部依赖离线可用。"},
				{"icon": "🖥", "name": "概览页仪表盘", "desc": "KPI + 服务器状态 + 用户分布 + 存储环形图 + 活动会话 + Top 文件 + 局域网访问卡。"},
				{"icon": "📦", "name": "单二进制分发", "desc": "前端 //go:embed web 嵌入，单文件即可运行。"},
				{"icon": "📱", "name": "移动端适配", "desc": "响应式布局；导航栏自动转为顶部横向；触控友好尺寸。"},
				{"icon": "📡", "name": "mDNS 局域网访问", "desc": "注册 philoftp.local；登录页与概览页展示 IP/主机名/二维码，扫码直达。"},
				{"icon": "🖥", "name": "Windows 系统托盘", "desc": "systray 常驻后台；托盘菜单查看状态/启停/打开Web/日志/退出；无闪退。"},
				{"icon": "📦", "name": "Windows 安装包", "desc": "NSIS 一键安装到用户目录；自动放行防火墙；完整卸载支持。"},
				{"icon": "📋", "name": "日志文件", "desc": "运行日志写入 ~/.philoftp/logs/philoftp.log，便于排查与追溯。"},
				{"icon": "⚙", "name": "热重载配置", "desc": "FTP 端口/PASV 即时生效；Web 端口修改需重启。"},
				{"icon": "🔧", "name": "命令行与环境变量", "desc": "支持 -web-port/-ftp-port 与 PHILOFTP_* 环境变量配置。"},
				{"icon": "🛡", "name": "路径越权防护", "desc": "safeJoin 防穿越；删除防护 home 根与回收站目录。"},
			},
		})
	}
}

// accessHandler 返回本机的局域网访问信息（IP / mDNS 主机名 / 端口），
// 供前端展示"其他电脑如何访问本服务"。
func accessHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ip":       localIPv4(),
			"hostname": "philoftp.local",
			"web_port": cfg.WebPort,
			"ftp_port": cfg.FTPPort,
		})
	}
}

// localIPv4 返回本机第一个非回环 IPv4 地址；无则回退 127.0.0.1。
func localIPv4() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	return "127.0.0.1"
}

// accessQRHandler 返回"访问地址"的二维码 PNG，供其他设备扫码直达管理端。
func accessQRHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		addr := fmt.Sprintf("http://%s:%d", localIPv4(), cfg.WebPort)
		png, err := qrcode.Encode(addr, qrcode.Medium, 256)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成二维码失败"})
			return
		}
		c.Header("Content-Type", "image/png")
		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, "image/png", png)
	}
}

// overviewHandler 返回概览页所需的完整统计快照：
// 用户分布 / 文件统计 / 存储用量 / 活跃会话 / 系统运行指标 / 最近更新时间。
func overviewHandler(cfg *config.Config, store *repository.DBStore, auth *AuthManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		dataDir := config.DataDirOf(cfg)
		users := store.List()

		// 1) 用户统计
		var userTotal = len(users)
		var userEnabled, userAdmins int
		for _, u := range users {
			if u.Enabled {
				userEnabled++
			}
			if u.Role == model.RoleAdmin {
				userAdmins++
			}
		}

		// 2) 活跃会话：合并同用户（取最近一次登录时间）
		loggedSet := make(map[string]time.Time)
		if auth != nil {
			for _, s := range auth.table {
				if t, ok := loggedSet[s.username]; !ok || s.createdAt.After(t) {
					loggedSet[s.username] = s.createdAt
				}
			}
		}
		loggedUsers := make([]map[string]interface{}, 0, len(loggedSet))
		for uname, t := range loggedSet {
			role := "user"
			for _, u := range users {
				if u.Username == uname {
					if u.Role == model.RoleAdmin {
						role = "admin"
					}
					break
				}
			}
			loggedUsers = append(loggedUsers, map[string]interface{}{
				"username":  uname,
				"role":      role,
				"login_at":  t.Format("2006-01-02 15:04:05"),
				"login_ago": time.Since(t).Round(time.Second).String(),
			})
		}
		sort.Slice(loggedUsers, func(i, j int) bool {
			return loggedUsers[i]["login_at"].(string) > loggedUsers[j]["login_at"].(string)
		})

		// 3) 文件统计（递归遍历 data 目录）
		var (
			fileTotal int64
			dirTotal  int64
			sizeTotal int64
		)
		extCount := make(map[string]int64)
		extSize := make(map[string]int64)
		var topFiles []map[string]interface{}
		var latestMod time.Time

		_ = filepath.WalkDir(dataDir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if p == dataDir {
				return nil
			}
			info, ierr := d.Info()
			if ierr != nil {
				return nil
			}
			if d.IsDir() {
				dirTotal++
				return nil
			}
			fileTotal++
			sizeTotal += info.Size()
			ext := strings.ToLower(filepath.Ext(p))
			name := filepath.Base(p)
			// 隐藏文件 / .gitkeep / 常见占位文件归为"其它"，避免扩展名被误判
			if ext == "" || strings.HasPrefix(name, ".") || name == ".gitkeep" || name == "Thumbs.db" || name == ".DS_Store" {
				ext = "其它"
			} else {
				ext = strings.TrimPrefix(ext, ".")
			}
			extCount[ext]++
			extSize[ext] += info.Size()
			if info.ModTime().After(latestMod) {
				latestMod = info.ModTime()
			}
			rel, _ := filepath.Rel(dataDir, p)
			topFiles = append(topFiles, map[string]interface{}{
				"path": rel,
				"size": info.Size(),
			})
			return nil
		})

		// Top 文件按大小降序，取前 5
		sort.Slice(topFiles, func(i, j int) bool {
			return topFiles[i]["size"].(int64) > topFiles[j]["size"].(int64)
		})
		if len(topFiles) > 5 {
			topFiles = topFiles[:5]
		}

		// 扩展名分布：按总大小降序
		type extKV struct {
			ext   string
			size  int64
			count int64
		}
		extList := make([]extKV, 0, len(extSize))
		for k, v := range extSize {
			extList = append(extList, extKV{k, v, extCount[k]})
		}
		sort.Slice(extList, func(i, j int) bool { return extList[i].size > extList[j].size })
		extDist := make([]map[string]interface{}, 0, len(extList))
		for _, e := range extList {
			extDist = append(extDist, map[string]interface{}{
				"ext":   e.ext,
				"size":  e.size,
				"count": e.count,
			})
		}

		// 4) 磁盘容量（macOS/Linux/Windows 通用 syscall.Statfs）
		storage := map[string]interface{}{}
		if total, free, ok := diskStatfs(dataDir); ok {
			storage["total"] = total
			storage["free"] = free
			storage["used"] = total - free
			if total > 0 {
				storage["used_pct"] = float64(total-free) * 100 / float64(total)
			}
			storage["path"] = dataDir
		}

		// 5) 运行时长 / 系统
		uptime := time.Since(config.StartTime).Round(time.Second).String()
		now := time.Now().Format("2006-01-02 15:04:05")
		var lastUpdateStr string
		if latestMod.IsZero() {
			lastUpdateStr = "—"
		} else {
			lastUpdateStr = latestMod.Format("2006-01-02 15:04:05")
		}

		c.JSON(http.StatusOK, gin.H{
			"now":         now,
			"uptime":      uptime,
			"last_update": lastUpdateStr,

			"users": gin.H{
				"total":       userTotal,
				"enabled":     userEnabled,
				"disabled":    userTotal - userEnabled,
				"admins":      userAdmins,
				"normal":      userTotal - userAdmins,
				"logged_in":   len(loggedSet),
				"logged_list": loggedUsers,
			},

			"files": gin.H{
				"file_count": fileTotal,
				"dir_count":  dirTotal,
				"total_size": sizeTotal,
				"top_files":  topFiles,
				"ext_dist":   extDist,
			},

			"storage": storage,
			"load": gin.H{
				"goroutines": runtime.NumGoroutine(),
				"go_version": runtime.Version(),
			},
			"server": gin.H{
				"ftp_port":   cfg.FTPPort,
				"web_port":   cfg.WebPort,
				"pasv_ports": fmt.Sprintf("%d-%d", cfg.PASVMinPort, cfg.PASVMaxPort),
				"ftps":       cfg.EnableFTPS,
				"data_dir":   dataDir,
			},
		})
	}
}

// configHandler 返回当前配置快照（供基础设置页读取）
func configHandler(c *gin.Context, cfg *config.Config) {
	c.JSON(http.StatusOK, cfg.ToAPI())
}

// updateConfigHandler 接收配置变更并持久化。FTP 相关字段（端口/PASV/FTPS）
// 会通过已注册回调热重载 FTP 服务，立即生效；Web 端口/数据目录等读取实时生效。
// 仅 Web 监听端口需在页面提示用户重启进程方可切换。
func updateConfigHandler(c *gin.Context, cfg *config.Config) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}
	// 记录是否涉及 Web 端口（该字段需重启进程才切换监听）
	_, webPortChanged := body["web_port"]
	if err := cfg.UpdateFromMap(body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := cfg.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存配置失败: " + err.Error()})
		return
	}
	msg := "配置已保存并即时生效（FTP 服务已热重载，无需重启）"
	if webPortChanged {
		msg = "配置已保存；Web 管理端口需重启服务进程后切换，其余项已即时生效"
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": msg, "web_port_changed": webPortChanged})
}

// registerHandler 处理 Web 端注册（创建一个普通用户，仅文件权限）
func registerHandler(cfg *config.Config, store *repository.DBStore, auth *AuthManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.AllowRegisterEnabled() {
			c.JSON(http.StatusForbidden, gin.H{"error": "当前已关闭自助注册"})
			return
		}
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Home     string `json:"home"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
			return
		}
		user := model.User{Username: req.Username, PasswordHash: req.Password, Home: req.Home}
		if err := validateUser(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if _, exists := store.Get(req.Username); exists {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已被占用"})
			return
		}
		user.Enabled = true
		user.Role = model.RoleUser // 自助注册仅产生普通用户
		if user.Home == "" {
			user.Home = user.Username
		}
		if err := store.Upsert(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
			return
		}
		// 自动创建 home 目录
		_ = os.MkdirAll(model.ResolveHome(config.DataDirOf(cfg), user.Home), 0755)
		// 注册成功后直接登录
		sid := newSID()
		auth.table[sid] = session{username: user.Username, createdAt: time.Now()}
		c.SetCookie(sessionCookie, sid, 60*60*24*7, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"username": user.Username, "role": user.Role, "is_admin": false})
	}
}

// validateUser 校验注册/新增用户输入（含密码强度）
func validateUser(u model.User) error {
	if len(u.Username) < 3 || len(u.Username) > 32 {
		return fmt.Errorf("用户名长度需为 3-32 个字符")
	}
	if !regexpMatch(`^[a-zA-Z0-9_\-]+$`, u.Username) {
		return fmt.Errorf("用户名仅可包含字母、数字、下划线和连字符")
	}
	if err := checkPasswordStrength(u.PasswordHash); err != nil {
		return err
	}
	return nil
}

// downloadHandler 文件下载：支持中文文件名、MIME 推断、Range 断点续传
func downloadHandler(cfg *config.Config, store *repository.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		path := c.Query("path")
		full, err := safeJoin(cfg, user, path)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		info, err := os.Stat(full)
		if err != nil || info.IsDir() {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
			return
		}
		ctype := mime.TypeByExtension(filepath.Ext(full))
		if ctype == "" {
			ctype = "application/octet-stream"
		}
		// 文件名编码（兼容中文）
		ascii := isASCII(info.Name())
		var disp string
		if ascii {
			disp = fmt.Sprintf("attachment; filename=\"%s\"", info.Name())
		} else {
			disp = fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
				strings.ReplaceAll(info.Name(), "\"", ""), url.QueryEscape(info.Name()))
		}

		// inline=1 时改为内联预览（浏览器直接展示而非下载）。
		// 安全限制：text/html 类型一律强制下载，避免预览时执行脚本（XSS）。
		if c.Query("inline") == "1" && !strings.HasPrefix(ctype, "text/html") {
			disp = strings.Replace(disp, "attachment", "inline", 1)
		}
		c.Header("Content-Type", ctype)
		c.Header("Content-Disposition", disp)
		c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
		c.Header("Accept-Ranges", "bytes")

		// Range 支持
		rangeHdr := c.GetHeader("Range")
		if rangeHdr != "" && strings.HasPrefix(rangeHdr, "bytes=") {
			spec := strings.TrimPrefix(rangeHdr, "bytes=")
			parts := strings.Split(spec, "-")
			if len(parts) == 2 {
				start, _ := strconv.ParseInt(parts[0], 10, 64)
				end := info.Size() - 1
				if parts[1] != "" {
					end, _ = strconv.ParseInt(parts[1], 10, 64)
				}
				if start < 0 || end >= info.Size() || start > end {
					c.Header("Content-Range", fmt.Sprintf("bytes */%d", info.Size()))
					c.AbortWithStatus(http.StatusRequestedRangeNotSatisfiable)
					return
				}
				f, err := os.Open(full)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "打开失败"})
					return
				}
				defer f.Close()
				_, _ = f.Seek(start, 0)
				c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, info.Size()))
				c.Header("Content-Length", strconv.FormatInt(end-start+1, 10))
				c.Status(http.StatusPartialContent)
				_, _ = io.CopyN(c.Writer, f, end-start+1)
				return
			}
		}

		c.File(full)
	}
}

// filesHandler 返回某用户目录下的文件/文件夹列表
func filesHandler(cfg *config.Config, store *repository.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		path := c.Query("path")
		full, err := safeJoin(cfg, user, path)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		entries, err := os.ReadDir(full)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "目录不存在"})
			return
		}
		items := make([]map[string]interface{}, 0)
		for _, e := range entries {
			// 过滤回收站等内部系统目录，不在文件列表中显示
			if e.Name() == ".trash" {
				continue
			}
			fi, err := e.Info()
			if err != nil {
				continue
			}
			items = append(items, map[string]interface{}{
				"name":     e.Name(),
				"is_dir":   e.IsDir(),
				"size":     fi.Size(),
				"mod_time": fi.ModTime().Format("2006-01-02 15:04"),
			})
		}
		sort.Slice(items, func(i, j int) bool {
			if items[i]["is_dir"].(bool) != items[j]["is_dir"].(bool) {
				return items[i]["is_dir"].(bool)
			}
			return items[i]["name"].(string) < items[j]["name"].(string)
		})
		c.JSON(http.StatusOK, gin.H{"path": path, "items": items})
	}
}

// deleteFileHandler 删除文件或目录（支持递归删除目录）
func deleteFileHandler(cfg *config.Config, store *repository.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		if !user.CanWrite() {
			c.JSON(http.StatusForbidden, gin.H{"error": "账户禁用，无法删除"})
			return
		}
		path := c.Query("path")
		if path == "" || path == "/" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "非法路径"})
			return
		}
		full, err := safeJoin(cfg, user, path)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 防止误删 home 根目录本身
		if filepath.Clean(full) == filepath.Clean(model.ResolveHome(config.DataDirOf(cfg), user.Home)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除主目录根路径"})
			return
		}
		// 删除改为移入回收站（可恢复）
		trashMu.Lock()
		item, err := moveToTrash(cfg, user, path)
		if err != nil {
			trashMu.Unlock()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items, _ := loadTrash(cfg, user)
		items = append(items, item)
		_ = saveTrash(cfg, user, items)
		trashMu.Unlock()
		c.JSON(http.StatusOK, gin.H{"ok": true, "trash_key": item.Key})
	}
}

// uploadHandler 文件上传（同时兼容单文件字段 file 与多文件字段 files）
// 上传同名文件冲突处理模式
const (
	uploadModeRename    = "rename"    // 自动重命名（默认）
	uploadModeOverwrite = "overwrite" // 覆盖原文件
	uploadModeCancel    = "cancel"    // 存在同名则取消整个上传
)

// uniqueName 返回不冲突的目标文件名：若 name 已存在，则在主名后加 _时间戳 再尝试，
// 仍冲突则递增序号，直到可用。
func uniqueName(dir, name string) string {
	if _, err := os.Stat(filepath.Join(dir, name)); os.IsNotExist(err) {
		return name
	}
	ext := filepath.Ext(name)
	base := name[:len(name)-len(ext)]
	for i := 0; i < 100; i++ {
		cand := fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano()%1000000, ext)
		if _, err := os.Stat(filepath.Join(dir, cand)); os.IsNotExist(err) {
			return cand
		}
	}
	// 兜底：追加随机序号
	return fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), ext)
}

func uploadHandler(cfg *config.Config, store *repository.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		if !user.CanWrite() {
			c.JSON(http.StatusForbidden, gin.H{"error": "账户禁用，无法上传"})
			return
		}
		path := c.PostForm("path")
		mode := c.PostForm("mode")
		if mode == "" {
			mode = uploadModeRename
		}
		full, err := safeJoin(cfg, user, path)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 多文件字段 files 优先；为空时回退到单文件字段 file
		var files []*multipart.FileHeader
		if form, ferr := c.MultipartForm(); ferr == nil {
			files = form.File["files"]
		}
		if len(files) == 0 {
			if f, ferr := c.FormFile("file"); ferr == nil {
				files = []*multipart.FileHeader{f}
			}
		}
		if len(files) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未收到文件"})
			return
		}

		// 冲突检测：收集与目标目录中已存在同名的文件
		var conflicts []map[string]string
		exists := make(map[string]bool)
		for _, f := range files {
			name := filepath.Base(f.Filename)
			if exists[name] {
				continue // 同批次内重复文件按一次处理
			}
			if _, err := os.Stat(filepath.Join(full, name)); err == nil {
				exists[name] = true
				conflicts = append(conflicts, map[string]string{"name": name})
			}
		}

		// mode=cancel：存在同名则取消整个上传，返回 409 与冲突清单
		if mode == uploadModeCancel && len(conflicts) > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"ok":       false,
				"error":    "存在同名文件，已取消上传",
				"conflicts": conflicts,
			})
			return
		}

		// 执行上传
		renamed := make([]map[string]string, 0)
		overwritten := 0
		for _, file := range files {
			name := filepath.Base(file.Filename)
			dstName := name
			if mode == uploadModeRename {
				if _, err := os.Stat(filepath.Join(full, name)); err == nil {
					dstName = uniqueName(full, name)
					renamed = append(renamed, map[string]string{"from": name, "to": dstName})
				}
			} else if mode == uploadModeOverwrite {
				if _, err := os.Stat(filepath.Join(full, name)); err == nil {
					overwritten++
				}
			}
			dst := filepath.Join(full, dstName)
			if err := c.SaveUploadedFile(file, dst); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败: " + file.Filename})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"ok":         true,
			"count":      len(files),
			"mode":       mode,
			"overwritten": overwritten,
			"renamed":    renamed,
		})
	}
}

// mkdirHandler 新建目录
func mkdirHandler(cfg *config.Config, store *repository.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		if !user.CanWrite() {
			c.JSON(http.StatusForbidden, gin.H{"error": "账户禁用，无法建目录"})
			return
		}
		var req struct {
			Path string `json:"path"`
			Name string `json:"name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "目录名不能为空"})
			return
		}
		base, err := safeJoin(cfg, user, req.Path)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		dst := filepath.Join(base, filepath.Base(req.Name))
		if err := os.MkdirAll(dst, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// upsertUserHandler 新增/更新用户（仅管理员可调用）
func upsertUserHandler(store *repository.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Home     string `json:"home"`
			Role     string `json:"role"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
			return
		}
		if req.Username == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码必填"})
			return
		}
		u := model.User{Username: req.Username, PasswordHash: req.Password, Home: req.Home, Role: req.Role, Enabled: true}
		if u.Home == "" {
			u.Home = u.Username
		}
		// 角色校验：仅允许 admin 或 user
		if u.Role != model.RoleAdmin && u.Role != model.RoleUser {
			u.Role = model.RoleUser
		}
		if err := store.Upsert(u); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// deleteUserHandler 删除用户（仅管理员可调用）
func deleteUserHandler(store *repository.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("username")
		if err := store.Delete(name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// usersHandler 返回用户列表（仅管理员可调用）
func usersHandler(store *repository.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		users := store.List()
		out := make([]map[string]interface{}, 0, len(users))
		for _, u := range users {
			out = append(out, map[string]interface{}{
				"username": u.Username,
				"home":     u.Home,
				"role":     u.Role,
				"enabled":  u.Enabled,
			})
		}
		// 前端统一按 { users: [...] } 结构读取
		c.JSON(http.StatusOK, gin.H{"users": out})
	}
}


