package viye

import (
	"testing"

	"olexsmir.xyz/x/is"
)

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

func TestSplitLeaf(t *testing.T) {
	for _, tt := range []struct {
		leaf, cmd string
		want      []string
	}{
		{"mkdir dir", "mkdir", []string{"dir"}},
		{"$ go run .", "$", []string{"go", "run", "."}},
		{"ip", "ip", []string{}},
	} {
		gotCmd, gotArgs := splitLeaf(tt.leaf)
		is.Equal(t, tt.cmd, gotCmd)
		is.Equal(t, tt.want, gotArgs)
	}
}
