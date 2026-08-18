package main

import "embed"

// webFS 嵌入前端静态资源目录（web/），实现前后端单二进制分发。
// 前端代码与后端 Go 代码物理分离：web/ 下为纯 HTML/CSS/JS，不内联于任何 .go 文件。
//
//go:embed web
var webFS embed.FS
