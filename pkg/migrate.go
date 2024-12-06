/*
MIT License

Copyright (c) 2023 Jonas Kaninda

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

package pkg

import (
	"fmt"
	"github.com/jkaninda/pg-bkup/utils"
	"github.com/spf13/cobra"
	"time"
)

func StartMigration(cmd *cobra.Command) {
	intro()
	utils.Info("Starting database migration...")
	// Get DB config
	dbConf = initDbConfig(cmd)
	targetDbConf = initTargetDbConfig()

	// Defining the target database variables
	newDbConfig := dbConfig{}
	newDbConfig.dbHost = targetDbConf.targetDbHost
	newDbConfig.dbPort = targetDbConf.targetDbPort
	newDbConfig.dbName = targetDbConf.targetDbName
	newDbConfig.dbUserName = targetDbConf.targetDbUserName
	newDbConfig.dbPassword = targetDbConf.targetDbPassword

	// Generate file name
	backupFileName := fmt.Sprintf("%s_%s.sql", dbConf.dbName, time.Now().Format("20060102_150405"))
	conf := &RestoreConfig{}
	conf.file = backupFileName
	// Backup source Database
	BackupDatabase(dbConf, backupFileName, true)
	// Restore source database into target database
	utils.Info("Restoring [%s] database into [%s] database...", dbConf.dbName, targetDbConf.targetDbName)
	RestoreDatabase(&newDbConfig, conf)
	utils.Info("[%s] database has been restored into [%s] database", dbConf.dbName, targetDbConf.targetDbName)
	utils.Info("Database migration completed.")
}
