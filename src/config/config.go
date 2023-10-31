package config

import (
	"errors"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	DefaultAjustBlockNum  = 2016
	DefaultBlockDuration  = 1 * time.Minute
	DefaultInitDifficulty = 8
)

type Config struct {
	DataDir        string        `yaml:"datadir,omitempty"`
	Endpoint       string        `yaml:"endpoint,omitempty"`
	Bootstraps     []string      `yaml:"bootstraps,omitempty"`
	AjustBlockNum  uint32        `yaml:"ajustblocknum,omitempty"`
	BlockDuration  time.Duration `yaml:"blockduation,omitempty"`
	InitDifficulty uint32        `yaml:"initdifficulty,omitempty"`
}

func Read(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	if strings.Trim(config.DataDir, "") == "" {
		return nil, errors.New("the data dir is empty")
	}
	if strings.Trim(config.Endpoint, "") == "" {
		return nil, errors.New("the listening endpoint is empty")
	}

	if config.AjustBlockNum == 0 {
		config.AjustBlockNum = DefaultAjustBlockNum
	}

	if config.BlockDuration == time.Duration(0) {
		config.BlockDuration = DefaultBlockDuration
	}

	if config.InitDifficulty == 0 {
		config.InitDifficulty = DefaultInitDifficulty
	}

	return &config, nil
}
