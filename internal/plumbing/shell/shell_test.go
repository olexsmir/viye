package shell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/olexsmir/viye/internal/viye"

	"olexsmir.xyz/x/is"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		cmd  string
		args []string
		want bool
	}{
		{"", nil, false},
		{"$", nil, true},
		{"$", []string{"go", "run", "."}, true},
		{"$$", nil, true},
		{"$$", []string{"sleep", "10"}, true},
		{"kill", []string{"1234"}, true},
		{"kill", []string{"abc"}, true},
		{"kill", nil, false},
		{"kill", []string{"1", "2"}, false},
		{"mkdir", []string{"dir"}, true},
		{"mkdir", []string{"dir1", "dir2"}, true},
		{"mkdir", nil, false},
	}
	for _, tt := range tests {
		got := (&Tool{}).Match(&viye.Context{Cmd: tt.cmd, Args: tt.args})
		is.Equal(t, tt.want, got)
	}
}

func TestExecute(t *testing.T) {
	t.Run("cmd", func(t *testing.T) {
		got, err := (&Tool{}).Execute(&viye.Context{
			Path: []string{"$ echo hello"},
			Cmd:  "$",
			Args: []string{"echo", "hello"},
			Dir:  ".",
		})
		is.Err(t, err, nil)
		is.Equal(t, "| hello\n", got)
	})

	t.Run("bg cmd", func(t *testing.T) {
		got, err := (&Tool{}).Execute(&viye.Context{
			Path: []string{"$$ sleep 0"},
			Cmd:  "$$",
			Args: []string{"sleep", "0"},
			Dir:  ".",
		})
		is.Err(t, err, nil)
		if len(got) < 10 {
			t.Fatalf("got short output %q, expected pid string", got)
		}
	})

	t.Run("mkdir", func(t *testing.T) {
		dir := t.TempDir()
		got, err := (&Tool{}).Execute(&viye.Context{
			Path: []string{"mkdir sub"},
			Cmd:  "mkdir",
			Args: []string{"sub"},
			Dir:  dir,
		})
		is.Err(t, err, nil)

		_, err = os.Stat(filepath.Join(dir, "sub"))
		is.Err(t, err, nil)
		is.Equal(t, "| created "+filepath.Join(dir, "sub")+"\n", got)
	})

	t.Run("kill", func(t *testing.T) {
		_, err := (&Tool{}).Execute(&viye.Context{
			Path: []string{"kill 999999"},
			Cmd:  "kill",
			Args: []string{"999999"},
			Dir:  ".",
		})
		is.NotEqual(t, nil, err)
	})
}
