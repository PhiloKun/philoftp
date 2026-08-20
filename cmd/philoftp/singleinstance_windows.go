//go:build windows

package main

import (
	"fmt"
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

// acquireWindowsMutex 创建/打开全局命名互斥量，作为单实例的第二道防线。
// 返回 held=true 表示成功创建（即当前唯一实例），held=false 表示已有实例。
func acquireWindowsMutex() (held bool, err error) {
	name, _ := windows.UTF16PtrFromString("Local\\PhiloFTP_SingleInstance_Mutex")
	h, err := windows.CreateMutex(nil, false, name)
	if err != nil {
		if err == windows.ERROR_ALREADY_EXISTS {
			return false, nil
		}
		return false, err
	}
	// 仅持有 handle 即可，进程退出时系统自动释放
	if windows.GetLastError() == windows.ERROR_ALREADY_EXISTS {
		_ = windows.CloseHandle(h)
		return false, nil
	}
	// 避免 handle 被 GC 关闭：保留到进程结束
	holdMutex(h)
	return true, nil
}

var globalMutex windows.Handle = windows.InvalidHandle

func holdMutex(h windows.Handle) {
	globalMutex = h
}

// acquireSingleInstance 在 Windows 使用文件锁 + 命名互斥量双保险。
// 任一方式检测到已有实例即视为重复启动。
func acquireSingleInstance() (alreadyRunning bool, err error) {
	lockPath, err := singleInstanceLockPath()
	if err != nil {
		// 锁路径拿不到时，仍尝试命名互斥量
		held, _ := acquireWindowsMutex()
		return !held, nil
	}
	// 文件锁：held=false 表示已有实例
	fileHeld, err := lockFile(lockPath)
	if err != nil {
		return false, fmt.Errorf("单实例文件锁异常: %w", err)
	}
	if !fileHeld {
		return true, nil
	}
	// 文件锁成功后再用互斥量确认一次
	mutexHeld, err := acquireWindowsMutex()
	if err != nil {
		// 互斥量仅作辅助，异常时不阻塞；记录到 stderr
		fmt.Fprintf(os.Stderr, "单实例互斥量检查异常（已忽略）: %v\n", err)
		return false, nil
	}
	if !mutexHeld {
		return true, nil
	}
	return false, nil
}

// osUserCacheDir 返回用于存放锁文件的用户目录（Windows 使用 %USERPROFILE%\.philoftp）。
func osUserCacheDir(home string) string {
	return filepath.Join(home, ".philoftp")
}

func joinPath(dir, name string) string {
	return filepath.Join(dir, name)
}
