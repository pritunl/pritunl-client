package metric

import (
	"math/rand"
	"time"
)

func GetBandwidthChart(prflId string, start, end time.Time,
	interval time.Duration) (chartData ChartData, err error) {

	chart := NewChart(start, end, interval)

	// TODO Return random data until bandwidth tracking is implemented
	intvSec := int64(interval / time.Second)
	sentRate := int64(100000 + rand.Int63n(200000))
	recvRate := int64(400000 + rand.Int63n(800000))

	for ts := start; !ts.After(end); ts = ts.Add(interval) {
		sentRate += rand.Int63n(60001) - 30000
		if sentRate < 20000 {
			sentRate = 20000
		}

		recvRate += rand.Int63n(240001) - 120000
		if recvRate < 80000 {
			recvRate = 80000
		}

		timestamp := ts.UnixMilli()

		chart.Add("bs", timestamp, uint64(sentRate*intvSec))
		chart.Add("br", timestamp, uint64(recvRate*intvSec))
	}

	chartData = chart.Export()

	return
}
