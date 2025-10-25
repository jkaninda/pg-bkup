/*
 *  MIT License
 *
 * Copyright (c) 2024 Jonas Kaninda
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 */

package pkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jkaninda/encryptor"
	"github.com/jkaninda/go-storage/pkg/local"
	goutils "github.com/jkaninda/go-utils"
	"github.com/jkaninda/logger"
	"github.com/jkaninda/pg-bkup/utils"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func StartBackup(cmd *cobra.Command) {
	intro()
	logger.Info("Starting backup process")
	logger.Info("Loading configuration")
	// Initialize backup configs
	config := initBackupConfig(cmd)
	// Load backup configuration file
	configFile, err := loadConfigFile()
	if err != nil {
		dbConf = initDbConfig(cmd)
		if config.cronExpression == "" {
			config.allowCustomName = true
			createBackupTask(dbConf, config)
		} else {
			if utils.IsValidCronExpression(config.cronExpression) {
				scheduledMode(dbConf, config)
			} else {
				logger.Fatal("Cron expression is not valid", "expression", config.cronExpression)
			}
		}
	} else {
		startMultiBackup(config, configFile)
	}

}

// scheduledMode Runs backup in scheduled mode
func scheduledMode(db *dbConfig, config *BackupConfig) {
	logger.Info("Running in Scheduled mode", "cron", config.cronExpression)
	logger.Info(fmt.Sprintf("The next scheduled time is: %v", utils.CronNextTime(config.cronExpression).Format(timeFormat)))
	logger.Info(fmt.Sprintf("Storage type %s ", config.storage))

	// Test backup
	logger.Info("Testing backup configurations...")
	err := testDatabaseConnection(db)
	if err != nil {
		logger.Error("Error connecting to database", "database", db.dbName)
		logger.Fatal("Error testing database connection", "error", err)
	}
	logger.Info("Testing backup configurations...done")
	logger.Info("Creating backup task", "database", db.dbName, "storage", config.storage)
	// Create a new cron instance
	c := cron.New()

	_, err = c.AddFunc(config.cronExpression, func() {
		createBackupTask(db, config)
		logger.Info("Backup task executed successfully; awaiting next scheduled time", "next_time", utils.CronNextTime(config.cronExpression).Format(timeFormat))
	})
	if err != nil {
		return
	}
	// Start the cron scheduler
	c.Start()
	logger.Info("Creating backup task...done")
	logger.Info("Backup task started")
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
		createBackupTask(getDatabase(db), bkConfig)
	}
}

// createBackupTask backup task
func createBackupTask(db *dbConfig, config *BackupConfig) {
	if config.all && !config.allInOne {
		backupAll(db, config)
	} else {
		if db.dbName == "" && !config.all {
			logger.Fatal("Database name is required, use DB_NAME environment variable or -d flag")
		}
		backupTask(db, config)
	}
}

// backupAll backup all databases
func backupAll(db *dbConfig, config *BackupConfig) {
	databases, err := listDatabases(*db)
	if err != nil {
		logger.Fatal("Error listing databases", "error", err)
	}
	logger.Info("Backing up all databases", "count", len(databases))
	for _, dbName := range databases {
		db.dbName = dbName
		config.backupFileName = fmt.Sprintf("%s_%s.sql.gz", dbName, time.Now().Format("20060102_150405"))
		backupTask(db, config)
	}

}

// backupTask handles database backup tasks based on the provided configuration.
func backupTask(db *dbConfig, config *BackupConfig) {
	logger.Info(
		"Initiating backup task",
		"database", db.dbName,
		"storage", config.storage,
		"compression", !config.disableCompression,
	)
	startTime = time.Now()
	// Determine file name prefix
	prefix := db.dbName
	if config.all && config.allInOne {
		prefix = "all_databases"
	}

	// Build backup filename
	timestamp := time.Now().Format("20060102_150405")
	config.backupFileName = generateBackupFileName(prefix, timestamp, config)

	// Storage handler
	switch config.storage {
	case LocalStorage:
		localBackup(db, config)
	case S3Storage:
		s3Backup(db, config)
	case SFTPStorage, SSHStorage, RemoteStorage:
		sshBackup(db, config)
	case FTPStorage:
		ftpBackup(db, config)
	case AzureStorage:
		azureBackup(db, config)
	default:
		localBackup(db, config)
	}
}

// generateBackupFileName creates the backup file name based on the configuration.
func generateBackupFileName(prefix, timestamp string, config *BackupConfig) string {
	var name string
	switch {
	case config.schemaOnly:
		config.disableCompression = true
		name = fmt.Sprintf("%s_schema_%s", prefix, timestamp)

	case len(config.tables) > 0:
		config.disableCompression = true
		name = fmt.Sprintf("%s_tables_%d_%s", prefix, len(config.tables), timestamp)

	case config.customName != "" && config.allowCustomName && !config.all:
		name = config.customName

	default:
		name = fmt.Sprintf("%s_%s", prefix, timestamp)
	}

	ext := ".sql"
	if !config.disableCompression {
		ext += ".gz"
	}

	return name + ext
}

