package pkg

import (
	"fmt"
	"github.com/jkaninda/go-storage/pkg/azure"
	"github.com/jkaninda/pg-bkup/utils"
	"os"
	"path/filepath"
	"time"
)

func azureBackup(db *dbConfig, config *BackupConfig) {
	utils.Info("Backup database to Azure Blob Storage")
	startTime = time.Now().Format(utils.TimeFormat())

	// Backup database
	BackupDatabase(db, config.backupFileName, disableCompression)
	finalFileName := config.backupFileName
	if config.encryption {
		encryptBackup(config)
		finalFileName = fmt.Sprintf("%s.%s", config.backupFileName, "gpg")
	}
	utils.Info("Uploading backup archive to Azure Blob storage ...")
	utils.Info("Backup name is %s", finalFileName)
	azureConfig := loadAzureConfig()
	azureStorage, err := azure.NewStorage(azure.Config{
		ContainerName: azureConfig.containerName,
		AccountName:   azureConfig.accountName,
		AccountKey:    azureConfig.accountKey,
		RemotePath:    config.remotePath,
		LocalPath:     tmpPath,
	})
	if err != nil {
		utils.Fatal("Error creating Azure Blob storage: %s", err)
	}
	err = azureStorage.Copy(finalFileName)
	if err != nil {
		utils.Fatal("Error copying backup file: %s", err)
	}
	utils.Info("Backup saved in %s", filepath.Join(config.remotePath, finalFileName))
	// Get backup info
	fileInfo, err := os.Stat(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		utils.Error("Error: %s", err)
	}
	backupSize = fileInfo.Size()
	// Delete backup file from tmp folder
	err = utils.DeleteFile(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		utils.Error("Error deleting file: %v", err)

	}
	if config.prune {
		err := azureStorage.Prune(config.backupRetention)
		if err != nil {
			utils.Fatal("Error deleting old backup from %s storage: %s ", config.storage, err)
		}

	}

	utils.Info("Uploading backup archive to Azure Blob storage ... done ")

	// Send notification
	utils.NotifySuccess(&utils.NotificationData{
		File:           finalFileName,
		BackupSize:     utils.BytesToMb(uint64(backupSize)),
		Database:       db.dbName,
		Storage:        config.storage,
		BackupLocation: filepath.Join(config.remotePath, finalFileName),
		StartTime:      startTime,
		EndTime:        time.Now().Format(utils.TimeFormat()),
	})
	// Delete temp
	deleteTemp()
	utils.Info("Backup completed successfully")
}
func azureRestore(db *dbConfig, conf *RestoreConfig) {
	utils.Info("Restore database from Azure Blob storage")
	azureConfig := loadAzureConfig()
	azureStorage, err := azure.NewStorage(azure.Config{
		ContainerName: azureConfig.containerName,
		AccountName:   azureConfig.accountName,
		AccountKey:    azureConfig.accountKey,
		RemotePath:    conf.remotePath,
		LocalPath:     tmpPath,
	})
	if err != nil {
		utils.Fatal("Error creating SSH storage: %s", err)
	}

	err = azureStorage.CopyFrom(conf.file)
	if err != nil {
		utils.Fatal("Error downloading backup file: %s", err)
	}
	RestoreDatabase(db, conf)
}
