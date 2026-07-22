package gobin

import "github.com/olexsmir/viye/internal/viye"

const menu = `
- build
- test
- generate
- doc
- intall
- mod
`

type Tool struct{}

func (Tool) Name() string { return "go" }
func (Tool) Match(ctx *viye.Context) bool {
	return ctx.Path[0] == "go"
}

func (Tool) Execute(ctx *viye.Context) (string, error) {
	switch ctx.Path[1] {
	case "build":
		return "| will build", nil
	case "test":
		return "| will test", nil
	case "generate":
		return "| will generate", nil
	case "doc":
		return "| will doc", nil
	case "install":
		return "| will install", nil
	case "mod":
		return "| will mod", nil
	default:
		return menu, nil
	}
}
