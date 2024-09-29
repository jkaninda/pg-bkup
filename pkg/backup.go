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
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func StartBackup(cmd *cobra.Command) {
	intro()
	dbConf = initDbConfig(cmd)
	//Initialize backup configs
	config := initBackupConfig(cmd)

	if config.cronExpression == "" {
		BackupTask(dbConf, config)
	} else {
		if utils.IsValidCronExpression(config.cronExpression) {
			scheduledMode(dbConf, config)
		} else {
			utils.Fatal("Cron expression is not valid: %s", config.cronExpression)
		}
	}

}

// Run in scheduled mode
func scheduledMode(db *dbConfig, config *BackupConfig) {
	utils.Info("Running in Scheduled mode")
	utils.Info("Backup cron expression:  %s", config.cronExpression)
	utils.Info("Storage type %s ", config.storage)

	//Test database connexion
	testDatabaseConnection(db)
	//Test backup
	utils.Info("Testing backup configurations...")
	BackupTask(db, config)
	utils.Info("Testing backup configurations...done")
	utils.Info("Creating backup job...")
	// Create a new cron instance
	c := cron.New()

	_, err := c.AddFunc(config.cronExpression, func() {
		BackupTask(db, config)
	})
	if err != nil {
		return
	}
	// Start the cron scheduler
	c.Start()
	utils.Info("Creating backup job...done")
	utils.Info("Backup job started")
	defer c.Stop()
	select {}
}
func BackupTask(db *dbConfig, config *BackupConfig) {
	utils.Info("Starting backup task...")
	//Generate file name
	backupFileName := fmt.Sprintf("%s_%s.sql.gz", db.dbName, time.Now().Format("20240102_150405"))
	if config.disableCompression {
		backupFileName = fmt.Sprintf("%s_%s.sql", db.dbName, time.Now().Format("20240102_150405"))
	}
	config.backupFileName = backupFileName
	switch config.storage {
	case "s3":
		s3Backup(db, config.backupFileName, config.disableCompression, config.prune, config.backupRetention, config.encryption)
	case "local":
		localBackup(db, config.backupFileName, config.disableCompression, config.prune, config.backupRetention, config.encryption)
	case "ssh", "remote":
		sshBackup(db, config.backupFileName, config.remotePath, config.disableCompression, config.prune, config.backupRetention, config.encryption)
	case "ftp":
		utils.Fatal("Not supported storage type: %s", config.storage)
	default:
		localBackup(db, config.backupFileName, config.disableCompression, config.prune, config.backupRetention, config.encryption)
	}
}
func intro() {
	utils.Info("Starting PostgreSQL Backup...")
	utils.Info("Copyright (c) 2024 Jonas Kaninda ")
}

// BackupDatabase backup database
func BackupDatabase(db *dbConfig, backupFileName string, disableCompression bool) {

	storagePath = os.Getenv("STORAGE_PATH")

	utils.Info("Starting database backup...")

	err := os.Setenv("PGPASSWORD", db.dbPassword)
	if err != nil {
		return
	}
	testDatabaseConnection(db)
	// Backup Database database
	utils.Info("Backing up database...")

	// Verify is compression is disabled
	if disableCompression {
		// Execute pg_dump
		cmd := exec.Command("pg_dump",
			"-h", db.dbHost,
			"-p", db.dbPort,
			"-U", db.dbUserName,
			"-d", db.dbName,
		)
		output, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}
		// save output
		file, err := os.Create(fmt.Sprintf("%s/%s", tmpPath, backupFileName))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		_, err = file.Write(output)
		if err != nil {
			log.Fatal(err)
		}

	} else {
		// Execute pg_dump
		cmd := exec.Command("pg_dump",
			"-h", db.dbHost,
			"-p", db.dbPort,
			"-U", db.dbUserName,
			"-d", db.dbName,
		)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		gzipCmd := exec.Command("gzip")
		gzipCmd.Stdin = stdout
		// save output
		gzipCmd.Stdout, err = os.Create(fmt.Sprintf("%s/%s", tmpPath, backupFileName))
		gzipCmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		if err := gzipCmd.Wait(); err != nil {
			log.Fatal(err)
		}

	}
	utils.Info("Database has been backed up")

}
func localBackup(db *dbConfig, backupFileName string, disableCompression bool, prune bool, backupRetention int, encrypt bool) {
	utils.Info("Backup database to local storage")
	BackupDatabase(db, backupFileName, disableCompression)
	finalFileName := backupFileName
	if encrypt {
		encryptBackup(backupFileName)
		finalFileName = fmt.Sprintf("%s.%s", backupFileName, gpgExtension)
	}
	utils.Info("Backup name is %s", finalFileName)
	moveToBackup(finalFileName, storagePath)
	//Send notification
	utils.NotifySuccess(finalFileName)
	//Delete old backup
	if prune {
		deleteOldBackup(backupRetention)
	}
	//Delete temp
	deleteTemp()
}

