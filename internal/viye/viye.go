package viye

import (
	"fmt"
	"io"
	"os"
	"strings"
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
	if args[1] == "--help" {
		return v.showHelp(out)
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

// splitArgs builds path by splitting each arg on "/" and body on "--".
// Everything before "--" is the path (split into components).
// Everything after "--" is the body (kept as-is, one element per line).
// If there's no "--", body is nil.
// for absolute paths(/...), leading / is preserved as part of first component.
// for home paths(~...), ~ is the first comonents
// urls are kept as single component
func splitArgs(args []string) (path []string, body []string) {
	sawSep := false
	for _, arg := range args {
		if !sawSep && arg == "--" {
			sawSep = true
			continue
		}
		if !sawSep {
			path = append(path, splitArg(arg)...)
		} else {
			body = append(body, arg)
		}
	}
	return
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
		i++
		for ; i < len(parts); i++ {
			if parts[i] != "" {
				result = append(result, parts[i])
			}
		}
		return result
	case strings.HasPrefix(arg, "~"):
		parts := strings.Split(arg, "/")
		var result []string
		for _, p := range parts {
			if p != "" {
				result = append(result, p)
			}
		}
		return result
	case strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://"):
		return []string{arg}
	default:
		s := strings.TrimPrefix(arg, "./")
		if s != arg {
			if s == "" {
				return []string{"."}
			}
			// No "/" left to split on — keep original (e.g. "./..." stays "./...")
			if !strings.Contains(s, "/") {
				return []string{arg}
			}
			return splitArg(s)
		}
		parts := strings.Split(s, "/")
		var result []string
		for _, p := range parts {
			if p != "" {
				result = append(result, p)
			}
		}
		return result
	}
}

func mustGetCwd() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get cwd: %v", err))
	}
	return dir
}
