package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port   int    `yaml:"port"`
	IP     string `yaml:"ip"`
	Bridge string `yaml:"bridge"`
}

func LoadConfigFile(path string) (*ServerConfig, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := new(ServerConfig)
	err = yaml.Unmarshal(f, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
