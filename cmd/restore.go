/*
MIT License

# Copyright (c) 2023 Jonas Kaninda

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

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
	// Restore
	RestoreCmd.PersistentFlags().StringP("file", "f", "", "File name of database")
	RestoreCmd.PersistentFlags().StringP("storage", "s", "local", "Define storage: local, s3, ssh, ftp")
	RestoreCmd.PersistentFlags().StringP("path", "P", "", "AWS S3 path without file name. eg: /custom_path or ssh remote path `/home/foo/backup`")

}
