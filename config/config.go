package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Pool struct {
		Nodes []struct {
			Name string `yaml:"name"`
			DSN  string `yaml:"dsn"`
		} `yaml:"nodes"`
	} `yaml:"pool"`
}

func New(path string) (*Config, error) {
	return newUnmarshal(path)
}

func newUnmarshal(path string) (*Config, error) {
	cfg := new(Config)
	cfgData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(cfgData, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}