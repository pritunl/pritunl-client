package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/event"
	"github.com/pritunl/pritunl-client/service/utils"
)

func wakeupPost(c *gin.Context) {
	evt := &event.Event{
		Id:   utils.Uuid(),
		Type: "wakeup",
	}
	evt.Init()

	if time.Since(event.LastPong) > 45*time.Second {
		c.String(404, "")
		return
	}

	for i := 0; i < 100; i++ {
		time.Sleep(5 * time.Millisecond)
		if utils.SinceAbs(event.LastAwake) < 300*time.Millisecond {
			c.String(200, "")
			return
		}
	}

	c.String(404, "")
}
