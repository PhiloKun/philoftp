package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/goftp/server"

	"github.com/philoftp"
	"github.com/philoftp/internal/config"
	"github.com/philoftp/internal/handler"
	"github.com/philoftp/internal/repository"
	"github.com/philoftp/internal/service"
)

// App 是 PhiloFTP 服务管理器，统一管理 FTP/Web 服务的启停与状态。
// 各平台入口（命令行 / 系统托盘）通过本结构控制服务生命周期。
type App struct {
	cfg   *config.Config
	store *repository.DBStore

	mu         sync.Mutex
	ftpSrv     *server.Server
	webSrv     *handler.WebServer
	mdnsSvc    *service.MDNS
	running    bool
	lastErr    error
	startedAt  time.Time
	configPath string
	dbPath     string
}

// NewApp 初始化 App（加载配置与数据库，但不启动服务）。
func NewApp(configPath, dbPath string) (*App, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}
	return &App{
		cfg:        cfg,
		configPath: configPath,
		dbPath:     dbPath,
	}, nil
}

// SetPortOverrides 应用命令行端口/数据目录覆盖
func (a *App) SetPortOverrides(ftpPort, webPort int, dataDir string) {
	if ftpPort != 0 {
		a.cfg.FTPPort = ftpPort
	}
	if webPort != 0 {
		a.cfg.WebPort = webPort
	}
	if dataDir != "" {
		a.cfg.DataDir = dataDir
	}
}

// Start 启动 FTP 与 Web 服务。任一失败返回错误，但不崩溃（供托盘展示）。
func (a *App) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.running {
		return nil
	}

	// 数据目录
	if err := a.cfg.EnsureDataDir(); err != nil {
		a.lastErr = fmt.Errorf("创建数据目录失败: %w", err)
		return a.lastErr
	}

	// 数据库
	store, err := repository.NewDBStore(a.dbPath, config.DataDirOf(a.cfg))
	if err != nil {
		a.lastErr = fmt.Errorf("初始化用户数据库失败: %w", err)
		return a.lastErr
	}
	a.store = store

	// 为默认用户创建 home
	dir := config.DataDirOf(a.cfg)
	for _, u := range store.List() {
		_ = os.MkdirAll(filepath.Join(dir, u.Home), 0755)
	}

	// FTP
	ftpSrv, err := service.StartFTP(a.cfg, store)
	if err != nil {
		a.lastErr = fmt.Errorf("%s", formatPortInUseMsg("FTP ", a.cfg.FTPPort, err))
		if !isAddrInUse(err) {
			a.lastErr = fmt.Errorf("启动 FTP 失败: %w", err)
		}
		return a.lastErr
	}
	a.ftpSrv = ftpSrv

	// Web
	webSrv, err := handler.StartWeb(a.cfg, store, embedfs.WebFS)
	if err != nil {
		if isAddrInUse(err) {
			a.lastErr = fmt.Errorf("%s", formatPortInUseMsg("Web 管理", a.cfg.WebPort, err))
		} else {
			a.lastErr = fmt.Errorf("启动 Web 失败: %w", err)
		}
		if a.ftpSrv != nil {
			_ = a.ftpSrv.Shutdown()
		}
		return a.lastErr
	}
	a.webSrv = webSrv

	// 注册 mDNS（philoftp.local），便于内网其他设备用主机名访问
	a.mdnsSvc = service.StartMDNS(a.cfg.WebPort)

	a.running = true
	a.startedAt = time.Now()
	a.lastErr = nil
	slog.Info("服务已启动",
		"ftp_port", a.cfg.FTPPort,
		"web_port", a.cfg.WebPort,
		"data_dir", dir,
	)
	return nil
}

// Stop 停止 FTP 与 Web 服务。
func (a *App) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.webSrv != nil {
		_ = a.webSrv.Shutdown()
		a.webSrv = nil
	}
	if a.ftpSrv != nil {
		_ = a.ftpSrv.Shutdown()
		a.ftpSrv = nil
	}
	if a.mdnsSvc != nil {
		a.mdnsSvc.Stop()
		a.mdnsSvc = nil
	}
	a.running = false
	slog.Info("服务已停止")
}

// IsRunning 返回服务是否运行中
func (a *App) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.running
}

// Status 返回状态描述（供托盘/日志展示）
func (a *App) Status() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.running {
		return fmt.Sprintf("运行中 · Web http://localhost:%d · FTP :%d（已运行 %s）",
			a.cfg.WebPort, a.cfg.FTPPort, time.Since(a.startedAt).Round(time.Second))
	}
	if a.lastErr != nil {
		return "已停止（上次错误: " + a.lastErr.Error() + "）"
	}
	return "已停止"
}

// WebURL 返回 Web 管理地址
func (a *App) WebURL() string {
	ip := localIP()
	return fmt.Sprintf("http://%s:%d", ip, a.cfg.WebPort)
}

// Config 返回配置
func (a *App) Config() *config.Config { return a.cfg }

// isAddrInUse 判断错误是否为端口已被占用（兼容 Windows / Unix 错误文本）。
func isAddrInUse(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	keywords := []string{
		"Only one usage of each socket address",
		"address already in use",
		"bind: address already in use",
		"通常每个套接字地址",
		"以一种访问权限不允许的方式",
	}
	for _, kw := range keywords {
		if containsCI(msg, kw) {
			return true
		}
	}
	return false
}

func containsCI(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
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
