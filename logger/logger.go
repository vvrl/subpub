package logger

import (
	"io"
	"os"
	"subpub-vk/config"

	"github.com/sirupsen/logrus"
)

func InitLogger(cfg *config.Config) {

	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	file, err := os.OpenFile(cfg.Log.FileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Warnf("can't open file %s for writing logs: %v", cfg.Log.FileName, err)
	} else {
		mw := io.MultiWriter(os.Stdout, file)
		logrus.SetOutput(mw)
	}
}
