package main

import (
	"fmt"
	"os"
)

// acquireSingleInstance 尝试获取全局单实例锁，确保同一时刻只有一个 PhiloFTP 进程运行。
// 成功返回 nil（并持有锁文件句柄，进程退出时由 OS 自动释放）；
// 若已有其他实例运行，返回带 alreadyRunning=true 的错误。
// 不同平台的具体加锁逻辑见 singleinstance_{windows,other}.go。
func acquireSingleInstance() (alreadyRunning bool, err error) {
	lockPath, err := singleInstanceLockPath()
	if err != nil {
		return false, err
	}
	held, err := lockFile(lockPath)
	if err != nil {
		// 锁文件本身出问题时，宁可放行（避免阻塞正常启动），但不视为“已运行”
		return false, err
	}
	if !held {
		return true, nil
	}
	return false, nil
}

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

// formatSingleInstanceMsg 生成已存在实例时的提示文案。
func formatSingleInstanceMsg() string {
	return "PhiloFTP 已在运行中，请勿重复启动。\n如需管理，请点击系统托盘图标或打开 Web 管理页。"
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
