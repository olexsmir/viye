package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olexsmir/viye/internal/viye"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		p string
		m bool
	}{
		{"~/foo", true},
		{"/tmp", true},
		{"/tmp/foo", true},
		{"./foo", true},
		{"./foo/bar", true},

		{"nope", false},
		{"$ ls", false},
		{"~", false},
		{".", false},
		{"..", false},
	}
	for _, tt := range tests {
		got := (&Tool{}).Match(&viye.Context{Path: []string{tt.p}})
		if got != tt.m {
			t.Errorf("Match(%q) = %v; want %v", tt.p, got, tt.m)
		}
	}
}

func TestExecute(t *testing.T) {
	t.Run("list dir", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "x"), nil, 0o644)
		ctx := &viye.Context{Path: []string{dir}, Dir: "."}
		got, err := (&Tool{}).Execute(ctx)
		if err != nil {
			t.Fatal(err)
		}
		want := "| x\n"
		if got != want {
			t.Errorf("Execute = %q; want %q", got, want)
		}
	})

	t.Run("navigate into subdir via full path", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "sub")
		os.Mkdir(sub, 0o755)
		os.WriteFile(filepath.Join(sub, "y"), nil, 0o644)

		v := viye.New()
		v.Register(&Tool{})
		var out strings.Builder
		if err := v.Run(&out, []string{"viye", sub}); err != nil {
			t.Fatal(err)
		}
		want := "| y\n"
		if out.String() != want {
			t.Errorf("Run = %q; want %q", out.String(), want)
		}
	})

	t.Run("non existent path", func(t *testing.T) {
		ctx := &viye.Context{Path: []string{"/nonexistent_foobar"}, Dir: "."}
		if _, err := (&Tool{}).Execute(ctx); err == nil {
			t.Error("expected error")
		}
	})
}

func TestResolve(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		p, dir string
		want   string
	}{
		{"~", ".", home},
		{"~/foo", ".", filepath.Join(home, "foo")},
		{"/abs/path", ".", "/abs/path"},
		{"rel", "/base", "/base/rel"},
		{".", "/base", "/base"},
		{"..", "/base/sub", "/base"},
	}
	for _, tt := range tests {
		got := resolve(tt.p, tt.dir)
		if got != tt.want {
			t.Errorf("resolve(%q, %q) = %q; want %q", tt.p, tt.dir, got, tt.want)
		}
	}
}

func TestListDir(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "sub"), 0o755)
	os.WriteFile(filepath.Join(dir, "a.txt"), nil, 0o644)

	got, err := listDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "a.txt\nsub/\n" {
		t.Errorf("listDir = %q; want %q", got, "a.txt\\nsub/\\n")
	}
}
