package json

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/olexsmir/viye/internal/viye"
)

type Tool struct{}

func (Tool) Name() string               { return "json" }
func (Tool) Match(c *viye.Context) bool { return c.Cmd == "json" }
func (Tool) Execute(c *viye.Context) (string, error) {
	if len(c.Body) == 0 {
		return ": name: your name\n" +
			": age: 0\n", nil
	}

	m := make(map[string]string)
	for _, line := range c.Body {
		line = strings.TrimPrefix(line, ": ")
		before, after, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		key := strings.TrimSpace(before)
		val := strings.TrimSpace(after)
		if key != "" {
			m[key] = val
		}
	}

	mashaled, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("demo: %w", err)
	}
	return fmt.Sprintf("| %s\n", string(mashaled)), nil
}
