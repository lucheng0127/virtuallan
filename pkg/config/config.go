package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	CFG_FILE string = "config.yaml"
	USER_DB  string = "users"
)

type ServerConfig struct {
	Port     int    `yaml:"port"`
	IP       string `yaml:"ip"`
	Bridge   string `yaml:"bridge"`
	LogLevel string `yaml:"log-level"`
}

func GetCfgPath(dir string) string {
	return fmt.Sprintf("%s/%s", dir, CFG_FILE)
}

func GetUserDBPath(dir string) string {
	return fmt.Sprintf("%s/%s", dir, USER_DB)
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
