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
	"os/exec"
	"path/filepath"
)

func StartRestore(cmd *cobra.Command) {
	intro()
	dbConf = initDbConfig(cmd)
	restoreConf := initRestoreConfig(cmd)

	switch restoreConf.storage {
	case "local":
		utils.Info("Restore database from local")
		copyToTmp(storagePath, restoreConf.file)
		RestoreDatabase(dbConf, restoreConf)
	case "s3", "S3":
		restoreFromS3(dbConf, restoreConf)
	case "ssh", "SSH", "remote":
		restoreFromRemote(dbConf, restoreConf)
	case "ftp", "FTP":
		restoreFromFTP(dbConf, restoreConf)
	default:
		utils.Info("Restore database from local")
		copyToTmp(storagePath, restoreConf.file)
		RestoreDatabase(dbConf, restoreConf)
	}
}

func restoreFromS3(db *dbConfig, conf *RestoreConfig) {
	utils.Info("Restore database from s3")
	err := DownloadFile(tmpPath, conf.file, conf.bucket, conf.s3Path)
	if err != nil {
		utils.Fatal("Error download file from s3 %s %v ", conf.file, err)
	}
	RestoreDatabase(db, conf)
}
func restoreFromRemote(db *dbConfig, conf *RestoreConfig) {
	utils.Info("Restore database from remote server")
	err := CopyFromRemote(conf.file, conf.remotePath)
	if err != nil {
		utils.Fatal("Error download file from remote server: %s %v", filepath.Join(conf.remotePath, conf.file), err)
	}
	RestoreDatabase(db, conf)
}
func restoreFromFTP(db *dbConfig, conf *RestoreConfig) {
	utils.Info("Restore database from FTP server")
	err := CopyFromFTP(conf.file, conf.remotePath)
	if err != nil {
		utils.Fatal("Error download file from FTP server: %s %v", filepath.Join(conf.remotePath, conf.file), err)
	}
	RestoreDatabase(db, conf)
}

// RestoreDatabase restore database
func RestoreDatabase(db *dbConfig, conf *RestoreConfig) {
	if conf.file == "" {
		utils.Fatal("Error, file required")
	}
	extension := filepath.Ext(filepath.Join(tmpPath, conf.file))
	if extension == ".gpg" {

		if conf.usingKey {
			utils.Warn("Backup decryption using a private key is not fully supported")
			err := decryptWithGPGPrivateKey(filepath.Join(tmpPath, conf.file), conf.privateKey, conf.passphrase)
			if err != nil {
				utils.Fatal("error during decrypting backup %v", err)
			}
		} else {
			if conf.passphrase == "" {
				utils.Error("Error, passphrase or private key required")
				utils.Fatal("Your file seems to be a GPG file.\nYou need to provide GPG keys. GPG_PASSPHRASE or GPG_PRIVATE_KEY environment variable is required.")
			} else {
				//decryptWithGPGSymmetric file
				err := decryptWithGPGSymmetric(filepath.Join(tmpPath, conf.file), conf.passphrase)
				if err != nil {
					utils.Fatal("Error decrypting file %s %v", file, err)
				}
				//Update file name
				file = RemoveLastExtension(file)
			}
		}

	}

	if utils.FileExists(filepath.Join(tmpPath, conf.file)) {

		err := os.Setenv("PGPASSWORD", db.dbPassword)
		if err != nil {
			return
		}
		testDatabaseConnection(db)
		utils.Info("Restoring database...")

		extension := filepath.Ext(conf.file)
		// Restore from compressed file / .sql.gz
		if extension == ".gz" {
			str := "zcat " + filepath.Join(tmpPath, conf.file) + " | psql -h " + db.dbHost + " -p " + db.dbPort + " -U " + db.dbUserName + " -v -d " + db.dbName
			_, err := exec.Command("sh", "-c", str).Output()
			if err != nil {
				utils.Fatal("Error, in restoring the database %v", err)
			}
			utils.Info("Restoring database... done")
			utils.Done("Database has been restored")
			//Delete temp
			deleteTemp()

		} else if extension == ".sql" {
			//Restore from sql file
			str := "cat " + filepath.Join(tmpPath, conf.file) + " | psql -h " + db.dbHost + " -p " + db.dbPort + " -U " + db.dbUserName + " -v -d " + db.dbName
			_, err := exec.Command("sh", "-c", str).Output()
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
		utils.Fatal("File not found in %s", filepath.Join(tmpPath, conf.file))
	}
}
