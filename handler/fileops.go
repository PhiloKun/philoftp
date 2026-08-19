package handler

import (
	"archive/zip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/philoftp/config"
	"github.com/philoftp/model"
)

// 搜索结果上限，避免目录过大时响应过长/卡顿
const searchMaxResults = 500

// renameHandler 文件/文件夹重命名（仅同级目录内改名，不支持移动）。
// 请求体：{"path": "相对home的路径", "new_name": "新名称"}
func renameHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		if !user.CanWrite() {
			c.JSON(http.StatusForbidden, gin.H{"error": "账户已禁用，无法操作"})
			return
		}
		var req struct {
			Path    string `json:"path"`
			NewName string `json:"new_name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
			return
		}
		if req.Path == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "路径不能为空"})
			return
		}
		newName := strings.TrimSpace(req.NewName)
		if newName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "新名称不能为空"})
			return
		}
		// 不允许包含路径分隔符或危险字符
		if strings.ContainsAny(newName, "/\\") || newName == "." || newName == ".." {
			c.JSON(http.StatusBadRequest, gin.H{"error": "名称非法"})
			return
		}
	src, err := safeJoin(cfg, user, req.Path)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		parent := filepath.Dir(src)
		dst := filepath.Join(parent, newName)
		if _, err := os.Stat(dst); err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "目标名称已存在"})
			return
		}
		if err := os.Rename(src, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "重命名失败: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "path": req.Path, "new_name": newName})
	}
}

// moveHandler 批量移动文件/文件夹到目标目录。
// 请求体：{"paths": ["相对home路径", ...], "dest": "目标目录(相对home，空串表示home根)"}
func moveHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		if !user.CanWrite() {
			c.JSON(http.StatusForbidden, gin.H{"error": "账户已禁用，无法操作"})
			return
		}
		var req struct {
			Paths []string `json:"paths"`
			Dest  string   `json:"dest"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
			return
		}
		if len(req.Paths) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未选择任何文件"})
			return
		}
		destFull, err := safeJoin(cfg, user, req.Dest)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "目标目录非法"})
			return
		}
		info, err := os.Stat(destFull)
		if err != nil || !info.IsDir() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "目标目录不存在"})
			return
		}
		var moved, skipped []string
		for _, p := range req.Paths {
			src, err := safeJoin(cfg, user, p)
			if err != nil {
				skipped = append(skipped, p)
				continue
			}
			name := filepath.Base(src)
			dst := filepath.Join(destFull, name)
			// 同名冲突自动重命名 name_1 / name_2 ...
			if _, err := os.Stat(dst); err == nil {
				ext := filepath.Ext(name)
				base := strings.TrimSuffix(name, ext)
				i := 1
				for {
					dst = filepath.Join(destFull, fmt.Sprintf("%s_%d%s", base, i, ext))
					if _, err := os.Stat(dst); err != nil {
						break
					}
					i++
				}
			}
			if err := os.Rename(src, dst); err != nil {
				skipped = append(skipped, p)
				continue
			}
			moved = append(moved, p)
		}
		if len(skipped) > 0 {
			c.JSON(http.StatusPartialContent, gin.H{
				"ok":      true,
				"moved":   moved,
				"skipped": skipped,
				"message": fmt.Sprintf("已移动 %d 项，%d 项失败", len(moved), len(skipped)),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "moved": moved, "message": fmt.Sprintf("已移动 %d 项", len(moved))})
	}
}

