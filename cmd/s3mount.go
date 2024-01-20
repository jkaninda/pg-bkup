package cmd

import (
	"github.com/jkaninda/pg-bkup/pkg"
	"github.com/spf13/cobra"
)

var S3MountCmd = &cobra.Command{
	Use:   "s3mount",
	Short: "Mount AWS S3 storage",
	Run: func(cmd *cobra.Command, args []string) {
		pkg.S3Mount()
	},
}
