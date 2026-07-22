package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/olexsmir/viye/internal/viye"
)

type Tool struct{}

func (Tool) Name() string { return "shell($, $$, kill, mkdir)" }
func (Tool) Match(ctx *viye.Context) bool {
	return isCmd(ctx) || isBGCmd(ctx) || isKill(ctx) || isMkdir(ctx)
}

func (Tool) Execute(ctx *viye.Context) (string, error) {
	switch {
	case isCmd(ctx):
		cmd := exec.Command("sh", "-c", strings.Join(ctx.Args, " "))
		cmd.Dir = ctx.Dir
		out, _ := cmd.CombinedOutput()
		return viye.Indent(string(out)), nil

	case isBGCmd(ctx):
		cmd := exec.Command("sh", "-c", strings.Join(ctx.Args, " "))
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Dir = ctx.Dir
		if err := cmd.Start(); err != nil {
			return "", err
		}
		return fmt.Sprintf("| pid: %d", cmd.Process.Pid), nil

	case isKill(ctx):
		pid, err := strconv.Atoi(ctx.Args[0])
		if err != nil {
			return "", fmt.Errorf("kill: invalid pid %q", ctx.Args[0])
		}
		p, err := os.FindProcess(pid)
		if err != nil {
			return "", fmt.Errorf("kill: %w", err)
		}
		if err := p.Kill(); err != nil {
			return "", fmt.Errorf("kill: %w", err)
		}
		return "| done\n", nil

	case isMkdir(ctx):
		path := filepath.Join(append([]string{ctx.Dir}, ctx.Args...)...)
		if err := os.MkdirAll(path, 0o755); err != nil {
			return "", fmt.Errorf("mkdir: %w", err)
		}
		return "| created " + path + "\n", nil

	default:
		panic("shouldn't have paniced")
	}
}

func isCmd(c *viye.Context) bool   { return c.Cmd == "$" }
func isBGCmd(c *viye.Context) bool { return c.Cmd == "$$" }
func isKill(c *viye.Context) bool  { return c.Cmd == "kill" && len(c.Args) == 1 }
func isMkdir(c *viye.Context) bool { return c.Cmd == "mkdir" && len(c.Args) >= 1 }
