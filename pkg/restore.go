package pkg

import (
	"fmt"
	"github.com/jkaninda/pg-bkup/utils"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
)

func StartRestore(cmd *cobra.Command) {

	//Set env
	utils.SetEnv("STORAGE_PATH", storagePath)
	utils.GetEnv(cmd, "dbname", "DB_NAME")
	utils.GetEnv(cmd, "port", "DB_PORT")

	//Get flag value and set env
	s3Path = utils.GetEnv(cmd, "path", "S3_PATH")
	storage = utils.GetEnv(cmd, "storage", "STORAGE")
	file = utils.GetEnv(cmd, "file", "FILE_NAME")
	executionMode, _ = cmd.Flags().GetString("mode")

	if storage == "s3" {
		utils.Info("Restore database from s3")
		s3Restore(file, s3Path)
	} else {
		utils.Info("Restore database from local")
		RestoreDatabase(file)

	}
}

// RestoreDatabase restore database
func RestoreDatabase(file string) {
	dbHost = os.Getenv("DB_HOST")
	dbPassword = os.Getenv("DB_PASSWORD")
	dbUserName = os.Getenv("DB_USERNAME")
	dbName = os.Getenv("DB_NAME")
	dbPort = os.Getenv("DB_PORT")
	storagePath = os.Getenv("STORAGE_PATH")
	if file == "" {
		utils.Fatal("Error required --file")
	}

	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_NAME") == "" || os.Getenv("DB_USERNAME") == "" || os.Getenv("DB_PASSWORD") == "" || file == "" {
		utils.Fatal("Please make sure all required environment variables are set")
	} else {

		if utils.FileExists(fmt.Sprintf("%s/%s", storagePath, file)) {

			err := os.Setenv("PGPASSWORD", dbPassword)
			if err != nil {
				return
			}
			utils.TestDatabaseConnection()

			extension := filepath.Ext(fmt.Sprintf("%s/%s", storagePath, file))
			// Restore from compressed file / .sql.gz
			if extension == ".gz" {
				str := "zcat " + fmt.Sprintf("%s/%s", storagePath, file) + " | psql -h " + os.Getenv("DB_HOST") + " -p " + os.Getenv("DB_PORT") + " -U " + os.Getenv("DB_USERNAME") + " -v -d " + os.Getenv("DB_NAME")
				_, err := exec.Command("bash", "-c", str).Output()
				if err != nil {
					utils.Fatal("Error, in restoring the database")
				}
				utils.Done("Database has been restored")

			} else if extension == ".sql" {
				//Restore from sql file
				str := "cat " + fmt.Sprintf("%s/%s", storagePath, file) + " | psql -h " + os.Getenv("DB_HOST") + " -p " + os.Getenv("DB_PORT") + " -U " + os.Getenv("DB_USERNAME") + " -v -d " + os.Getenv("DB_NAME")
				_, err := exec.Command("bash", "-c", str).Output()
				if err != nil {
					utils.Fatal("Error in restoring the database", err)
				}
				utils.Done("Database has been restored")
			} else {
				utils.Fatal("Unknown file extension ", extension)
			}

		} else {
			utils.Fatal("File not found in ", fmt.Sprintf("%s/%s", storagePath, file))
		}
	}
}
func s3Restore(file, s3Path string) {
	// Restore database from S3
	MountS3Storage(s3Path)
	RestoreDatabase(file)
}
