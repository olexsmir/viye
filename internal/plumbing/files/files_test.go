package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olexsmir/viye/internal/viye"
)

func TestMatch(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "sub"), 0o755)
	os.WriteFile(filepath.Join(dir, "f"), nil, 0o644)

	tests := []struct {
		p, dir string
		m      bool
	}{
		{"~", ".", true},
		{"/tmp", ".", true},
		{dir, ".", true},
		{"sub", dir, true},
		{"f", dir, true},
		{"nope", dir, false},
		{"$ ls", dir, false},
		{".", dir, true},
		{"..", dir, true},
	}
	for _, tt := range tests {
		got := (&Tool{}).Match(&viye.Context{Path: []string{tt.p}, Dir: tt.dir})
		if got != tt.m {
			t.Errorf("Match(%q, dir=%q) = %v; want %v", tt.p, tt.dir, got, tt.m)
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
		if got != "x\n" {
			t.Errorf("Execute = %q; want %q", got, "x\\n")
		}
	})

	t.Run("navigate into subdir", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "sub")
		os.Mkdir(sub, 0o755)
		os.WriteFile(filepath.Join(sub, "y"), nil, 0o644)

		v := viye.New()
		v.Register(&Tool{})
		var out strings.Builder
		if err := v.Run(&out, []string{"viye", dir, "sub"}); err != nil {
			t.Fatal(err)
		}
		if out.String() != "y\n" {
			t.Errorf("Run = %q; want %q", out.String(), "y\\n")
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
