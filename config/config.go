package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Pool struct {
		Nodes []struct {
			ID           string `yaml:"id"`
			DSN          string `yaml:"dsn"`
			InternalHost string `yaml:"internal_host"`
			InternalPort string `yaml:"internal_port"`
		} `yaml:"nodes"`
	} `yaml:"pool"`
	Zookeeper []string `yaml:"zookeeper"`
	NodesPath string   `yaml:"nodes_path"`
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
