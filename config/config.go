package config

import (
	"time"

	"github.com/spf13/viper"
)

const configPath = "config/config.yaml"

type Config struct {
	Logger LoggerConfig `yaml:"logger"`
	Worker WorkerConfig `yaml:"worker"`
}

type LoggerConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

type WorkerConfig struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
	Threads int64         `yaml:"threads"`
}

func NewConfig() (*Config, error) {
	var err error
	var config Config

	viper.SetConfigFile(configPath)

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
