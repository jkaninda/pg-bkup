// Package cmd /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var appVersion = os.Getenv("VERSION")

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		Version()
	},
}

func Version() {
	fmt.Printf("Version: %s \n", appVersion)
	fmt.Println()
}
