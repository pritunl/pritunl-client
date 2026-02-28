package handlers

import (
	"runtime/debug"

	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/connection"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/sprofile"
	"github.com/pritunl/pritunl-client/service/types"
	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/sirupsen/logrus"
)

// TODO Add SyncHosts on client

type profileData struct {
	Id                 string                      `json:"id"`
	Mode               string                      `json:"mode"`
	OrgId              string                      `json:"org_id"`
	UserId             string                      `json:"user_id"`
	ServerId           string                      `json:"server_id"`
	SyncHosts          []string                    `json:"sync_hosts"`
	SyncToken          string                      `json:"sync_token"`
	SyncSecret         string                      `json:"sync_secret"`
	Data               string                      `json:"data"`
	Username           string                      `json:"username"`
	Password           string                      `json:"password"`
	RemotesData        map[string]types.RemoteData `json:"remotes_data"`
	HideOvpn           bool                        `json:"hide_ovpn"`
	DynamicFirewall    bool                        `json:"dynamic_firewall"`
	GeoSort            string                      `json:"geo_sort"`
	ForceConnect       bool                        `json:"force_connect"`
	DeviceAuth         bool                        `json:"device_auth"`
	DisableGateway     bool                        `json:"disable_gateway"`
	DisableDns         bool                        `json:"disable_dns"`
	Dco                bool                        `json:"dco"`
	RestrictClient     bool                        `json:"restrict_client"`
	ForceDns           bool                        `json:"force_dns"`
	SsoAuth            bool                        `json:"sso_auth"`
	ServerPublicKey    string                      `json:"server_public_key"`
	ServerBoxPublicKey string                      `json:"server_box_public_key"`
	TokenTtl           int                         `json:"token_ttl"`
	Reconnect          bool                        `json:"reconnect"`
	Timeout            bool                        `json:"timeout"`
}

func profilesGet(c *gin.Context) {
	c.JSON(200, connection.GlobalStore.GetAllData())
}

func profileGet(c *gin.Context) {
	prflId := utils.FilterStr(c.Param("profile_id"))
	if prflId == "" {
		err := &errortypes.ParseError{
			errors.New("handler: Invalid profile ID"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	prfl := connection.GlobalStore.GetData(prflId)
	if prfl == nil {
		utils.AbortWithStatus(c, 404)
		return
	}

	c.JSON(200, prfl)
}

func profilePost(c *gin.Context) {
	data := &profileData{}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	data.Id = utils.FilterStr(data.Id)
	if data.Id == "" {
		err = &errortypes.ParseError{
			errors.New("handler: Invalid profile ID"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	sprfl := sprofile.Get(data.Id)
	if sprfl != nil {
		err = sprofile.Activate(data.Id, data.Mode, data.Password)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		c.JSON(200, nil)
		return
	}

	conn := connection.GlobalStore.Get(data.Id)
	if conn != nil {
		conn.StopWait()
	}

	prfl := &connection.Profile{
		Id:                 data.Id,
		Mode:               data.Mode,
		OrgId:              data.OrgId,
		UserId:             data.UserId,
		ServerId:           data.ServerId,
		SyncHosts:          data.SyncHosts,
		SyncToken:          data.SyncToken,
		SyncSecret:         data.SyncSecret,
		Data:               data.Data,
		Username:           data.Username,
		Password:           data.Password,
		RemotesData:        data.RemotesData,
		HideOvpn:           data.HideOvpn,
		DynamicFirewall:    data.DynamicFirewall,
		GeoSort:            data.GeoSort,
		ForceConnect:       data.ForceConnect,
		DeviceAuth:         data.DeviceAuth,
		DisableGateway:     data.DisableGateway,
		DisableDns:         data.DisableDns,
		Dco:                data.Dco,
		RestrictClient:     data.RestrictClient,
		ForceDns:           data.ForceDns,
		SsoAuth:            data.SsoAuth,
		ServerPublicKey:    data.ServerPublicKey,
		ServerBoxPublicKey: data.ServerBoxPublicKey,
		TokenTtl:           data.TokenTtl,
		Reconnect:          data.Reconnect,
	}

	conn, err = connection.NewConnection(prfl)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	go func() {
		defer func() {
			panc := recover()
			if panc != nil {
				logrus.WithFields(logrus.Fields{
					"trace": string(debug.Stack()),
					"panic": panc,
				}).Error("handlers: Profile start panic")
			}
		}()

		err := conn.Start(connection.Options{
			Interactive: true,
		})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"profile_id": prfl.Id,
				"error":      err,
			}).Error("profile: Failed to start profile")
		}
	}()

	c.JSON(200, nil)
}

func profileDel(c *gin.Context) {
	data := &profileData{}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 400, err)
		return
	}

	prflId := utils.FilterStr(data.Id)
	if prflId == "" {
		err = &errortypes.ParseError{
			errors.New("handler: Invalid profile ID"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	connection.GlobalStore.SetStop(prflId)

	sprfl := sprofile.Get(prflId)
	if sprfl != nil {
		sprofile.Deactivate(prflId)
		c.JSON(200, nil)
		return
	}

	conn := connection.GlobalStore.Get(prflId)
	if conn != nil {
		conn.Stop()
	}

	c.JSON(200, nil)
}

func profileDel2(c *gin.Context) {
	prflId := utils.FilterStr(c.Param("profile_id"))
	if prflId == "" {
		err := &errortypes.ParseError{
			errors.New("handler: Invalid profile ID"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	connection.GlobalStore.SetStop(prflId)

	sprfl := sprofile.Get(prflId)
	if sprfl != nil {
		sprofile.Deactivate(prflId)
		c.JSON(200, nil)
		return
	}

	conn := connection.GlobalStore.Get(prflId)
	if conn != nil {
		conn.Stop()
	}

	c.JSON(200, nil)
}
