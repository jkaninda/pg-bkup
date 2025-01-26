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
	"github.com/jkaninda/go-storage/pkg/ftp"
	"github.com/jkaninda/go-storage/pkg/ssh"
	goutils "github.com/jkaninda/go-utils"
	"github.com/jkaninda/pg-bkup/utils"
	"os"
	"path/filepath"
	"time"
)

func sshBackup(db *dbConfig, config *BackupConfig) {
	utils.Info("Backup database to Remote server")
	// Backup database
	err := BackupDatabase(db, config.backupFileName, disableCompression)
	if err != nil {
		recoverMode(err, "Error backing up database")
		return
	}
	finalFileName := config.backupFileName
	if config.encryption {
		encryptBackup(config)
		finalFileName = fmt.Sprintf("%s.%s", config.backupFileName, "gpg")
	}
	utils.Info("Uploading backup archive to remote storage ... ")
	utils.Info("Backup name is %s", finalFileName)
	sshConfig, err := loadSSHConfig()
	if err != nil {
		utils.Fatal("Error loading ssh config: %s", err)
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
		utils.Fatal("Error creating SSH storage: %s", err)
	}
	err = sshStorage.Copy(finalFileName)
	if err != nil {
		utils.Fatal("Error copying backup file: %s", err)
	}
	// Get backup info
	fileInfo, err := os.Stat(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		utils.Error("Error: %s", err)
	}
	backupSize = fileInfo.Size()
	utils.Info("Backup saved in %s", filepath.Join(config.remotePath, finalFileName))

	// Delete backup file from tmp folder
	err = utils.DeleteFile(filepath.Join(tmpPath, finalFileName))
	if err != nil {
		utils.Error("Error deleting file: %v", err)

	}
	if config.prune {
		err := sshStorage.Prune(config.backupRetention)
		if err != nil {
			utils.Fatal("Error deleting old backup from %s storage: %s ", config.storage, err)
		}

	}
	utils.Info("Backup name is %s", finalFileName)
	utils.Info("Backup size: %s", goutils.ConvertBytes(uint64(backupSize)))
	utils.Info("Uploading backup archive to remote storage ... done ")
	duration := goutils.FormatDuration(time.Since(startTime), 0)

	// Send notification
	utils.NotifySuccess(&utils.NotificationData{
		File:           finalFileName,
		BackupSize:     goutils.ConvertBytes(uint64(backupSize)),
		Database:       db.dbName,
		Storage:        config.storage,
		BackupLocation: filepath.Join(config.remotePath, finalFileName),
		Duration:       duration,
	})
	// Delete temp
	deleteTemp()
	utils.Info("Backup completed successfully in %s", duration)

}
func remoteRestore(db *dbConfig, conf *RestoreConfig) {
	utils.Info("Restore database from remote server")
	sshConfig, err := loadSSHConfig()
	if err != nil {
		utils.Fatal("Error loading ssh config: %s", err)
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
		utils.Fatal("Error creating SSH storage: %s", err)
	}
	err = sshStorage.CopyFrom(conf.file)
	if err != nil {
		utils.Fatal("Error copying backup file: %s", err)
	}
	RestoreDatabase(db, conf)
}
func ftpRestore(db *dbConfig, conf *RestoreConfig) {
	utils.Info("Restore database from FTP server")
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
		utils.Fatal("Error creating SSH storage: %s", err)
	}
	err = ftpStorage.CopyFrom(conf.file)
	if err != nil {
		utils.Fatal("Error copying backup file: %s", err)
	}
	RestoreDatabase(db, conf)
}
func ftpBackup(db *dbConfig, config *BackupConfig) {
	utils.Info("Backup database to the remote FTP server")
	// Backup database
	err := BackupDatabase(db, config.backupFileName, disableCompression)
	if err != nil {
		recoverMode(err, "Error backing up database")
		return
	}
	finalFileName := config.backupFileName
	if config.encryption {
		encryptBackup(config)
		finalFileName = fmt.Sprintf("%s.%s", config.backupFileName, "gpg")
	}
	utils.Info("Uploading backup archive to the remote FTP server ... ")
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
		utils.Fatal("Error creating SSH storage: %s", err)
	}
	err = ftpStorage.Copy(finalFileName)
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
		err := ftpStorage.Prune(config.backupRetention)
		if err != nil {
			utils.Fatal("Error deleting old backup from %s storage: %s ", config.storage, err)
		}

	}
	utils.Info("Backup name is %s", finalFileName)
	utils.Info("Backup size: %s", goutils.ConvertBytes(uint64(backupSize)))
	utils.Info("Uploading backup archive to the remote FTP server ... done ")
	duration := goutils.FormatDuration(time.Since(startTime), 0)

	// Send notification
	utils.NotifySuccess(&utils.NotificationData{
		File:           finalFileName,
		BackupSize:     goutils.ConvertBytes(uint64(backupSize)),
		Database:       db.dbName,
		Storage:        config.storage,
		BackupLocation: filepath.Join(config.remotePath, finalFileName),
		Duration:       duration,
	})
	// Delete temp
	deleteTemp()
	utils.Info("Backup successfully completed in %s", duration)
}
