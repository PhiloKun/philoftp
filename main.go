package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"philoftp/config"
	"philoftp/handler"
	"philoftp/model"
	"philoftp/repository"
	"philoftp/service"
)

func main() {
	configPath := flag.String("config", "configs/config.json", "配置文件路径")
	usersPath := flag.String("users", "configs/users.json", "用户文件路径")
	ftpPort := flag.Int("ftp-port", 0, "FTP 端口(覆盖配置)")
	webPort := flag.Int("web-port", 0, "Web 管理端口(覆盖配置)")
	dataDir := flag.String("data", "", "数据根目录(覆盖配置)")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	if *ftpPort != 0 {
		cfg.FTPPort = *ftpPort
	}
	if *webPort != 0 {
		cfg.WebPort = *webPort
	}
	if *dataDir != "" {
		cfg.DataDir = *dataDir
	}

	if err := cfg.EnsureDataDir(); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}

	store, err := repository.NewUserStore(*usersPath)
	if err != nil {
		log.Fatalf("加载用户失败: %v", err)
	}

	// 为默认用户创建 home
	for _, u := range store.List() {
		_ = os.MkdirAll(model.ResolveHome(config.DataDirOf(cfg), u.Home), 0755)
	}

	if _, err := service.StartFTP(cfg, store); err != nil {
		log.Fatalf("启动 FTP 失败: %v", err)
	}

	_ = handler.StartWeb(cfg, store)

	ip := localIP()
	fmt.Println("========================================")
	fmt.Println("  PhiloFTP 内网 FTP 服务器已启动")
	fmt.Println("========================================")
	fmt.Printf("  FTP  地址: ftp://%s:%d\n", ip, cfg.FTPPort)
	fmt.Printf("  Web  管理: http://%s:%d\n", ip, cfg.WebPort)
	fmt.Printf("  数据目录: %s\n", config.DataDirOf(cfg))
	fmt.Println("  默认用户: admin/admin123 (可写)  guest/guest123 (只读)")
	fmt.Println("  按 Ctrl+C 停止")
	fmt.Println("========================================")

	// 阻塞主协程，等待服务运行（信号在 goroutine 中处理）
	select {}
}

// localIP 返回本机非回环 IPv4 地址
func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
