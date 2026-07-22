package gobin

import (
	"testing"

	"github.com/olexsmir/viye/internal/viye"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		path []string
		want bool
	}{
		{[]string{"go"}, true},
		{[]string{"go", "build"}, true},
		{[]string{"go", "test", "./..."}, true},
		{[]string{"other"}, false},
		{[]string{}, false},
	}
	for _, tt := range tests {
		got := (Tool{}).Match(&viye.Context{Path: tt.path})
		if got != tt.want {
			t.Errorf("Match(%v) = %v; want %v", tt.path, got, tt.want)
		}
	}
}

func TestExecute(t *testing.T) {
	t.Skipf("TODO implement")
}
