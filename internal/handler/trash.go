package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/philoftp/internal/config"
	"github.com/philoftp/internal/model"
)

// TrashItem 回收站索引项
type TrashItem struct {
	Key       string `json:"key"`        // 回收站内唯一标识（子目录名）
	Name      string `json:"name"`       // 原始文件名/目录名
	OrigPath  string `json:"orig_path"`  // 原始相对 home 的路径（如 /doc/a.txt）
	DeletedAt string `json:"deleted_at"` // 删除时间
}

// trashMu 回收站索引访问互斥
var trashMu sync.Mutex

// trashDir 返回用户回收站目录
func trashDir(cfg *config.Config, user model.User) string {
	home := model.ResolveHome(config.DataDirOf(cfg), user.Home)
	return filepath.Join(home, ".trash")
}

// trashIndexPath 返回回收站索引文件路径
func trashIndexPath(cfg *config.Config, user model.User) string {
	return filepath.Join(trashDir(cfg, user), "index.json")
}

// loadTrash 读取回收站索引
func loadTrash(cfg *config.Config, user model.User) ([]TrashItem, error) {
	data, err := os.ReadFile(trashIndexPath(cfg, user))
	if err != nil {
		if os.IsNotExist(err) {
			return []TrashItem{}, nil
		}
		return nil, err
	}
	var items []TrashItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// saveTrash 写回收站索引
func saveTrash(cfg *config.Config, user model.User, items []TrashItem) error {
	dir := trashDir(cfg, user)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(trashIndexPath(cfg, user), data, 0644)
}

// relToHome 将用户相对路径转换为相对 home 的规范化路径（以 / 开头）
func relToHome(cfg *config.Config, user model.User, rel string) string {
	return filepath.ToSlash(filepath.Clean("/" + rel))
}

// moveToTrash 将 home 内的路径移动到回收站，返回索引项
func moveToTrash(cfg *config.Config, user model.User, rel string) (TrashItem, error) {
	home := model.ResolveHome(config.DataDirOf(cfg), user.Home)
	src, err := safeJoin(cfg, user, rel)
	if err != nil {
		return TrashItem{}, err
	}
	// 防止删除 home 根或回收站本身
	origRel := relToHome(cfg, user, rel)
	if origRel == "/" || strings.HasPrefix(origRel, "/.trash") {
		return TrashItem{}, fmt.Errorf("不能删除该路径")
	}
	// 唯一 key
	key := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizeName(filepath.Base(src)))
	trash := trashDir(cfg, user)
	dst := filepath.Join(trash, key)
	if err := os.MkdirAll(trash, 0755); err != nil {
		return TrashItem{}, err
	}
	if err := os.Rename(src, dst); err != nil {
		return TrashItem{}, fmt.Errorf("删除失败: %w", err)
	}
	_ = os.RemoveAll(filepath.Join(home, origRel)) // 清理可能的残留
	return TrashItem{
		Key:       key,
		Name:      filepath.Base(origRel),
		OrigPath:  origRel,
		DeletedAt: time.Now().Format("2006-01-02 15:04"),
	}, nil
}

// sanitizeName 清洗名称以作为回收站 key 的一部分（去除路径分隔符）
func sanitizeName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	return name
}

// restoreFromTrash 从回收站恢复指定项
func restoreFromTrash(cfg *config.Config, user model.User, key string) error {
	home := model.ResolveHome(config.DataDirOf(cfg), user.Home)
	items, err := loadTrash(cfg, user)
	if err != nil {
		return err
	}
	var target *TrashItem
	var kept []TrashItem
	for i := range items {
		if items[i].Key == key {
			target = &items[i]
			continue
		}
		kept = append(kept, items[i])
	}
	if target == nil {
		return fmt.Errorf("回收站中不存在该项")
	}
	// 原路径目标
	dst := filepath.Join(home, target.OrigPath)
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	src := filepath.Join(trashDir(cfg, user), target.Key)
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("恢复失败: %w", err)
	}
	return saveTrash(cfg, user, kept)
}

// batchDeleteHandler 批量删除（移至回收站）权限：可写用户
func batchDeleteHandler(cfg *config.Config) gin.HandlerFunc {
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
		var req struct {
			Paths []string `json:"paths"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
			return
		}
		if len(req.Paths) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未选择任何文件"})
			return
		}
		trashMu.Lock()
		defer trashMu.Unlock()
		var moved []TrashItem
		var failed []string
		for _, p := range req.Paths {
			item, err := moveToTrash(cfg, user, p)
			if err != nil {
				failed = append(failed, p)
				continue
			}
			moved = append(moved, item)
		}
		// 合并到索引
		items, _ := loadTrash(cfg, user)
		items = append(items, moved...)
		_ = saveTrash(cfg, user, items)
		c.JSON(http.StatusOK, gin.H{
			"ok":      true,
			"moved":   len(moved),
			"failed":  failed,
			"trashed": moved,
		})
	}
}

// trashHandler 回收站列表
func trashHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		items, _ := loadTrash(cfg, user)
		// 过滤掉实际文件已不存在的孤儿项
		var valid []TrashItem
		for _, it := range items {
			if _, err := os.Stat(filepath.Join(trashDir(cfg, user), it.Key)); err == nil {
				valid = append(valid, it)
			}
		}
		sort.Slice(valid, func(i, j int) bool { return valid[i].DeletedAt > valid[j].DeletedAt })
		c.JSON(http.StatusOK, gin.H{"items": valid})
	}
}

// trashRestoreHandler 从回收站恢复
func trashRestoreHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		if !user.CanWrite() {
			c.JSON(http.StatusForbidden, gin.H{"error": "账户禁用，无法恢复"})
			return
		}
		var req struct {
			Keys []string `json:"keys"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
			return
		}
		trashMu.Lock()
		defer trashMu.Unlock()
		restored := 0
		for _, k := range req.Keys {
			if err := restoreFromTrash(cfg, user, k); err == nil {
				restored++
			}
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "restored": restored})
	}
}

// trashClearHandler 清空回收站（物理删除）
func trashClearHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUserOf(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		if !user.CanWrite() {
			c.JSON(http.StatusForbidden, gin.H{"error": "账户禁用"})
			return
		}
		trashMu.Lock()
		defer trashMu.Unlock()
		dir := trashDir(cfg, user)
		if err := os.RemoveAll(dir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "清空失败"})
			return
		}
		_ = os.MkdirAll(dir, 0755)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}
