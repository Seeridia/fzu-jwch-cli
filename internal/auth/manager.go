package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/seeridia/fzu-jwch-cli/internal/client"
)

type Manager struct {
	Store       Store
	Factory     client.Factory
	NoAutoLogin bool
	Timeout     time.Duration
}

func (m Manager) Login(id, password string) (*Config, error) {
	if id == "" {
		return nil, fmt.Errorf("missing id: pass --id or set FZU_JWCH_ID")
	}
	if password == "" {
		return nil, fmt.Errorf("missing password: pass --password, --password-stdin, or set FZU_JWCH_PASSWORD")
	}

	service := m.factory()(client.Credentials{ID: id, Password: password})
	if _, err := client.WithTimeout(m.Timeout, func() (struct{}, error) {
		return struct{}{}, service.Login()
	}); err != nil {
		return nil, err
	}

	data, err := client.WithTimeout(m.Timeout, func() (sessionData, error) {
		identifier, cookies, err := service.SessionData()
		return sessionData{identifier: identifier, cookies: cookies}, err
	})
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		ID:         id,
		Password:   password,
		Identifier: data.identifier,
		Cookies:    data.cookies,
		LastLogin:  time.Now(),
	}
	return cfg, m.Store.Save(cfg)
}

func (m Manager) Service() (client.Service, *Config, error) {
	cfg, err := m.Store.Load()
	if err != nil {
		return nil, nil, err
	}

	service := m.factory()(client.Credentials{
		ID:         cfg.ID,
		Password:   cfg.Password,
		Identifier: cfg.Identifier,
		Cookies:    cfg.Cookies,
	})

	if _, err := client.WithTimeout(m.Timeout, func() (struct{}, error) {
		return struct{}{}, service.CheckSession()
	}); err == nil {
		return service, cfg, nil
	} else if m.NoAutoLogin {
		return nil, nil, err
	}

	if _, err := client.WithTimeout(m.Timeout, func() (struct{}, error) {
		return struct{}{}, service.Login()
	}); err != nil {
		return nil, nil, err
	}

	data, err := client.WithTimeout(m.Timeout, func() (sessionData, error) {
		identifier, cookies, err := service.SessionData()
		return sessionData{identifier: identifier, cookies: cookies}, err
	})
	if err != nil {
		return nil, nil, err
	}

	cfg.Identifier = data.identifier
	cfg.Cookies = data.cookies
	cfg.LastLogin = time.Now()
	if err := m.Store.Save(cfg); err != nil {
		return nil, nil, err
	}

	return service, cfg, nil
}

func (m Manager) factory() client.Factory {
	if m.Factory != nil {
		return m.Factory
	}
	return client.NewJWCHService
}

type sessionData struct {
	identifier string
	cookies    []*http.Cookie
}
