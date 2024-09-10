// Package pkg /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package pkg

import (
	"github.com/jkaninda/pg-bkup/utils"
	"github.com/spf13/cobra"
	"os"
)

type Config struct {
}

type dbConfig struct {
	dbHost     string
	dbPort     string
	dbName     string
	dbUserName string
	dbPassword string
}
type targetDbConfig struct {
	targetDbHost     string
	targetDbPort     string
	targetDbUserName string
	targetDbPassword string
	targetDbName     string
}
type TgConfig struct {
	Token  string
	ChatId string
}

func getDbConfig(cmd *cobra.Command) *dbConfig {
	//Set env
	utils.GetEnv(cmd, "dbname", "DB_NAME")
	dConf := dbConfig{}
	dConf.dbHost = os.Getenv("DB_HOST")
	dConf.dbPort = os.Getenv("DB_PORT")
	dConf.dbName = os.Getenv("DB_NAME")
	dConf.dbUserName = os.Getenv("DB_USERNAME")
	dConf.dbPassword = os.Getenv("DB_PASSWORD")

	err := utils.CheckEnvVars(dbHVars)
	if err != nil {
		utils.Error("Please make sure all required environment variables for database are set")
		utils.Fatal("Error checking environment variables: %s", err)
	}
	return &dConf
}
func getTargetDbConfig() *targetDbConfig {
	tdbConfig := targetDbConfig{}
	tdbConfig.targetDbHost = os.Getenv("TARGET_DB_HOST")
	tdbConfig.targetDbPort = os.Getenv("TARGET_DB_PORT")
	tdbConfig.targetDbName = os.Getenv("TARGET_DB_NAME")
	tdbConfig.targetDbUserName = os.Getenv("TARGET_DB_USERNAME")
	tdbConfig.targetDbPassword = os.Getenv("TARGET_DB_PASSWORD")

	err := utils.CheckEnvVars(tdbRVars)
	if err != nil {
		utils.Error("Please make sure all required environment variables for the target database are set")
		utils.Fatal("Error checking target database environment variables: %s", err)
	}
	return &tdbConfig
}
