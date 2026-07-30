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

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		args  []string
		want  []string
		wantB []string
	}{
		{[]string{"/tmp/foo", "bar/baz"}, []string{"/tmp/foo", "bar/baz"}, nil},
		{[]string{"./foo", "bar"}, []string{"./foo", "bar"}, nil},
		{[]string{"./"}, []string{"./"}, nil},
		{[]string{"././foo", "./bar"}, []string{"././foo", "./bar"}, nil},
		{[]string{"demo"}, []string{"demo"}, nil},
		{[]string{"demo", "--", ": name: olex"}, []string{"demo"}, []string{": name: olex"}},
		{[]string{"demo", "--", ": name: olex", ": age: 30"}, []string{"demo"}, []string{": name: olex", ": age: 30"}},
		{[]string{"--", ": body"}, nil, []string{": body"}},
	}
	for _, tt := range tests {
		got, gotB := splitArgs(tt.args)
		if len(got) != len(tt.want) {
			t.Errorf("splitArgs(%v) path = %v; want %v", tt.args, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitArgs(%v) path = %v; want %v", tt.args, got, tt.want)
				break
			}
		}
		if len(gotB) != len(tt.wantB) {
			t.Errorf("splitArgs(%v) body = %v; want %v", tt.args, gotB, tt.wantB)
			continue
		}
		for i := range gotB {
			if gotB[i] != tt.wantB[i] {
				t.Errorf("splitArgs(%v) body = %v; want %v", tt.args, gotB, tt.wantB)
				break
			}
		}
	}
}
