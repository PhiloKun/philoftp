package service

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/goftp/server"

	"github.com/philoftp/internal/config"
	"github.com/philoftp/internal/model"
	"github.com/philoftp/internal/repository"
)

// ftpAuth 实现 server.Auth，校验数据库中的用户
type ftpAuth struct {
	store *repository.DBStore
}

func (a *ftpAuth) CheckPasswd(name, pass string) (bool, error) {
	_, err := a.store.Authenticate(name, pass)
	return err == nil, err
}

// ftpDriverFactory 为每个连接创建 Driver
type ftpDriverFactory struct {
	store *repository.DBStore
	cfg   *config.Config
}

func (f *ftpDriverFactory) NewDriver() (server.Driver, error) {
	return &ftpDriver{store: f.store, cfg: f.cfg}, nil
}

// ftpDriver 单个连接的文件系统驱动，按用户 chroot + 角色控制
type ftpDriver struct {
	store *repository.DBStore
	cfg   *config.Config
	conn  *server.Conn // 保存连接引用，操作时实时取当前登录用户
}

// Init 在连接建立时调用（此时尚未登录），仅保存 conn 引用
func (d *ftpDriver) Init(c *server.Conn) {
	d.conn = c
}

// currentUser 返回当前已登录用户及其 home 绝对路径
func (d *ftpDriver) currentUser() (model.User, string, bool) {
	if d.conn == nil {
		return model.User{}, "", false
	}
	name := d.conn.LoginUser()
	u, ok := d.store.Get(name)
	if !ok {
		return model.User{}, "", false
	}
	home := model.ResolveHome(config.DataDirOf(d.cfg), u.Home)
	_ = os.MkdirAll(home, 0755)
	return u, home, true
}

// resolve 将 FTP 路径映射到用户 home 下的绝对路径（防越界）
// FTP 路径以 "/" 开头表示用户根目录，需去除前导 "/" 后拼接
func (d *ftpDriver) resolve(path string, home string) string {
	if home == "" {
		return "."
	}
	// 去除 FTP 路径的前导 "/"，统一当作相对 home 的路径
	clean := path
	for len(clean) > 0 && (clean[0] == '/' || clean[0] == filepath.Separator) {
		clean = clean[1:]
	}
	// 防止通过 ".." 越出 home
	joined := filepath.Join(home, clean)
	rel, err := filepath.Rel(home, joined)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return home
	}
	return joined
}

func (d *ftpDriver) Stat(path string) (server.FileInfo, error) {
	_, home, ok := d.currentUser()
	if !ok {
		return nil, os.ErrPermission
	}
	f, err := os.Stat(d.resolve(path, home))
	if err != nil {
		return nil, err
	}
	return &fileInfo{FileInfo: f}, nil
}

func (d *ftpDriver) ChangeDir(path string) error {
	_, home, ok := d.currentUser()
	if !ok {
		return os.ErrPermission
	}
	p := d.resolve(path, home)
	info, err := os.Stat(p)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return os.ErrInvalid
	}
	return nil
}

func (d *ftpDriver) ListDir(path string, cb func(server.FileInfo) error) error {
	_, home, ok := d.currentUser()
	if !ok {
		return os.ErrPermission
	}
	p := d.resolve(path, home)
	entries, err := os.ReadDir(p)
	if err != nil {
		return err
	}
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		if err := cb(&fileInfo{FileInfo: info}); err != nil {
			return err
		}
	}
	return nil
}

func (d *ftpDriver) DeleteDir(path string) error {
	u, home, ok := d.currentUser()
	if !ok {
		return os.ErrPermission
	}
	if !u.CanWrite() {
		return os.ErrPermission
	}
	return os.RemoveAll(d.resolve(path, home))
}

func (d *ftpDriver) DeleteFile(path string) error {
	u, home, ok := d.currentUser()
	if !ok {
		return os.ErrPermission
	}
	if !u.CanWrite() {
		return os.ErrPermission
	}
	return os.Remove(d.resolve(path, home))
}

func (d *ftpDriver) Rename(from, to string) error {
	u, home, ok := d.currentUser()
	if !ok {
		return os.ErrPermission
	}
	if !u.CanWrite() {
		return os.ErrPermission
	}
	return os.Rename(d.resolve(from, home), d.resolve(to, home))
}

