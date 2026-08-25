package makefile

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/olexsmir/viye/internal/viye"
)

var ErrMakefileNotFound = errors.New("makefile not found")

type Tool struct{}

func (Tool) Name() string               { return "make" }
func (Tool) Match(c *viye.Context) bool { return c.Cmd == "make" }
func (Tool) Execute(c *viye.Context) (string, error) {
	fpath, found := findMakefile(c)
	if !found {
		return "", ErrMakefileNotFound
	}

	contents, err := os.ReadFile(fpath)
	if err != nil {
		return "", err
	}

	tasks, err := listTasksFromMakefile(contents)
	if err != nil {
		return "", err
	}

	switch len(c.Args) {
	case 0: // list make file tasks
		return viye.FormatBulletList(tasks), nil

	case 1: // run specified task
		if !slices.Contains(tasks, c.Args[0]) {
			return "", ErrTaskNotFound
		}
		cmd := exec.Command("make", c.Args[0])
		cmd.Dir = c.Dir
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return viye.FormatOutput(string(out)), nil

	default:
		return "", errors.New("invalid command, make usage: make [task]")
	}
}

var ErrTaskNotFound = errors.New("task not found")

func findMakefile(c *viye.Context) (name string, found bool) {
	for _, mfile := range []string{"makefile", "Makefile", "GNUMakefile"} {
		path := filepath.Join(c.Dir, mfile)
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

func listTasksFromMakefile(contents []byte) ([]string, error) {
	var tasks []string
	scanner := bufio.NewScanner(bytes.NewReader(contents))
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || line[0] == '\t' || line[0] == ' ' {
			continue
		}
		if strings.HasPrefix(line, "#") { // comments
			continue
		}

		colon := strings.IndexByte(line, ':')
		if colon <= 0 {
			continue
		}

		name := strings.TrimSpace(line[:colon])
		if name == "" {
			continue
		}

		rest := line[colon+1:]
		if strings.HasPrefix(rest, "=") { // variable assigments
			continue
		}
		if strings.HasPrefix(rest, ":") { // ::= or double-colon rule
			rest = rest[1:]
			if strings.HasPrefix(rest, "=") {
				continue
			}
		}

		if strings.Contains(name, "%") || strings.HasPrefix(name, ".") { // pattern rules and special targets
			continue
		}

		for n := range strings.FieldsSeq(name) {
			tasks = append(tasks, n)
		}
	}
	return tasks, nil
}
