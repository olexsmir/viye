package shell

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/olexsmir/viye/core"
)

type Tool struct{}

func (Tool) Name() string                 { return "shell($)" }
func (Tool) Match(ctx *core.Context) bool { return isCmd(ctx.Path[0]) || isBGCmd(ctx.Path[0]) }
func (Tool) Execute(ctx *core.Context) (string, error) {
	switch {
	case isCmd(ctx.Path[0]):
		cmd := exec.Command("sh", "-c", strings.TrimPrefix(ctx.Path[0], "$ "))
		cmd.Dir = ctx.Dir
		out, _ := cmd.CombinedOutput()

		lines := strings.Split(string(out), "\n")
		for i, line := range lines {
			if line != "" {
				lines[i] = "| " + line
			}
		}
		return strings.Join(lines, "\n"), nil

	case isBGCmd(ctx.Path[0]):
		cmd := exec.Command("sh", "-c", strings.TrimPrefix(ctx.Path[0], "$$ "))
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Dir = ctx.Dir
		if err := cmd.Start(); err != nil {
			return "", err
		}
		return fmt.Sprintf("| pid: %d", cmd.Process.Pid), nil

	default:
		panic("shouldn't have paniced")
	}
}

func isCmd(s string) bool   { return strings.HasPrefix(s, "$ ") }
func isBGCmd(s string) bool { return strings.HasPrefix(s, "$$ ") }
