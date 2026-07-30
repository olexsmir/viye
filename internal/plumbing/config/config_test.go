package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/olexsmir/viye/internal/viye"

	"olexsmir.xyz/x/is"
)

func TestMatch(t *testing.T) {
	got := (Tool{}).Match(&viye.Context{Cmd: "config"})
	is.Equal(t, true, got)
}

func TestExecute(t *testing.T) {
	t.Run("no body empty config", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		got, err := (Tool{}).Execute(&viye.Context{Path: []string{"config"}})
		is.Err(t, err, nil)
		is.Equal(t, "\n", got)
	})

	t.Run("no body with existing key", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		cfgPath := filepath.Join(dir, "viye.conf")
		os.WriteFile(cfgPath, []byte("k = v\n"), 0o644)

		got, err := (Tool{}).Execute(&viye.Context{Path: []string{"config"}})
		is.Err(t, err, nil)
		is.Equal(t, ": k = v\n\n", got)
	})

	t.Run("set key", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		got, err := (Tool{}).Execute(&viye.Context{
			Path: []string{"config"},
			Body: []string{": name = olex"},
		})
		is.Err(t, err, nil)
		is.Equal(t, ": name = olex\n\n", got)
	})

	t.Run("update and delete", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		cfgPath := filepath.Join(dir, "viye.conf")
		os.WriteFile(cfgPath, []byte("a = 1\nb = 2\n"), 0o644)

		got, err := (Tool{}).Execute(&viye.Context{
			Path: []string{"config"},
			Body: []string{": a = updated"},
		})
		is.Err(t, err, nil)
		is.Equal(t, ": a = updated\n\n", got)
	})

	t.Run("key with spaces", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", dir)

		_, err := (Tool{}).Execute(&viye.Context{
			Path: []string{"config"},
			Body: []string{": my key = val"},
		})
		is.Err(t, err, `contains spaces`)
	})
}
