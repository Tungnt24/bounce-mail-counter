package utils

import (
	"github.com/sirupsen/logrus"
)

func InitLog() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)
}
