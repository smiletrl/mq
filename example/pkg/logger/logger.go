package logger

import (
	"log"

	"github.com/smiletrl/mq"
)

type logger struct{}

func NewLogger() mq.Logger {
	return &logger{}
}

func (l *logger) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}
