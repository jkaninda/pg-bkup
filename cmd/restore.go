// Package cmd /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
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
			utils.Fatal(`"restore" accepts no argument %q`, args)

		}

	},
}

func init() {
	//Restore
	RestoreCmd.PersistentFlags().StringP("file", "f", "", "File name of database")
	RestoreCmd.PersistentFlags().StringP("storage", "s", "local", "Define storage: local, s3, ssh, ftp")
	RestoreCmd.PersistentFlags().StringP("path", "P", "", "AWS S3 path without file name. eg: /custom_path or ssh remote path `/home/foo/backup`")

}
