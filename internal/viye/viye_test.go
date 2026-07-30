package viye

import (
	"testing"

	"olexsmir.xyz/x/is"
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
		is.Equal(t, tt.cmd, cmd)
		is.Equal(t, len(tt.args), len(args))
		if len(args) == len(tt.args) {
			for i := range args {
				is.Equal(t, tt.args[i], args[i])
			}
		}
	}
}

func TestSplitArgs(t *testing.T) {
	tests := []struct{ args, want, wantB []string }{
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
		is.Equal(t, len(tt.want), len(got))
		if len(got) == len(tt.want) {
			for i := range got {
				is.Equal(t, tt.want[i], got[i])
			}
		}
		is.Equal(t, len(tt.wantB), len(gotB))
		if len(gotB) == len(tt.wantB) {
			for i := range gotB {
				is.Equal(t, tt.wantB[i], gotB[i])
			}
		}
	}
}
