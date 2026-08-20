package main

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// initLogging 配置全局日志：同时输出到文件与标准输出。
// 日志文件位于用户数据目录下 logs/philoftp.log（Windows 为 %USERPROFILE%\\.philoftp\logs）。
// 返回日志目录路径。文件打开失败时退化为仅控制台输出，不影响程序运行。
func initLogging(logDir string) string {
	if logDir == "" {
		logDir = defaultLogDir()
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// 无法创建日志目录时不阻塞启动
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
		return logDir
	}
	logPath := filepath.Join(logDir, "philoftp.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
		return logDir
	}
	// 文件 + 控制台双写
	mw := io.MultiWriter(os.Stdout, f)
	slog.SetDefault(slog.New(slog.NewTextHandler(mw, nil)))
	return logDir
}

// defaultLogDir 返回日志默认目录。
// Windows 使用 %USERPROFILE%\.philoftp\logs（避免 Program Files 写权限问题），
// 其他平台使用可执行文件同目录下 logs/。
func defaultLogDir() string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".philoftp", "logs")
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), "logs")
	}
	return "logs"
}
