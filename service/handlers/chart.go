package handlers

import (
	"strconv"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/metric"
	"github.com/pritunl/pritunl-client/service/utils"
)

type chartData struct {
	HasData bool             `json:"has_data"`
	Data    metric.ChartData `json:"data"`
}

func chartGet(c *gin.Context) {
	prflId := utils.FilterStr(c.Param("profile_id"))
	if prflId == "" {
		err := &errortypes.ParseError{
			errors.New("handler: Invalid profile ID"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	resource := c.Query("resource")
	if resource == "" {
		resource = "bandwidth"
	}

	period, _ := strconv.ParseInt(c.Query("period"), 10, 0)
	if period == 0 {
		period = 1440
	}

	interval, _ := strconv.ParseInt(c.Query("interval"), 10, 0)
	if interval == 0 {
		interval = 30
	}

	endTime := time.Now().UTC()
	startTime := endTime.Add(time.Duration(-period) * time.Minute)

	data, err := metric.GetChart(prflId, resource, startTime, endTime,
		time.Duration(interval)*time.Minute)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	chrtData := &chartData{
		HasData: len(data) > 0,
		Data:    data,
	}

	c.JSON(200, chrtData)
}
