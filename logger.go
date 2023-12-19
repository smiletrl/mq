package mq

type Logger interface {
	Errorf(format string, args ...interface{})
}
