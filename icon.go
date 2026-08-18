package main

import _ "embed"

//go:embed assets/icon.png
var iconPNG []byte

// iconBytes 返回托盘/窗口图标字节（PNG）。
func iconBytes() []byte {
	return iconPNG
}
