package handler

import (
	"archive/zip"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/philoftp/internal/config"
	"github.com/philoftp/internal/model"
	"github.com/philoftp/internal/repository"
)

// publicBaseURL 返回用于分享链接公开访问的基础地址（与局域网二维码地址策略一致）
func publicBaseURL(cfg *config.Config) string {
	ip := localIPv4()
	return fmt.Sprintf("http://%s:%d", ip, cfg.WebPort)
}

// shareURL 拼接完整公开分享链接
func shareURL(cfg *config.Config, token string) string {
	return publicBaseURL(cfg) + "/s/" + token
}

// ShareRecord 表示一条文件分享记录（由分享者显式授权，公开但受 token 保护）
type ShareRecord struct {
	Token     string `json:"token"`      // 公开访问令牌（URL 路径段）
	Owner     string `json:"owner"`      // 分享者用户名
	RelPath   string `json:"rel_path"`   // 相对 home 的路径（如 /doc/a.txt，目录亦可）
	Name      string `json:"name"`       // 展示名（文件名/目录名）
	IsDir     bool   `json:"is_dir"`     // 是否为目录
	Code      string `json:"code"`       // 提取码（空表示无需）
	CreatedAt string `json:"created_at"` // 创建时间
	ExpiresAt string `json:"expires_at"` // 过期时间（空表示永不过期）
}

// shareMu 分享索引访问互斥
var shareMu sync.Mutex

// shareDir 返回用户分享元数据目录（与回收站同处 home 下隐藏目录）
func shareDir(cfg *config.Config, user model.User) string {
	home := model.ResolveHome(config.DataDirOf(cfg), user.Home)
	return filepath.Join(home, ".shares")
}

// shareIndexPath 返回分享索引文件路径
func shareIndexPath(cfg *config.Config, user model.User) string {
	return filepath.Join(shareDir(cfg, user), "index.json")
}

// loadShares 读取分享索引
func loadShares(cfg *config.Config, user model.User) ([]ShareRecord, error) {
	data, err := os.ReadFile(shareIndexPath(cfg, user))
	if err != nil {
		if os.IsNotExist(err) {
			return []ShareRecord{}, nil
		}
		return nil, err
	}
	var items []ShareRecord
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// saveShares 写分享索引
func saveShares(cfg *config.Config, user model.User, items []ShareRecord) error {
	dir := shareDir(cfg, user)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(shareIndexPath(cfg, user), data, 0644)
}

// genToken 生成 16 字节十六进制随机令牌
func genToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

// createShareHandler 创建分享链接（需登录，分享自己 home 内的文件/目录）
func createShareHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		var req struct {
			Path     string `json:"path"`      // 相对 home 的路径
			Code     string `json:"code"`      // 提取码（可选）
			ExpireIn int    `json:"expire_in"` // 有效期（小时），0 表示永不过期
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
			return
		}
		req.Path = relToHome(cfg, user, req.Path)
		if req.Path == "/" || strings.HasPrefix(req.Path, "/.trash") || strings.HasPrefix(req.Path, "/.shares") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不能分享该路径"})
			return
		}
		// 校验目标确实存在且位于 home 内
		src, err := safeJoin(cfg, user, req.Path)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "路径非法"})
			return
		}
		info, err := os.Stat(src)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
			return
		}
		token, err := genToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
			return
		}
		rec := ShareRecord{
			Token:     token,
			Owner:     user.Username,
			RelPath:   req.Path,
			Name:      filepath.Base(req.Path),
			IsDir:     info.IsDir(),
			Code:      strings.TrimSpace(req.Code),
			CreatedAt: time.Now().Format("2006-01-02 15:04"),
		}
		if req.ExpireIn > 0 {
			rec.ExpiresAt = time.Now().Add(time.Duration(req.ExpireIn) * time.Hour).Format("2006-01-02 15:04")
		}
		shareMu.Lock()
		defer shareMu.Unlock()
		items, _ := loadShares(cfg, user)
		items = append(items, rec)
		if err := saveShares(cfg, user, items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存分享失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "share": rec, "url": shareURL(cfg, token)})
	}
}

