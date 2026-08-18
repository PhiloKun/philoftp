package service

import (
	"log/slog"
	"net"

	"github.com/hashicorp/mdns"
)

// MDNS 封装 mDNS 服务注册（Bonjour/Avahi），使内网其他设备可通过
// 主机名（如 philoftp.local）访问本机，无需记忆 IP。
type MDNS struct {
	server *mdns.Server
}

// StartMDNS 注册一个 _http._tcp 的 mDNS 服务，实例名 philoftp，
// 主机名 philoftp.local，端口为 webPort。
// 返回 nil 表示注册成功或无需注册（无可用 IP 时静默跳过，不阻塞启动）。
func StartMDNS(webPort int) *MDNS {
	ips := localIPv4s()
	if len(ips) == 0 {
		slog.Warn("mDNS 未注册：未检测到局域网 IP")
		return &MDNS{}
	}
	// 服务类型：_http._tcp，实例 philoftp，主机 philoftp.local
	svc, err := mdns.NewMDNSService("philoftp", "_http._tcp", "", "", webPort, ips, nil)
	if err != nil {
		slog.Warn("mDNS 服务创建失败", "error", err)
		return &MDNS{}
	}
	server, err := mdns.NewServer(&mdns.Config{Zone: svc})
	if err != nil {
		slog.Warn("mDNS 服务启动失败", "error", err)
		return &MDNS{}
	}
	slog.Info("mDNS 已注册", "host", "philoftp.local", "port", webPort)
	return &MDNS{server: server}
}

// Stop 停止并注销 mDNS 服务
func (m *MDNS) Stop() {
	if m != nil && m.server != nil {
		if err := m.server.Shutdown(); err != nil {
			slog.Warn("mDNS 关闭失败", "error", err)
		}
		slog.Info("mDNS 已注销")
	}
}

// localIPv4s 返回本机非回环 IPv4 地址列表
func localIPv4s() []net.IP {
	var ips []net.IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				ips = append(ips, ip4)
			}
		}
	}
	return ips
}
