package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/connection"
)

type statusData struct {
	Status bool `json:"status"`
}

func statusGet(c *gin.Context) {
	data := &statusData{
		Status: connection.GlobalStore.IsConnected(),
	}

	c.JSON(200, data)
}
