package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/olexsmir/viye/internal/viye"

	"olexsmir.xyz/x/is"
)

func TestMatch(t *testing.T) {
	for _, tt := range []struct {
		cmd  string
		args []string
		want bool
	}{
		{"", nil, false},
		{"get", []string{"http://example.com"}, true},
		{"post", []string{"https://x.io"}, true},
		{"put", []string{"http://x.io/a"}, true},
		{"patch", []string{"https://x.io/a"}, true},
		{"delete", []string{"http://example.com"}, true},
		{"GET", []string{"http://example.com"}, false}, // case-sensitive
		{"get", nil, false},
		{"post", []string{"not-a-url"}, false},
		{"delete", nil, false},
		{"empty", nil, false},
	} {
		got := (Tool{}).Match(&viye.Context{Cmd: tt.cmd, Args: tt.args})
		is.Equal(t, tt.want, got)
	}
}

func TestExecute(t *testing.T) {
	t.Run("no args returns error", func(t *testing.T) {
		_, err := (Tool{}).Execute(&viye.Context{Cmd: "get", Args: []string{}})
		is.Err(t, err, "provide an url")
	})

	t.Run("invalid url returns error", func(t *testing.T) {
		_, err := (Tool{}).Execute(&viye.Context{Cmd: "get", Args: []string{"not-a-url"}})
		is.Err(t, err, "provide an url")
	})

	t.Run("successful GET request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			is.Equal(t, "GET", r.Method)
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "response body")
		}))
		defer server.Close()

		got, err := (Tool{}).Execute(&viye.Context{
			Cmd:  "get",
			Args: []string{server.URL},
		})
		is.Err(t, err, nil)
		is.Equal(t, "| 200\tresponse body\n", got)
	})

	t.Run("POST request with body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			is.Equal(t, "POST", r.Method)
			body, _ := io.ReadAll(r.Body)
			is.Equal(t, "test body", string(body))
			w.WriteHeader(http.StatusCreated)
			io.WriteString(w, "created")
		}))
		defer server.Close()

		got, err := (Tool{}).Execute(&viye.Context{
			Cmd:  "post",
			Args: []string{server.URL},
			Body: []string{"test body"},
		})
		is.Err(t, err, nil)
		is.Equal(t, "| 201\tcreated\n", got)
	})

	t.Run("request with headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			is.Equal(t, "application/json", r.Header.Get("Accept"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		_, err := (Tool{}).Execute(&viye.Context{
			Cmd:  "get",
			Args: []string{server.URL},
			Body: []string{"Accept: application/json"},
		})
		is.Err(t, err, nil)
	})

	t.Run("nonexistent server returns error", func(t *testing.T) {
		_, err := (Tool{}).Execute(&viye.Context{
			Cmd:  "get",
			Args: []string{"http://localhost:1"},
		})
		is.Err(t, err, "failed to make a request:")
	})
}

func TestParseBody(t *testing.T) {
	tests := map[string]struct {
		inp         []string
		wantHeaders map[string]string
		wantBody    string
	}{
		"empty input": {
			inp:         []string{},
			wantHeaders: nil,
			wantBody:    "",
		},
		"headers only": {
			inp: []string{"Header: Value", "Content-Type: no"},
			wantHeaders: map[string]string{
				"Header":       "Value",
				"Content-Type": "no",
			},
		},
		"headers with body": {
			inp:      []string{"Header: Value", "Content-Type: no", "some input"},
			wantBody: "some input",
			wantHeaders: map[string]string{
				"Header":       "Value",
				"Content-Type": "no",
			},
		},
		"body with colon in value": {
			inp:         []string{"Accept: application/json", "plain text body"},
			wantHeaders: map[string]string{"Accept": "application/json"},
			wantBody:    "plain text body",
		},
	}
	for tname, tt := range tests {
		t.Run(tname, func(t *testing.T) {
			headers, body, err := parseBody(tt.inp)
			is.Err(t, err, nil)
			is.Equal(t, tt.wantHeaders, headers)
			is.Equal(t, tt.wantBody, body)
		})
	}
}
