package ip

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/olexsmir/viye/internal/viye"
)

type Tool struct{}

func (Tool) Name() string                 { return "ip" }
func (Tool) Match(ctx *viye.Context) bool { return len(ctx.Path) == 1 && ctx.Path[0] == "ip" }
func (Tool) Execute(ctx *viye.Context) (string, error) {
	// TODO: if it's not possile to get  public ip, report only local

	var buf strings.Builder
	buf.WriteString("| ")

	pip, err := getPublicIP()
	if err != nil {
		return "", err
	}
	_, _ = buf.WriteString(pip)
	_, _ = buf.WriteString(" / ")

	lip, err := getLocalIP()
	if err != nil {
		return "", err
	}
	_, _ = buf.WriteString(lip)

	_ = buf.WriteByte('\n')
	return buf.String(), nil
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("couldn't get local error")
}

func getPublicIP() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://ifconfig.me/ip", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(ip)), nil
}