// searchHandler 递归搜索当前用户 home 内的文件/文件夹（按名称子串匹配，大小写不敏感）。
// 查询参数：q=关键词（必填），path=起始目录（相对home，可选，默认home根）
func searchHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		q := strings.TrimSpace(c.Query("q"))
		if q == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
			return
		}
		root, err := safeJoin(cfg, user, c.Query("path"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "起始目录非法"})
			return
		}
		keyword := strings.ToLower(q)
		home := model.ResolveHome(config.DataDirOf(cfg), user.Home)
		results := make([]map[string]interface{}, 0, searchMaxResults)
		_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if p == root {
				return nil
			}
			// 超出上限提前结束
			if len(results) >= searchMaxResults {
				return filepath.SkipAll
			}
			name := d.Name()
			if !strings.Contains(strings.ToLower(name), keyword) {
				return nil
			}
			rel, rerr := filepath.Rel(home, p)
			if rerr != nil {
				return nil
			}
			relSlash := filepath.ToSlash(rel)
			info, ierr := d.Info()
			if ierr != nil {
				return nil
			}
			dir := relSlash
			if !d.IsDir() {
				dir = filepath.ToSlash(filepath.Dir(relSlash))
				if dir == "" || dir == "." {
					dir = "/"
				}
			}
			results = append(results, map[string]interface{}{
				"path":     relSlash,
				"dir":      dir,
				"name":     name,
				"is_dir":   d.IsDir(),
				"size":     info.Size(),
				"mod_time": info.ModTime().Format("2006-01-02 15:04"),
			})
			return nil
			})
		c.JSON(http.StatusOK, gin.H{
			"q":       q,
			"count":   len(results),
			"truncated": len(results) >= searchMaxResults,
			"items":  results,
		})
	}
}

// downloadZipHandler 将多个文件/文件夹打包为 zip 流式返回。
// 查询参数：paths=相对home路径（可多次，或逗号分隔），或 path=单个路径。
func downloadZipHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		var paths []string
		if ps := c.QueryArray("paths"); len(ps) > 0 {
			paths = ps
		} else if p := c.Query("path"); p != "" {
			paths = []string{p}
		} else if ps := c.Query("paths"); ps != "" {
			// 逗号分隔兜底
			for _, s := range strings.Split(ps, ",") {
				if t := strings.TrimSpace(s); t != "" {
					paths = append(paths, t)
				}
			}
		}
		if len(paths) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未指定要下载的文件"})
			return
		}
		fulls := make([]string, 0, len(paths))
		home := model.ResolveHome(config.DataDirOf(cfg), user.Home)
		for _, p := range paths {
			f, err := safeJoin(cfg, user, p)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "路径非法: " + p})
				return
			}
			fulls = append(fulls, f)
		}

		// zip 文件名：单文件用原名，多文件用时间戳
		zipName := "philoftp-" + time.Now().Format("20060102-150405") + ".zip"
		if len(paths) == 1 {
			if info, err := os.Stat(fulls[0]); err == nil {
				if info.IsDir() {
					zipName = filepath.Base(fulls[0]) + ".zip"
				} else {
					zipName = strings.TrimSuffix(filepath.Base(fulls[0]), filepath.Ext(fulls[0])) + ".zip"
				}
			}
		}
		disp := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
			url.QueryEscape(zipName), url.QueryEscape(zipName))
		c.Header("Content-Type", "application/zip")
		c.Header("Content-Disposition", disp)

		zw := zip.NewWriter(c.Writer)
		defer zw.Close()
		// 以所有文件的公共目录前缀作为 zip 内相对路径基准，避免跨目录时出现 ../ 前缀
		base := commonDir(fulls)
		for _, f := range fulls {
			if err := addToZip(zw, f, home, base); err != nil {
				// 单个文件失败不影响其它（已写入部分头部，无法改状态码，仅记录）
				slog.Warn("zip 打包跳过", "path", f, "error", err)
			}
		}
	}
}

// commonDir 计算多个绝对路径的最长公共目录前缀（用于 zip 内相对路径基准），
// 使跨目录打包时不出现 ../ 前缀。若无公共目录则返回空串（回退为按 home 相对路径）。
func commonDir(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	// 取第一个路径的目录作为初始前缀
	prefix := filepath.Dir(paths[0])
	for _, p := range paths[1:] {
		dir := filepath.Dir(p)
		// 逐段缩短 prefix，直到它是 dir 的前缀
		for prefix != dir && !strings.HasPrefix(dir, prefix+string(os.PathSeparator)) && prefix != "." {
			prefix = filepath.Dir(prefix)
			if prefix == "." || prefix == string(os.PathSeparator) {
				return "" // 无公共目录
			}
		}
		if prefix == "." {
			return ""
		}
	}
	if prefix == string(os.PathSeparator) || prefix == "." {
		return ""
	}
	return prefix
}

