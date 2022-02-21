package config

import "testing"

func TestDefaultPath(t *testing.T) {
	t.Log(DefaultConfigPath("demoapp", "app.conf"))
}
