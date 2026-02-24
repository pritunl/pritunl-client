package handlers

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/log"
	"github.com/pritunl/pritunl-client/service/utils"
)

func logGet(c *gin.Context) {
	logId := utils.FilterStr(c.Param("log_id"))
	if logId == "" {
		err := &errortypes.ParseError{
			errors.New("handler: Invalid log ID"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	var err error
	var data string

	if logId == "service" {
		data, err = utils.GetServiceLog()
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
	} else {
		data, err = log.GetProfileLog(logId)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
	}

	c.String(200, data)
}

func logDel(c *gin.Context) {
	logId := utils.FilterStr(c.Param("log_id"))
	if logId == "" {
		err := &errortypes.ParseError{
			errors.New("handler: Invalid log ID"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	if logId == "service" {
		err := utils.ClearServiceLog()
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
	} else {
		err := log.ClearProfileLog(logId)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
	}

	c.JSON(200, nil)
}