func (d *ftpDriver) MakeDir(path string) error {
	u, home, ok := d.currentUser()
	if !ok {
		return os.ErrPermission
	}
	if !u.CanWrite() {
		return os.ErrPermission
	}
	return os.MkdirAll(d.resolve(path, home), 0755)
}

func (d *ftpDriver) GetFile(path string, offset int64) (int64, io.ReadCloser, error) {
	_, home, ok := d.currentUser()
	if !ok {
		return 0, nil, os.ErrPermission
	}
	f, err := os.Open(d.resolve(path, home))
	if err != nil {
		return 0, nil, err
	}
	if offset > 0 {
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			f.Close()
			return 0, nil, err
		}
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return 0, nil, err
	}
	return info.Size(), f, nil
}

func (d *ftpDriver) PutFile(path string, r io.Reader, appendData bool) (int64, error) {
	u, home, ok := d.currentUser()
	if !ok {
		return 0, os.ErrPermission
	}
	if !u.CanWrite() {
		return 0, os.ErrPermission
	}
	p := d.resolve(path, home)
	if appendData {
		f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return 0, err
		}
		defer f.Close()
		return io.Copy(f, r)
	}
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, r)
}

// fileInfo 包装 os.FileInfo 满足 server.FileInfo
type fileInfo struct{ os.FileInfo }

func (f *fileInfo) Owner() string { return "owner" }
func (f *fileInfo) Group() string { return "group" }

// 确保 fileInfo 实现 server.FileInfo（编译期检查）
var _ server.FileInfo = (*fileInfo)(nil)

// ftpController 保存 FTP 服务器的可热重载句柄
type ftpController struct {
	cfg   *config.Config
	store *repository.DBStore
	mu    sync.Mutex
	srv   *server.Server
}

// StartFTP 启动 FTP 服务器，并注册热重载回调，使配置变更（端口/PASV/FTPS 等）
// 能够在不重启整个进程的前提下即时生效。返回当前 server 与错误。
func StartFTP(cfg *config.Config, store *repository.DBStore) (*server.Server, error) {
	ctl := &ftpController{cfg: cfg, store: store}
	if err := ctl.start(); err != nil {
		return nil, err
	}
	// 注册热重载：配置变更时重建 FTP 服务
	cfg.RegisterFTPReloader(func(c *config.Config) error {
		log.Println("检测到 FTP 配置变更，执行热重载...")
		return ctl.reload()
	})
	return ctl.srv, nil
}

// start 依据当前配置创建并启动一个 FTP server（调用方需保证并发安全）
func (ctl *ftpController) start() error {
	opts := &server.ServerOpts{
		Factory:        &ftpDriverFactory{store: ctl.store, cfg: ctl.cfg},
		Auth:           &ftpAuth{store: ctl.store},
		Name:           "PhiloFTP",
		Port:           ctl.cfg.FTPPort,
		PassivePorts:   passivePortRange(ctl.cfg.PASVMinPort, ctl.cfg.PASVMaxPort),
		WelcomeMessage: "Welcome to PhiloFTP",
	}
	if ctl.cfg.EnableFTPS && ctl.cfg.TLSCert != "" && ctl.cfg.TLSKey != "" {
		opts.TLS = true
		opts.CertFile = ctl.cfg.TLSCert
		opts.KeyFile = ctl.cfg.TLSKey
		opts.ExplicitFTPS = true
	}
	s := server.NewServer(opts)
	ctl.srv = s
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("FTP server panic: %v", r)
			}
		}()
		if err := s.ListenAndServe(); err != nil {
			log.Printf("FTP server stopped: %v", err)
		}
	}()
	// 给一点启动时间
	time.Sleep(200 * time.Millisecond)
	return nil
}

// reload 关闭旧 FTP 服务并以最新配置重建（热重载，无需重启进程）
func (ctl *ftpController) reload() error {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()
	if ctl.srv != nil {
		_ = ctl.srv.Shutdown()
		// 等待端口释放，避免新实例绑定冲突
		time.Sleep(300 * time.Millisecond)
	}
	return ctl.start()
}

func passivePortRange(min, max int) string {
	return itoa(min) + "-" + itoa(max)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
