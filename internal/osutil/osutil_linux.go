//go:build linux

package osutil

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func Open(path string) error {
	cmd := exec.Command("xdg-open", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Start()
	return nil
}

func GetKernel() (string, error) {
	// or read from: /proc/sys/kernel/osrelease
	k, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return "?", nil
	}
	return strings.TrimSpace(string(k)), nil
}

func GetMemory() (string, error) {
	f, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return "", err
	}
	var totalKB, availKB int64
	for line := range strings.SplitSeq(string(f), "\n") {
		switch {
		case strings.HasPrefix(line, "MemTotal:"):
			p := strings.Fields(line)
			if len(p) >= 2 {
				totalKB, _ = strconv.ParseInt(p[1], 10, 64)
			}
		case strings.HasPrefix(line, "MemAvailable:"):
			lp := strings.Fields(line)
			if len(lp) >= 2 {
				availKB, _ = strconv.ParseInt(lp[1], 10, 64)
			}
		}
	}
	if totalKB == 0 {
		return "", fmt.Errorf("no MemTotal in /proc/meminfo")
	}
	used := float64(totalKB-availKB) / (1024 * 1024)
	total := float64(totalKB) / (1024 * 1024)
	return fmt.Sprintf("%.1f/%.1f GiB", used, total), nil
}

func GetCPU() (string, error) {
	f, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(string(f), "\n") {
		if !strings.HasPrefix(line, "model name") {
			continue
		}
		_, after, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		return fmt.Sprintf("%s (%d)", strings.TrimSpace(after), runtime.NumCPU()), nil
	}
	return "", fmt.Errorf("no model name in /proc/cpuinfo")
}

func GetUptime() (string, error) {
	f, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "", err
	}
	p := strings.Fields(string(f))
	if len(p) == 0 {
		return "", fmt.Errorf("empty /proc/uptime")
	}
	s, err := strconv.ParseFloat(p[0], 64)
	if err != nil {
		return "", err
	}

	d := time.Duration(s) * time.Second
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours), nil
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, mins), nil
	default:
		return fmt.Sprintf("%dm", mins), nil
	}
}

func GetOS() (string, error) {
	f, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(string(f), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			p := strings.Split(line, "=")
			if len(p) != 2 {
				return "Linux", err
			}
			return strings.Trim(p[1], `" `), nil
		}
	}
	return "Linux", nil
}
