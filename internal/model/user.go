package model

import "path/filepath"

// Role 定义用户权限等级
const (
	RoleAdmin = "admin" // 最高权限：用户管理、系统配置、文件读写、系统管理
	RoleUser  = "user"  // 普通用户：仅文件上传/下载/删除/新建目录/浏览
)

// User 表示一个 FTP / Web 用户
type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"-"`       // bcrypt 哈希，绝不暴露给前端
	Home         string `json:"home"`    // 用户根目录（相对 DataDir 或绝对路径）
	Role         string `json:"role"`    // "admin" | "user"
	Enabled      bool   `json:"enabled"` // 是否启用
}

// IsAdmin 判断是否为管理员
func (u *User) IsAdmin() bool { return u.Role == RoleAdmin }

// CanWrite 判断用户是否可执行写操作（上传/删除/建目录）。
// 按 RBAC 设计：普通用户与管理员均可写，仅被禁用账户受限。
func (u *User) CanWrite() bool { return u.Enabled }

// ResolveHome 将用户 home 解析为绝对路径（基于 DataDir）
func ResolveHome(dataDir, home string) string {
	if filepath.IsAbs(home) {
		return home
	}
	return filepath.Join(dataDir, home)
}
