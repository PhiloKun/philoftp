package handler

import (
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/philoftp/config"
)

// servePage 从嵌入的前端文件系统（web/）读取并返回对应 HTML 页面。
// 前端与后端完全分离：这里只负责把静态资源托管出去，不含任何界面逻辑。
func servePage(webFS fs.FS, name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := fs.ReadFile(webFS, "web/"+name)
		if err != nil {
			c.String(http.StatusNotFound, "页面未找到")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	}
}

// registerStatic 注册前端静态资源路由（登录页 / 注册页 / 控制台 / 资源目录）。
// webFS 必须包含 web/ 目录（由 main 包通过 //go:embed web 嵌入后传入）。
func registerStatic(r *gin.Engine, cfg *config.Config, webFS fs.FS, auth *AuthManager) {
	sub, err := fs.Sub(webFS, "web/assets")
	if err != nil {
		// 缺少前端资源时直接 panic，避免静默运行出空白页
		panic("无法挂载前端静态资源: " + err.Error())
	}
	r.StaticFS("/assets", http.FS(sub))

	// 公开页面
	r.GET("/login", servePage(webFS, "index.html"))
	r.GET("/register", func(c *gin.Context) {
		if !cfg.AllowRegisterEnabled() {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		servePage(webFS, "register.html")(c)
	})

	// 受保护的控制台页（需登录）
	protected := r.Group("")
	protected.Use(auth.RequireAuth())
	protected.GET("/", servePage(webFS, "app.html"))
	protected.GET("/dashboard", servePage(webFS, "app.html"))
}
