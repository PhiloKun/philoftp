package main

import (
	"fmt"
	"os"
)

// singleInstanceLockPath 返回锁文件路径（用户目录下的 .philoftp/philoftp.lock）。
func singleInstanceLockPath() (string, error) {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		dir := osUserCacheDir(home)
		_ = os.MkdirAll(dir, 0755)
		return joinPath(dir, "philoftp.lock"), nil
	}
	tmp := os.TempDir()
	return joinPath(tmp, "philoftp.lock"), nil
}

// guardSingleInstance 在 main 最前调用：若已有实例则提示并退出进程。
// notify 为可选的平台专属提示函数（如 Windows 气泡/MessageBox），可为 nil。
func guardSingleInstance(notify func(title, msg string)) {
	alreadyRunning, err := acquireSingleInstance()
	if err != nil {
		// 锁文件异常：记录但不阻塞，避免误杀正常启动
		fmt.Fprintln(os.Stderr, "单实例锁检查失败（已忽略）:", err)
		return
	}
	if !alreadyRunning {
		return
	}
	msg := formatSingleInstanceMsg()
	fmt.Fprintln(os.Stderr, msg)
	if notify != nil {
		notify("PhiloFTP", msg)
	}
	os.Exit(0)
}
