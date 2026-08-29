package config

import (
	"bytes"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	_ "embed"
)

var (
	ErrConfigFileNotFound = errors.New("config file not found")
	ErrInvalidConfig      = errors.New("config file is invalid")
)

//go:embed default.conf
var defaultConfig []byte

type Config struct {
	cfgPath string
	cfg     map[string]string
}

func Load() (*Config, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	cfgPath := filepath.Join(cfgDir, "viye.conf")
	c := Config{cfgPath: cfgPath}

	// write default config
	if _, serr := os.Stat(cfgPath); serr != nil {
		if !os.IsNotExist(serr) {
			return nil, serr
		}
		if cerr := os.WriteFile(cfgPath, defaultConfig, 0o644); cerr != nil {
			return nil, cerr
		}
		c.cfg = make(map[string]string)
		if cerr := c.read(defaultConfig); cerr != nil {
			return nil, cerr
		}
		return &c, nil
	}

	file, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	c.cfg = make(map[string]string)
	if err := c.read(file); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Config) GetAll() map[string]string { return c.cfg }
func (c *Config) Get(key string) (val string, ok bool) {
	val, ok = c.cfg[key]
	return val, ok
}

func (c *Config) Set(key, value string) error {
	c.cfg[key] = value
	return c.save()
}

func (c *Config) Delete(key string) error {
	delete(c.cfg, key)
	return c.save()
}

func (c *Config) read(inp []byte) error {
	for line := range strings.SplitSeq(string(inp), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, " = ")
		if len(parts) < 2 || len(parts) > 2 {
			return ErrInvalidConfig // TODO: improve error reporting
		}
		key := strings.TrimSpace(parts[0])
		if strings.Contains(key, " ") {
			return ErrInvalidConfig
		}
		c.cfg[key] = strings.TrimSpace(parts[1])
	}
	return nil
}

func (c *Config) save() error {
	var buf bytes.Buffer
	for _, key := range slices.Sorted(maps.Keys(c.cfg)) {
		val := c.cfg[key]
		_, _ = buf.WriteString(key)
		_, _ = buf.WriteString(" = ")
		_, _ = buf.WriteString(val)
		_ = buf.WriteByte('\n')
	}
	err := os.WriteFile(c.cfgPath, buf.Bytes(), 0o644)
	return err
}
