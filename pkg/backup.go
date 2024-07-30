// Package pkg /*
/*
Copyright © 2024 Jonas Kaninda
*/
package pkg

import (
	"fmt"
	"github.com/hpcloud/tail"
	"github.com/jkaninda/pg-bkup/utils"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func StartBackup(cmd *cobra.Command) {
	_, _ = cmd.Flags().GetString("operation")
	//Set env
	utils.SetEnv("STORAGE_PATH", storagePath)
	utils.GetEnv(cmd, "dbname", "DB_NAME")
	utils.GetEnv(cmd, "port", "DB_PORT")
	utils.GetEnv(cmd, "period", "SCHEDULE_PERIOD")

	//Get flag value and set env
	s3Path = utils.GetEnv(cmd, "path", "S3_PATH")
	storage = utils.GetEnv(cmd, "storage", "STORAGE")
	file = utils.GetEnv(cmd, "file", "FILE_NAME")
	backupRetention, _ := cmd.Flags().GetInt("keep-last")
	prune, _ := cmd.Flags().GetBool("prune")
	disableCompression, _ = cmd.Flags().GetBool("disable-compression")
	executionMode, _ = cmd.Flags().GetString("mode")
	dbName = os.Getenv("DB_NAME")
	gpgPassphrase := os.Getenv("GPG_PASSPHRASE")
	//
	if gpgPassphrase != "" {
		encryption = true
	}

	//Generate file name
	backupFileName := fmt.Sprintf("%s_%s.sql.gz", dbName, time.Now().Format("20060102_150405"))
	if disableCompression {
		backupFileName = fmt.Sprintf("%s_%s.sql", dbName, time.Now().Format("20060102_150405"))
	}

	if executionMode == "default" {
		switch storage {
		case "s3":
			s3Backup(backupFileName, s3Path, disableCompression, prune, backupRetention, encryption)
		case "local":
			localBackup(backupFileName, disableCompression, prune, backupRetention, encryption)
		case "ssh":
			sshBackup(backupFileName, s3Path, disableCompression, prune, backupRetention, encryption)
		case "ftp":
			fmt.Println("x is 3")
		default:
			localBackup(backupFileName, disableCompression, prune, backupRetention, encryption)
		}

	} else if executionMode == "scheduled" {
		scheduledMode()
	} else {
		utils.Fatal("Error, unknown execution mode!")
	}

}

// Run in scheduled mode
func scheduledMode() {

	fmt.Println()
	fmt.Println("**********************************")
	fmt.Println("     Starting PostgreSQL Bkup...   ")
	fmt.Println("***********************************")
	utils.Info("Running in Scheduled mode")
	utils.Info("Execution period ", os.Getenv("SCHEDULE_PERIOD"))

	//Test database connexion
	utils.TestDatabaseConnection()

	utils.Info("Creating backup job...")
	CreateCrontabScript(disableCompression, storage)

	supervisorConfig := "/etc/supervisor/supervisord.conf"

	// Start Supervisor
	cmd := exec.Command("supervisord", "-c", supervisorConfig)
	err := cmd.Start()
	if err != nil {
		utils.Fatal("Failed to start supervisord: %v", err)
	}
	utils.Info("Backup job started")
	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			utils.Info("Failed to kill supervisord process: %v", err)
		} else {
			utils.Info("Supervisor stopped.")
		}
	}()
	if _, err := os.Stat(cronLogFile); os.IsNotExist(err) {
		utils.Fatal("Log file %s does not exist.", cronLogFile)
	}
	t, err := tail.TailFile(cronLogFile, tail.Config{Follow: true})
	if err != nil {
		utils.Fatalf("Failed to tail file: %v", err)
	}

	// Read and print new lines from the log file
	for line := range t.Lines {
		fmt.Println(line.Text)
	}
}

// BackupDatabase backup database
func BackupDatabase(backupFileName string, disableCompression bool) {
	dbHost = os.Getenv("DB_HOST")
	dbPassword = os.Getenv("DB_PASSWORD")
	dbUserName = os.Getenv("DB_USERNAME")
	dbName = os.Getenv("DB_NAME")
	dbPort = os.Getenv("DB_PORT")
	storagePath = os.Getenv("STORAGE_PATH")

	utils.Info("Starting database backup...")
	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_NAME") == "" || os.Getenv("DB_USERNAME") == "" || os.Getenv("DB_PASSWORD") == "" {
		utils.Fatal("Please make sure all required environment variables for database are set")
	} else {
		err := os.Setenv("PGPASSWORD", dbPassword)
		if err != nil {
			return
		}
		utils.TestDatabaseConnection()
		// Backup Database database
		utils.Info("Backing up database...")

		// Verify is compression is disabled
		if disableCompression {
			// Execute pg_dump
			cmd := exec.Command("pg_dump",
				"-h", dbHost,
				"-p", dbPort,
				"-U", dbUserName,
				"-d", dbName,
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
				"-h", dbHost,
				"-p", dbPort,
				"-U", dbUserName,
				"-d", dbName,
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

}
func localBackup(backupFileName string, disableCompression bool, prune bool, backupRetention int, encrypt bool) {
	utils.Info("Backup database to local storage")
	BackupDatabase(backupFileName, disableCompression)
	finalFileName := backupFileName
	if encrypt {
		encryptBackup(backupFileName)
		finalFileName = fmt.Sprintf("%s.%s", backupFileName, gpgExtension)
	}
	utils.Info("Backup name is ", finalFileName)
	moveToBackup(finalFileName, storagePath)
	//Delete old backup
	if prune {
		deleteOldBackup(backupRetention)
	}
}

func s3Backup(backupFileName string, s3Path string, disableCompression bool, prune bool, backupRetention int, encrypt bool) {
	bucket := utils.GetEnvVariable("AWS_S3_BUCKET_NAME", "BUCKET_NAME")
	utils.Info("Backup database to s3 storage")
	//Backup database
	BackupDatabase(backupFileName, disableCompression)
	finalFileName := backupFileName
	if encrypt {
		encryptBackup(backupFileName)
		finalFileName = fmt.Sprintf("%s.%s", backupFileName, "gpg")
	}
	utils.Info("Uploading backup file to S3 storage...")
	utils.Info("Backup name is ", backupFileName)
	err := utils.UploadFileToS3(tmpPath, finalFileName, bucket, s3Path)
	if err != nil {
		utils.Fatalf("Error uploading file to S3: %s ", err)

	}

	//Delete backup file from tmp folder
	err = utils.DeleteFile(filepath.Join(tmpPath, backupFileName))
	if err != nil {
		fmt.Println("Error deleting file:", err)

	}
	// Delete old backup
	if prune {
		err := utils.DeleteOldBackup(bucket, s3Path, backupRetention)
		if err != nil {
			utils.Fatalf("Error deleting old backup from S3: %s ", err)
		}
	}
	utils.Done("Database has been backed up and uploaded to s3 ")
}
func sshBackup(backupFileName string, s3Path string, disableCompression bool, prune bool, backupRetention int, encrypt bool) {

}

func encryptBackup(backupFileName string) {
	gpgPassphrase := os.Getenv("GPG_PASSPHRASE")

	err := Encrypt(filepath.Join(tmpPath, backupFileName), gpgPassphrase)
	if err != nil {
		utils.Fatalf("Error during encrypting backup %s", err)
	}

}
