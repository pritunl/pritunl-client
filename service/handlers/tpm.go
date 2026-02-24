package handlers

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/tpm"
	"github.com/pritunl/pritunl-client/service/utils"
)

type tpmCallbackData struct {
	Id         string `json:"id"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Signature  string `json:"signature"`
	Error      string `json:"error"`
}

func tpmCallbackPost(c *gin.Context) {
	data := &tpmCallbackData{}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	callerId := data.Id
	pubKey := data.PublicKey
	privKey := data.PrivateKey
	signature := data.Signature
	callerErr := data.Error

	tpm.RemoteCallback(callerId, pubKey, privKey, signature, callerErr)

	c.JSON(200, nil)
}
