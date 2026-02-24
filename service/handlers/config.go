package handlers

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/config"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/utils"
)

type configData struct {
	DisableDnsWatch  bool `json:"disable_dns_watch"`
	EnableDnsRefresh bool `json:"enable_dns_refresh"`
	DisableWakeWatch bool `json:"disable_wake_watch"`
	DisableNetClean  bool `json:"disable_net_clean"`
	DisableWgDns     bool `json:"disable_wg_dns"`
	InterfaceMetric  int  `json:"interface_metric"`
}

func configGet(c *gin.Context) {
	data := &configData{
		DisableDnsWatch:  config.Config.DisableDnsWatch,
		EnableDnsRefresh: config.Config.EnableDnsRefresh,
		DisableWakeWatch: config.Config.DisableWakeWatch,
		DisableNetClean:  config.Config.DisableNetClean,
		DisableWgDns:     config.Config.DisableWgDns,
		InterfaceMetric:  config.Config.InterfaceMetric,
	}

	c.JSON(200, data)
}

func configPut(c *gin.Context) {
	data := &configData{}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	config.Config.DisableDnsWatch = data.DisableDnsWatch
	config.Config.EnableDnsRefresh = data.EnableDnsRefresh
	config.Config.DisableWakeWatch = data.DisableWakeWatch
	config.Config.DisableNetClean = data.DisableNetClean
	config.Config.DisableWgDns = data.DisableWgDns
	config.Config.InterfaceMetric = data.InterfaceMetric

	err = config.Save()
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, data)
}
