package main

import (
	"fmt"
	"os"

	"github.com/olexsmir/viye/internal/plumbing/config"
	"github.com/olexsmir/viye/internal/plumbing/curl"
	"github.com/olexsmir/viye/internal/plumbing/files"
	"github.com/olexsmir/viye/internal/plumbing/gobin"
	"github.com/olexsmir/viye/internal/plumbing/http"
	"github.com/olexsmir/viye/internal/plumbing/ip"
	"github.com/olexsmir/viye/internal/plumbing/json"
	"github.com/olexsmir/viye/internal/plumbing/makefile"
	"github.com/olexsmir/viye/internal/plumbing/neofetch"
	"github.com/olexsmir/viye/internal/plumbing/shell"
	"github.com/olexsmir/viye/internal/plumbing/tldr"
	"github.com/olexsmir/viye/internal/plumbing/url"
	"github.com/olexsmir/viye/internal/plumbing/weather"
	"github.com/olexsmir/viye/internal/viye"
)

func main() {
	v := viye.New()
	v.Register(&files.Tool{})
	v.Register(&shell.Tool{})
	v.Register(&http.Tool{})
	v.Register(&url.Tool{})
	v.Register(&ip.Tool{})
	v.Register(&gobin.Tool{})
	v.Register(&json.Tool{})
	v.Register(&config.Tool{})
	v.Register(&neofetch.Tool{})
	v.Register(&makefile.Tool{})
	v.Register(&curl.Tool{})
	v.Register(&tldr.Tool{})
	v.Register(&weather.Tool{})

	if err := v.Run(os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "viye: %v\n", err)
		os.Exit(1)
	}
}
