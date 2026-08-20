//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// formatSingleInstanceMsg 生成已存在实例时的提示文案（Windows 用户视角）。
func formatSingleInstanceMsg() string {
	return "PhiloFTP 已经在运行，请勿重复启动。\n" +
		"如果右下角看不到托盘图标，请点击系统托盘的小箭头展开隐藏图标。\n" +
		"或者直接在浏览器访问：http://本机IP:8080"
}

// formatPortInUseMsg 生成端口被占用时的中文提示文案。
// svc 为 "Web 管理" 或 "FTP"，port 为端口号，rawErr 为原始错误（用于日志）。
func formatPortInUseMsg(svc string, port int, rawErr error) string {
	return fmt.Sprintf(
		"%s端口 %d 已被其他程序占用，PhiloFTP 无法启动。\n\n"+
			"可能原因：\n"+
			"· 已经打开了一个 PhiloFTP 实例；\n"+
			"· 其他软件（如另一个 FTP/Web 服务）占用了该端口。\n\n"+
			"解决方法：\n"+
			"1. 在系统托盘找到 PhiloFTP 图标，右键“退出 PhiloFTP”；\n"+
			"2. 如找不到图标，打开任务管理器结束 philoftp.exe 后再启动；\n"+
			"3. 或修改配置文件 configs/config.json 里的端口后重新启动。\n\n"+
			"技术信息：%v",
		svc, port, rawErr,
	)
}

// winNotify 通过 MessageBox 弹出提示框（Windows GUI 无控制台时使用）。
func winNotify(title, msg string) {
	// user32!MessageBoxW
	user32 := windows.NewLazySystemDLL("user32.dll")
	proc := user32.NewProc("MessageBoxW")
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	msgPtr, _ := syscall.UTF16PtrFromString(msg)
	// 0 = MB_OK
	_, _, _ = proc.Call(0,
		uintptr(unsafe.Pointer(msgPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		0)
}
