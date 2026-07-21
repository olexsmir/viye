package main

import (
	"fmt"
	"os"

	"github.com/olexsmir/viye/core"
	"github.com/olexsmir/viye/plumbing/files"
	"github.com/olexsmir/viye/plumbing/gobin"
	"github.com/olexsmir/viye/plumbing/ip"
	"github.com/olexsmir/viye/plumbing/shell"
	"github.com/olexsmir/viye/plumbing/url"
)

func main() {
	v := core.New()
	v.Register(&files.Tool{})
	v.Register(&shell.Tool{})
	v.Register(&url.Tool{})
	v.Register(&ip.Tool{})
	v.Register(&gobin.Tool{})

	if err := v.Run(os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "viye: %v\n", err)
		os.Exit(1)
	}
}
