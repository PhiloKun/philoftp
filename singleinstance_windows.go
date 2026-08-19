//go:build windows

package main

import (
	"os"
	"path/filepath"
	"golang.org/x/sys/windows"
)

// lockFile 在 Windows 平台使用 LockFileEx 实现互斥锁（支持进程退出自动释放）。
// 返回 held=true 表示成功获得锁；held=false 表示已被其他进程持有。
func lockFile(path string) (held bool, err error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return false, err
	}
	// 不关闭文件，保持锁至进程退出
	ol := new(windows.Overlapped)
	err = windows.LockFileEx(
		windows.Handle(f.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0, 1, 0, ol,
	)
	if err != nil {
		_ = f.Close()
		if err == windows.ERROR_LOCK_VIOLATION {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// osUserCacheDir 返回用于存放锁文件的用户目录（Windows 使用 %USERPROFILE%\.philoftp）。
func osUserCacheDir(home string) string {
	return filepath.Join(home, ".philoftp")
}

func joinPath(dir, name string) string {
	return filepath.Join(dir, name)
}
