package shell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/olexsmir/viye/internal/viye"
)

type Tool struct{}

func (Tool) Name() string { return "shell($, $$, ls, kill, mkdir)" }
func (Tool) Match(c *viye.Context) bool {
	return isCmd(c) || isBGCmd(c) || isLs(c) || isKill(c) || isMkdir(c)
}

func (Tool) Execute(c *viye.Context) (string, error) {
	switch {
	case isCmd(c):
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "sh", "-c", strings.Join(c.Args, " "))
		cmd.Dir = c.Dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return "", fmt.Errorf("timeout: command exceeded 5s")
			}
		}
		return viye.FormatOutput(string(out)), nil

	case isBGCmd(c):
		cmd := exec.Command("sh", "-c", strings.Join(c.Args, " "))
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Dir = c.Dir
		if err := cmd.Start(); err != nil {
			return "", err
		}
		return fmt.Sprintf("| pid: %d", cmd.Process.Pid), nil

	case isLs(c):
		entries, err := os.ReadDir(c.Dir)
		if err != nil {
			return "", err
		}
		var res []string
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() {
				name += "/"
			}
			res = append(res, name)
		}
		return viye.FormatOutputList(res), nil

	case isKill(c):
		pid, err := strconv.Atoi(c.Args[0])
		if err != nil {
			return "", fmt.Errorf("kill: invalid pid %q", c.Args[0])
		}
		p, err := os.FindProcess(pid)
		if err != nil {
			return "", fmt.Errorf("kill: %w", err)
		}
		if err := p.Kill(); err != nil {
			return "", fmt.Errorf("kill: %w", err)
		}
		return "| done\n", nil

	case isMkdir(c):
		path := filepath.Join(append([]string{c.Dir}, c.Args...)...)
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
func isLs(c *viye.Context) bool    { return c.Cmd == "ls" }
func isKill(c *viye.Context) bool  { return c.Cmd == "kill" && len(c.Args) == 1 }
func isMkdir(c *viye.Context) bool { return c.Cmd == "mkdir" && len(c.Args) >= 1 }
