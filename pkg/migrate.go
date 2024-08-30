package pkg

import (
	"fmt"
	"github.com/jkaninda/pg-bkup/utils"
	"github.com/spf13/cobra"
	"time"
)

func StartMigration(cmd *cobra.Command) {
	utils.Info("Starting database migration...")
	//Get DB config
	dbConf = getDbConfig(cmd)
	sDbConf = getSourceDbConfig()

	//Generate file name
	backupFileName := fmt.Sprintf("%s_%s.sql", sDbConf.sourceDbName, time.Now().Format("20060102_150405"))
	//Backup Source Database
	newDbConfig := dbConfig{}
	newDbConfig.dbHost = sDbConf.sourceDbHost
	newDbConfig.dbPort = sDbConf.sourceDbPort
	newDbConfig.dbName = sDbConf.sourceDbName
	newDbConfig.dbUserName = sDbConf.sourceDbUserName
	newDbConfig.dbPassword = sDbConf.sourceDbPassword
	BackupDatabase(&newDbConfig, backupFileName, true)
	//Restore source database into target database
	utils.Info("Restoring [%s] database into [%s] database...", sDbConf.sourceDbName, dbConf.dbName)
	RestoreDatabase(dbConf, backupFileName)
	utils.Info("[%s] database has been restored into [%s] database", sDbConf.sourceDbName, dbConf.dbName)
	utils.Info("Database migration completed!")
}
