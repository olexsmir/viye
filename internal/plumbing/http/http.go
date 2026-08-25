package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/olexsmir/viye/internal/viye"
)

// get http://localhost
//	: Header: Value
//	: Header2: value
//	: {"json": "body"}

const timeout = 5 * time.Second

type Tool struct{}

func (Tool) Name() string { return "http(get, post, put, patch, delete)" }
func (Tool) Match(c *viye.Context) bool {
	return (c.Cmd == "get" || c.Cmd == "post" || c.Cmd == "put" || c.Cmd == "patch" || c.Cmd == "delete")
}

func (Tool) Execute(c *viye.Context) (string, error) {
	if len(c.Args) < 1 || !isURL(c.Args[0]) {
		return "", errors.New("please provide an url")
	}

	headers, body, err := parseBody(viye.ParseData(c.Body))
	if err != nil {
		return "", fmt.Errorf("body parsing error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, cmdToMethod(c.Cmd), c.Args[0], strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	for header, value := range headers {
		req.Header.Set(header, value)
	}

	resp, err := (&http.Client{Timeout: timeout}).Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make a request: %v", err)
	}

	respBody, _ := io.ReadAll(resp.Body)
	out := fmt.Sprintf("%d\t%s", resp.StatusCode, respBody)
	return viye.FormatOutput(out), nil
}

func parseBody(inp []string) (headers map[string]string, body string, err error) {
	if len(inp) == 0 {
		return nil, "", nil
	}
	headers = make(map[string]string)

	for _, line := range inp {
		key, val, ok := strings.Cut(line, ": ")
		if ok {
			headers[key] = val
		}
	}

	last := inp[len(inp)-1]
	if _, _, ok := strings.Cut(last, ": "); !ok {
		body = last
	}

	return headers, body, nil
}

func isURL(u string) bool {
	return strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")
}

func cmdToMethod(c string) string {
	switch c {
	case "get":
		return http.MethodGet
	case "post":
		return http.MethodPost
	case "put":
		return http.MethodPut
	case "patch":
		return http.MethodPatch
	case "delete":
		return http.MethodDelete
	default:
		panic("unreachable")
	}
}
