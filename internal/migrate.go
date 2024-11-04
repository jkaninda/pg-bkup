// Package internal /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package internal

import (
	"fmt"
	"github.com/jkaninda/pg-bkup/utils"
	"github.com/spf13/cobra"
	"time"
)

func StartMigration(cmd *cobra.Command) {
	intro()
	utils.Info("Starting database migration...")
	//Get DB config
	dbConf = initDbConfig(cmd)
	targetDbConf = initTargetDbConfig()

	//Defining the target database variables
	newDbConfig := dbConfig{}
	newDbConfig.dbHost = targetDbConf.targetDbHost
	newDbConfig.dbPort = targetDbConf.targetDbPort
	newDbConfig.dbName = targetDbConf.targetDbName
	newDbConfig.dbUserName = targetDbConf.targetDbUserName
	newDbConfig.dbPassword = targetDbConf.targetDbPassword

	//Generate file name
	backupFileName := fmt.Sprintf("%s_%s.sql", dbConf.dbName, time.Now().Format("20060102_150405"))
	conf := &RestoreConfig{}
	conf.file = backupFileName
	//Backup source Database
	BackupDatabase(dbConf, backupFileName, true)
	//Restore source database into target database
	utils.Info("Restoring [%s] database into [%s] database...", dbConf.dbName, targetDbConf.targetDbName)
	RestoreDatabase(&newDbConfig, conf)
	utils.Info("[%s] database has been restored into [%s] database", dbConf.dbName, targetDbConf.targetDbName)
	utils.Info("Database migration completed.")
}
