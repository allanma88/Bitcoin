package config

import (
	"Bitcoin/src/config"
	"testing"
)

func Test_Read(t *testing.T) {
	cfg, err := config.Read("config.yml")
	if err != nil {
		t.Fatalf("read config error: %v", err)
	}
	expect := "Bitcoin"
	if cfg.DataDir != expect {
		t.Fatalf("unexpect config, expect: %v, actual: %v", expect, cfg.DataDir)
	}
}
