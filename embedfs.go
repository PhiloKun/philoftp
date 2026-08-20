// Package embedfs 集中管理需要嵌入二进制的静态资源（前端 web/ 与图标 assets/）。
//
// 设计为仓库根级独立包，规避 Go //go:embed 不支持 ".." 路径的限制：
// web/ 与 assets/ 保留在仓库根目录（与前端代码物理分离、互不耦合），
// 由本文件（位于根目录，与 web/、assets/ 同级）直接嵌入后对外导出，
// 供 cmd/philoftp 入口引用。
package embedfs

import "embed"

// WebFS 嵌入前端静态资源目录（web/），实现前后端单二进制分发。
// 前端代码与后端 Go 代码物理分离：web/ 下为纯 HTML/CSS/JS，不内联于任何 .go 文件。
//
//go:embed web
var WebFS embed.FS

//go:embed assets/icon.png
var iconPNG []byte

//go:embed assets/icon.ico
var iconICO []byte

// IconPNG 返回托盘图标字节（PNG）。
func IconPNG() []byte {
	return iconPNG
}

// IconICO 返回 .ico 格式图标字节（Windows 托盘用，兼容性最好）。
func IconICO() []byte {
	return iconICO
}