// addToZip 将 src（文件或目录）递归加入 zip 写入器。
// home 用于计算 zip 内相对路径（避免泄露绝对路径）；base 为公共前缀，用于多文件打包时去掉冗余上层目录。
func addToZip(zw *zip.Writer, src, home, base string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// zip 内相对路径：优先基于 base 去公共前缀，否则用相对 home
		var rel string
		if base != "" {
			if r, rerr := filepath.Rel(base, p); rerr == nil && r != "." {
				rel = r
			} else {
				rel = filepath.Base(p)
			}
		} else {
			r, rerr := filepath.Rel(home, p)
			if rerr != nil {
				return rerr
			}
			rel = r
		}
		rel = filepath.ToSlash(rel)
		if info.IsDir() {
			if rel == "" {
				return nil
			}
			rel += "/"
			_, err := zw.Create(rel)
			return err
		}
		w, err := zw.Create(rel)
		if err != nil {
			return err
		}
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})
}

// unzipHandler 解压已存在于用户 home 内的 zip 文件到目标目录。
// 请求体：{"path": "相对home的zip路径", "dest": "目标目录(相对home，空串表示zip同名目录)", "mode": "rename|overwrite|cancel"(默认rename)}
func unzipHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		if !user.CanWrite() {
			c.JSON(http.StatusForbidden, gin.H{"error": "账户已禁用，无法操作"})
			return
		}
		var req struct {
			Path string `json:"path"`
			Dest string `json:"dest"`
			Mode string `json:"mode"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
			return
		}
		if req.Path == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未指定 zip 文件"})
			return
		}
		zipFull, err := safeJoin(cfg, user, req.Path)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "zip 路径非法"})
			return
		}
		zr, err := zip.OpenReader(zipFull)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无法打开 zip: " + err.Error()})
			return
		}
		defer zr.Close()

		// 目标目录：默认 zip 同名目录；需确保仍在 home 内
		destFull := filepath.Dir(zipFull)
		if req.Dest != "" {
			destFull, err = safeJoin(cfg, user, req.Dest)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "目标目录非法"})
				return
			}
		} else {
			destFull = filepath.Join(filepath.Dir(zipFull), strings.TrimSuffix(filepath.Base(zipFull), filepath.Ext(zipFull)))
		}
		if err := os.MkdirAll(destFull, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目标目录失败: " + err.Error()})
			return
		}
		mode := req.Mode
		if mode == "" {
			mode = "rename"
		}

		var extracted int
		var conflicts []string
		for _, f := range zr.File {
			name := f.Name
			// 防止 zip 穿越（zip slip）
			if strings.Contains(name, "..") || path.IsAbs(name) {
				continue
			}
			outPath := filepath.Join(destFull, name)
			// 二次校验仍在 destFull 内
			if !strings.HasPrefix(outPath, filepath.Clean(destFull)+string(os.PathSeparator)) && outPath != filepath.Clean(destFull) {
				continue
			}
			if f.FileInfo().IsDir() {
				if err := os.MkdirAll(outPath, 0755); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "解压失败: " + err.Error()})
					return
				}
				continue
			}
			if _, err := os.Stat(outPath); err == nil {
				switch mode {
				case "cancel":
					conflicts = append(conflicts, name)
					continue
				case "overwrite":
					// 直接覆盖
				default: // rename
					ext := filepath.Ext(name)
					base := strings.TrimSuffix(name, ext)
					i := 1
					for {
						newName := fmt.Sprintf("%s_%d%s", base, i, ext)
						cand := filepath.Join(destFull, newName)
						if _, err := os.Stat(cand); err != nil {
							outPath = cand
							break
						}
						i++
					}
				}
			}
			if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "解压失败: " + err.Error()})
				return
			}
			if err := extractZipFile(f, outPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "解压失败: " + err.Error()})
				return
			}
			extracted++
		}
		if len(conflicts) > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"ok":        true,
				"extracted": extracted,
				"conflicts": conflicts,
				"message":   fmt.Sprintf("解压 %d 项，%d 项因冲突按取消跳过", extracted, len(conflicts)),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "extracted": extracted, "message": fmt.Sprintf("已解压 %d 个条目", extracted)})
	}
}

// extractZipFile 将 zip 内的单个文件写出到磁盘。
func extractZipFile(f *zip.File, outPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}
