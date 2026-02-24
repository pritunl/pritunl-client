package cmd

import (
	"github.com/pritunl/pritunl-client/cli/sprofile"
	"github.com/spf13/cobra"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove [profile_id]",
	Short: "Remove profile",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cobra.CheckErr("cmd: Missing profile ID")
		}

		err := sprofile.Delete(args[0])
		cobra.CheckErr(err)
	},
}
