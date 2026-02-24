package handlers

import (
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/connection"
	"github.com/pritunl/pritunl-client/service/constants"
	"github.com/pritunl/pritunl-client/service/update"
)

type stateData struct {
	Wg       bool   `json:"wg"`
	Version  string `json:"version"`
	Upgrade  bool   `json:"upgrade"`
	Security bool   `json:"security"`
}

func stateGet(c *gin.Context) {
	data := &stateData{
		Wg:       false,
		Version:  constants.Version,
		Upgrade:  update.Upgrade,
		Security: update.Security,
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		if connection.GetWgPath() != "" && connection.GetWgQuickPath() != "" {
			data.Wg = true
		}

		break
	case "windows":
		if connection.GetWgPath() != "" {
			data.Wg = true
		}

		break
	default:
		panic("handlers: Not implemented")
	}

	c.JSON(200, data)
}
