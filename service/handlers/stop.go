package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/autoclean"
	"github.com/pritunl/pritunl-client/service/connection"
)

func stopPost(c *gin.Context) {
	conns := connection.GlobalStore.GetAll()
	for _, conn := range conns {
		conn.StopBackground()
	}

	for _, conn := range conns {
		conn.StopWait()
	}

	autoclean.CheckAndCleanWatch()

	c.JSON(200, nil)
}

func cleanupPost(c *gin.Context) {
	autoclean.CheckAndCleanWatch()

	c.JSON(200, nil)
}
