package cmd

import (
	"fmt"

	"github.com/pritunl/pritunl-client/cli/sprofile"
	"github.com/spf13/cobra"
)

var LogsCmd = &cobra.Command{
	Use:   "logs [profile_id]",
	Short: "Show logs for profile",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cobra.CheckErr("cmd: Missing profile ID")
		}

		sprfl, err := sprofile.Match(args[0])
		cobra.CheckErr(err)

		data, err := sprfl.GetLogs()
		cobra.CheckErr(err)

		fmt.Print(data)
	},
}
