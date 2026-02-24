package handlers

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/token"
	"github.com/pritunl/pritunl-client/service/utils"
)

type tokenData struct {
	Profile            string `json:"profile"`
	ServerPublicKey    string `json:"server_public_key"`
	ServerBoxPublicKey string `json:"server_box_public_key"`
	Ttl                int    `json:"ttl"`
}

func tokenPut(c *gin.Context) {
	data := &tokenData{}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	tokn, err := token.Update(
		data.Profile,
		data.ServerPublicKey,
		data.ServerBoxPublicKey,
		data.Ttl,
	)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, tokn)
}

func tokenDelete(c *gin.Context) {
	data := &tokenData{}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	token.Clear(data.Profile)

	c.JSON(200, nil)
}

func tokenDelete2(c *gin.Context) {
	prflId := utils.FilterStr(c.Param("profile_id"))
	if prflId == "" {
		err := &errortypes.ParseError{
			errors.New("handler: Invalid profile ID"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	token.Clear(prflId)

	c.JSON(200, nil)
}
