//go:build !windows

package main

import (
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

// osUserCacheDir 返回用于存放锁文件的用户目录（其他平台使用 .philoftp 根目录）。
func osUserCacheDir(home string) string {
	return filepath.Join(home, ".philoftp")
}

func joinPath(dir, name string) string {
	return filepath.Join(dir, name)
}
