package json

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/olexsmir/viye/internal/viye"
)

type Tool struct{}

func (Tool) Name() string                 { return "json" }
func (Tool) Match(ctx *viye.Context) bool { return ctx.Cmd == "json" }
func (Tool) Execute(ctx *viye.Context) (string, error) {
	if len(ctx.Body) == 0 {
		return ": name: your name\n" +
			": age: 0\n", nil
	}

	m := make(map[string]string)
	for _, line := range ctx.Body {
		line = strings.TrimPrefix(line, ": ")
		col := strings.Index(line, ":")
		if col < 0 {
			continue
		}

		key := strings.TrimSpace(line[:col])
		val := strings.TrimSpace(line[col+1:])
		if key != "" {
			m[key] = val
		}
	}

	b, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("demo: %w", err)
	}
	return fmt.Sprintf("| %s\n", string(b)), nil
}
