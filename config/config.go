package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// StartTime 记录程序启动时间，供运行时长统计使用
var StartTime = time.Now()

// Config 是程序主配置，保存在 config.json
type Config struct {
	FTPPort     int    `json:"ftp_port"` // FTP 控制端口
	PASVMinPort int    `json:"pasv_min_port"` // 被动模式端口范围（最小）
	PASVMaxPort int    `json:"pasv_max_port"` // 被动模式端口范围（最大）
	WebPort     int    `json:"web_port"` // Web 管理端端口
	DataDir     string `json:"data_dir"` // 文件根目录（用户 home 的相对基准）
	TLSCert     string `json:"tls_cert"` // FTPS 证书（留空则禁用）
	TLSKey      string `json:"tls_key"`  // FTPS 私钥
	EnableFTPS  bool   `json:"enable_ftps"` // 是否启用 FTPS
	AllowRegister bool `json:"allow_register"` // 是否允许 Web 端自助注册（默认 true，生产可关闭）
	mu          sync.RWMutex
	configPath  string // 配置文件路径，供运行时保存使用
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		FTPPort:     2121,
		PASVMinPort: 21100,
		PASVMaxPort: 21110,
		WebPort:     8080,
		DataDir:     "data",
		EnableFTPS:  false,
		AllowRegister: true,
	}
}

// LoadConfig 读取 config.json，不存在则写入默认
func LoadConfig(path string) (*Config, error) {
	c := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := SaveConfig(path, c); err != nil {
				return nil, err
			}
			return c, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("解析 config.json 失败: %w", err)
	}
	// 补全字段
	if c.FTPPort == 0 {
		c.FTPPort = 2121
	}
	if c.PASVMinPort == 0 {
		c.PASVMinPort = 21100
	}
	if c.PASVMaxPort == 0 {
		c.PASVMaxPort = 21110
	}
	if c.WebPort == 0 {
		c.WebPort = 8080
	}
	if c.DataDir == "" {
		c.DataDir = "data"
	}
	c.configPath = path
	return c, nil
}

// Save 将当前配置持久化回加载时的配置文件路径（写锁保护）
func (c *Config) Save() error {
	c.mu.RLock()
	data, err := json.MarshalIndent(c, "", "  ")
	c.mu.RUnlock()
	if err != nil {
		return err
	}
	return os.WriteFile(c.configPath, data, 0644)
}

// SaveConfig 将配置持久化为 JSON
func SaveConfig(path string, c *Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// EnsureDataDir 确保数据根目录存在
func (c *Config) EnsureDataDir() error {
	c.mu.RLock()
	dir := c.DataDir
	c.mu.RUnlock()
	if !filepath.IsAbs(dir) {
		if exe, err := os.Executable(); err == nil {
			dir = filepath.Join(filepath.Dir(exe), dir)
		}
	}
	return os.MkdirAll(dir, 0755)
}

// DataDirOf 返回数据根目录的绝对路径（相对可执行文件时自动补全）
func DataDirOf(c *Config) string {
	c.mu.RLock()
	dir := c.DataDir
	c.mu.RUnlock()
	if !filepath.IsAbs(dir) {
		if exe, err := os.Executable(); err == nil {
			dir = filepath.Join(filepath.Dir(exe), dir)
		}
	}
	return dir
}

// ToAPI 返回对外暴露的配置快照（读锁保护，供 Web API 使用）
func (c *Config) ToAPI() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return map[string]interface{}{
		"ftp_port":      c.FTPPort,
		"pasv_min_port": c.PASVMinPort,
		"pasv_max_port": c.PASVMaxPort,
		"web_port":      c.WebPort,
		"data_dir":      c.DataDir,
		"enable_ftps":   c.EnableFTPS,
		"allow_register": c.AllowRegister,
		"tls_cert":      c.TLSCert,
		"tls_key":       c.TLSKey,
	}
}

// AllowRegisterEnabled 返回是否允许自助注册（读锁保护）
func (c *Config) AllowRegisterEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AllowRegister
}

// UpdateFromMap 在写锁保护下，从字段映射中更新可热改的配置项。
// 仅接受白名单字段，避免被注入未知字段。端口/数据目录等修改需重启生效，
// 但会被持久化到 config.json，下次启动自动应用。
func (c *Config) UpdateFromMap(m map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := m["ftp_port"]; ok {
		if n, ok := toInt(v); ok && n > 0 && n < 65536 {
			c.FTPPort = n
		} else {
			return fmt.Errorf("FTP 端口非法")
		}
	}
	if v, ok := m["web_port"]; ok {
		if n, ok := toInt(v); ok && n > 0 && n < 65536 {
			c.WebPort = n
		} else {
			return fmt.Errorf("Web 端口非法")
		}
	}
	if v, ok := m["pasv_min_port"]; ok {
		if n, ok := toInt(v); ok && n > 0 && n < 65536 {
			c.PASVMinPort = n
		} else {
			return fmt.Errorf("PASV 起始端口非法")
		}
	}
	if v, ok := m["pasv_max_port"]; ok {
		if n, ok := toInt(v); ok && n > 0 && n < 65536 {
			c.PASVMaxPort = n
		} else {
			return fmt.Errorf("PASV 结束端口非法")
		}
	}
	if v, ok := m["data_dir"]; ok {
		if s, ok := v.(string); ok && s != "" {
			c.DataDir = s
		} else {
			return fmt.Errorf("数据目录非法")
		}
	}
	if v, ok := m["enable_ftps"]; ok {
		c.EnableFTPS = toBool(v)
	}
	if v, ok := m["allow_register"]; ok {
		c.AllowRegister = toBool(v)
	}
	if v, ok := m["tls_cert"]; ok {
		if s, ok := v.(string); ok {
			c.TLSCert = s
		} else {
			return fmt.Errorf("TLS 证书路径非法")
		}
	}
	if v, ok := m["tls_key"]; ok {
		if s, ok := v.(string); ok {
			c.TLSKey = s
		} else {
			return fmt.Errorf("TLS 私钥路径非法")
		}
	}
	return nil
}

// toInt 从各种数字类型安全转为 int
func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0, false
		}
		return int(i), true
	}
	return 0, false
}

func toBool(v interface{}) bool {
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return b == "true" || b == "1"
	}
	return false
}