func s3Backup(db *dbConfig, backupFileName string, disableCompression bool, prune bool, backupRetention int, encrypt bool) {
	bucket := utils.GetEnvVariable("AWS_S3_BUCKET_NAME", "BUCKET_NAME")
	s3Path := utils.GetEnvVariable("AWS_S3_PATH", "S3_PATH")
	utils.Info("Backup database to s3 storage")
	//Backup database
	BackupDatabase(db, backupFileName, disableCompression)
	finalFileName := backupFileName
	if encrypt {
		encryptBackup(backupFileName)
		finalFileName = fmt.Sprintf("%s.%s", backupFileName, "gpg")
	}
	utils.Info("Uploading backup archive to remote storage S3 ... ")

	utils.Info("Backup name is %s", finalFileName)
	err := utils.UploadFileToS3(tmpPath, finalFileName, bucket, s3Path)
	if err != nil {
		utils.Fatal("Error uploading backup archive to S3: %s ", err)

	}

	//Delete backup file from tmp folder
	err = utils.DeleteFile(filepath.Join(tmpPath, backupFileName))
	if err != nil {
		fmt.Println("Error deleting file: ", err)

	}
	// Delete old backup
	if prune {
		err := utils.DeleteOldBackup(bucket, s3Path, backupRetention)
		if err != nil {
			utils.Fatal("Error deleting old backup from S3: %s ", err)
		}
	}
	utils.Done("Uploading backup archive to remote storage S3 ... done ")
	//Send notification
	utils.NotifySuccess(finalFileName)
	//Delete temp
	deleteTemp()
}
func sshBackup(db *dbConfig, backupFileName, remotePath string, disableCompression bool, prune bool, backupRetention int, encrypt bool) {
	utils.Info("Backup database to Remote server")
	//Backup database
	BackupDatabase(db, backupFileName, disableCompression)
	finalFileName := backupFileName
	if encrypt {
		encryptBackup(backupFileName)
		finalFileName = fmt.Sprintf("%s.%s", backupFileName, "gpg")
	}
	utils.Info("Uploading backup archive to remote storage ... ")
	utils.Info("Backup name is %s", finalFileName)
	err := CopyToRemote(finalFileName, remotePath)
	if err != nil {
		utils.Fatal("Error uploading file to the remote server: %s ", err)

	}

	//Delete backup file from tmp folder
	err = utils.DeleteFile(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		utils.Error("Error deleting file: %v", err)

	}
	if prune {
		//TODO: Delete old backup from remote server
		utils.Info("Deleting old backup from a remote server is not implemented yet")

	}

	utils.Done("Uploading backup archive to remote storage ... done ")
	//Send notification
	utils.NotifySuccess(finalFileName)
	//Delete temp
	deleteTemp()
}

func encryptBackup(backupFileName string) {
	gpgPassphrase := os.Getenv("GPG_PASSPHRASE")
	err := Encrypt(filepath.Join(tmpPath, backupFileName), gpgPassphrase)
	if err != nil {
		utils.Fatal("Error during encrypting backup %v", err)
	}

}
