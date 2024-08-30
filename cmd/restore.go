package cmd

import (
	"github.com/jkaninda/pg-bkup/pkg"
	"github.com/jkaninda/pg-bkup/utils"
	"github.com/spf13/cobra"
)

var RestoreCmd = &cobra.Command{
	Use:     "restore",
	Short:   "Restore database operation",
	Example: utils.RestoreExample,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			pkg.StartRestore(cmd)
		} else {
			utils.Fatal("Error, no argument required")

		}

	},
}

func init() {
	//Restore
	RestoreCmd.PersistentFlags().StringP("file", "f", "", "File name of database")
	RestoreCmd.PersistentFlags().StringP("storage", "s", "local", "Storage. local or s3")
	RestoreCmd.PersistentFlags().StringP("path", "P", "", "AWS S3 path without file name. eg: /custom_path or ssh remote path `/home/foo/backup`")

}
