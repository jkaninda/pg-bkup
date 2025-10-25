/*
 *  MIT License
 *
 * Copyright (c) 2024 Jonas Kaninda
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 */

package cmd

import (
	"github.com/jkaninda/logger"
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
			return
		}
		logger.Fatal(`"backup" accepts no argument`, "args", args)

	},
}

func init() {
	// Backup
	BackupCmd.PersistentFlags().StringP("storage", "s", "local", "Define storage: local, s3, ssh, ftp, azure")
	BackupCmd.PersistentFlags().StringP("path", "P", "", "Storage path without file name. e.g: /custom_path or ssh remote path `/home/foo/backup`")
	BackupCmd.PersistentFlags().StringP("cron-expression", "e", "", "Backup cron expression (e.g., `0 0 * * *` or `@daily`)")
	BackupCmd.PersistentFlags().StringP("config", "c", "", "Configuration file for multi database backup. (e.g: `/backup/config.yaml`)")
	BackupCmd.PersistentFlags().BoolP("disable-compression", "", false, "Disable backup compression")
	BackupCmd.PersistentFlags().BoolP("all-databases", "a", false, "Backup all databases")
	BackupCmd.PersistentFlags().BoolP("all-in-one", "A", false, "Backup all databases in a single file")
	BackupCmd.PersistentFlags().String("custom-name", "", "Custom backup name")
	BackupCmd.PersistentFlags().Bool("schema-only", false, "Backup database schema only")
	BackupCmd.PersistentFlags().Bool("data-only", false, "Backup database data only")
	BackupCmd.PersistentFlags().StringSliceP("tables", "t", []string{}, "List of tables to include in the backup")
}
