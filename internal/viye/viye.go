package viye

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const (
	Version = "0.0.1-alpha"
	Timeout = 5 * time.Second
)

var (
	ErrToolNotFound = errors.New("tool not found")
	ErrTimeout      = fmt.Errorf("timeout: command exceeded %s", Timeout)
)

type Context struct {
	Dir  string   // working directory
	Path []string // ancestor chain [root, ..., target]
	Cmd  string
	Args []string
	Body []string // body lines, ": " prefix stripped

	dispatch func(*Context) (string, error) // set by [Viye.dispatch] for tool chaining
}

// Next continues dispatch with currect [Context] state.
func (c *Context) Next() (string, error) { return c.dispatch(c) }

type Tool interface {
	Name() string
	Match(*Context) bool
	Execute(*Context) (string, error)
}

type Viye struct {
	tools []Tool
}

func New() *Viye { return &Viye{} }

func (v *Viye) Register(tool Tool) {
	v.tools = append(v.tools, tool)
}

func (v *Viye) Run(out io.Writer, args []string) error {
	if len(args) <= 1 {
		return nil
	}
	if args[1] == "help" {
		fmt.Fprintln(out, "| Available tools:")
		for _, tool := range v.tools {
			fmt.Fprintf(out, "|   %s\n", tool.Name())
		}
		return nil
	}
	if args[1] == "version" {
		fmt.Fprint(out, "| viye version: "+Version)
		return nil
	}

	path, body := splitArgs(args[1:])
	if len(path) == 0 {
		os.Exit(0)
	}

	res, err := v.dispatch(&Context{
		Path: path,
		Dir:  mustGetCwd(),
		Body: body,
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(out, res)
	return err
}

func (v *Viye) dispatch(ctx *Context) (string, error) {
	if len(ctx.Path) == 0 {
		return "", nil
	}
	for i := range ctx.Path {
		ctx.Path[i] = expandEnv(ctx.Path[i])
	}
	ctx.Cmd, ctx.Args = splitLeaf(ctx.Path)
	ctx.dispatch = v.dispatch
	for _, tool := range v.tools {
		if tool.Match(ctx) {
			return tool.Execute(ctx)
		}
	}
	return "", ErrToolNotFound
}

func expandEnv(s string) string {
	if s == "$$" {
		return s
	}
	return os.ExpandEnv(s)
}

// splitLeaf splits a path into command name and args.
// ["get", "http://site.com"] → ("get", ["http://site.com"]).
func splitLeaf(path []string) (cmd string, args []string) {
	if len(path) == 0 {
		return "", nil
	}
	return path[0], path[1:]
}

func splitArgs(args []string) (path []string, body []string) {
	sawSep := false
	for _, arg := range args {
		if !sawSep && arg == "--" {
			sawSep = true
			continue
		}
		if !sawSep {
			path = append(path, arg)
		} else {
			body = append(body, strings.TrimPrefix(arg, ": "))
		}
	}
	return
}

func mustGetCwd() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get cwd: %v", err))
	}
	return dir
}
