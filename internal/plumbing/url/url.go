package url

import (
	"strings"

	"github.com/olexsmir/viye/internal/osutil"
	"github.com/olexsmir/viye/internal/viye"
)

type Tool struct{}

func (Tool) Name() string { return "url(http)" }
func (Tool) Match(ctx *viye.Context) bool {
	return strings.HasPrefix(ctx.Path[0], "http://") || strings.HasPrefix(ctx.Path[0], "https://")
}

func (Tool) Execute(ctx *viye.Context) (string, error) {
	return "", osutil.Open(ctx.Path[0])
}
