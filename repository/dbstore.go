package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"

	"philoftp/model"
)

// DBStore 是基于 SQLite 的用户存储实现（纯 Go 驱动，无 cgo，单二进制可分发）。
// 同时提供用户认证与基于角色的授权数据支撑。
type DBStore struct {
	dbPath string
	mu     sync.RWMutex
}

// legacyUser 兼容旧版 configs/users.json（明文密码 / read_only 字段）
type legacyUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Home     string `json:"home"`
	ReadOnly bool   `json:"read_only"`
	Enabled  bool   `json:"enabled"`
}

// NewDBStore 打开（必要时创建）SQLite 数据库并初始化表结构。
// 若同目录下存在旧版 users.json，将自动迁移为 bcrypt 哈希记录。
// dataDir 用于为新用户解析 home 相对路径。
func NewDBStore(dbPath, dataDir string) (*DBStore, error) {
	if dir := filepath.Dir(dbPath); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		username      TEXT PRIMARY KEY,
		password_hash TEXT NOT NULL,
		home          TEXT NOT NULL,
		role          TEXT NOT NULL,
		enabled       INTEGER NOT NULL
	)`); err != nil {
		db.Close()
		return nil, fmt.Errorf("create table: %w", err)
	}
	s := &DBStore{dbPath: dbPath}
	if err := s.migrateFromJSON(dbPath, dataDir); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.ensureSeed(dataDir); err != nil {
		db.Close()
		return nil, err
	}
	return s, db.Close()
}

// migrateFromJSON 将旧版 configs/users.json 迁移到 SQLite（明文密码转 bcrypt）。
func (s *DBStore) migrateFromJSON(dbPath, dataDir string) error {
	jsonPath := filepath.Join(filepath.Dir(dbPath), "users.json")
	if _, err := os.Stat(jsonPath); err != nil {
		return nil // 无旧文件，跳过
	}
	raw, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("read legacy users.json: %w", err)
	}
	var legacy []legacyUser
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return fmt.Errorf("parse legacy users.json: %w", err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	for _, lu := range legacy {
		role := model.RoleUser
		if lu.ReadOnly {
			role = model.RoleUser // 旧只读用户仍为普通用户（按新需求普通用户可写，保持 user）
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(lu.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash legacy password: %w", err)
		}
		home := lu.Home
		if home == "" {
			home = lu.Username
		}
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO users(username,password_hash,home,role,enabled) VALUES(?,?,?,?,?)`,
			lu.Username, string(hash), home, role, boolToInt(lu.Enabled),
		); err != nil {
			return fmt.Errorf("migrate user %s: %w", lu.Username, err)
		}
	}
	// 迁移完成后备份旧文件，避免重复迁移
	_ = os.Rename(jsonPath, jsonPath+".migrated")
	return nil
}

