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

	if cfg.BlocksPerDifficulty != config.DefaultBlocksPerDifficulty {
		t.Fatalf("default AjustBlockNum should be %v, actual: %v", config.DefaultBlocksPerDifficulty, cfg.BlocksPerDifficulty)
	}

	if cfg.BlockInterval != config.DefaultBlockInterval {
		t.Fatalf("default BlockDuration should be %v, actual: %v", config.DefaultBlockInterval, cfg.BlockInterval)
	}

	if cfg.InitDifficultyLevel != config.DefaultInitDifficultyLevel {
		t.Fatalf("default InitDifficulty should be %v, actual: %v", config.DefaultInitDifficultyLevel, cfg.InitDifficultyLevel)
	}
}
