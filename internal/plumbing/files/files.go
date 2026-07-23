package files

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
			ls, err := listDir(path)
			if err != nil {
				return "", err
			}
			return viye.FormatOutput(ls), nil
		}
		ctx.Dir = path
		ctx.Path = ctx.Path[1:]
		return ctx.Next()

	// case info.Mode()&0o111 != 0:
	// 	return runExec(path)

	default:
		s, err := head(path, 10)
		if err != nil {
			return "", err
		}
		return s, nil
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

func head(path string, n int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	var b strings.Builder
	for i := 0; i < n && sc.Scan(); i++ {
		fmt.Fprintf(&b, "| %s\n", sc.Text())
	}
	if sc.Scan() {
		b.WriteString("| ...\n")
	}
	return b.String(), sc.Err()
}
