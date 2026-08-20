//go:build !windows

package main

import (
	"fmt"
	"log/slog"

	"github.com/philoftp/internal/config"
)

// runTray 在非 Windows 平台保持传统控制台模式（不启用系统托盘）。
// 服务以阻塞方式运行，通过 Ctrl+C 停止。headless 参数在该平台忽略。
func runTray(app *App, headless bool) {
	if err := app.Start(); err != nil {
		slog.Error("服务启动失败", "error", err)
		fmt.Println("启动失败，程序即将退出:", err)
		return
	}
	ip := localIP()
	fmt.Println("========================================")
	fmt.Println("  PhiloFTP 内网 FTP 服务器已启动")
	fmt.Println("========================================")
	fmt.Printf("  FTP  地址: ftp://%s:%d\n", ip, app.Config().FTPPort)
	fmt.Printf("  Web  管理: %s\n", app.WebURL())
	fmt.Printf("  数据目录: %s\n", config.DataDirOf(app.Config()))
	fmt.Println("  默认用户: admin/admin123 (管理员)")
	fmt.Println("  按 Ctrl+C 停止")
	fmt.Println("========================================")

	// 阻塞主协程，等待服务运行（信号在 goroutine 中处理）
	select {}
}
