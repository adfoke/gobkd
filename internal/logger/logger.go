package logger

import (
	"strings"

	"github.com/sirupsen/logrus"
)

func New(level string) (*logrus.Logger, error) {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	parsed, err := logrus.ParseLevel(strings.ToLower(level))
	if err != nil {
		return nil, err
	}
	log.SetLevel(parsed)
	return log, nil
}
