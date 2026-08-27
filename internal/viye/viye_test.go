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
		{[]string{"demo", "--", ": name: olex"}, []string{"demo"}, []string{"name: olex"}},
		{[]string{"demo", "--", ": name: olex", ": age: 30"}, []string{"demo"}, []string{"name: olex", "age: 30"}},
		{[]string{"--", ": body"}, nil, []string{"body"}},
		{[]string{"demo", "--", "raw body"}, []string{"demo"}, []string{"raw body"}},
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
		path     []string
		cmd      string
		wantArgs []string
	}{
		{[]string{"get", "http://site.com"}, "get", []string{"http://site.com"}},
		{[]string{"$", "go", "run", "."}, "$", []string{"go", "run", "."}},
		{[]string{"ip"}, "ip", []string{}},
	} {
		gotCmd, gotArgs := splitLeaf(tt.path)
		is.Equal(t, tt.cmd, gotCmd)
		is.Equal(t, tt.wantArgs, gotArgs)
	}
}

func TestExpandEnv(t *testing.T) {
	t.Setenv("VIYE_TEST_ONE", "alpha")
	for _, tt := range []struct{ in, want string }{
		{"plain", "plain"},
		{"$VIYE_TEST_ONE", "alpha"},
		{"${VIYE_TEST_ONE}/x", "alpha/x"},
		{"$VIYE_TEST_UNSET", ""},
		{"$$", "$$"},
	} {
		is.Equal(t, tt.want, expandEnv(tt.in))
	}
}

func TestDispatch(t *testing.T) {
	for _, tt := range []struct {
		path     []string
		wantCmd  string
		wantArgs []string
	}{
		{[]string{"get", "http://site.com"}, "get", []string{"http://site.com"}},
		{[]string{"ip"}, "ip", []string{}},
		{[]string{"$", "go", "run", "."}, "$", []string{"go", "run", "."}},
		{[]string{"$$", "sleep", "5"}, "$$", []string{"sleep", "5"}},
	} {
		tool := &mockTool{cmd: tt.wantCmd}
		v := &Viye{tools: []Tool{tool}}
		ctx := &Context{Path: tt.path}
		v.dispatch(ctx)
		is.Equal(t, tt.wantCmd, ctx.Cmd)
		is.Equal(t, tt.wantArgs, ctx.Args)
	}
}

type mockTool struct {
	name, cmd string
	args      []string
}

func (m *mockTool) Name() string          { return m.name }
func (m *mockTool) Match(c *Context) bool { return c.Cmd == m.cmd }
func (m *mockTool) Execute(c *Context) (string, error) {
	m.args = c.Args
	return c.Cmd, nil
}
