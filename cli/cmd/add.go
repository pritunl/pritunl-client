package cmd

import (
	"github.com/pritunl/pritunl-client/cli/sprofile"
	"github.com/spf13/cobra"
)

var AddCmd = &cobra.Command{
	Use:   "add [profile_uri|tar_path|ovpn_path]",
	Short: "Add profile",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cobra.CheckErr("cmd: Missing profile URI or path")
		}

		err := sprofile.ImportPath(args[0])
		cobra.CheckErr(err)
	},
}
