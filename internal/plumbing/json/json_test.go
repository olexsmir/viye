package json

import (
	"testing"

	"github.com/olexsmir/viye/internal/viye"

	"olexsmir.xyz/x/is"
)

func TestMatch(t *testing.T) {
	got := (Tool{}).Match(&viye.Context{Cmd: "json"})
	is.Equal(t, true, got)
}

func TestExecute(t *testing.T) {
	tests := map[string]struct {
		body []string
		want string
	}{
		"no body":          {want: ": name: your name\n: age: 0\n"},
		"empty body slice": {want: ": name: your name\n: age: 0\n"},
		"single key": {
			body: []string{"name: olex"},
			want: `| {"name":"olex"}` + "\n",
		},
		"multiple keys": {
			body: []string{"name: olex", "age: 30"},
			want: `| {"age":"30","name":"olex"}` + "\n",
		},
		"values with spaces": {
			body: []string{"title: hello world"},
			want: `| {"title":"hello world"}` + "\n",
		},
		"keys with spaces": {
			body: []string{"my key: value"},
			want: `| {"my key":"value"}` + "\n",
		},
		"line without colon separator": {
			body: []string{"justtext"},
			want: `| {}` + "\n",
		},
		"line without : prefix": {
			body: []string{"name: olex"},
			want: `| {"name":"olex"}` + "\n",
		},
		"mixed valid and invalid": {
			body: []string{"name: olex", "garbage", "age: 30"},
			want: `| {"age":"30","name":"olex"}` + "\n",
		},
		"empty key": {
			body: []string{": value"},
			want: `| {}` + "\n",
		},
		"trailing whitespace in key": {
			body: []string{"name  : olex"},
			want: `| {"name":"olex"}` + "\n",
		},
		"trailing whitespace in value": {
			body: []string{"name: olex  "},
			want: `| {"name":"olex"}` + "\n",
		},
		"empty value": {
			body: []string{"name: "},
			want: `| {"name":""}` + "\n",
		},
	}
	for tname, tt := range tests {
		t.Run(tname, func(t *testing.T) {
			got, err := (Tool{}).Execute(&viye.Context{Cmd: "json", Body: tt.body})
			is.Err(t, err, nil)
			is.Equal(t, tt.want, got)
		})
	}
}
