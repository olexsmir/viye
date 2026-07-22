package viye

import (
	"testing"
)

func TestSplitLeaf(t *testing.T) {
	tests := []struct {
		leaf, cmd string
		args      []string
	}{
		{"", "", nil},
		{"ip", "ip", nil},
		{"$", "$", nil},
		{"$$", "$$", nil},
		{"mkdir", "mkdir", nil},
		{"$ go run .", "$", []string{"go", "run", "."}},
		{"$$ sleep 10", "$$", []string{"sleep", "10"}},
		{"mkdir dir", "mkdir", []string{"dir"}},
		{"mkdir dir1 dir2", "mkdir", []string{"dir1", "dir2"}},
		{"kill 1234", "kill", []string{"1234"}},
		{"  spaced  ", "spaced", nil},
		{"mkdir  spaced  dir", "mkdir", []string{"spaced", "dir"}},
		{"   ", "", nil},
	}
	for _, tt := range tests {
		cmd, args := splitLeaf(tt.leaf)
		if cmd != tt.cmd {
			t.Errorf("splitLeaf(%q) cmd = %q; want %q", tt.leaf, cmd, tt.cmd)
		}
		if len(args) != len(tt.args) {
			t.Errorf("splitLeaf(%q) len(args) = %d; want %d", tt.leaf, len(args), len(tt.args))
		} else {
			for i := range args {
				if args[i] != tt.args[i] {
					t.Errorf("splitLeaf(%q) args[%d] = %q; want %q", tt.leaf, i, args[i], tt.args[i])
				}
			}
		}
	}
}

func TestSplitArg(t *testing.T) {
	tests := []struct {
		arg  string
		want []string
	}{
		{"foo", []string{"foo"}},
		{"foo/bar", []string{"foo", "bar"}},
		{"foo/bar/", []string{"foo", "bar"}},
		{"a/b/c", []string{"a", "b", "c"}},

		// absolute paths
		{"/", []string{"/"}},
		{"/tmp", []string{"/tmp"}},
		{"/tmp/foo", []string{"/tmp", "foo"}},
		{"/tmp/foo/", []string{"/tmp", "foo"}},
		{"/a/b/c", []string{"/a", "b", "c"}},

		// home paths
		{"~", []string{"~"}},
		{"~/foo", []string{"~", "foo"}},
		{"~/a/b", []string{"~", "a", "b"}},

		// urls
		{"http://example.com", []string{"http://example.com"}},
		{"https://example.com/path", []string{"https://example.com/path"}},
	}
	for _, tt := range tests {
		got := splitArg(tt.arg)
		if len(got) != len(tt.want) {
			t.Errorf("splitArg(%q) = %v; want %v", tt.arg, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitArg(%q) = %v; want %v", tt.arg, got, tt.want)
				break
			}
		}
	}
}

func TestSplitArgs(t *testing.T) {
	got := splitArgs([]string{"/tmp/foo", "bar/baz"})
	want := []string{"/tmp", "foo", "bar", "baz"}
	if len(got) != len(want) {
		t.Errorf("splitArgs = %v; want %v", got, want)
	} else {
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("splitArgs = %v; want %v", got, want)
				break
			}
		}
	}
}
