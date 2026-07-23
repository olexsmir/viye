package config

import (
	"fmt"
	"strings"

	"github.com/olexsmir/viye/internal/config"
	"github.com/olexsmir/viye/internal/viye"
)

type Tool struct{}

func (Tool) Name() string               { return "config" }
func (Tool) Match(c *viye.Context) bool { return c.Path[0] == "config" }
func (Tool) Execute(c *viye.Context) (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("couldn't load config")
	}

	if len(c.Body) > 0 {
		seen := make(map[string]bool, len(c.Body))
		for _, line := range c.Body {
			line = strings.TrimPrefix(line, ": ")
			key, val, ok := strings.Cut(line, " = ")
			if !ok {
				continue
			}
			key, val = strings.TrimSpace(key), strings.TrimSpace(val)
			if strings.Contains(key, " ") {
				return "", fmt.Errorf("key %q contains spaces", key)
			}
			seen[key] = true
			if err := cfg.Set(key, val); err != nil {
				return "", err
			}
		}
		for key := range cfg.GetAll() {
			if !seen[key] {
				if err := cfg.Delete(key); err != nil {
					return "", err
				}
			}
		}
	}

	conf := cfg.GetAll()
	maxLen := 0
	for key := range conf {
		if len(key) > maxLen {
			maxLen = len(key)
		}
	}
	var buf strings.Builder
	for key, val := range conf {
		fmt.Fprintf(&buf, ": %-*s = %s\n", maxLen, key, val)
	}

	buf.WriteByte('\n')
	return buf.String(), nil
}
