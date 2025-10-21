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
	"github.com/jkaninda/go-storage/pkg/ftp"
	"github.com/jkaninda/go-storage/pkg/ssh"
	goutils "github.com/jkaninda/go-utils"
	"github.com/jkaninda/logger"
	"github.com/jkaninda/pg-bkup/utils"

	"os"
	"path/filepath"
	"time"
)

func sshBackup(db *dbConfig, config *BackupConfig) {
	logger.Info("Backup database to Remote server")
	// Backup database
	err := BackupDatabase(db, config)
	if err != nil {
		recoverMode(err, "Error backing up database")
		return
	}
	finalFileName := config.backupFileName
	if config.encryption {
		encryptBackup(config)
		finalFileName = fmt.Sprintf("%s.%s", config.backupFileName, "gpg")
	}
	logger.Info("Uploading backup archive to remote storage ... ")
	sshConfig, err := loadSSHConfig()
	if err != nil {
		logger.Fatal("Error loading ssh config", "error", err)
	}

	sshStorage, err := ssh.NewStorage(ssh.Config{
		Host:         sshConfig.hostName,
		Port:         sshConfig.port,
		User:         sshConfig.user,
		Password:     sshConfig.password,
		IdentifyFile: sshConfig.identifyFile,
		RemotePath:   config.remotePath,
		LocalPath:    tmpPath,
	})
	if err != nil {
		logger.Fatal("Error creating SSH storage", "error", err)
	}
	err = sshStorage.Copy(finalFileName)
	if err != nil {
		logger.Fatal("Error copying backup file", "error", err)
	}
	// Get backup info
	fileInfo, err := os.Stat(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		logger.Error("Error get backup info", "error", err)
	}
	backupSize = fileInfo.Size()
	logger.Info("Backup saved", "location", filepath.Join(config.remotePath, finalFileName))
	logger.Info("Uploading backup archive to SFTP storage ... done", "filename", finalFileName, "size", goutils.ConvertBytes(uint64(backupSize)))

	// Delete backup file from tmp folder
	err = utils.DeleteFile(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		logger.Error("Error deleting file", "error", err)

	}
	if config.prune {
		err := sshStorage.Prune(config.backupRetention)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Error deleting old backup from %s storage: %s ", config.storage, err))
		}

	}
	duration := goutils.FormatDuration(time.Since(startTime), 0)
	logger.Info("Backup file uploaded to  Remote storage", "file", finalFileName, "destination", storagePath)
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
func remoteRestore(db *dbConfig, conf *RestoreConfig) {
	logger.Info("Restore database from remote server")
	sshConfig, err := loadSSHConfig()
	if err != nil {
		logger.Fatal("Error loading ssh config", "error", err)
	}

	sshStorage, err := ssh.NewStorage(ssh.Config{
		Host:         sshConfig.hostName,
		Port:         sshConfig.port,
		User:         sshConfig.user,
		Password:     sshConfig.password,
		IdentifyFile: sshConfig.identifyFile,
		RemotePath:   conf.remotePath,
		LocalPath:    tmpPath,
	})
	if err != nil {
		logger.Fatal("Error creating SSH storage", "error", err)
	}
	err = sshStorage.CopyFrom(conf.file)
	if err != nil {
		logger.Fatal("Error uploading backup file", "error", err)
	}
	RestoreDatabase(db, conf)
}
func ftpRestore(db *dbConfig, conf *RestoreConfig) {
	logger.Info("Restore database from FTP server")
	ftpConfig := loadFtpConfig()
	ftpStorage, err := ftp.NewStorage(ftp.Config{
		Host:       ftpConfig.host,
		Port:       ftpConfig.port,
		User:       ftpConfig.user,
		Password:   ftpConfig.password,
		RemotePath: conf.remotePath,
		LocalPath:  tmpPath,
	})
	if err != nil {
		logger.Fatal("Error creating SSH storage", "error", err)
	}
	err = ftpStorage.CopyFrom(conf.file)
	if err != nil {
		logger.Fatal("Error uploading backup file", "error", err)
	}
	RestoreDatabase(db, conf)
}
func ftpBackup(db *dbConfig, config *BackupConfig) {
	logger.Info("Backup database to the remote FTP server")

	// Backup database
	err := BackupDatabase(db, config)
	if err != nil {
		recoverMode(err, "Error backing up database")
		return
	}
	finalFileName := config.backupFileName
	if config.encryption {
		encryptBackup(config)
		finalFileName = fmt.Sprintf("%s.%s", config.backupFileName, "gpg")
	}
	logger.Info("Uploading backup archive to the remote FTP server ... ")
	logger.Info(fmt.Sprintf("Backup name is %s", finalFileName))
	ftpConfig := loadFtpConfig()
	ftpStorage, err := ftp.NewStorage(ftp.Config{
		Host:       ftpConfig.host,
		Port:       ftpConfig.port,
		User:       ftpConfig.user,
		Password:   ftpConfig.password,
		RemotePath: config.remotePath,
		LocalPath:  tmpPath,
	})
	if err != nil {
		logger.Fatal("Error creating SSH storage", "error", err)
	}
	err = ftpStorage.Copy(finalFileName)
	if err != nil {
		logger.Fatal("Error uploading backup file", "error", err)
	}
	logger.Info(fmt.Sprintf("Backup saved in %s", filepath.Join(config.remotePath, finalFileName)))
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
		err := ftpStorage.Prune(config.backupRetention)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Error deleting old backup from %s storage: %s ", config.storage, err))
		}

	}
	duration := goutils.FormatDuration(time.Since(startTime), 0)
	logger.Info("Backup file uploaded to  FTP storage", "file", finalFileName, "destination", storagePath)
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
