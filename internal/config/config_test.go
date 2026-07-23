package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("creates default config when file missing", func(t *testing.T) {
		want := Config{cfg: make(map[string]string)}
		if err := want.read(defaultConfig); err != nil {
			t.Fatal(err)
		}

		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() = _, %v", err)
		}
		for key, val := range want.cfg {
			if cfg.Get(key) != val {
				t.Fatalf("got %s=%q, want %q", key, cfg.Get(key), val)
			}
		}
	})

	t.Run("loads existing config file", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		cfgDir := filepath.Join(dir, "viye.conf")
		if err := os.WriteFile(cfgDir, []byte("custom = value\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() = _, %v", err)
		}
		if cfg.Get("custom") != "value" {
			t.Fatalf("got custom=%q, want %q", cfg.Get("custom"), "value")
		}
	})

	t.Run("invalid config file returns error", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		cfgDir := filepath.Join(dir, "viye.conf")
		if err := os.WriteFile(cfgDir, []byte("bad line\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		_, err := Load()
		if err != ErrInvalidConfig {
			t.Fatalf("got %v, want ErrInvalidConfig", err)
		}
	})
}

func TestSet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "conf")
	c := Config{
		cfgPath: path,
		cfg:     map[string]string{},
	}

	if err := c.Set("key", "val"); err != nil {
		t.Fatal(err)
	}
	if c.cfg["key"] != "val" {
		t.Fatalf("in-memory: got %v, want key=val", c.cfg)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "key = val\n" {
		t.Fatalf("on disk: got %q, want %q", string(b), "key = val\n")
	}
}

func TestRead(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		c := Config{cfg: make(map[string]string)}
		err := c.read([]byte("key1 = value1\nkey2 = value2\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.cfg["key1"] != "value1" || c.cfg["key2"] != "value2" {
			t.Fatalf("got %v, want key1=value1 key2=value2", c.cfg)
		}
	})

	t.Run("trailing newline", func(t *testing.T) {
		c := Config{cfg: make(map[string]string)}
		err := c.read([]byte("key = val\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.cfg["key"] != "val" {
			t.Fatalf("got %v, want key=val", c.cfg)
		}
	})

	t.Run("empty lines", func(t *testing.T) {
		c := Config{cfg: make(map[string]string)}
		err := c.read([]byte("\nkey = val\n\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.cfg["key"] != "val" {
			t.Fatalf("got %v, want key=val", c.cfg)
		}
	})

	t.Run("multiple equals signs", func(t *testing.T) {
		c := Config{cfg: make(map[string]string)}
		err := c.read([]byte("key = val = extra"))
		if err != ErrInvalidConfig {
			t.Fatalf("got %v, want ErrInvalidConfig", err)
		}
	})

	t.Run("no separator", func(t *testing.T) {
		c := Config{cfg: make(map[string]string)}
		err := c.read([]byte("justtext"))
		if err != ErrInvalidConfig {
			t.Fatalf("got %v, want ErrInvalidConfig", err)
		}
	})
}
