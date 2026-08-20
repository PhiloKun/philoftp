//go:build !windows

package handler

import (
	"path/filepath"
	"syscall"
)

// diskStatfs 获取路径所在磁盘的总容量与剩余容量（字节）。
// Unix 平台（macOS/Linux）使用 syscall.Statfs。
// 不可用时返回 ok=false，由调用方决定是否省略存储指标。
func diskStatfs(path string) (total, free int64, ok bool) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(abs, &stat); err != nil {
		return 0, 0, false
	}
	bsize := stat.Bsize
	if bsize <= 0 {
		return 0, 0, false
	}
	total = int64(stat.Blocks) * int64(bsize)
	avail := stat.Bavail
	if avail <= 0 {
		avail = stat.Bfree
	}
	free = int64(avail) * int64(bsize)
	return total, free, true
}
