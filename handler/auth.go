package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/philoftp/config"
	"github.com/philoftp/model"
	"github.com/philoftp/repository"
)

const sessionCookie = "philoftp_sid"

// session 表示一个登录态
type session struct {
	username  string
	createdAt time.Time
}

// AuthManager 管理内存会话
type AuthManager struct {
	store *repository.DBStore
	cfg   *config.Config
	// sid -> session
	table map[string]session
}

// NewAuthManager 创建鉴权管理器
func NewAuthManager(cfg *config.Config, store *repository.DBStore) *AuthManager {
	return &AuthManager{store: store, cfg: cfg, table: make(map[string]session)}
}

func newSID() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Login 处理登录
func (a *AuthManager) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}
	user, err := a.store.Authenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	sid := newSID()
	a.table[sid] = session{username: user.Username, createdAt: time.Now()}
	c.SetCookie(sessionCookie, sid, 60*60*24*7, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"username": user.Username,
		"role":     user.Role,
		"is_admin": user.IsAdmin(),
	})
}

// Logout 处理登出
func (a *AuthManager) Logout(c *gin.Context) {
	if sid, err := c.Cookie(sessionCookie); err == nil {
		delete(a.table, sid)
	}
	c.SetCookie(sessionCookie, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// currentUser 从请求中取当前登录用户
func (a *AuthManager) currentUser(c *gin.Context) (model.User, bool) {
	sid, err := c.Cookie(sessionCookie)
	if err != nil {
		return model.User{}, false
	}
	s, ok := a.table[sid]
	if !ok {
		return model.User{}, false
	}
	return a.store.Get(s.username)
}

// RequireAuth 是 Gin 中间件，未登录跳转登录页（HTML 请求）或返回 401（API 请求）
func (a *AuthManager) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := a.currentUser(c)
		if ok {
			c.Set("user", user)
			c.Next()
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录或会话已过期"})
			return
		}
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
	}
}

// CurrentUserOf 供 handler 包内从上下文取已登录用户（需经 RequireAuth 中间件）
func (a *AuthManager) CurrentUserOf(c *gin.Context) (model.User, bool) {
	if v, ok := c.Get("user"); ok {
		if u, ok := v.(model.User); ok {
			return u, true
		}
	}
	return model.User{}, false
}

// RequireRole 是 Gin 中间件，在 RequireAuth 鉴权基础上校验当前用户是否具备指定角色。
// 未登录返回 401，角色不符返回 403，保证权限严格分级。
func (a *AuthManager) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		a.RequireAuth()(c)
		if c.IsAborted() {
			return
		}
		u, _ := c.MustGet("user").(model.User)
		if u.Role != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "权限不足：需要 " + role + " 角色"})
			return
		}
		c.Next()
	}
}