// ensureSeed 保证至少存在一个 admin 账户。
func (s *DBStore) ensureSeed(dataDir string) error {
	db, err := sql.Open("sqlite", s.dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE role=?`, model.RoleAdmin).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	// 创建默认管理员 admin/admin123
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(
		`INSERT OR IGNORE INTO users(username,password_hash,home,role,enabled) VALUES(?,?,?,?,?)`,
		"admin", string(hash), "admin", model.RoleAdmin, 1,
	)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// List 返回所有用户（不含密码哈希）
func (s *DBStore) List() []model.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	db, err := sql.Open("sqlite", s.dbPath)
	if err != nil {
		return nil
	}
	defer db.Close()
	rows, err := db.Query(`SELECT username,home,role,enabled FROM users ORDER BY username`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var users []model.User
	for rows.Next() {
		var u model.User
		var enabled int
		if err := rows.Scan(&u.Username, &u.Home, &u.Role, &enabled); err != nil {
			continue
		}
		u.Enabled = enabled == 1
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil
	}
	return users
}

// Get 按用户名获取用户（不含密码哈希）
func (s *DBStore) Get(username string) (model.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	db, err := sql.Open("sqlite", s.dbPath)
	if err != nil {
		return model.User{}, false
	}
	defer db.Close()
	var u model.User
	var enabled int
	err = db.QueryRow(`SELECT username,home,role,enabled FROM users WHERE username=?`, username).
		Scan(&u.Username, &u.Home, &u.Role, &enabled)
	if err != nil {
		return model.User{}, false
	}
	u.Enabled = enabled == 1
	return u, true
}

// Authenticate 校验用户名/密码，成功返回用户；失败返回错误。
func (s *DBStore) Authenticate(username, password string) (model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	db, err := sql.Open("sqlite", s.dbPath)
	if err != nil {
		return model.User{}, err
	}
	defer db.Close()
	var u model.User
	var hash string
	var enabled int
	err = db.QueryRow(`SELECT username,password_hash,home,role,enabled FROM users WHERE username=?`, username).
		Scan(&u.Username, &hash, &u.Home, &u.Role, &enabled)
	if err != nil {
		return model.User{}, errors.New("用户名或密码错误")
	}
	u.Enabled = enabled == 1
	if !u.Enabled {
		return model.User{}, errors.New("账户已禁用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return model.User{}, errors.New("用户名或密码错误")
	}
	return u, nil
}

// Upsert 新增或更新用户（若提供密码明文则重新哈希；否则保留原哈希）。
func (s *DBStore) Upsert(u model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	db, err := sql.Open("sqlite", s.dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	// 防止用户名冲突时覆盖为管理员：若尝试把唯一 admin 降级，需保证至少一个 admin
	if u.Role != model.RoleAdmin {
		if err := s.guardLastAdmin(db, u.Username); err != nil {
			return err
		}
	}
	hash := u.PasswordHash
	if !strings.HasPrefix(hash, "$2") { // 明文或空，重新哈希
		if hash == "" {
			// 新增用户必须设置密码
			return errors.New("密码不能为空")
		}
		h, err := bcrypt.GenerateFromPassword([]byte(hash), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		hash = string(h)
	}
	home := u.Home
	if home == "" {
		home = u.Username
	}
	role := u.Role
	if role == "" {
		role = model.RoleUser
	}
	_, err = db.Exec(
		`INSERT INTO users(username,password_hash,home,role,enabled) VALUES(?,?,?,?,?)
		 ON CONFLICT(username) DO UPDATE SET password_hash=excluded.password_hash,
		                                     home=excluded.home,
		                                     role=excluded.role,
		                                     enabled=excluded.enabled`,
		u.Username, hash, home, role, boolToInt(u.Enabled),
	)
	return err
}

// guardLastAdmin 防止将最后一个管理员降级/删除，避免系统锁死。
func (s *DBStore) guardLastAdmin(db *sql.DB, username string) error {
	var n, isAdmin int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE role=?`, model.RoleAdmin).Scan(&n); err != nil {
		return err
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE role=? AND username=?`, model.RoleAdmin, username).Scan(&isAdmin); err != nil {
		return err
	}
	if n == 1 && isAdmin == 1 {
		return errors.New("至少需保留一个管理员账户")
	}
	return nil
}

// Delete 删除用户（禁止删除最后一个管理员）。
func (s *DBStore) Delete(username string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	db, err := sql.Open("sqlite", s.dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := s.guardLastAdmin(db, username); err != nil {
		return err
	}
	_, err = db.Exec(`DELETE FROM users WHERE username=?`, username)
	return err
}

// SetPassword 修改指定用户密码。
func (s *DBStore) SetPassword(username, newPlain string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	db, err := sql.Open("sqlite", s.dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	hash, err := bcrypt.GenerateFromPassword([]byte(newPlain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	res, err := db.Exec(`UPDATE users SET password_hash=? WHERE username=?`, string(hash), username)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("用户不存在")
	}
	return nil
}
