// Package cmd /*
/*
Copyright © 2024 Jonas Kaninda  <jonaskaninda@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pg-bkup",
	Short: "PostgreSQL Backup tool, backup database to S3 or Object Storage",
	Long:  `PostgreSQL Backup and Restoration tool. Backup database to AWS S3 storage or any S3 Alternatives for Object Storage.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mysql-bkup.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().StringP("operation", "o", "backup", "Set operation")
	rootCmd.PersistentFlags().StringP("storage", "s", "local", "Set storage. local or s3")
	rootCmd.PersistentFlags().StringP("file", "f", "", "Set file name")
	rootCmd.PersistentFlags().StringP("path", "P", "/mysql-bkup", "Set s3 path, without file name")
	rootCmd.PersistentFlags().StringP("dbname", "d", "", "Set database name")
	rootCmd.PersistentFlags().StringP("mode", "m", "default", "Set execution mode. default or scheduled")
	rootCmd.PersistentFlags().StringP("period", "", "0 1 * * *", "Set schedule period time")
	rootCmd.PersistentFlags().IntP("timeout", "t", 30, "Set timeout")
	rootCmd.PersistentFlags().BoolP("disable-compression", "", false, "Disable backup compression")
	rootCmd.PersistentFlags().IntP("port", "p", 5432, "Set database port")
	rootCmd.PersistentFlags().BoolP("help", "h", false, "Print this help message")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "shows version information")
	rootCmd.AddCommand(VersionCmd)

}
