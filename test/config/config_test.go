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

	if cfg.AjustBlockNum != config.DefaultAjustBlockNum {
		t.Fatalf("default AjustBlockNum should be %v, actual: %v", config.DefaultAjustBlockNum, cfg.AjustBlockNum)
	}

	if cfg.BlockDuration != config.DefaultBlockDuration {
		t.Fatalf("default BlockDuration should be %v, actual: %v", config.DefaultBlockDuration, cfg.BlockDuration)
	}

	if cfg.InitDifficulty != config.DefaultInitDifficulty {
		t.Fatalf("default InitDifficulty should be %v, actual: %v", config.DefaultInitDifficulty, cfg.InitDifficulty)
	}
}
