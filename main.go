package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	configPath := flag.String("config", envOr("PHILOFTP_CONFIG", defaultConfigPath()), "配置文件路径")
	dbPath := flag.String("db", envOr("PHILOFTP_DB", defaultDBPath()), "SQLite 用户数据库路径")
	ftpPort := flag.Int("ftp-port", envInt("PHILOFTP_FTP_PORT", 0), "FTP 端口(覆盖配置)")
	webPort := flag.Int("web-port", envInt("PHILOFTP_WEB_PORT", 0), "Web 管理端口(覆盖配置)")
	dataDir := flag.String("data", envOr("PHILOFTP_DATA_DIR", ""), "数据根目录(覆盖配置)")
	headless := flag.Bool("headless", false, "无托盘模式（仅控制台运行）")
	flag.Parse()

	// 初始化日志：文件 + 控制台
	initLogging("")

	app, err := NewApp(*configPath, *dbPath)
	if err != nil {
		log.Printf("初始化失败: %v", err)
		// 不闪退：若 config 路径不可用，提示并等待用户操作
		fmt.Println("初始化失败:", err)
		fmt.Println("程序即将退出。请检查配置目录是否有写入权限。")
		return
	}
	app.SetPortOverrides(*ftpPort, *webPort, *dataDir)

	// 平台分发：由各平台的 runTray 实现决定（Windows 系统托盘 / 其他控制台）
	runTray(app, *headless)
}

// envOr 返回环境变量值，未设置或为空时返回 fallback。
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envInt 返回环境变量对应的整数，未设置或解析失败时返回 fallback。
func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// defaultConfigPath 返回默认配置文件路径。
// 优先使用可执行文件同目录的 configs/config.json（安装包场景，目录可写），
// 否则回退到用户目录 ~/.philoftp/configs/config.json（规避 Program Files 等无写权限目录）。
func defaultConfigPath() string {
	exeDir := executableDir()
	exeCandidate := filepath.Join(exeDir, "configs", "config.json")
	// 若 exe 同目录存在 configs，且该目录可写，优先使用
	if fileExists(exeCandidate) && dirWritable(filepath.Dir(exeCandidate)) {
		return exeCandidate
	}
	// 回退用户目录
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		userCandidate := filepath.Join(home, ".philoftp", "configs", "config.json")
		_ = os.MkdirAll(filepath.Dir(userCandidate), 0755)
		return userCandidate
	}
	return exeCandidate
}

// defaultDBPath 返回默认用户数据库路径（与配置文件同目录）
func defaultDBPath() string {
	return filepath.Join(filepath.Dir(defaultConfigPath()), "users.db")
}

// executableDir 返回可执行文件所在目录
func executableDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	wd, _ := os.Getwd()
	return wd
}

// fileExists 判断文件是否存在
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// dirWritable 粗略判断目录是否可写（尝试创建临时文件）
func dirWritable(dir string) bool {
	f, err := os.CreateTemp(dir, ".write-test-*")
	if err != nil {
		return false
	}
	_ = f.Close()
	_ = os.Remove(f.Name())
	return true
}
