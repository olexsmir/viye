package gocli

import (
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
		{"go", nil, true},
		{"go", []string{"test"}, true},
		{"go", []string{"test", "./..."}, true},
		{"other", nil, false},
		{"", nil, false},
	}
	for _, tt := range tests {
		got := (Tool{}).Match(&viye.Context{Cmd: tt.cmd, Args: tt.args})
		is.Equal(t, tt.want, got)
	}
}

func TestExecute(t *testing.T) {
	t.Skipf("TODO implement")
}
