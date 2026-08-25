package url

import (
	"strings"

	"github.com/olexsmir/viye/internal/osutil"
	"github.com/olexsmir/viye/internal/viye"
)

type Tool struct{}

func (Tool) Name() string { return "url(http://.*, https://.*)" }
func (Tool) Match(c *viye.Context) bool {
	return strings.HasPrefix(c.Cmd, "http://") || strings.HasPrefix(c.Cmd, "https://")
}

func (Tool) Execute(c *viye.Context) (string, error) {
	return "", osutil.Open(c.Cmd)
}
