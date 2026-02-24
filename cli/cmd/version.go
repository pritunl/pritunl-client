package cmd

import (
	"fmt"

	"github.com/pritunl/pritunl-client/cli/constants"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Pritunl Client v%s\n", constants.Version)
	},
}
