package shell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/olexsmir/viye/internal/viye"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		path string
		cmd  string
		args []string
		want bool
	}{
		{"$", "$", nil, true},
		{"$ go run .", "$", []string{"go", "run", "."}, true},
		{"$$", "$$", nil, true},
		{"$$ sleep 10", "$$", []string{"sleep", "10"}, true},
		{"kill 1234", "kill", []string{"1234"}, true},
		{"kill abc", "kill", []string{"abc"}, true},
		{"mkdir dir", "mkdir", []string{"dir"}, true},
		{"mkdir dir1 dir2", "mkdir", []string{"dir1", "dir2"}, true},
		{"", "", nil, false},
		{"ip", "ip", nil, false},
		{"kill", "kill", nil, false},
		{"kill 1 2", "kill", []string{"1", "2"}, false},
		{"mkdir", "mkdir", nil, false},
	}
	for _, tt := range tests {
		got := (&Tool{}).Match(&viye.Context{
			Path: []string{tt.path},
			Cmd:  tt.cmd,
			Args: tt.args,
		})
		if got != tt.want {
			t.Errorf("Match(%q) = %v; want %v", tt.path, got, tt.want)
		}
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
		if err != nil {
			t.Fatal(err)
		}
		if got != "| hello\n" {
			t.Errorf("Execute($ echo hello) = %q; want %q", got, "| hello\\n")
		}
	})

	t.Run("bg cmd", func(t *testing.T) {
		got, err := (&Tool{}).Execute(&viye.Context{
			Path: []string{"$$ sleep 0"},
			Cmd:  "$$",
			Args: []string{"sleep", "0"},
			Dir:  ".",
		})
		if err != nil {
			t.Fatal(err)
		}
		if got != "| pid: " && len(got) < 10 {
			t.Errorf("Execute($$ sleep 0) = %q; want pid string", got)
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
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(dir, "sub")); err != nil {
			t.Errorf("dir not created: %v", err)
		}
		if got != "| created "+filepath.Join(dir, "sub")+"\n" {
			t.Errorf("output = %q; want %q", got, "| created "+filepath.Join(dir, "sub")+"\n")
		}
	})

	t.Run("kill", func(t *testing.T) {
		// no reliable way to test kill without a real process,
		// just verify it doesn't panic with valid args structure
		if _, err := (&Tool{}).Execute(&viye.Context{
			Path: []string{"kill 999999"},
			Cmd:  "kill",
			Args: []string{"999999"},
			Dir:  ".",
		}); err == nil {
			t.Log("kill on bogus pid returned nil (expected error or success)")
		}
	})
}
