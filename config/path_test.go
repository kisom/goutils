package config_test

import (
	"testing"

	"git.wntrmute.dev/kyle/goutils/config"
)

func TestDefaultPath(t *testing.T) {
	t.Log(config.DefaultConfigPath("demoapp", "app.conf"))
}
