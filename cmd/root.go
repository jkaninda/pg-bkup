// Package cmd /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package cmd

import (
	"github.com/jkaninda/pg-bkup/utils"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "pg-bkup [Command]",
	Short:   "PostgreSQL Backup tool, backup database to AWS S3 or SSH Remote Server",
	Long:    `PostgreSQL Database backup and restoration tool. Backup database to AWS S3 storage, any S3 Alternatives for Object Storage or SSH remote server.`,
	Example: utils.MainExample,
	Version: appVersion,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("dbname", "d", "", "Database name")
	rootCmd.AddCommand(VersionCmd)
	rootCmd.AddCommand(BackupCmd)
	rootCmd.AddCommand(RestoreCmd)
	rootCmd.AddCommand(MigrateCmd)
}
