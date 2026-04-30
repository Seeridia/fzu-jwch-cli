package auth

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestStoreSaveLoadUses0600(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fzu-jwch", "config.json")
	store := Store{Path: path}
	cfg := &Config{
		ID:         "102400000",
		Password:   "secret",
		Identifier: "identifier",
		Cookies: []*http.Cookie{
			{Name: "ASP.NET_SessionId", Value: "session"},
		},
	}

	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("config mode = %v, want 0600", got)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.ID != cfg.ID || loaded.Password != cfg.Password || loaded.Identifier != cfg.Identifier {
		t.Fatalf("loaded config = %#v, want %#v", loaded, cfg)
	}
	if len(loaded.Cookies) != 1 || loaded.Cookies[0].Value != "session" {
		t.Fatalf("loaded cookies = %#v", loaded.Cookies)
	}
}

func TestStoreLoadMissingConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	_, err := (Store{Path: path}).Load()
	if err == nil {
		t.Fatal("Load() error = nil, want ErrConfigNotFound")
	}
	if _, ok := err.(ErrConfigNotFound); !ok {
		t.Fatalf("Load() error = %T, want ErrConfigNotFound", err)
	}
}
