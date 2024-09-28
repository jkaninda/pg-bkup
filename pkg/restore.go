// Package pkg /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
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
	intro()
	dbConf = initDbConfig(cmd)
	restoreConf := initRestoreConfig(cmd)

	switch restoreConf.storage {
	case "s3":
		restoreFromS3(dbConf, restoreConf.file, restoreConf.bucket, restoreConf.s3Path)
	case "local":
		utils.Info("Restore database from local")
		copyToTmp(storagePath, restoreConf.file)
		RestoreDatabase(dbConf, restoreConf.file)
	case "ssh":
		restoreFromRemote(dbConf, restoreConf.file, restoreConf.remotePath)
	case "ftp":
		utils.Fatal("Restore from FTP is not yet supported")
	default:
		utils.Info("Restore database from local")
		copyToTmp(storagePath, restoreConf.file)
		RestoreDatabase(dbConf, restoreConf.file)
	}
}

func restoreFromS3(db *dbConfig, file, bucket, s3Path string) {
	utils.Info("Restore database from s3")
	err := utils.DownloadFile(tmpPath, file, bucket, s3Path)
	if err != nil {
		utils.Fatal("Error download file from s3 %s %v ", file, err)
	}
	RestoreDatabase(db, file)
}
func restoreFromRemote(db *dbConfig, file, remotePath string) {
	utils.Info("Restore database from remote server")
	err := CopyFromRemote(file, remotePath)
	if err != nil {
		utils.Fatal("Error download file from remote server: %s %v", filepath.Join(remotePath, file), err)
	}
	RestoreDatabase(db, file)
}

// RestoreDatabase restore database
func RestoreDatabase(db *dbConfig, file string) {
	gpgPassphrase := os.Getenv("GPG_PASSPHRASE")
	if file == "" {
		utils.Fatal("Error, file required")
	}
	extension := filepath.Ext(fmt.Sprintf("%s/%s", tmpPath, file))
	if extension == ".gpg" {
		if gpgPassphrase == "" {
			utils.Fatal("Error: GPG passphrase is required, your file seems to be a GPG file.\nYou need to provide GPG keys. GPG_PASSPHRASE environment variable is required.")

		} else {
			//Decrypt file
			err := Decrypt(filepath.Join(tmpPath, file), gpgPassphrase)
			if err != nil {
				utils.Fatal("Error decrypting file %s %v", file, err)
			}
			//Update file name
			file = RemoveLastExtension(file)
		}

	}

	err := utils.CheckEnvVars(dbHVars)
	if err != nil {
		utils.Error("Please make sure all required environment variables for database are set")
		utils.Fatal("Error checking environment variables: %s", err)
	}

	if utils.FileExists(fmt.Sprintf("%s/%s", tmpPath, file)) {

		err := os.Setenv("PGPASSWORD", db.dbPassword)
		if err != nil {
			return
		}
		testDatabaseConnection(db)
		utils.Info("Restoring database...")

		extension := filepath.Ext(fmt.Sprintf("%s/%s", tmpPath, file))
		// Restore from compressed file / .sql.gz
		if extension == ".gz" {
			str := "zcat " + fmt.Sprintf("%s/%s", tmpPath, file) + " | psql -h " + db.dbHost + " -p " + db.dbPort + " -U " + db.dbUserName + " -v -d " + db.dbName
			_, err := exec.Command("bash", "-c", str).Output()
			if err != nil {
				utils.Fatal("Error, in restoring the database %v", err)
			}
			utils.Info("Restoring database... done")
			utils.Done("Database has been restored")
			//Delete temp
			deleteTemp()

		} else if extension == ".sql" {
			//Restore from sql file
			str := "cat " + fmt.Sprintf("%s/%s", tmpPath, file) + " | psql -h " + db.dbHost + " -p " + db.dbPort + " -U " + db.dbUserName + " -v -d " + db.dbName
			_, err := exec.Command("bash", "-c", str).Output()
			if err != nil {
				utils.Fatal("Error in restoring the database %v", err)
			}
			utils.Info("Restoring database... done")
			utils.Done("Database has been restored")
			//Delete temp
			deleteTemp()
		} else {
			utils.Fatal("Unknown file extension: %s", extension)
		}

	} else {
		utils.Fatal("File not found in %s", fmt.Sprintf("%s/%s", tmpPath, file))
	}
}
