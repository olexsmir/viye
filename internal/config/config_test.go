package config

import (
	"os"
	"path/filepath"
	"testing"

	"olexsmir.xyz/x/is"
)

func TestLoad(t *testing.T) {
	t.Run("creates default config when file missing", func(t *testing.T) {
		want := Config{cfg: make(map[string]string)}
		is.Err(t, want.read(defaultConfig), nil)

		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		cfg, err := Load()
		is.Err(t, err, nil)
		for key, val := range want.cfg {
			is.Equal(t, val, cfg.Get(key))
		}
	})

	t.Run("loads existing config file", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		cfgDir := filepath.Join(dir, "viye.conf")
		is.Err(t, os.WriteFile(cfgDir, []byte("custom = value\n"), 0o644), nil)

		cfg, err := Load()
		is.Err(t, err, nil)
		is.Equal(t, "value", cfg.Get("custom"))
	})

	t.Run("invalid config file returns error", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		cfgDir := filepath.Join(dir, "viye.conf")
		is.Err(t, os.WriteFile(cfgDir, []byte("bad line\n"), 0o644), nil)

		_, err := Load()
		is.Err(t, err, ErrInvalidConfig)
	})
}

func TestSet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "conf")
	c := Config{
		cfgPath: path,
		cfg:     map[string]string{},
	}

	is.Err(t, c.Set("key", "val"), nil)
	is.Equal(t, "val", c.cfg["key"])

	b, err := os.ReadFile(path)
	is.Err(t, err, nil)
	is.Equal(t, "key = val\n", string(b))
}

func TestRead(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		c := Config{cfg: make(map[string]string)}
		is.Err(t, c.read([]byte("key1 = value1\nkey2 = value2\n")), nil)
		is.Equal(t, "value1", c.cfg["key1"])
		is.Equal(t, "value2", c.cfg["key2"])
	})

	t.Run("trailing newline", func(t *testing.T) {
		c := Config{cfg: make(map[string]string)}
		is.Err(t, c.read([]byte("key = val\n")), nil)
		is.Equal(t, "val", c.cfg["key"])
	})

	t.Run("empty lines", func(t *testing.T) {
		c := Config{cfg: make(map[string]string)}
		is.Err(t, c.read([]byte("\nkey = val\n\n")), nil)
		is.Equal(t, "val", c.cfg["key"])
	})

	t.Run("multiple equals signs", func(t *testing.T) {
		c := Config{cfg: make(map[string]string)}
		is.Err(t, c.read([]byte("key = val = extra")), ErrInvalidConfig)
	})

	t.Run("no separator", func(t *testing.T) {
		c := Config{cfg: make(map[string]string)}
		is.Err(t, c.read([]byte("justtext")), ErrInvalidConfig)
	})
}
