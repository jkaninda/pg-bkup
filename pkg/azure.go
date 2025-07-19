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
	"fmt"
	"github.com/jkaninda/go-storage/pkg/azure"
	goutils "github.com/jkaninda/go-utils"
	"github.com/jkaninda/logger"
	"github.com/jkaninda/pg-bkup/utils"

	"os"
	"path/filepath"
	"time"
)

func azureBackup(db *dbConfig, config *BackupConfig) {
	logger.Info("Backup database to Azure Blob Storage")

	// Backup database
	err := BackupDatabase(db, config.backupFileName, disableCompression, config.all, config.allInOne)
	if err != nil {
		recoverMode(err, "Error backing up database")
		return
	}
	finalFileName := config.backupFileName
	if config.encryption {
		encryptBackup(config)
		finalFileName = fmt.Sprintf("%s.%s", config.backupFileName, "gpg")
	}
	logger.Info("Uploading backup archive to Azure Blob storage ...", "filename", finalFileName)
	azureConfig := loadAzureConfig()
	azureStorage, err := azure.NewStorage(azure.Config{
		ContainerName: azureConfig.containerName,
		AccountName:   azureConfig.accountName,
		AccountKey:    azureConfig.accountKey,
		RemotePath:    config.remotePath,
		LocalPath:     tmpPath,
	})
	if err != nil {
		logger.Fatal("Error creating Azure storage", "error", err)
	}
	err = azureStorage.Copy(finalFileName)
	if err != nil {
		logger.Fatal("Error copying backup file", "error", err)
	}
	logger.Info("Backup uploaded", "location", filepath.Join(config.remotePath, finalFileName))
	// Get backup info
	fileInfo, err := os.Stat(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		logger.Error("Error getting backup info", "error", err)
	}
	backupSize = fileInfo.Size()
	// Delete backup file from tmp folder
	err = utils.DeleteFile(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		logger.Error("Error deleting file", "error", err)

	}
	if config.prune {
		err := azureStorage.Prune(config.backupRetention)
		if err != nil {
			logger.Fatal("Error deleting old backup", "storage", config.storage, "error", err)
		}

	}

	duration := goutils.FormatDuration(time.Since(startTime), 0)

	logger.Info("Backup file uploaded to  Azure Blob storage", "file", finalFileName, "destination", storagePath)
	logger.Info("Backup completed", "file", finalFileName, "size", goutils.ConvertBytes(uint64(backupSize)), "duration", duration)

	// Send notification
	utils.NotifySuccess(&utils.NotificationData{
		File:           finalFileName,
		BackupSize:     goutils.ConvertBytes(uint64(backupSize)),
		Database:       db.dbName,
		Storage:        string(config.storage),
		BackupLocation: filepath.Join(config.remotePath, finalFileName),
		Duration:       duration,
	})
	// Delete temp
	deleteTemp()
	logger.Info(fmt.Sprintf("The backup of the %s database has been completed in %s", db.dbName, duration))
}
func azureRestore(db *dbConfig, conf *RestoreConfig) {
	logger.Info("Restore database from Azure Blob storage")
	azureConfig := loadAzureConfig()
	azureStorage, err := azure.NewStorage(azure.Config{
		ContainerName: azureConfig.containerName,
		AccountName:   azureConfig.accountName,
		AccountKey:    azureConfig.accountKey,
		RemotePath:    conf.remotePath,
		LocalPath:     tmpPath,
	})
	if err != nil {
		logger.Fatal("Error creating Azure Blob storage", "error", err)
	}

	err = azureStorage.CopyFrom(conf.file)
	if err != nil {
		logger.Fatal("Error downloading backup file", "error", err)
	}
	RestoreDatabase(db, conf)
}
