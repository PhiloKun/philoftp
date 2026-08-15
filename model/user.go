package model

import "path/filepath"

// User 表示一个 FTP / Web 用户
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Home     string `json:"home"`      // 用户根目录（相对 DataDir 或绝对路径）
	ReadOnly bool   `json:"read_only"` // true=只读（不可上传/删除/建目录）
	Enabled  bool   `json:"enabled"`   // 是否启用
}

// ResolveHome 将用户 home 解析为绝对路径（基于 DataDir）
func ResolveHome(dataDir, home string) string {
	if filepath.IsAbs(home) {
		return home
	}
	return filepath.Join(dataDir, home)
}
