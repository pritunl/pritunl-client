package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/config"
	"github.com/pritunl/pritunl-client/service/utils"
)

func resetEnclave(c *gin.Context) {
	config.Config.EnclavePrivateKey = ""

	err := config.Config.Save()
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, nil)
}
