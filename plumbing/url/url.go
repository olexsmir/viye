package url

import (
	"strings"

	"github.com/olexsmir/viye/core"
	"github.com/olexsmir/viye/core/osutil"
)

type Tool struct{}

func (Tool) Name() string { return "url(http)" }
func (Tool) Match(ctx *core.Context) bool {
	return strings.HasPrefix(ctx.Path[0], "http://") || strings.HasPrefix(ctx.Path[0], "https://")
}

func (Tool) Execute(ctx *core.Context) (string, error) {
	return "", osutil.Open(ctx.Path[0])
}
