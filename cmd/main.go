package main

import (
	"subpub-vk/config"
	"subpub-vk/logger"

	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.InitConfig()

	logger.InitLogger(cfg)

	logrus.Infof("Config is loaded: %v", cfg)

}
