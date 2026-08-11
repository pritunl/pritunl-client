package metric

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/errortypes"
)

type Point struct {
	X int64       `json:"x"`
	Y interface{} `json:"y"`
}

type ChartData = map[string][]*Point

func GetChart(prflId string, resource string, start, end time.Time,
	interval time.Duration) (ChartData, error) {

	start = start.Add(time.Duration(start.UnixMilli()%
		interval.Milliseconds()) * -time.Millisecond)
	end = end.Add(time.Duration(end.UnixMilli()%
		interval.Milliseconds()) * -time.Millisecond)

	switch resource {
	case "bandwidth":
		return GetBandwidthChart(prflId, start, end, interval)
	default:
		return nil, &errortypes.UnknownError{
			errors.New("metric: Unknown resource type"),
		}
	}
}

type Chart struct {
	start    int64
	end      int64
	intv     int64
	data     ChartData
	curTimes map[string]int64
}

func (c *Chart) add(resource string, timestamp int64, value interface{}) {
	c.data[resource] = append(c.data[resource], &Point{
		X: timestamp,
		Y: value,
	})
}

func (c *Chart) Add(resource string, timestamp int64, value interface{}) {
	cur := c.curTimes[resource]
	if cur == 0 {
		cur = c.start - c.intv
	}

	for timestamp-c.intv > cur {
		cur += c.intv
		c.add(resource, cur, 0)
	}

	c.add(resource, timestamp, value)
	c.curTimes[resource] = timestamp
}

func (c *Chart) Export() map[string][]*Point {
	for resource, cur := range c.curTimes {
		for c.end > cur {
			cur += c.intv
			c.add(resource, cur, 0)
		}
	}

	return c.data
}

func NewChart(start, end time.Time, interval time.Duration) (chrt *Chart) {
	chrt = &Chart{
		start:    start.UnixMilli(),
		end:      end.UnixMilli(),
		intv:     interval.Milliseconds(),
		data:     ChartData{},
		curTimes: map[string]int64{},
	}

	if interval == time.Minute {
		chrt.end -= time.Minute.Milliseconds()
	}

	return
}
