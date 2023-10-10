package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	DB         string   `yaml:"db,omitempty"`
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
	return &config, nil
}
