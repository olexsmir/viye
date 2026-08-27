package curl

import (
	"os/exec"
	"testing"

	"github.com/olexsmir/viye/internal/viye"

	"olexsmir.xyz/x/is"
)

func TestMatch(t *testing.T) {
	is.Equal(t, true, (Tool{}).Match(&viye.Context{Cmd: "curl"}))
	is.Equal(t, false, (Tool{}).Match(&viye.Context{Cmd: "get"}))
}

func TestExecute(t *testing.T) {
	if _, err := exec.LookPath("curl"); err != nil {
		t.Skip("curl not installed")
	}

	t.Run("version", func(t *testing.T) {
		got, err := (Tool{}).Execute(&viye.Context{
			Cmd:  "curl",
			Args: []string{"--version"},
			Dir:  ".",
		})
		is.Err(t, err, nil)
		is.NotEqual(t, "", got)
	})

	t.Run("extra tags from body", func(t *testing.T) {
		got, err := (Tool{}).Execute(&viye.Context{
			Cmd:  "curl",
			Args: []string{"--version"},
			Body: []string{": --help"},
			Dir:  ".",
		})
		is.Err(t, err, nil)
		is.NotEqual(t, "", got)
	})
}
