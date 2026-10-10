package logger

import (
	"strings"

	"github.com/sirupsen/logrus"
)

type logHook struct{}

func (h *logHook) Fire(entry *logrus.Entry) (err error) {
	if strings.HasPrefix(entry.Message, "logger:") {
		return
	}

	if len(buffer) <= 64 {
		buffer <- cloneEntry(entry)
	}

	return
}

func (h *logHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

func cloneEntry(entry *logrus.Entry) (clone *logrus.Entry) {
	clone = entry.Dup()
	clone.Level = entry.Level
	clone.Message = entry.Message
	clone.Caller = entry.Caller

	return
}
