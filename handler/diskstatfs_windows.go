//go:build windows

package handler

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

// diskStatfs 获取路径所在磁盘的总容量与剩余容量（字节）。
// Windows 平台使用 GetDiskFreeSpaceEx 获取卷容量。
// 不可用时返回 ok=false，由调用方决定是否省略存储指标。
func diskStatfs(path string) (total, free int64, ok bool) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	// 文件路径转 UTF-16，供 Windows API 使用
	pathPtr, err := windows.UTF16PtrFromString(abs)
	if err != nil {
		return 0, 0, false
	}
	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	err = windows.GetDiskFreeSpaceEx(pathPtr, &freeBytesAvailable, &totalBytes, &totalFreeBytes)
	if err != nil {
		return 0, 0, false
	}
	total = int64(totalBytes)
	free = int64(totalFreeBytes)
	if total <= 0 {
		return 0, 0, false
	}
	return total, free, true
}
