package main

import _ "embed"

//go:embed assets/icon.png
var iconPNG []byte

//go:embed assets/icon.ico
var iconICO []byte

// iconBytes 返回托盘图标字节（PNG）。
func iconBytes() []byte {
	return iconPNG
}

// iconBytesICO 返回 .ico 格式图标字节（Windows 托盘用，兼容性最好）。
func iconBytesICO() []byte {
	return iconICO
}
