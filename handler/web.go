package handler

import (
	"fmt"
	"io"
	"io/fs"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"philoftp/config"
	"philoftp/model"
	"philoftp/repository"
)

// StartWeb 启动 Web 管理端（Gin）。webFS 为嵌入的前端静态资源（web/ 目录），返回监听地址。
func StartWeb(cfg *config.Config, store *repository.DBStore, webFS fs.FS) (string, error) {
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

	// 受保护接口（所有已登录用户均可访问文件操作）
	authed := r.Group("")
	authed.Use(auth.RequireAuth())
	authed.GET("/api/status", statusHandler(cfg, store))
	authed.GET("/api/files", filesHandler(cfg, store))
	authed.DELETE("/api/files", deleteFileHandler(cfg, store))
	authed.POST("/api/mkdir", mkdirHandler(cfg, store))
	authed.POST("/api/upload", uploadHandler(cfg, store))
	authed.GET("/api/download", downloadHandler(cfg, store))
	authed.POST("/api/logout", auth.Logout)
	authed.GET("/api/me", func(c *gin.Context) {
		u, _ := auth.CurrentUserOf(c)
		c.JSON(http.StatusOK, gin.H{
			"username": u.Username,
			"role":     u.Role,
			"is_admin": u.IsAdmin(),
		})
	})

	// 管理员专属接口（用户管理 / 系统配置 / 系统信息）
	admin := r.Group("")
	admin.Use(auth.RequireAuth(), auth.RequireRole(model.RoleAdmin))
	admin.GET("/api/system", systemHandler)
	admin.GET("/api/config", func(c *gin.Context) { configHandler(c, cfg) })
	admin.PUT("/api/config", func(c *gin.Context) { updateConfigHandler(c, cfg) })
	admin.GET("/api/users", usersHandler(store))
	admin.POST("/api/users", upsertUserHandler(store))
	admin.DELETE("/api/users/:username", deleteUserHandler(store))

	// 前端静态资源与页面（登录 / 注册 / 控制台）
	registerStatic(r, cfg, webFS, auth)

	addr := fmt.Sprintf(":%d", cfg.WebPort)
	errCh := make(chan error, 1)
	go func() { errCh <- r.Run(addr) }()
	// 等待至多 1.5s 确认是否成功绑定；绑定失败（如端口被占用）立即返回错误
	select {
	case err := <-errCh:
		return addr, fmt.Errorf("Web 服务启动失败: %w", err)
	case <-time.After(1500 * time.Millisecond):
		return addr, nil
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
		if err := os.RemoveAll(full); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// uploadHandler 文件上传（同时兼容单文件字段 file 与多文件字段 files）
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
		for _, file := range files {
			dst := filepath.Join(full, filepath.Base(file.Filename))
			if err := c.SaveUploadedFile(file, dst); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败: " + file.Filename})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "count": len(files)})
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
		c.JSON(http.StatusOK, out)
	}
}
