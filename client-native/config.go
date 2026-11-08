package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Config is everything the user is allowed to tune from the login UI.
// The database user / database name stay fixed, exactly like the old CLI.
type Config struct {
	Server     string `json:"server"`
	Port       int    `json:"port"`
	DBPassword string `json:"dbPassword"`
	GamePath   string `json:"gamePath,omitempty"`
	LastUser   string `json:"lastUser,omitempty"`
}

func defaultConfig() Config {
	return Config{
		Server:     "192.168.200.131",
		Port:       3306,
		DBPassword: "uu5!^%jg",
	}
}

// ConfigStore keeps the on-disk settings in %APPDATA%\DOFLogin\config.json.
type ConfigStore struct {
	mu   sync.RWMutex
	path string
	cfg  Config
}

func newConfigStore(dir string) (*ConfigStore, error) {
	s := &ConfigStore{path: filepath.Join(dir, "config.json"), cfg: defaultConfig()}

	data, err := os.ReadFile(s.path)
	if err == nil {
		var cfg Config
		if json.Unmarshal(data, &cfg) == nil {
			s.cfg = normalizeConfig(cfg)
			return s, nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	s.cfg = normalizeConfig(s.cfg)
	_ = s.saveLocked()
	return s, nil
}

func normalizeConfig(c Config) Config {
	c.Server = strings.TrimSpace(c.Server)
	if c.Server == "" {
		c.Server = defaultConfig().Server
	}
	if c.Port <= 0 || c.Port > 65535 {
		c.Port = defaultConfig().Port
	}
	if c.DBPassword == "" {
		c.DBPassword = defaultConfig().DBPassword
	}
	c.GamePath = strings.Trim(strings.TrimSpace(c.GamePath), `"`)
	c.LastUser = strings.TrimSpace(c.LastUser)
	return c
}

func (s *ConfigStore) Get() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

// Update keeps the current password when an empty password is supplied.
func (s *ConfigStore) Update(c Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c.Server = strings.TrimSpace(c.Server)
	if c.Server == "" {
		return errors.New("服务器地址不能为空")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("端口必须是 1-65535 之间的数字")
	}
	if c.DBPassword == "" {
		c.DBPassword = s.cfg.DBPassword
	}
	c.GamePath = strings.Trim(strings.TrimSpace(c.GamePath), `"`)
	c.LastUser = s.cfg.LastUser
	if c.GamePath != "" {
		if st, err := os.Stat(c.GamePath); err != nil || st.IsDir() {
			return errors.New("游戏路径无效，请填写完整的 DNF.exe 路径")
		}
	}

	s.cfg = c
	return s.saveLocked()
}

// SetLastUser remembers the account name when "记住账号" is checked.
func (s *ConfigStore) SetLastUser(user string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.LastUser = strings.TrimSpace(user)
	return s.saveLocked()
}

func (s *ConfigStore) saveLocked() error {
	data, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *ConfigStore) Addr() string {
	c := s.Get()
	return c.Server + ":" + strconv.Itoa(c.Port)
}
