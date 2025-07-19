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
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jkaninda/logger"
	"github.com/spf13/cobra"
	"time"
)

func StartMigration(cmd *cobra.Command) {
	intro()
	logger.Info("Starting database migration task...")
	all, _ := cmd.Flags().GetBool("all-databases")
	instance, _ := cmd.Flags().GetBool("entire-instance")

	// Get DB config
	dbConf = initDbConfig(cmd)
	targetDbConf = initTargetDbConfig()

	if targetDbConf.targetDbName == "" && !all {
		logger.Fatal("Target database name is required, use TARGET_DB_NAME environment variable")
	}

	// Defining the target database variables
	newDbConfig := dbConfig{}
	newDbConfig.dbHost = targetDbConf.targetDbHost
	newDbConfig.dbPort = targetDbConf.targetDbPort
	newDbConfig.dbName = targetDbConf.targetDbName
	newDbConfig.dbUserName = targetDbConf.targetDbUserName
	newDbConfig.dbPassword = targetDbConf.targetDbPassword

	if all {
		migrateAllDatabases(dbConf, &newDbConfig)
	} else if instance {
		migrate(dbConf, &newDbConfig, true)

	} else {
		migrate(dbConf, &newDbConfig, false)
	}
	logger.Info("Database migration process finished successfully.")

}

func migrate(dbConf, targetDb *dbConfig, allInstance bool) {
	// Generate a timestamped backup file name
	backupFileName := fmt.Sprintf("%s_%s.sql", dbConf.dbName, time.Now().Format("20060102_150405"))
	conf := &RestoreConfig{file: backupFileName}

	// Backup the source database
	logger.Info(fmt.Sprintf("Starting backup for database [%s]...", dbConf.dbName))
	err := BackupDatabase(dbConf, backupFileName, true, allInstance, allInstance)
	if err != nil {
		logger.Fatal("Failed to back up database", "name", dbConf.dbName, "error", err)
	}

	logger.Info("Backup completed", "filename", backupFileName)

	// Restore the backup into the target database
	logger.Info(fmt.Sprintf("Starting restoration: [%s] → [%s]...", dbConf.dbName, targetDb.dbName))
	RestoreDatabase(targetDb, conf)
	logger.Info(fmt.Sprintf("Restoration completed: [%s] successfully migrated to [%s]", dbConf.dbName, targetDb.dbName))

}

func migrateAllDatabases(dbConf, targetDb *dbConfig) {
	databases, err := listDatabases(*dbConf)
	if err != nil {
		logger.Fatal("Error listing databases", "error", err)
	}

	for _, dbName := range databases {
		dbConf.dbName = dbName
		targetDb.dbName = dbName

		exists, err := targetDb.databaseExists()
		if err != nil {
			logger.Fatal("Error checking database existence", "error", err)
		}

		if !exists {
			logger.Info(fmt.Sprintf("Database [%s] does not exist, creating...", dbName))
			if err := targetDb.createDatabase(); err != nil {
				logger.Fatal("Error creating database", "error", err)
			}
		} else {
			logger.Info(fmt.Sprintf("Database [%s] already exists, skipping creation...", dbName))
		}

		migrate(dbConf, targetDb, false)
	}
	logger.Info("All databases have been migrated.")
}

func (db *dbConfig) databaseExists() (bool, error) {
	adminDb := *db
	adminDb.dbName = "postgres" // Connect to default "postgres"
	dbConn, err := dbConnect(&adminDb)
	if err != nil {
		return false, fmt.Errorf("error connecting to the database: %w", err)
	}
	defer func(dbConn *pgx.Conn, ctx context.Context) {
		err := dbConn.Close(ctx)
		if err != nil {
			logger.Error("Error closing connection", "error", err)
		}
	}(dbConn, context.Background())

	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)"
	err = dbConn.QueryRow(context.Background(), query, db.dbName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error querying database existence: %w", err)
	}

	return exists, nil
}

func (db *dbConfig) createDatabase() error {
	adminDb := *db
	adminDb.dbName = "postgres" // Connect to default "postgres" database to create a new one

	dbConn, err := dbConnect(&adminDb)
	if err != nil {
		return fmt.Errorf("error connecting to create database: %w", err)
	}
	defer func(dbConn *pgx.Conn, ctx context.Context) {
		err := dbConn.Close(ctx)
		if err != nil {
			logger.Error("Error closing connection", "error", err)
		}
	}(dbConn, context.Background())

	_, err = dbConn.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE \"%s\"", db.dbName))
	if err != nil {
		return fmt.Errorf("error creating database: %w", err)
	}
	return nil
}

func dbConnect(db *dbConfig) (*pgx.Conn, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", db.dbUserName, db.dbPassword, db.dbHost, db.dbPort, db.dbName)
	return pgx.Connect(context.Background(), connString)
}
