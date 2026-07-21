package core

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Context struct {
	Path []string // ancestor chain [root, ..., target]
	Dir  string   // working directory

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
	if args[1] == "--help" {
		return v.showHelp(out)
	}

	path := splitArgs(args[1:])
	if len(path) == 0 {
		os.Exit(0)
	}

	res, err := v.dispatch(&Context{
		Path: path,
		Dir:  ".", // TODO: get cwd
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
		fmt.Fprintf(out, "  - %s\n", tool.Name())
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

// splitArgs builds path by splitting each arg on "/"
// for absolute paths(/...), leading / is preserved as part of first component.
// for home paths(~...), ~ is the first comonents
// urls are kept as single component
func splitArgs(args []string) []string {
	var path []string
	for _, arg := range args {
		path = append(path, splitArg(arg)...)
	}
	return path
}

func splitArg(arg string) []string {
	switch {
	case strings.HasPrefix(arg, "/"):
		parts := strings.Split(arg, "/")
		// parts[0] is always "" for absolute paths
		i := 0
		for i < len(parts) && parts[i] == "" {
			i++
		}
		if i >= len(parts) {
			return []string{"/"}
		}
		// parts[i] is the first non-empty segment → attach it to "/"
		result := []string{"/" + parts[i]}
		result = append(result, parts[i+1:]...)
		return result
	case strings.HasPrefix(arg, "~"):
		return strings.Split(arg, "/")
	case strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://"):
		return []string{arg}
	default:
		return strings.Split(arg, "/")
	}
}
