package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dropbox/godropbox/errors"
	"github.com/olekukonko/tablewriter"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/sprofile"
	"github.com/spf13/cobra"
)

type Profile struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	State           string `json:"state"`
	RunState        string `json:"run_state"`
	RegistrationKey string `json:"registration_key"`
	Connected       bool   `json:"connected"`
	Uptime          int64  `json:"uptime"`
	Status          string `json:"status"`
	ServerAddress   string `json:"server_address"`
	ClientAddress   string `json:"client_address"`
}

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List profiles",
	Run: func(cmd *cobra.Command, args []string) {
		sprfls, err := sprofile.GetAll()
		cobra.CheckErr(err)

		if jsonFormat || jsonFormated {
			prfls := []*Profile{}

			for _, sprfl := range sprfls {
				if sprfl.Profile != nil {
					prfls = append(prfls, &Profile{
						Id:              sprfl.Id,
						Name:            sprfl.FormatedName(),
						State:           sprfl.FormatedState(),
						RunState:        sprfl.FormatedRunState(),
						RegistrationKey: sprfl.RegistrationKey,
						Connected:       sprfl.Profile.ClientAddr != "",
						Uptime:          sprfl.Profile.Uptime(),
						Status:          sprfl.Profile.FormatedTime(),
						ServerAddress:   sprfl.Profile.ServerAddr,
						ClientAddress:   sprfl.Profile.ClientAddr,
					})
				} else {
					prfls = append(prfls, &Profile{
						Id:              sprfl.Id,
						Name:            sprfl.FormatedName(),
						State:           sprfl.FormatedState(),
						RunState:        sprfl.FormatedRunState(),
						RegistrationKey: sprfl.RegistrationKey,
						Uptime:          0,
						Status:          "Disconnected",
						ServerAddress:   "",
						ClientAddress:   "",
					})
				}
			}

			var output []byte
			if jsonFormated {
				output, err = json.MarshalIndent(prfls, "", "  ")
				if err != nil {
					err = &errortypes.ParseError{
						errors.Wrap(err, "utils: Failed to marshal profile"),
					}
					cobra.CheckErr(err)
				}
			} else {
				output, err = json.Marshal(prfls)
				if err != nil {
					err = &errortypes.ParseError{
						errors.Wrap(err, "utils: Failed to marshal profile"),
					}
					cobra.CheckErr(err)
				}
			}

			fmt.Println(string(output))
		} else {
			hasRegKey := false
			for _, sprfl := range sprfls {
				if sprfl.RegistrationKey != "" {
					hasRegKey = true
					break
				}
			}

			table := tablewriter.NewWriter(os.Stdout)

			fields := []string{
				"ID",
				"Name",
				"State",
				"Autostart",
				"Online For",
				"Server Address",
				"Client Address",
			}
			if hasRegKey {
				fields = append(fields, "Registration Key")
			}

			table.SetHeader(fields)
			table.SetBorder(true)

			for _, sprfl := range sprfls {
				if sprfl.Profile != nil && sprfl.State {
					fields := []string{
						sprfl.Id,
						sprfl.FormatedName(),
						sprfl.FormatedRunState(),
						sprfl.FormatedState(),
						sprfl.Profile.FormatedTime(),
						sprfl.Profile.ServerAddr,
						sprfl.Profile.ClientAddr,
					}
					if hasRegKey {
						fields = append(fields, sprfl.RegistrationKey)
					}

					table.Append(fields)
				} else {
					fields := []string{
						sprfl.Id,
						sprfl.FormatedName(),
						sprfl.FormatedRunState(),
						sprfl.FormatedState(),
						"Disconnected",
						"-",
						"-",
					}
					if hasRegKey {
						fields = append(fields, sprfl.RegistrationKey)
					}

					table.Append(fields)
				}
			}

			table.Render()
		}
	},
}
