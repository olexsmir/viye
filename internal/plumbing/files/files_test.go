package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olexsmir/viye/internal/viye"

	"olexsmir.xyz/x/is"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		p    string
		want bool
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
		is.Equal(t, tt.want, got)
	}
}

func TestExecute(t *testing.T) {
	t.Run("list dir", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "x"), nil, 0o644)
		ctx := &viye.Context{Path: []string{dir}, Dir: "."}
		got, err := (&Tool{}).Execute(ctx)
		is.Err(t, err, nil)
		is.Equal(t, "| x\n", got)
	})

	t.Run("navigate into subdir via full path", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "sub")
		os.Mkdir(sub, 0o755)
		os.WriteFile(filepath.Join(sub, "y"), nil, 0o644)

		v := viye.New()
		v.Register(&Tool{})
		var out strings.Builder
		err := v.Run(&out, []string{"viye", sub})
		is.Err(t, err, nil)
		is.Equal(t, "| y\n", out.String())
	})

	t.Run("non existent path", func(t *testing.T) {
		ctx := &viye.Context{Path: []string{"/nonexistent_foobar"}, Dir: "."}
		_, err := (&Tool{}).Execute(ctx)
		is.NotEqual(t, nil, err)
	})
}

func TestResolve(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct{ p, dir, want string }{
		{"~", ".", home},
		{"~/foo", ".", filepath.Join(home, "foo")},
		{"/abs/path", ".", "/abs/path"},
		{"rel", "/base", "/base/rel"},
		{".", "/base", "/base"},
		{"..", "/base/sub", "/base"},
	}
	for _, tt := range tests {
		got := resolve(tt.p, tt.dir)
		is.Equal(t, tt.want, got)
	}
}

func TestListDir(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "sub"), 0o755)
	os.WriteFile(filepath.Join(dir, "a.txt"), nil, 0o644)

	got, err := listDir(dir)
	is.Err(t, err, nil)
	is.Equal(t, "a.txt\nsub/\n", got)
}
