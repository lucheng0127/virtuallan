package config

import (
	"fmt"
	"os"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const (
	CFG_FILE string = "config.yaml"
	USER_DB  string = "users"
)

type WebConfig struct {
	Enable bool `yaml:"enable"`
	Port   int  `yaml:"port"`
}

// TODO: Add routes
type ServerConfig struct {
	Port      int    `yaml:"port" default:"6123"`
	IP        string `yaml:"ip" validate:"required"`
	Bridge    string `yaml:"bridge" validate:"required"`
	LogLevel  string `yaml:"log-level" default:"info"`
	Key       string `yaml:"key" validate:"required,validKeyLen"`
	DHCPRange string `yaml:"dhcp-range" validate:"required"`
	WebConfig `yaml:"web"`
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

	validate := validator.New()
	validate.RegisterValidation("validKeyLen", IsValidKeyLength)

	if err := validate.Struct(cfg); err != nil {
		return nil, err
	}

	if err := defaults.Set(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func IsValidKeyLength(fl validator.FieldLevel) bool {
	k := len([]byte(fl.Field().String()))

	switch k {
	default:
		return false
	case 16, 24, 32:
		return true
	}
}
