//go:build !windows

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// lockFile 在非 Windows 平台使用 flock(LOCK_EX|LOCK_NB) 实现互斥锁。
// 返回 held=true 表示成功获得锁；held=false 表示已被其他进程持有。
func lockFile(path string) (held bool, err error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return false, err
	}
	// 注意：不关闭文件，保持锁至进程退出
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = f.Close()
		if err == unix.EWOULDBLOCK {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// acquireSingleInstance 尝试获取全局单实例锁，确保同一时刻只有一个 PhiloFTP 进程运行。
// 成功返回 nil（并持有锁文件句柄，进程退出时由 OS 自动释放）；
// 若已有其他实例运行，返回带 alreadyRunning=true 的错误。
func acquireSingleInstance() (alreadyRunning bool, err error) {
	lockPath, err := singleInstanceLockPath()
	if err != nil {
		return false, err
	}
	held, err := lockFile(lockPath)
	if err != nil {
		// 锁文件本身出问题时，宁可放行（避免阻塞正常启动），但不视为“已运行”
		return false, fmt.Errorf("单实例文件锁异常: %w", err)
	}
	if !held {
		return true, nil
	}
	return false, nil
}

// osUserCacheDir 返回用于存放锁文件的用户目录（其他平台使用 .philoftp 根目录）。
func osUserCacheDir(home string) string {
	return filepath.Join(home, ".philoftp")
}

func joinPath(dir, name string) string {
	return filepath.Join(dir, name)
}
