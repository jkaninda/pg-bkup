/*
MIT License

Copyright (c) 2023 Jonas Kaninda

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package pkg

import (
	"fmt"
	"github.com/jkaninda/encryptor"
	"github.com/jkaninda/go-storage/pkg/local"
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
	// Initialize backup configs
	config := initBackupConfig(cmd)
	// Load backup configuration file
	configFile, err := loadConfigFile()
	if err != nil {
		dbConf = initDbConfig(cmd)
		if config.cronExpression == "" {
			BackupTask(dbConf, config)
		} else {
			if utils.IsValidCronExpression(config.cronExpression) {
				scheduledMode(dbConf, config)
			} else {
				utils.Fatal("Cron expression is not valid: %s", config.cronExpression)
			}
		}
	} else {
		startMultiBackup(config, configFile)
	}

}

// scheduledMode Runs backup in scheduled mode
func scheduledMode(db *dbConfig, config *BackupConfig) {
	utils.Info("Running in Scheduled mode")
	utils.Info("Backup cron expression:  %s", config.cronExpression)
	utils.Info("The next scheduled time is: %v", utils.CronNextTime(config.cronExpression).Format(timeFormat))
	utils.Info("Storage type %s ", config.storage)

	// Test backup
	utils.Info("Testing backup configurations...")
	testDatabaseConnection(db)
	utils.Info("Testing backup configurations...done")
	utils.Info("Creating backup job...")
	// Create a new cron instance
	c := cron.New()

	_, err := c.AddFunc(config.cronExpression, func() {
		BackupTask(db, config)
		utils.Info("Next backup time is: %v", utils.CronNextTime(config.cronExpression).Format(timeFormat))

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

// multiBackupTask backup multi database
func multiBackupTask(databases []Database, bkConfig *BackupConfig) {
	for _, db := range databases {
		// Check if path is defined in config file
		if db.Path != "" {
			bkConfig.remotePath = db.Path
		}
		BackupTask(getDatabase(db), bkConfig)
	}
}

// BackupTask backups database
func BackupTask(db *dbConfig, config *BackupConfig) {
	utils.Info("Starting backup task...")
	// Generate file name
	backupFileName := fmt.Sprintf("%s_%s.sql.gz", db.dbName, time.Now().Format("20060102_150405"))
	if config.disableCompression {
		backupFileName = fmt.Sprintf("%s_%s.sql", db.dbName, time.Now().Format("20060102_150405"))
	}
	config.backupFileName = backupFileName
	switch config.storage {
	case "local":
		localBackup(db, config)
	case "s3", "S3":
		s3Backup(db, config)
	case "ssh", "SSH", "remote":
		sshBackup(db, config)
	case "ftp", "FTP":
		ftpBackup(db, config)
	case "azure":
		azureBackup(db, config)
	default:
		localBackup(db, config)
	}
}
func startMultiBackup(bkConfig *BackupConfig, configFile string) {
	utils.Info("Starting Multi backup task...")
	conf, err := readConf(configFile)
	if err != nil {
		utils.Fatal("Error reading config file: %s", err)
	}
	// Check if cronExpression is defined in config file
	if conf.CronExpression != "" {
		bkConfig.cronExpression = conf.CronExpression
	}
	if len(conf.Databases) == 0 {
		utils.Fatal("No databases found")
	}
	// Check if cronExpression is defined
	if bkConfig.cronExpression == "" {
		multiBackupTask(conf.Databases, bkConfig)
	} else {
		// Check if cronExpression is valid
		if utils.IsValidCronExpression(bkConfig.cronExpression) {
			utils.Info("Running backup in Scheduled mode")
			utils.Info("Backup cron expression:  %s", bkConfig.cronExpression)
			utils.Info("The next scheduled time is: %v", utils.CronNextTime(bkConfig.cronExpression).Format(timeFormat))
			utils.Info("Storage type %s ", bkConfig.storage)

			// Test backup
			utils.Info("Testing backup configurations...")
			for _, db := range conf.Databases {
				testDatabaseConnection(getDatabase(db))
			}
			utils.Info("Testing backup configurations...done")
			utils.Info("Creating backup job...")
			// Create a new cron instance
			c := cron.New()

			_, err := c.AddFunc(bkConfig.cronExpression, func() {
				multiBackupTask(conf.Databases, bkConfig)
				utils.Info("Next backup time is: %v", utils.CronNextTime(bkConfig.cronExpression).Format(timeFormat))

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

		} else {
			utils.Fatal("Cron expression is not valid: %s", bkConfig.cronExpression)
		}
	}

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
		file, err := os.Create(filepath.Join(tmpPath, backupFileName))
		if err != nil {
			log.Fatal(err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				return

			}
		}(file)

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
		gzipCmd.Stdout, err = os.Create(filepath.Join(tmpPath, backupFileName))
		err2 := gzipCmd.Start()
		if err2 != nil {
			return
		}
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
func localBackup(db *dbConfig, config *BackupConfig) {
	utils.Info("Backup database to local storage")
	startTime = time.Now().Format(utils.TimeFormat())
	BackupDatabase(db, config.backupFileName, disableCompression)
	finalFileName := config.backupFileName
	if config.encryption {
		encryptBackup(config)
		finalFileName = fmt.Sprintf("%s.%s", config.backupFileName, gpgExtension)
	}
	fileInfo, err := os.Stat(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		utils.Error("Error: %s", err)
	}
	backupSize = fileInfo.Size()
	localStorage := local.NewStorage(local.Config{
		LocalPath:  tmpPath,
		RemotePath: storagePath,
	})
	err = localStorage.Copy(finalFileName)
	if err != nil {
		utils.Fatal("Error copying backup file: %s", err)
	}
	utils.Info("Backup name is %s", finalFileName)
	utils.Info("Backup size: %s", utils.ConvertBytes(uint64(backupSize)))
	utils.Info("Backup saved in %s", filepath.Join(storagePath, finalFileName))
	// Send notification
	utils.NotifySuccess(&utils.NotificationData{
		File:           finalFileName,
		BackupSize:     utils.ConvertBytes(uint64(backupSize)),
		Database:       db.dbName,
		Storage:        config.storage,
		BackupLocation: filepath.Join(storagePath, finalFileName),
		StartTime:      startTime,
		EndTime:        time.Now().Format(utils.TimeFormat()),
	})
	// Delete old backup
	if config.prune {
		err = localStorage.Prune(config.backupRetention)
		if err != nil {
			utils.Fatal("Error deleting old backup from %s storage: %s ", config.storage, err)
		}

	}
	// Delete temp
	deleteTemp()
	utils.Info("Backup completed successfully")
}
func encryptBackup(config *BackupConfig) {
	backupFile, err := os.ReadFile(filepath.Join(tmpPath, config.backupFileName))
	outputFile := fmt.Sprintf("%s.%s", filepath.Join(tmpPath, config.backupFileName), gpgExtension)
	if err != nil {
		utils.Fatal("Error reading backup file: %s ", err)
	}
	if config.usingKey {
		utils.Info("Encrypting backup using public key...")
		pubKey, err := os.ReadFile(config.publicKey)
		if err != nil {
			utils.Fatal("Error reading public key: %s ", err)
		}
		err = encryptor.EncryptWithPublicKey(backupFile, fmt.Sprintf("%s.%s", filepath.Join(tmpPath, config.backupFileName), gpgExtension), pubKey)
		if err != nil {
			utils.Fatal("Error encrypting backup file: %v ", err)
		}
		utils.Info("Encrypting backup using public key...done")

	} else if config.passphrase != "" {
		utils.Info("Encrypting backup using passphrase...")
		err := encryptor.Encrypt(backupFile, outputFile, config.passphrase)
		if err != nil {
			utils.Fatal("error during encrypting backup %v", err)
		}
		utils.Info("Encrypting backup using passphrase...done")

	}

}
