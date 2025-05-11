package config

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port    string    `yaml:"port"`
		Timeout time.Time `yaml:"shutdown_timeout"`
	}
	Log struct {
		Level    string `yaml:"level"`
		FileName string `yaml:"file_name"`
	}
}

func InitConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("config read error: %v", err)
	}

	var cfg Config

	err := viper.Unmarshal(&cfg)
	if err != nil {
		logrus.Fatalf("parsing in struct error: %v", err)
	}

	return &cfg
}