// listSharesHandler 列出当前用户的分享
func listSharesHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		shareMu.Lock()
		items, _ := loadShares(cfg, user)
		shareMu.Unlock()
		// 过滤已过期项（仅用于展示，不在此清理）
		now := time.Now()
		var valid []ShareRecord
		for _, it := range items {
			if it.ExpiresAt != "" {
				if t, err := time.Parse("2006-01-02 15:04", it.ExpiresAt); err == nil && now.After(t) {
					continue
				}
			}
			valid = append(valid, it)
		}
		var out []gin.H
		for _, it := range valid {
			out = append(out, gin.H{
				"token":      it.Token,
				"owner":      it.Owner,
				"rel_path":   it.RelPath,
				"name":       it.Name,
				"is_dir":     it.IsDir,
				"code":       it.Code,
				"created_at": it.CreatedAt,
				"expires_at": it.ExpiresAt,
				"url":        shareURL(cfg, it.Token),
			})
		}
		c.JSON(http.StatusOK, gin.H{"items": out})
	}
}

// revokeShareHandler 撤销（删除）指定分享
func revokeShareHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		token := c.Param("token")
		shareMu.Lock()
		defer shareMu.Unlock()
		items, _ := loadShares(cfg, user)
		var kept []ShareRecord
		var found bool
		for _, it := range items {
			if it.Token == token {
				found = true
				continue
			}
			kept = append(kept, it)
		}
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "分享不存在"})
			return
		}
		if err := saveShares(cfg, user, kept); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "撤销失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// lookupShare 在全体用户中按 token 查找有效分享（公开访问用）
// 返回分享记录与对应 owner 用户；若过期或不存在返回 nil
func lookupShare(cfg *config.Config, store *repository.DBStore, token string) (*ShareRecord, *model.User, error) {
	users := store.List()
	now := time.Now()
	for i := range users {
		u := users[i]
		items, err := loadShares(cfg, u)
		if err != nil {
			continue
		}
		for j := range items {
			it := items[j]
			if it.Token != token {
				continue
			}
			if it.ExpiresAt != "" {
				if t, err := time.Parse("2006-01-02 15:04", it.ExpiresAt); err == nil && now.After(t) {
					return nil, nil, nil // 已过期
				}
			}
			cu := u
			return &it, &cu, nil
		}
	}
	return nil, nil, nil
}

// publicShareHandler 处理 /s/:token 公开访问
// 支持 ?code= 提取码、?inline=1 预览、?dl=1 下载（默认显示文件信息页）
func publicShareHandler(cfg *config.Config, store *repository.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		rec, owner, err := lookupShare(cfg, store, token)
		if err != nil {
			c.String(http.StatusInternalServerError, "服务器错误")
			return
		}
		if rec == nil {
			c.String(http.StatusNotFound, "分享链接不存在或已过期")
			return
		}
		// 提取码校验
		codeProvided := c.Query("code")
		if rec.Code != "" && codeProvided == "" {
			// 未带码：返回提取码输入框
			serveCodeEntry(c, rec)
			return
		}
		if rec.Code != "" && codeProvided != rec.Code {
			c.String(http.StatusForbidden, "提取码错误")
			return
		}
		src, err := safeJoin(cfg, *owner, rec.RelPath)
		if err != nil {
			c.String(http.StatusNotFound, "文件不可访问")
			return
		}
		info, err := os.Stat(src)
		if err != nil {
			c.String(http.StatusNotFound, "分享的文件已不存在")
			return
		}
		// 目录直接打包下载
		if info.IsDir() {
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", sanitizeName(rec.Name)))
			c.Header("Content-Type", "application/zip")
			if err := zipDir(src, c.Writer); err != nil {
				c.String(http.StatusInternalServerError, "打包失败")
				return
			}
			return
		}
		// 明确请求下载/预览时直接返回文件流
		if c.Query("dl") == "1" {
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", urlEncodeUTF8(info.Name())))
			c.File(src)
			return
		}
		if c.Query("inline") == "1" {
			ctype := detectContentType(info.Name())
			if ctype != "" {
				c.Header("Content-Type", ctype)
			}
			c.Header("Content-Disposition", fmt.Sprintf("inline; filename*=UTF-8''%s", urlEncodeUTF8(info.Name())))
			c.File(src)
			return
		}
		// 默认：展示文件信息页，由用户决定下载或预览
		serveShareInfoPage(c, rec, info)
	}
}

