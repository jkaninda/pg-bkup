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
			utils.Fatal("Error, no argument required")
		}
	},
}

func init() {
	//Backup
	BackupCmd.PersistentFlags().StringP("mode", "m", "default", "Execution mode. default or scheduled")
	BackupCmd.PersistentFlags().StringP("period", "", "0 1 * * *", "Schedule period time")
	BackupCmd.PersistentFlags().BoolP("prune", "", false, "Delete old backup, default disabled")
	BackupCmd.PersistentFlags().IntP("keep-last", "", 7, "Delete files created more than specified days ago, default 7 days")
	BackupCmd.PersistentFlags().BoolP("disable-compression", "", false, "Disable backup compression")

}
