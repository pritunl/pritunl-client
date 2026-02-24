package logger

import (
	"hash/fnv"
	"time"

	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/sirupsen/logrus"
)

type limiter map[uint32]time.Time

func (l limiter) Check(entry *logrus.Entry, limit time.Duration) bool {
	hash := fnv.New32a()
	hash.Write([]byte(entry.Message))
	key := hash.Sum32()

	if timestamp, ok := l[key]; ok &&
		utils.SinceAbs(timestamp) < limit {

		return false
	}
	l[key] = time.Now()

	return true
}
