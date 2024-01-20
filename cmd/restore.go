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

}
