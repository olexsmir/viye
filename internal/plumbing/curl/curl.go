package curl

import (
	"context"
	"os/exec"
	"strings"

	"github.com/olexsmir/viye/internal/viye"
)

type Tool struct{}

func (Tool) Name() string               { return "curl" }
func (Tool) Match(c *viye.Context) bool { return c.Cmd == "curl" }
func (Tool) Execute(c *viye.Context) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), viye.Timeout)
	defer cancel()

	args := append([]string{"-sSL"}, c.Args...)
	args = append(args, c.Body...)
	cmd := exec.CommandContext(ctx, "curl", args...)
	cmd.Dir = c.Dir

	output, err := cmd.CombinedOutput()
	if err != nil && ctx.Err() == context.DeadlineExceeded {
		return "", viye.ErrTimeout
	}

	out := strings.ReplaceAll(string(output), "\r", "")
	return viye.FormatOutput(out), nil
}
