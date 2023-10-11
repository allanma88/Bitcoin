package config

import (
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	DataDir    string   `yaml:"datadir,omitempty"`
	Endpoint   string   `yaml:"endpoint,omitempty"`
	Bootstraps []string `yaml:"bootstraps,omitempty"`
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
	return &config, nil
}
