package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/connection"
	"github.com/sirupsen/logrus"
)

func restartPost(c *gin.Context) {
	logrus.Warn("handlers: Restarting...")

	connection.RestartProfiles(false)

	c.JSON(200, nil)
}
