package handlers

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/connection"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/sprofile"
	"github.com/pritunl/pritunl-client/service/types"
	"github.com/pritunl/pritunl-client/service/utils"
)

type sprofileData struct {
	Id                 string                      `json:"id"`
	Name               string                      `json:"name"`
	State              bool                        `json:"state"`
	Wg                 bool                        `json:"wg"`
	LastMode           string                      `json:"last_mode"`
	OrganizationId     string                      `json:"organization_id"`
	Organization       string                      `json:"organization"`
	ServerId           string                      `json:"server_id"`
	Server             string                      `json:"server"`
	UserId             string                      `json:"user_id"`
	User               string                      `json:"user"`
	PreConnectMsg      string                      `json:"pre_connect_msg"`
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
	PasswordMode       string                      `json:"password_mode"`
	Token              bool                        `json:"token"`
	TokenTtl           int                         `json:"token_ttl"`
	Disabled           bool                        `json:"disabled"`
	SyncTime           int64                       `json:"sync_time"`
	SyncHosts          []string                    `json:"sync_hosts"`
	SyncHash           string                      `json:"sync_hash"`
	SyncSecret         string                      `json:"sync_secret"`
	SyncToken          string                      `json:"sync_token"`
	ServerPublicKey    []string                    `json:"server_public_key"`
	ServerBoxPublicKey string                      `json:"server_box_public_key"`
	RegistrationKey    string                      `json:"registration_key"`
	OvpnData           string                      `json:"ovpn_data"`
}

func sprofilesGet(c *gin.Context) {
	err := sprofile.Reload()
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	prfls, err := sprofile.GetAllClient()
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, prfls)
}

func sprofileGet(c *gin.Context) {
	prflId := utils.FilterStr(c.Param("profile_id"))
	if prflId == "" {
		err := &errortypes.ParseError{
			errors.New("handler: Invalid profile ID"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	prfl := sprofile.Get(prflId)
	if prfl == nil {
		utils.AbortWithStatus(c, 505)
		return
	}

	c.JSON(200, prfl.Client())
}

func sprofilePut(c *gin.Context) {
	data := &sprofileData{}

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

	prfl := &sprofile.Sprofile{
		Id:                 data.Id,
		Name:               data.Name,
		Wg:                 data.Wg,
		LastMode:           data.LastMode,
		OrganizationId:     data.OrganizationId,
		Organization:       data.Organization,
		ServerId:           data.ServerId,
		Server:             data.Server,
		UserId:             data.UserId,
		User:               data.User,
		PreConnectMsg:      data.PreConnectMsg,
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
		PasswordMode:       data.PasswordMode,
		Token:              data.Token,
		TokenTtl:           data.TokenTtl,
		Disabled:           data.Disabled,
		SyncTime:           data.SyncTime,
		SyncHosts:          data.SyncHosts,
		SyncHash:           data.SyncHash,
		SyncSecret:         data.SyncSecret,
		SyncToken:          data.SyncToken,
		ServerPublicKey:    data.ServerPublicKey,
		ServerBoxPublicKey: data.ServerBoxPublicKey,
		RegistrationKey:    data.RegistrationKey,
		OvpnData:           data.OvpnData,
	}

	err = prfl.Commit()
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, prfl.Client())
}

func sprofileDel(c *gin.Context) {
	data := &profileData{}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
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

	conn := connection.GlobalStore.Get(prflId)
	if conn != nil {
		conn.Stop()
	}

	sprofile.Remove(prflId)

	c.JSON(200, nil)
}

func sprofileDel2(c *gin.Context) {
	prflId := utils.FilterStr(c.Param("profile_id"))
	if prflId == "" {
		err := &errortypes.ParseError{
			errors.New("handler: Invalid profile ID"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	connection.GlobalStore.SetStop(prflId)

	conn := connection.GlobalStore.Get(prflId)
	if conn != nil {
		conn.Stop()
	}

	sprofile.Remove(prflId)

	c.JSON(200, nil)
}

func sprofileLogGet(c *gin.Context) {
	prflId := utils.FilterStr(c.Param("profile_id"))
	if prflId == "" {
		err := &errortypes.ParseError{
			errors.New("handler: Invalid profile ID"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	sprfl := sprofile.Get(prflId)
	if sprfl == nil {
		utils.AbortWithStatus(c, 404)
		return
	}

	output, err := sprfl.GetOutput()
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.String(200, output)
}

func sprofileLogDel(c *gin.Context) {
	prflId := utils.FilterStr(c.Param("profile_id"))
	if prflId == "" {
		err := &errortypes.ParseError{
			errors.New("handler: Invalid profile ID"),
		}
		utils.AbortWithError(c, 400, err)
		return
	}

	err := sprofile.ClearLog(prflId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, nil)
}
