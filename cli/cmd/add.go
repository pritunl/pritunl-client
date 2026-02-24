package cmd

import (
	"strings"

	"github.com/pritunl/pritunl-client/cli/sprofile"
	"github.com/spf13/cobra"
)

var AddCmd = &cobra.Command{
	Use:   "add [profile_uri|tar_path]",
	Short: "Add profile",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cobra.CheckErr("cmd: Missing profile URI or path")
		}

		path := args[0]
		if strings.HasPrefix(path, "http://") ||
			strings.HasPrefix(path, "https://") ||
			strings.HasPrefix(path, "pritunl://") ||
			strings.HasPrefix(path, "pritunls://") {

			err := sprofile.ImportUri(path)
			cobra.CheckErr(err)
		} else {
			err := sprofile.ImportTar(path)
			cobra.CheckErr(err)
		}
	},
}