// startMultiBackup start multi backup
func startMultiBackup(bkConfig *BackupConfig, configFile string) {
	logger.Info("Starting Multi backup task...")
	conf, err := readConf(configFile)
	if err != nil {
		logger.Fatal("Error reading config file", "error", err)
	}
	// Check if cronExpression is defined in config file
	if conf.CronExpression != "" {
		bkConfig.cronExpression = conf.CronExpression
	}
	if len(conf.Databases) == 0 {
		logger.Fatal("No databases found")
	}
	// Check if cronExpression is defined
	if bkConfig.cronExpression == "" {
		multiBackupTask(conf.Databases, bkConfig)
	} else {
		backupRescueMode = conf.BackupRescueMode
		// Check if cronExpression is valid
		if utils.IsValidCronExpression(bkConfig.cronExpression) {
			logger.Info("Running in Scheduled mode", "cron", bkConfig.cronExpression)
			logger.Info(fmt.Sprintf("The next scheduled time is: %v", utils.CronNextTime(bkConfig.cronExpression).Format(timeFormat)))
			logger.Info(fmt.Sprintf("Storage type %s ", bkConfig.storage))

			// Test backup
			logger.Info("Testing backup configurations...")
			for _, db := range conf.Databases {
				err = testDatabaseConnection(getDatabase(db))
				if err != nil {
					recoverMode(err, fmt.Sprintf("Error connecting to database: %s", db.Name))
					continue
				}
			}
			logger.Info("Testing backup configurations...done")
			logger.Info("Creating backup job...")
			// Create a new cron instance
			c := cron.New()

			_, err := c.AddFunc(bkConfig.cronExpression, func() {
				multiBackupTask(conf.Databases, bkConfig)
				logger.Info("Next scheduled time", "time", utils.CronNextTime(bkConfig.cronExpression).Format(timeFormat))

			})
			if err != nil {
				return
			}
			// Start the cron scheduler
			c.Start()
			logger.Info("Creating backup job...done")
			logger.Info("Backup job started")
			defer c.Stop()
			select {}

		} else {
			logger.Fatal("Cron expression is not valid", "cron", bkConfig.cronExpression)
		}
	}

}

// BackupDatabase backs up the database, selected tables, or schema only.
func BackupDatabase(db *dbConfig, config *BackupConfig) error {
	storagePath = os.Getenv("STORAGE_PATH")

	if err := testDatabaseConnection(db); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	var (
		dumpCmd  string
		dumpArgs []string
	)

	dumpArgs = []string{
		"-h", db.dbHost,
		"-p", db.dbPort,
		"-U", db.dbUserName,
	}

	if config.all && config.allInOne {
		logger.Info("Backing up all databases...")
		dumpCmd = "pg_dumpall"
	} else {
		dumpCmd = "pg_dump"
		dumpArgs = append(dumpArgs, db.dbName)

		if config.schemaOnly {
			dumpArgs = append(dumpArgs, "--schema-only")
			logger.Info(fmt.Sprintf("Backing up schema for database: %s", db.dbName))
		} else if config.dataOnly {
			dumpArgs = append(dumpArgs, "--data-only")
			logger.Info(fmt.Sprintf("Backing up data only for database: %s", db.dbName))
		}

		if len(config.tables) > 0 {
			logger.Info("Backing up specified tables...")
			for _, table := range config.tables {
				dumpArgs = append(dumpArgs, "-t", table)
			}
			logger.Info(fmt.Sprintf("Backing up tables: %v", config.tables))
		} else if !config.schemaOnly && !config.dataOnly {
			logger.Info(fmt.Sprintf("Backing up full database: %s", db.dbName))
		}
	}

	backupPath := filepath.Join(tmpPath, config.backupFileName)

	// Handle compression
	if config.disableCompression {
		return runCommandAndSaveOutput(dumpCmd, dumpArgs, backupPath)
	}
	return runCommandWithCompression(dumpCmd, dumpArgs, backupPath)
}

// runCommandAndSaveOutput runs a command and saves the output to a file
func runCommandAndSaveOutput(command string, args []string, outputPath string) error {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute %s: %v, output: %s", command, err, output)
	}

	return os.WriteFile(outputPath, output, 0644)
}

// runCommandWithCompression runs a command and compresses the output
func runCommandWithCompression(command string, args []string, outputPath string) error {
	cmd := exec.Command(command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	gzipCmd := exec.Command("gzip")
	gzipCmd.Stdin = stdout
	gzipFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create gzip file: %w", err)
	}
	defer func(gzipFile *os.File) {
		err = gzipFile.Close()
		if err != nil {
			logger.Error("Error closing gzip file", "error", err)
		}
	}(gzipFile)
	gzipCmd.Stdout = gzipFile

	if err = gzipCmd.Start(); err != nil {
		return fmt.Errorf("failed to start gzip: %w", err)
	}
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute %s: %w", command, err)
	}
	if err = gzipCmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait for gzip completion: %w", err)
	}

	logger.Info("Database has been backed up")
	return nil
}

