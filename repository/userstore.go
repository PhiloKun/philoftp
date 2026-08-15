package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"philoftp/model"
)

// UserStore 管理用户列表，带读写锁与持久化
type UserStore struct {
	path  string
	users []model.User
	mu    sync.RWMutex
}

// NewUserStore 加载 users.json，不存在则创建默认管理员与访客
func NewUserStore(path string) (*UserStore, error) {
	s := &UserStore{path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.users = []model.User{
				{Username: "admin", Password: "admin123", Home: "admin", ReadOnly: false, Enabled: true},
				{Username: "guest", Password: "guest123", Home: "guest", ReadOnly: true, Enabled: true},
			}
			if err := s.save(); err != nil {
				return nil, err
			}
			return s, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &s.users); err != nil {
		return nil, fmt.Errorf("解析 users.json 失败: %w", err)
	}
	return s, nil
}

func (s *UserStore) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := json.MarshalIndent(s.users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

// List 返回用户副本
func (s *UserStore) List() []model.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.User, len(s.users))
	copy(out, s.users)
	return out
}

// Get 按用户名查找
func (s *UserStore) Get(username string) (model.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.Username == username {
			return u, true
		}
	}
	return model.User{}, false
}

// Authenticate 校验用户名密码
func (s *UserStore) Authenticate(username, password string) (model.User, bool) {
	u, ok := s.Get(username)
	if !ok || !u.Enabled || u.Password != password {
		return model.User{}, false
	}
	return u, true
}

// Upsert 新增或更新用户
func (s *UserStore) Upsert(u model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.users {
		if existing.Username == u.Username {
			s.users[i] = u
			return s.save()
		}
	}
	s.users = append(s.users, u)
	return s.save()
}

// Delete 删除用户
func (s *UserStore) Delete(username string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.users[:0]
	for _, u := range s.users {
		if u.Username != username {
			out = append(out, u)
		}
	}
	s.users = out
	return s.save()
}
