package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"

	"philoftp/config"
	"philoftp/model"
)

// appAuth 保存当前鉴权管理器实例，供辅助函数取登录用户
var appAuth *AuthManager

// currentUserOf 从 Gin 上下文解析登录用户（由 RequireAuth 中间件注入）
func currentUserOf(c *gin.Context) (model.User, bool) {
	if appAuth == nil {
		return model.User{}, false
	}
	return appAuth.CurrentUserOf(c)
}

// safeJoin 将用户相对路径拼接到其 home 目录，并防止越权穿越
func safeJoin(cfg *config.Config, user model.User, rel string) (string, error) {
	home := model.ResolveHome(config.DataDirOf(cfg), user.Home)
	if rel == "" {
		return home, nil
	}
	clean := filepath.Clean("/" + rel) // 规范化，去除 .. 等
	full := filepath.Join(home, clean)
	// 越权检测：必须仍在 home 内
	if !strings.HasPrefix(full, filepath.Clean(home)+string(os.PathSeparator)) && full != filepath.Clean(home) {
		return "", fmt.Errorf("非法路径")
	}
	return full, nil
}

func regexpMatch(pattern, s string) bool {
	ok, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false
	}
	return ok
}

// checkPasswordStrength 校验密码强度，返回错误说明（注册与改密复用）
func checkPasswordStrength(pw string) error {
	if len(pw) < 8 {
		return fmt.Errorf("密码至少 8 位")
	}
	if len(pw) > 64 {
		return fmt.Errorf("密码过长（最多 64 位）")
	}
	var hasLower, hasUpper, hasDigit, hasSymbol bool
	for _, r := range pw {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}
	score := 0
	if hasLower {
		score++
	}
	if hasUpper {
		score++
	}
	if hasDigit {
		score++
	}
	if hasSymbol {
		score++
	}
	if score < 3 {
		return fmt.Errorf("密码需包含大小写字母、数字、符号中至少 3 类")
	}
	return nil
}

// isASCII 判断字符串是否全部为 ASCII 字符（用于选择下载响应头编码方式）
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

// systemHandler 返回运行时系统信息
func systemHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"go_version": runtime.Version(),
		"goroutines": runtime.NumGoroutine(),
		"uptime":     time.Since(config.StartTime).Round(time.Second).String(),
	})
}
