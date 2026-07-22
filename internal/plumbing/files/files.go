package files

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/olexsmir/viye/internal/osutil"
	"github.com/olexsmir/viye/internal/viye"
)

type Tool struct{}

func (Tool) Name() string { return "files(~/..., /...)" }

func (Tool) Match(ctx *viye.Context) bool {
	p := ctx.Path[0]
	if p == "~" || strings.HasPrefix(p, "/") {
		return true
	}
	_, err := os.Stat(filepath.Join(ctx.Dir, p))
	return err == nil
}

func (Tool) Execute(ctx *viye.Context) (string, error) {
	path := resolve(ctx.Path[0], ctx.Dir)

	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	switch {
	case info.IsDir():
		if len(ctx.Path) == 1 {
			return listDir(path)
		}
		ctx.Dir = path
		ctx.Path = ctx.Path[1:]
		return ctx.Next()

	// case info.Mode()&0o111 != 0:
	// 	return runExec(path)

	default:
		return "", osutil.Open(path)
	}
}

func resolve(p, dir string) string {
	if p == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[2:])
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Join(dir, p)
}

func listDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		b.WriteString(name)
		b.WriteByte('\n')
	}
	return b.String(), nil
}

// func runExec(path string) (string, error) {
// 	cmd := exec.Command(path)
// 	out, err := cmd.Output()
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(out), nil
// }
