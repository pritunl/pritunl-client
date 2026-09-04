package handlers

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/tpm"
	"github.com/pritunl/pritunl-client/service/utils"
)

type tpmClaimData struct {
	ClientId string `json:"client_id"`
}

type tpmResultData struct {
	ClientId  string `json:"client_id"`
	KeyData   string `json:"key_data"`
	PublicKey string `json:"public_key"`
	Signature string `json:"signature"`
	Error     string `json:"error"`
}

func tpmClaimPost(c *gin.Context) {
	requestId := c.Param("request_id")

	data := &tpmClaimData{}
	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	if data.ClientId == "" {
		err = &errortypes.ParseError{
			errors.New("handler: Missing client id"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	status := tpm.Claim(requestId, data.ClientId)

	c.JSON(status, nil)
}

func tpmRequestPost(c *gin.Context) {
	requestId := c.Param("request_id")

	data := &tpmResultData{}
	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	if data.ClientId == "" {
		err = &errortypes.ParseError{
			errors.New("handler: Missing client id"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	status := tpm.Complete(requestId, data.ClientId, &tpm.Result{
		KeyData:   data.KeyData,
		PublicKey: data.PublicKey,
		Signature: data.Signature,
	}, data.Error)

	c.JSON(status, nil)
}
