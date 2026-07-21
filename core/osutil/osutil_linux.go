//go:build linux

package osutil

import (
	"os/exec"
	"syscall"
)

func Open(path string) error {
	cmd := exec.Command("xdg-open", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Start()
	return nil
}
