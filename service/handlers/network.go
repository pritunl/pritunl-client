package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/connection"
	"github.com/pritunl/pritunl-client/service/utils"
)

func networkDnsReset(c *gin.Context) {
	utils.ResetDns()
	utils.ClearDNSCache()

	c.JSON(200, nil)
}

func networkAllReset(c *gin.Context) {
	utils.ResetDns()
	utils.ClearDns()
	utils.ResetNetworking()
	utils.ClearDNSCache()

	_ = connection.RestartProfiles(true)

	c.JSON(200, nil)
}
