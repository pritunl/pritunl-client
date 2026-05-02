package cmd

import (
	"fmt"

	"github.com/pritunl/pritunl-client/cli/sprofile"
	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:   "status [profile_id]",
	Short: "Profile status",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cobra.CheckErr("cmd: Missing profile ID")
		}

		status, err := sprofile.Status(args[0])
		if err != nil {
			cobra.CheckErr(err)
		}

		fmt.Println(status)
	},
}
