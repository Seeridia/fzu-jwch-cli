package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	AppDirName     = "fzu-jwch"
	ConfigFileName = "config.json"
)

type Config struct {
	ID         string         `json:"id"`
	Password   string         `json:"password"`
	Identifier string         `json:"identifier"`
	Cookies    []*http.Cookie `json:"cookies"`
	LastLogin  time.Time      `json:"last_login"`
}

type Store struct {
	Path string
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, AppDirName, ConfigFileName), nil
}

func (s Store) ResolvePath() (string, error) {
	if s.Path != "" {
		return s.Path, nil
	}
	return DefaultPath()
}

func (s Store) Load() (*Config, error) {
	path, err := s.ResolvePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrConfigNotFound{Path: path}
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s Store) Save(cfg *Config) error {
	path, err := s.ResolvePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

type ErrConfigNotFound struct {
	Path string
}

func (e ErrConfigNotFound) Error() string {
	return "config not found at " + e.Path + "; run `fzu-jwch login` first"
}
