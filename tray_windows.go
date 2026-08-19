//go:build windows

package main

import (
	"log/slog"
	"os/exec"
	"sync"

	"github.com/getlantern/systray"
)

// runTray 在 Windows 上启动系统托盘：常驻后台，提供启动/停止/打开Web/退出菜单。
// headless 为 true 时仍使用控制台模式（便于调试/无托盘环境）。
func runTray(app *App, headless bool) {
	if headless {
		runConsole(app)
		return
	}
	systray.Run(func() {
		onTrayReady(app)
	}, func() {
		// 托盘退出时确保服务停止
		app.Stop()
	})
}

// runConsole Windows 控制台模式（headless）
func runConsole(app *App) {
	if err := app.Start(); err != nil {
		slog.Error("服务启动失败", "error", err)
		return
	}
	select {}
}

var (
	trayMu     sync.Mutex
	trayStatus *systray.MenuItem
	trayStart  *systray.MenuItem
	trayStop   *systray.MenuItem
)

func onTrayReady(app *App) {
	// Windows 托盘优先使用 .ico 图标（兼容性最好）
	systray.SetIcon(iconBytesICO())
	systray.SetTitle("PhiloFTP")
	systray.SetTooltip("PhiloFTP 内网 FTP 服务器")

	trayStatus = systray.AddMenuItem("状态：未启动", "")
	trayStatus.Disable()
	systray.AddSeparator()

	openWeb := systray.AddMenuItem("打开 Web 管理页", "在浏览器打开管理界面")
	trayStart = systray.AddMenuItem("启动服务", "启动 FTP 与 Web 服务")
	trayStop = systray.AddMenuItem("停止服务", "停止 FTP 与 Web 服务")
	systray.AddSeparator()
	showLog := systray.AddMenuItem("打开日志目录", "在文件管理器中打开日志目录")
	quit := systray.AddMenuItem("退出 PhiloFTP", "退出并停止服务")

	// 异步启动服务，避免阻塞托盘初始化（托盘图标应立即显示）
	go func() {
		if err := app.Start(); err != nil {
			slog.Error("服务启动失败", "error", err)
			winNotify("PhiloFTP 启动失败", err.Error())
			return
		}
		refreshTrayStatus(app)
		// 启动确认提示：GUI 模式无控制台，弹窗让用户明确知道程序已启动
		winNotify("PhiloFTP 已启动",
			"Web 管理: "+app.WebURL()+"\n点击托盘图标可打开管理页或退出。")
	}()

	// 打开 Web
	go func() {
		for range openWeb.ClickedCh {
			_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", app.WebURL()).Start()
		}
	}()
	// 启动
	go func() {
		for range trayStart.ClickedCh {
			if err := app.Start(); err != nil {
				slog.Error("启动服务失败", "error", err)
			}
			refreshTrayStatus(app)
		}
	}()
	// 停止
	go func() {
		for range trayStop.ClickedCh {
			app.Stop()
			refreshTrayStatus(app)
		}
	}()
	// 打开日志目录
	go func() {
		for range showLog.ClickedCh {
			_ = exec.Command("explorer", defaultLogDir()).Start()
		}
	}()
	// 退出
	go func() {
		for range quit.ClickedCh {
			systray.Quit()
		}
	}()
}

// refreshTrayStatus 更新托盘菜单中的状态与启停按钮可用性
func refreshTrayStatus(app *App) {
	trayMu.Lock()
	defer trayMu.Unlock()
	status := app.Status()
	if trayStatus != nil {
		trayStatus.SetTitle("状态：" + status)
		trayStatus.SetTooltip(status)
	}
	running := app.IsRunning()
	if trayStart != nil {
		trayStart.Enable()
		if running {
			trayStart.Disable()
		}
	}
	if trayStop != nil {
		trayStop.Enable()
		if !running {
			trayStop.Disable()
		}
	}
}
