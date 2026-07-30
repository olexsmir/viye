package viye

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/olexsmir/viye/internal/version"
)

type Context struct {
	Path []string // ancestor chain [root, ..., target]
	Dir  string   // working directory

	Cmd  string   // first word of the leaf path component
	Args []string // remaining words of the leaf path component
	Body []string // body lines from : children (edited data for two-phase execution)

	dispatch func(*Context) (string, error) // set by [Viye.dispatch] for tool chaining
}

// Next continues dispatch with currect [Context] state.
// Call after consuming .Path[0] and updating .Dir.
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
		return v.showHelp(out)
	}
	if args[1] == "--version" {
		fmt.Fprint(out, "| viye version: "+version.Version)
		return nil
	}

	path, body := splitArgs(args[1:])
	if len(path) == 0 {
		os.Exit(0)
	}

	cmd, cmdArgs := splitLeaf(path[len(path)-1])
	res, err := v.dispatch(&Context{
		Path: path,
		Dir:  mustGetCwd(),
		Cmd:  cmd,
		Args: cmdArgs,
		Body: body,
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(out, res)
	return err
}

func (v *Viye) showHelp(out io.Writer) error {
	fmt.Fprintln(out, "Available tools:")
	for _, tool := range v.tools {
		fmt.Fprintf(out, "- %s\n", tool.Name())
	}
	return nil
}

// dispatch executes the path against registered tools, first-match-wins
func (v *Viye) dispatch(ctx *Context) (string, error) {
	if len(ctx.Path) == 0 {
		return "", nil
	}
	ctx.dispatch = v.dispatch
	for _, tool := range v.tools {
		if tool.Match(ctx) {
			return tool.Execute(ctx)
		}
	}
	return "", nil
}

// splitLeaf splits a leaf path component into command name and args.
// For "mkdir dir" returns ("mkdir", []string{"dir"}).
// For "$ go run ." returns ("$", []string{"go", "run", "."}).
// For "ip" returns ("ip", nil).
func splitLeaf(leaf string) (cmd string, args []string) {
	parts := strings.Fields(leaf)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

// splitArgs builds path and body from args.
// Everything before "--" is the path (each arg is one component).
// Everything after "--" is the body (kept as-is, one element per line).
// If there's no "--", body is nil.
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
			body = append(body, arg)
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