// localBackup backup database to local storage
func localBackup(db *dbConfig, config *BackupConfig) {
	logger.Info("Backup database to local storage")
	err := BackupDatabase(db, config)
	if err != nil {
		recoverMode(err, "Error backing up database")
		return
	}
	finalFileName := config.backupFileName
	if config.encryption {
		encryptBackup(config)
		finalFileName = fmt.Sprintf("%s.%s", config.backupFileName, gpgExtension)
	}
	fileInfo, err := os.Stat(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		logger.Error("Error getting backup info", "error", err)
	}
	backupSize = fileInfo.Size()
	localStorage := local.NewStorage(local.Config{
		LocalPath:  tmpPath,
		RemotePath: storagePath,
	})
	err = localStorage.Copy(finalFileName)
	if err != nil {
		logger.Fatal("Error copying backup file", "error", err)
	}

	duration := goutils.FormatDuration(time.Since(startTime), 0)
	logger.Info("Backup file copied to local storage", "file", finalFileName, "destination", storagePath)
	logger.Info("Backup completed", "file", finalFileName, "size", goutils.ConvertBytes(uint64(backupSize)), "duration", duration)

	// Send notification
	utils.NotifySuccess(&utils.NotificationData{
		File:           finalFileName,
		BackupSize:     goutils.ConvertBytes(uint64(backupSize)),
		Database:       db.dbName,
		Storage:        string(config.storage),
		BackupLocation: filepath.Join(storagePath, finalFileName),
		Duration:       duration,
	})
	// Delete old backup
	if config.prune {
		err = localStorage.Prune(config.backupRetention)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Error deleting old backup from %s storage: %s ", config.storage, err))
		}

	}
	// Delete temp
	deleteTemp()
	logger.Info(fmt.Sprintf("The backup of the %s database has been completed in %s", db.dbName, duration))
}

// encryptBackup encrypt backup
func encryptBackup(config *BackupConfig) {
	logger.Info("Starting backup encryption", "file", config.backupFileName)
	backupFile, err := os.ReadFile(filepath.Join(tmpPath, config.backupFileName))
	outputFile := fmt.Sprintf("%s.%s", filepath.Join(tmpPath, config.backupFileName), gpgExtension)
	if err != nil {
		logger.Fatal("Error reading backup file", "error", err)
	}
	if config.usingKey {
		logger.Info("Encrypting backup using public key...")
		pubKey, err := os.ReadFile(config.publicKey)
		if err != nil {
			logger.Fatal("Error reading public key", "error", err)
		}
		err = encryptor.EncryptWithPublicKey(backupFile, fmt.Sprintf("%s.%s", filepath.Join(tmpPath, config.backupFileName), gpgExtension), pubKey)
		if err != nil {
			logger.Fatal("Error encrypting backup file", "error", err)
		}
		logger.Info("Encrypting backup using public key...done")

	} else if config.passphrase != "" {
		logger.Info("Encrypting backup using passphrase...")
		err := encryptor.Encrypt(backupFile, outputFile, config.passphrase)
		if err != nil {
			logger.Fatal("error during encrypting backup", "error", err)
		}
		logger.Info("Encrypting backup using passphrase...done")

	}
	logger.Info("Encryption completed", "output", outputFile)

}

// listDatabases lists all databases in the PostgreSQL server
func listDatabases(db dbConfig) ([]string, error) {
	databases := []string{}

	logger.Info("Listing databases...")
	// Create the PostgresSQL client config file
	if err := createPGConfigFile(db); err != nil {
		return databases, errors.New(err.Error())
	}
	// Construct the connection string
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable", db.dbUserName, db.dbPassword, db.dbHost, db.dbPort)

	// Connect to the PostgreSQL server
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer func(conn *pgx.Conn, ctx context.Context) {
		err = conn.Close(ctx)
		if err != nil {
			logger.Error("Error closing connexion", "error", err)
		}
	}(conn, context.Background())

	// Query to list all non-template databases
	query := `SELECT datname 
         FROM pg_database 
         WHERE datistemplate = false 
           AND datname NOT IN ('postgres', 'template0', 'template1')`

	// Execute the query to list databases
	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Collect database names from the result
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, fmt.Errorf("failed to scan database name: %w", err)
		}
		databases = append(databases, dbName)
	}

	// Check for errors during iteration
	if rows.Err() != nil {
		return nil, fmt.Errorf("error during row iteration: %w", rows.Err())
	}

	logger.Info("Found databases", "count", len(databases))
	return databases, nil

}
func recoverMode(err error, msg string) {
	if err != nil {
		if backupRescueMode {
			utils.NotifyError(fmt.Sprintf("%s : %v", msg, err))
			logger.Error("Backup failed", "reason", msg, "error", err)
			logger.Warn("Backup rescue mode is enabled,Backup will continue")
		} else {
			logger.Error("Backup failed", "reason", msg, "error", err)
			logger.Fatal("An occurred error", "error", err)
			return
		}
	}

}