// serveCodeEntry 返回提取码输入页
func serveCodeEntry(c *gin.Context, rec *ShareRecord) {
	html := `<!doctype html>
<html lang="zh"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>访问分享</title>
<style>
  :root{--cyan:#22d3ee;--ok:#34d399;--bg:#06121a;--txt:#e6f6ff;--muted:#9fb6c4;--line:rgba(34,211,238,.18);--card:rgba(255,255,255,.06)}
  *{box-sizing:border-box}
  body{margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;
       background:radial-gradient(1200px 600px at 50% -10%,#0b2230,var(--bg));
       font-family:"JetBrains Mono",ui-monospace,monospace;color:var(--txt);padding:20px}
  .card{max-width:420px;width:100%;background:var(--card);border:1px solid var(--line);border-radius:18px;
        padding:34px;backdrop-filter:blur(14px);text-align:center;box-shadow:0 0 45px rgba(34,211,238,.12)}
  .lock{font-size:38px;margin-bottom:10px}
  h2{margin:0 0 6px;color:var(--cyan);font-size:22px}
  p{color:var(--muted);margin:0 0 20px;font-size:13px;line-height:1.6}
  .file-name{color:var(--txt);font-weight:700;margin-bottom:18px;word-break:break-all;padding:10px 12px;background:rgba(7,10,18,.45);border-radius:10px;border:1px solid var(--line)}
  input{width:100%;padding:12px;border-radius:12px;border:1px solid rgba(34,211,238,.35);
        background:rgba(0,0,0,.25);color:var(--txt);font-size:16px;text-align:center;letter-spacing:3px}
  input:focus{outline:none;border-color:var(--cyan);box-shadow:0 0 14px rgba(34,211,238,.25)}
  button{margin-top:16px;width:100%;padding:12px;border:0;border-radius:12px;cursor:pointer;font-size:15px;
         background:linear-gradient(135deg,var(--cyan),#0ea5b7);color:#04121a;font-weight:800}
  .err{color:#f87171;min-height:18px;font-size:12px;margin-top:8px}
</style></head><body>
<form class="card" method="get">
  <div class="lock">🔒</div>
  <h2>受保护的分享</h2>
  <p>该分享需要提取码，验证通过后进入文件详情页。</p>
  <div class="file-name">` + escHTML(rec.Name) + `</div>
  <input name="code" type="text" autocomplete="off" placeholder="请输入提取码" autofocus>
  <div class="err" id="err"></div>
  <button type="submit">进入文件详情</button>
</form>
</body></html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// serveShareInfoPage 返回分享文件信息页（含下载/预览按钮）
func serveShareInfoPage(c *gin.Context, rec *ShareRecord, info os.FileInfo) {
	icon := "📄"
	if info.IsDir() {
		icon = "📁"
	}
	ctype := detectContentType(info.Name())
	previewable := ctype != "" && !info.IsDir()
	size := formatBytes(info.Size())
	expires := "永不过期"
	if rec.ExpiresAt != "" {
		expires = "有效期至 " + escHTML(rec.ExpiresAt)
	}
	codeHidden := ""
	codeParam := ""
	if rec.Code != "" {
		codeHidden = `<div class="meta-row"><span>提取码</span><span class="code">` + escHTML(rec.Code) + `</span></div>`
		// 信息页的下载/预览链接必须携带提取码，否则后端会判定"未带码"而跳回提取码页
		codeParam = "&code=" + urlEncodeUTF8(rec.Code)
	}
	previewBtn := ""
	if previewable {
		previewBtn = `<a class="btn btn-ghost" href="?inline=1` + codeParam + `" target="_blank">👁 预览</a>`
	}
	html := `<!doctype html>
<html lang="zh"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>分享文件 - ` + escHTML(info.Name()) + `</title>
<style>
  :root{--cyan:#22d3ee;--ok:#34d399;--bg:#06121a;--txt:#e6f6ff;--muted:#9fb6c4;--line:rgba(34,211,238,.18);--card:rgba(255,255,255,.06)}
  *{box-sizing:border-box}
  body{margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;
       background:radial-gradient(1200px 600px at 50% -10%,#0b2230,var(--bg));
       font-family:"JetBrains Mono",ui-monospace,monospace;color:var(--txt);padding:20px}
  .card{max-width:520px;width:100%;background:var(--card);border:1px solid var(--line);border-radius:20px;
        padding:34px;backdrop-filter:blur(14px);box-shadow:0 0 50px rgba(34,211,238,.12)}
  .head{display:flex;align-items:flex-start;gap:14px;margin-bottom:22px}
  .icon{font-size:42px;line-height:1}
  .name-wrap{min-width:0;flex:1}
  .name{font-size:20px;font-weight:800;color:var(--txt);word-break:break-all;line-height:1.35}
  .type{font-size:12px;color:var(--muted);margin-top:4px}
  .meta{border-top:1px solid var(--line);border-bottom:1px solid var(--line);padding:16px 0;margin:18px 0}
  .meta-row{display:flex;justify-content:space-between;align-items:center;padding:8px 0;font-size:13px}
  .meta-row span:first-child{color:var(--muted)}
  .meta-row span:last-child{color:var(--txt);font-weight:600}
  .meta-row .code{font-family:"JetBrains Mono",monospace;letter-spacing:1px;background:rgba(34,211,238,.12);padding:3px 10px;border-radius:6px;border:1px solid rgba(34,211,238,.25)}
  .actions{display:flex;gap:12px;margin-top:22px}
  .btn{flex:1;padding:12px;border-radius:12px;text-align:center;text-decoration:none;font-weight:800;font-size:14px;cursor:pointer;display:inline-block;transition:transform .08s}
  .btn-primary{background:linear-gradient(135deg,var(--cyan),#0ea5b7);color:#04121a;border:0}
  .btn-ghost{background:rgba(255,255,255,.05);color:var(--txt);border:1px solid var(--line)}
  .btn:hover{transform:translateY(-2px)}
  .footer{margin-top:20px;text-align:center;font-size:11px;color:var(--muted);opacity:.7}
  .tag{display:inline-block;margin-left:8px;padding:2px 8px;border-radius:999px;background:rgba(52,211,153,.12);border:1px solid rgba(52,211,153,.3);color:var(--ok);font-size:11px}
</style></head><body>
<div class="card">
  <div class="head">
    <div class="icon">` + icon + `</div>
    <div class="name-wrap">
      <div class="name">` + escHTML(info.Name()) + `</div>
      <div class="type">` + escHTML(strings.ToUpper(filepath.Ext(info.Name()))) + ` 文件 · ` + size + `</div>
    </div>
  </div>
  <div class="meta">
    <div class="meta-row"><span>分享类型</span><span>` + escHTML(rec.CreatedAt) + ` 分享</span></div>
    <div class="meta-row"><span>有效期</span><span>` + expires + `</span></div>
    ` + codeHidden + `
    <div class="meta-row"><span>访问状态</span><span>有效<span class="tag">已验证</span></span></div>
  </div>
  <div class="actions">
    <a class="btn btn-primary" href="?dl=1` + codeParam + `">⬇ 下载文件</a>` + previewBtn + `
  </div>
  <div class="footer">由 PhiloFTP 生成 · 本链接由分享者显式授权</div>
</div>
</body></html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// escHTML 简易 HTML 转义
func escHTML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(s)
}

// formatBytes 返回人类可读的文件大小
func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return strconv.FormatInt(n, 10) + " B"
	}
	div, exp := int64(unit), 0
	for n >= div*unit && exp < 4 {
		div *= unit
		exp++
	}
	prefix := []string{"KB", "MB", "GB", "TB"}[exp]
	return fmt.Sprintf("%.2f %s", float64(n)/float64(div), prefix)
}

// zipDir 将目录 src 打包为 zip 写入 w（顶层保留目录名本身）
func zipDir(src string, w io.Writer) error {
	zw := zip.NewWriter(w)
	defer zw.Close()
	base := filepath.Dir(src)
	if err := addToZip(zw, src, base, base); err != nil {
		return err
	}
	return zw.Close()
}

// urlEncodeUTF8 对文件名做 RFC 5987 百分号编码（用于 Content-Disposition filename*=）
func urlEncodeUTF8(s string) string {
	return url.QueryEscape(s)
}

// detectContentType 根据扩展名返回适合 inline 预览的 Content-Type
func detectContentType(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".txt", ".md", ".log", ".json", ".csv", ".go", ".js", ".ts", ".css", ".html", ".xml", ".yaml", ".yml":
		return "text/plain; charset=utf-8"
	case ".mp3":
		return "audio/mpeg"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	default:
		return ""
	}
}
