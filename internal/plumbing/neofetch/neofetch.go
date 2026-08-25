package neofetch

import (
	"fmt"
	"strings"

	"github.com/olexsmir/viye/internal/osutil"
	"github.com/olexsmir/viye/internal/viye"
)

type Tool struct{}

func (Tool) Name() string               { return "neofetch" }
func (Tool) Match(c *viye.Context) bool { return c.Cmd == "neofetch" }
func (Tool) Execute(c *viye.Context) (string, error) {
	kernel, _ := osutil.GetKernel()
	osname, _ := osutil.GetOS()
	cpu, _ := osutil.GetCPU()
	mem, _ := osutil.GetMemory()
	uptime, _ := osutil.GetUptime()

	var buf strings.Builder
	_, _ = fmt.Fprintf(&buf, "| OS: %s (%s)\n", osname, kernel)
	_, _ = fmt.Fprintf(&buf, "| SYS: %s · %s\n", cpu, mem)
	_, _ = fmt.Fprintf(&buf, "| Uptime: %s\n", uptime)
	_, _ = fmt.Fprintf(&buf, "| Viye: %s\n", viye.Version)
	return buf.String(), nil
}
