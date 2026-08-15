package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"philoftp/config"
	"philoftp/model"
	"philoftp/repository"
)

// StartWeb 启动 Web 管理端
func StartWeb(cfg *config.Config, store *repository.UserStore) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(DashboardHTML))
	})

	api := r.Group("/api")
	{
		// 状态
		api.GET("/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"ftp_port":   cfg.FTPPort,
				"web_port":   cfg.WebPort,
				"pasv_ports": fmt.Sprintf("%d-%d", cfg.PASVMinPort, cfg.PASVMaxPort),
				"data_dir":   cfg.DataDir,
				"user_count": len(store.List()),
				"ftps":       cfg.EnableFTPS,
				"uptime":     time.Since(config.StartTime).Round(time.Second).String(),
			})
		})

		// 用户列表
		api.GET("/users", func(c *gin.Context) {
			c.JSON(200, store.List())
		})

		// 新增/更新用户
		api.POST("/users", func(c *gin.Context) {
			var u struct {
				Username string `json:"username"`
				Password string `json:"password"`
				Home     string `json:"home"`
				ReadOnly bool   `json:"read_only"`
				Enabled  bool   `json:"enabled"`
			}
			if err := c.ShouldBindJSON(&u); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			if u.Username == "" || u.Password == "" {
				c.JSON(400, gin.H{"error": "用户名和密码不能为空"})
				return
			}
			if u.Home == "" {
				u.Home = u.Username
			}
			user := model.User{
				Username: u.Username,
				Password: u.Password,
				Home:     u.Home,
				ReadOnly: u.ReadOnly,
				Enabled:  u.Enabled,
			}
			if err := store.Upsert(user); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			// 确保目录存在
			_ = os.MkdirAll(model.ResolveHome(config.DataDirOf(cfg), user.Home), 0755)
			c.JSON(200, gin.H{"ok": true})
		})

		// 删除用户
		api.DELETE("/users/:name", func(c *gin.Context) {
			name := c.Param("name")
			if err := store.Delete(name); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"ok": true})
		})

		// 文件浏览：?path=相对某用户的路径&user=用户名
		api.GET("/files", func(c *gin.Context) {
			user := c.Query("user")
			relPath := c.Query("path")
			u, ok := store.Get(user)
			if !ok {
				c.JSON(400, gin.H{"error": "用户不存在"})
				return
			}
			root := model.ResolveHome(config.DataDirOf(cfg), u.Home)
			abs := filepath.Join(root, relPath)
			entries, err := os.ReadDir(abs)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			type item struct {
				Name  string `json:"name"`
				IsDir bool   `json:"is_dir"`
				Size  int64  `json:"size"`
			}
			var items []item
			for _, e := range entries {
				info, err := e.Info()
				if err != nil {
					continue
				}
				items = append(items, item{
					Name:  e.Name(),
					IsDir: e.IsDir(),
					Size:  info.Size(),
				})
			}
			c.JSON(200, gin.H{
				"path":  relPath,
				"root":  root,
				"items": items,
			})
		})

		// 文件下载：?user=&path=相对root的完整路径
		api.GET("/download", func(c *gin.Context) {
			user := c.Query("user")
			relPath := c.Query("path")
			u, ok := store.Get(user)
			if !ok {
				c.JSON(400, gin.H{"error": "用户不存在"})
				return
			}
			root := model.ResolveHome(config.DataDirOf(cfg), u.Home)
			abs := filepath.Join(root, relPath)
			f, err := os.Open(abs)
			if err != nil {
				c.JSON(400, gin.H{"error": "文件不存在"})
				return
			}
			defer f.Close()
			info, _ := f.Stat()
			c.Header("Content-Disposition", "attachment; filename="+filepath.Base(abs))
			c.Data(200, "application/octet-stream", readAll(f, info.Size()))
		})

		// 文件上传：form user, path(目录), file
		api.POST("/upload", func(c *gin.Context) {
			user := c.PostForm("user")
			dir := c.PostForm("path")
			u, ok := store.Get(user)
			if !ok {
				c.JSON(400, gin.H{"error": "用户不存在"})
				return
			}
			if u.ReadOnly {
				c.JSON(403, gin.H{"error": "该用户为只读"})
				return
			}
			file, err := c.FormFile("file")
			if err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			root := model.ResolveHome(config.DataDirOf(cfg), u.Home)
			destDir := filepath.Join(root, dir)
			_ = os.MkdirAll(destDir, 0755)
			dest := filepath.Join(destDir, filepath.Base(file.Filename))
			if err := c.SaveUploadedFile(file, dest); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"ok": true, "saved": dest})
		})

		// 批量上传
		api.POST("/upload/batch", func(c *gin.Context) {
			user := c.PostForm("user")
			dir := c.PostForm("path")
			u, ok := store.Get(user)
			if !ok {
				c.JSON(400, gin.H{"error": "用户不存在"})
				return
			}
			if u.ReadOnly {
				c.JSON(403, gin.H{"error": "该用户为只读"})
				return
			}
			form, err := c.MultipartForm()
			if err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			files := form.File["files"]
			if len(files) == 0 {
				c.JSON(400, gin.H{"error": "请选择文件"})
				return
			}
			root := model.ResolveHome(config.DataDirOf(cfg), u.Home)
			destDir := filepath.Join(root, dir)
			_ = os.MkdirAll(destDir, 0755)
			var saved []string
			for _, file := range files {
				dest := filepath.Join(destDir, filepath.Base(file.Filename))
				if err := c.SaveUploadedFile(file, dest); err != nil {
					continue
				}
				saved = append(saved, filepath.Base(file.Filename))
			}
			c.JSON(200, gin.H{"ok": true, "count": len(saved), "files": saved})
		})

		// 删除文件/目录
		api.DELETE("/files", func(c *gin.Context) {
			user := c.Query("user")
			relPath := c.Query("path")
			u, ok := store.Get(user)
			if !ok {
				c.JSON(400, gin.H{"error": "用户不存在"})
				return
			}
			if u.ReadOnly {
				c.JSON(403, gin.H{"error": "该用户为只读"})
				return
			}
			root := model.ResolveHome(config.DataDirOf(cfg), u.Home)
			abs := filepath.Join(root, relPath)
			if err := os.RemoveAll(abs); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"ok": true})
		})

		// 批量删除
		api.POST("/files/batch-delete", func(c *gin.Context) {
			var req struct {
				User  string   `json:"user"`
				Paths []string `json:"paths"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			u, ok := store.Get(req.User)
			if !ok {
				c.JSON(400, gin.H{"error": "用户不存在"})
				return
			}
			if u.ReadOnly {
				c.JSON(403, gin.H{"error": "该用户为只读"})
				return
			}
			root := model.ResolveHome(config.DataDirOf(cfg), u.Home)
			var deleted int
			for _, p := range req.Paths {
				abs := filepath.Join(root, p)
				if err := os.RemoveAll(abs); err == nil {
					deleted++
				}
			}
			c.JSON(200, gin.H{"ok": true, "deleted": deleted})
		})

		// 文件搜索
		api.GET("/search", func(c *gin.Context) {
			user := c.Query("user")
			keyword := c.Query("q")
			if keyword == "" {
				c.JSON(400, gin.H{"error": "搜索关键词不能为空"})
				return
			}
			u, ok := store.Get(user)
			if !ok {
				c.JSON(400, gin.H{"error": "用户不存在"})
				return
			}
			root := model.ResolveHome(config.DataDirOf(cfg), u.Home)
			type item struct {
				Name  string `json:"name"`
				Path  string `json:"path"`
				IsDir bool   `json:"is_dir"`
				Size  int64  `json:"size"`
			}
			var results []item
			filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if len(results) >= 100 {
					return filepath.SkipAll
				}
				if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(keyword)) {
					rel, _ := filepath.Rel(root, path)
					results = append(results, item{
						Name:  info.Name(),
						Path:  rel,
						IsDir: info.IsDir(),
						Size:  info.Size(),
					})
				}
				return nil
			})
			c.JSON(200, gin.H{"keyword": keyword, "results": results, "count": len(results)})
		})

		// 创建目录
		api.POST("/mkdir", func(c *gin.Context) {
			var req struct {
				User string `json:"user"`
				Path string `json:"path"`
				Name string `json:"name"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			u, ok := store.Get(req.User)
			if !ok {
				c.JSON(400, gin.H{"error": "用户不存在"})
				return
			}
			if u.ReadOnly {
				c.JSON(403, gin.H{"error": "该用户为只读"})
				return
			}
			root := model.ResolveHome(config.DataDirOf(cfg), u.Home)
			dir := filepath.Join(root, req.Path, req.Name)
			if err := os.MkdirAll(dir, 0755); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"ok": true})
		})

		// 获取配置
		api.GET("/config", func(c *gin.Context) {
			c.JSON(200, cfg.ToAPI())
		})

		// 系统信息
		api.GET("/system", func(c *gin.Context) {
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			c.JSON(200, gin.H{
				"go_version": runtime.Version(),
				"goroutines": runtime.NumGoroutine(),
				"cpu_count":  runtime.NumCPU(),
				"alloc_mb":   float64(mem.Alloc) / 1024 / 1024,
				"sys_mb":     float64(mem.Sys) / 1024 / 1024,
				"gc_count":   mem.NumGC,
				"uptime":     time.Since(config.StartTime).Round(time.Second).String(),
			})
		})
	}

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.WebPort),
		Handler: r,
	}
	go func() {
		_ = srv.ListenAndServe()
	}()
	time.Sleep(200 * time.Millisecond)
	return srv
}

// readAll 读取整个文件内容（受 size 限制）
func readAll(f *os.File, size int64) []byte {
	if size <= 0 {
		return []byte{}
	}
	buf := make([]byte, size)
	n, _ := f.Read(buf)
	return buf[:n]
}
