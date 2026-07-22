package gobin

import (
	"os/exec"
	"strings"

	"github.com/olexsmir/viye/internal/viye"
)

const menu = `- build
- test
- run
- generate
- doc
- mod/
`

const modMenu = `- tidy
- download
- verify
- init`

type Tool struct{}

func (Tool) Name() string { return "go" }
func (Tool) Match(ctx *viye.Context) bool {
	return ctx.Cmd == "go" || (len(ctx.Path) > 0 && ctx.Path[0] == "go")
}

func (Tool) Execute(ctx *viye.Context) (string, error) {
	var subcmd string
	var args []string

	if ctx.Cmd == "go" {
		// cmd-based: 'go build ./...' → Cmd="go", Args=["build","./..."]
		if len(ctx.Args) == 0 {
			return menu, nil
		}
		subcmd = ctx.Args[0]
		args = ctx.Args[1:]
	} else {
		// path-based: join remaining path as a command line, then re-split
		if len(ctx.Path) < 2 {
			return menu, nil
		}
		parts := strings.Fields(strings.Join(ctx.Path[1:], " "))
		if len(parts) == 0 {
			return menu, nil
		}
		subcmd = parts[0]
		args = parts[1:]
	}

	switch subcmd {
	case "mod":
		if len(args) == 0 {
			return modMenu, nil
		}
	case "build", "test", "generate", "doc", "install", "run", "fmt", "vet", "clean", "env", "fix", "list", "tool", "version":
	default:
		return menu, nil
	}

	return runGo(ctx.Dir, subcmd, args...)
}

func runGo(dir, subcmd string, args ...string) (string, error) {
	cmd := exec.Command("go", append([]string{subcmd}, args...)...)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return viye.Indent(string(out)), nil
}
