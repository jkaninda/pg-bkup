// Package pkg /*
/*
Copyright © 2024 Jonas Kaninda  <jonaskaninda.gmail.com>
*/
package pkg

import (
	"fmt"
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
	keepLast, _ := cmd.Flags().GetInt("keep-last")
	prune, _ := cmd.Flags().GetBool("prune")
	disableCompression, _ = cmd.Flags().GetBool("disable-compression")
	executionMode, _ = cmd.Flags().GetString("mode")

	if executionMode == "default" {
		if storage == "s3" {
			utils.Info("Backup database to s3 storage")
			s3Backup(disableCompression, s3Path, prune, keepLast)
		} else {
			utils.Info("Backup database to local storage")
			BackupDatabase(disableCompression, prune, keepLast)

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
	utils.Info("Log file in /var/log/pg-bkup.log")
	utils.Info("Execution period ", os.Getenv("SCHEDULE_PERIOD"))

	//Test database connexion
	utils.TestDatabaseConnection()

	//Test database connexion
	utils.TestDatabaseConnection()

	utils.Info("Creating backup job...")
	CreateCrontabScript(disableCompression, storage)

	//Start Supervisor
	supervisordCmd := exec.Command("supervisord", "-c", "/etc/supervisor/supervisord.conf")
	if err := supervisordCmd.Run(); err != nil {
		utils.Fatalf("Error starting supervisord: %v\n", err)
	}
}

// BackupDatabase backup database
func BackupDatabase(disableCompression bool, prune bool, keepLast int) {
	dbHost = os.Getenv("DB_HOST")
	dbPassword = os.Getenv("DB_PASSWORD")
	dbUserName = os.Getenv("DB_USERNAME")
	dbName = os.Getenv("DB_NAME")
	dbPort = os.Getenv("DB_PORT")
	storagePath = os.Getenv("STORAGE_PATH")

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
		//Generate file name
		bkFileName := fmt.Sprintf("%s_%s.sql.gz", dbName, time.Now().Format("20060102_150405"))

		// Verify is compression is disabled
		if disableCompression {
			bkFileName = fmt.Sprintf("%s_%s.sql", dbName, time.Now().Format("20060102_150405"))
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
			file, err := os.Create(fmt.Sprintf("%s/%s", storagePath, bkFileName))
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			_, err = file.Write(output)
			if err != nil {
				log.Fatal(err)
			}
			utils.Done("Database has been backed up")

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
			gzipCmd.Stdout, err = os.Create(fmt.Sprintf("%s/%s", storagePath, bkFileName))
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
			utils.Done("Database has been backed up")

		}
		//Delete old backup
		if prune {
			deleteOldBackup(keepLast)
		}

		historyFile, err := os.OpenFile(fmt.Sprintf("%s/history.txt", storagePath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer historyFile.Close()
		if _, err := historyFile.WriteString(bkFileName + "\n"); err != nil {
			log.Fatal(err)
		}
	}

}

func s3Backup(disableCompression bool, s3Path string, prune bool, keepLast int) {
	// Backup Database to S3 storage
	MountS3Storage(s3Path)
	BackupDatabase(disableCompression, prune, keepLast)
}
func deleteOldBackup(keepLast int) {
	utils.Info("Deleting old backups...")
	storagePath = os.Getenv("STORAGE_PATH")
	// Define the directory path
	backupDir := storagePath + "/"
	// Get current time
	currentTime := time.Now()
	// Delete file
	deleteFile := func(filePath string) error {
		err := os.Remove(filePath)
		if err != nil {
			utils.Fatal("Error:", err)
		} else {
			utils.Done("File ", filePath, " deleted successfully")
		}
		return err
	}

	// Walk through the directory and delete files modified more than specified days ago
	err := filepath.Walk(backupDir, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if it's a regular file and if it was modified more than specified days ago
		if fileInfo.Mode().IsRegular() {
			timeDiff := currentTime.Sub(fileInfo.ModTime())
			if timeDiff.Hours() > 24*float64(keepLast) {
				err := deleteFile(filePath)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		utils.Fatal("Error:", err)
		return
	}
}
