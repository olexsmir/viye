package gobin

import (
	"context"
	"os/exec"

	"github.com/olexsmir/viye/internal/viye"
)

const menu = `- build
- test
- run
- generate
- doc
+ mod/
`

const modMenu = `- tidy
- download
- verify
- init`

type Tool struct{}

func (Tool) Name() string               { return "go" }
func (Tool) Match(c *viye.Context) bool { return c.Cmd == "go" }
func (Tool) Execute(c *viye.Context) (string, error) {
	if len(c.Args) == 0 {
		return menu, nil
	}

	args := c.Args[1:]
	switch c.Args[0] {
	case "mod":
		if len(args) == 0 {
			return modMenu, nil
		}
	case "build", "test", "generate", "doc", "install", "run", "fmt", "vet", "clean", "env", "fix", "list", "tool", "version":
	default:
		return menu, nil
	}

	return runGo(c.Dir, c.Args[0], args...)
}

func runGo(dir, subcmd string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), viye.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", append([]string{subcmd}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil && ctx.Err() == context.DeadlineExceeded {
		return "", viye.ErrTimeout
	}
	return viye.FormatOutput(out), nil
}
