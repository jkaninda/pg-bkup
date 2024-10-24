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

var BackupCmd = &cobra.Command{
	Use:     "backup ",
	Short:   "Backup database operation",
	Example: utils.BackupExample,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			pkg.StartBackup(cmd)
		} else {
			utils.Fatal(`"backup" accepts no argument %q`, args)

		}
	},
}

func init() {
	//Backup
	BackupCmd.PersistentFlags().StringP("storage", "s", "local", "Define storage: local, s3, ssh, ftp")
	BackupCmd.PersistentFlags().StringP("path", "P", "", "AWS S3 path without file name. eg: /custom_path or ssh remote path `/home/foo/backup`")
	BackupCmd.PersistentFlags().StringP("cron-expression", "", "", "Backup cron expression")
	BackupCmd.PersistentFlags().BoolP("disable-compression", "", false, "Disable backup compression")

}
