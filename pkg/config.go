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
type dbSourceConfig struct {
	sourceDbHost     string
	sourceDbPort     string
	sourceDbUserName string
	sourceDbPassword string
	sourceDbName     string
}

func getDbConfig(cmd *cobra.Command) *dbConfig {
	//Set env
	utils.GetEnv(cmd, "dbname", "DB_NAME")
	utils.GetEnv(cmd, "port", "DB_PORT")
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
func getSourceDbConfig() *dbSourceConfig {
	sdbConfig := dbSourceConfig{}
	sdbConfig.sourceDbHost = os.Getenv("SOURCE_DB_HOST")
	sdbConfig.sourceDbPort = os.Getenv("SOURCE_DB_PORT")
	sdbConfig.sourceDbName = os.Getenv("SOURCE_DB_NAME")
	sdbConfig.sourceDbUserName = os.Getenv("SOURCE_DB_USERNAME")
	sdbConfig.sourceDbPassword = os.Getenv("SOURCE_DB_PASSWORD")

	err := utils.CheckEnvVars(sdbRVars)
	if err != nil {
		utils.Error("Please make sure all required environment variables for source database are set")
		utils.Fatal("Error checking environment variables: %s", err)
	}
	return &sdbConfig
}
