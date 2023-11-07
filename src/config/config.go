package config

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"

	"github.com/peteprogrammer/go-automapper"
	"gopkg.in/yaml.v2"
)

const (
	DefaultBlocksPerDifficulty = 2016
	DefaultBlocksPerRewrad     = 210 * 1000
	DefaultBlockInterval       = 60
	DefaultInitDifficultyLevel = 8
)

type Config struct {
	DataDir             string   `yaml:"data_dir,omitempty"`
	Endpoint            string   `yaml:"endpoint,omitempty"`
	Bootstraps          []string `yaml:"bootstraps,omitempty"`
	BlocksPerDifficulty uint64   `yaml:"blocks_per_difficulty,omitempty"` //TODO: switch to int
	BlocksPerRewrad     uint64   `yaml:"blocks_per_reward,omitempty"`     //TODO: switch to int
	BlockInterval       uint64   `yaml:"block_interval,omitempty"`
	InitDifficultyLevel uint64   `yaml:"init_difficulty_level,omitempty"`
	MinerPubkey         []byte   `yaml:"miner_address,omitempty"`
}

// TODO: need more test cases
func Read(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s struct {
		DataDir             string   `yaml:"data_dir,omitempty"`
		Endpoint            string   `yaml:"endpoint,omitempty"`
		Bootstraps          []string `yaml:"bootstraps,omitempty"`
		BlocksPerDifficulty uint64   `yaml:"blocks_per_difficulty,omitempty"`
		BlocksPerRewrad     uint64   `yaml:"blocks_per_reward,omitempty"`
		BlockInterval       uint64   `yaml:"block_interval,omitempty"`
		InitDifficultyLevel uint64   `yaml:"init_difficulty_level,omitempty"`
		MinerAddress        string   `yaml:"miner_address,omitempty"`
	}
	err = yaml.Unmarshal(file, &s)
	if err != nil {
		return nil, err
	}

	if strings.Trim(s.DataDir, "") == "" {
		return nil, errors.New("the data dir is empty")
	}
	if strings.Trim(s.Endpoint, "") == "" {
		return nil, errors.New("the listening endpoint is empty")
	}
	if strings.Trim(s.MinerAddress, "") == "" {
		return nil, errors.New("the miner address is empty")
	}

	var config Config
	automapper.MapLoose(&config, &s)

	if config.BlocksPerDifficulty == 0 {
		config.BlocksPerDifficulty = DefaultBlocksPerDifficulty
	}

	if config.BlocksPerRewrad == 0 {
		config.BlocksPerRewrad = DefaultBlocksPerRewrad
	}

	if config.BlockInterval == 0 {
		config.BlockInterval = DefaultBlockInterval
	}

	if config.InitDifficultyLevel == 0 {
		config.InitDifficultyLevel = DefaultInitDifficultyLevel
	}

	pubkey, err := base64.RawStdEncoding.DecodeString(s.MinerAddress)
	if err != nil {
		return nil, err
	}

	config.MinerPubkey = pubkey
	return &config, nil
}
