package service

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goftp/server"

	"philoftp/config"
	"philoftp/model"
	"philoftp/repository"
)

// ftpAuth 实现 server.Auth，校验 users.json 中的用户
type ftpAuth struct {
	store *repository.UserStore
}

func (a *ftpAuth) CheckPasswd(name, pass string) (bool, error) {
	_, ok := a.store.Authenticate(name, pass)
	return ok, nil
}

// ftpDriverFactory 为每个连接创建 Driver
type ftpDriverFactory struct {
	store *repository.UserStore
	cfg   *config.Config
}

func (f *ftpDriverFactory) NewDriver() (server.Driver, error) {
	return &ftpDriver{store: f.store, cfg: f.cfg}, nil
}

// ftpDriver 单个连接的文件系统驱动，按用户 chroot + 只读控制
type ftpDriver struct {
	store *repository.UserStore
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
	if u.ReadOnly {
		return os.ErrPermission
	}
	return os.RemoveAll(d.resolve(path, home))
}

func (d *ftpDriver) DeleteFile(path string) error {
	u, home, ok := d.currentUser()
	if !ok {
		return os.ErrPermission
	}
	if u.ReadOnly {
		return os.ErrPermission
	}
	return os.Remove(d.resolve(path, home))
}

func (d *ftpDriver) Rename(from, to string) error {
	u, home, ok := d.currentUser()
	if !ok {
		return os.ErrPermission
	}
	if u.ReadOnly {
		return os.ErrPermission
	}
	return os.Rename(d.resolve(from, home), d.resolve(to, home))
}

func (d *ftpDriver) MakeDir(path string) error {
	u, home, ok := d.currentUser()
	if !ok {
		return os.ErrPermission
	}
	if u.ReadOnly {
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
	if u.ReadOnly {
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

// StartFTP 启动 FTP 服务器
func StartFTP(cfg *config.Config, store *repository.UserStore) (*server.Server, error) {
	opts := &server.ServerOpts{
		Factory: &ftpDriverFactory{store: store, cfg: cfg},
		Auth:    &ftpAuth{store: store},
		Name:    "PhiloFTP",
		Port:    cfg.FTPPort,
		PassivePorts: passivePortRange(cfg.PASVMinPort, cfg.PASVMaxPort),
		WelcomeMessage: "Welcome to PhiloFTP",
	}
	if cfg.EnableFTPS && cfg.TLSCert != "" && cfg.TLSKey != "" {
		opts.TLS = true
		opts.CertFile = cfg.TLSCert
		opts.KeyFile = cfg.TLSKey
		opts.ExplicitFTPS = true
	}
	s := server.NewServer(opts)
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
	return s, nil
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
