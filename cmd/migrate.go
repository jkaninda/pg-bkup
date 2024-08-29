package cmd

import (
	"github.com/jkaninda/pg-bkup/pkg"
	"github.com/jkaninda/pg-bkup/utils"
	"github.com/spf13/cobra"
)

var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate database from a source database to a target database",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			pkg.StartMigration(cmd)
		} else {
			utils.Fatal("Error, no argument required")

		}

	},
}
