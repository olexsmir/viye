package tldr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/olexsmir/viye/internal/viye"
)

type Tool struct{}

func (Tool) Name() string               { return "tldr" }
func (Tool) Match(c *viye.Context) bool { return c.Cmd == "tldr" }
func (Tool) Execute(c *viye.Context) (string, error) {
	if len(c.Args) == 0 {
		return "", fmt.Errorf("usage: tldr <command> [subcommand...]")
	}
	content, err := page(c.Args)
	if err != nil {
		return "", err
	}
	return viye.FormatOutput(render(content)), nil
}

const baseURL = "https://raw.githubusercontent.com/tldr-pages/tldr/main"

func page(args []string) ([]byte, error) {
	cache := filepath.Join(mustUserCacheDir(), "tealdeer", "tldr-pages", "pages.en")
	var remote []string
	for _, dir := range []string{"common", platformDir()} {
		for _, name := range names(args) {
			if b, err := os.ReadFile(filepath.Join(cache, dir, name)); err == nil {
				return b, nil
			}
			remote = append(remote, baseURL+"/pages.en/"+dir+"/"+name)
		}
	}
	for _, url := range remote {
		if body, err := fetch(url); err == nil {
			return body, nil
		} else if !errors.Is(err, errNotFound) {
			return nil, err
		}
	}
	return nil, fmt.Errorf("no tldr found for %q", strings.Join(args, " "))
}

var errNotFound = errors.New("page not found")

func fetch(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), viye.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch tldr page: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func names(args []string) []string {
	n := []string{strings.Join(args, "-") + ".md"}
	if len(args) > 1 {
		n = append(n, args[0]+".md")
	}
	return n
}

// platformDir maps the running OS to its tealdeer page directory.
func platformDir() string {
	switch runtime.GOOS {
	case "darwin":
		return "osx"
	default:
		return runtime.GOOS
	}
}

func mustUserCacheDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get user cache dir: %v", err))
	}
	return dir
}

// render strips tldr markdown to plain text: "# name" -> "name", "> text" ->
// "text", "`code`" -> 2-space-indented code directly under its description.
func render(content []byte) string {
	var out []string
	for line := range strings.SplitSeq(string(content), "\n") {
		switch {
		case strings.HasPrefix(line, "# "):
			out = append(out, strings.TrimPrefix(line, "# "))
		case strings.HasPrefix(line, "> "):
			out = append(out, strings.TrimPrefix(line, "> "))
		case strings.HasPrefix(line, "`") && strings.HasSuffix(line, "`"):
			if out[len(out)-1] == "" { // close the gap between desc and code
				out = out[:len(out)-1]
			}
			out = append(out, "  "+strings.Trim(line, "`"))
		default:
			out = append(out, line)
		}
	}
	for len(out) > 0 && out[len(out)-1] == "" { // drop trailing blank
		out = out[:len(out)-1]
	}
	return strings.Join(out, "\n")
}
