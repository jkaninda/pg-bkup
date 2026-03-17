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
	"github.com/jkaninda/go-storage/pkg/s3"
	goutils "github.com/jkaninda/go-utils"
	"github.com/jkaninda/logger"
	"github.com/jkaninda/pg-bkup/utils"

	"os"
	"path/filepath"
	"time"
)

func s3Backup(db *dbConfig, config *BackupConfig) {

	logger.Info("Backup database to s3 storage")
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

	// Handle multipart split if enabled
	filesToUpload, err := handleMultipartBackup(config, finalFileName)
	if err != nil {
		logger.Fatal("Error splitting backup file", "error", err)
	}

	logger.Info("Uploading backup archive to remote storage S3 ... ")
	awsConfig := initAWSConfig()
	if config.remotePath == "" {
		config.remotePath = awsConfig.remotePath
	}
	logger.Info(fmt.Sprintf("Backup name is %s", finalFileName))
	s3Storage, err := s3.NewStorage(s3.Config{
		Endpoint:       awsConfig.endpoint,
		Bucket:         awsConfig.bucket,
		AccessKey:      awsConfig.accessKey,
		SecretKey:      awsConfig.secretKey,
		Region:         awsConfig.region,
		DisableSsl:     awsConfig.disableSsl,
		ForcePathStyle: awsConfig.forcePathStyle,
		RemotePath:     config.remotePath,
		LocalPath:      tmpPath,
	})
	if err != nil {
		logger.Fatal("Error creating s3 storage", "error", err)
	}

	// Upload all files (single file or multiple parts)
	totalSize := int64(0)
	for _, fileName := range filesToUpload {
		err = s3Storage.Copy(fileName)
		if err != nil {
			logger.Fatal("Error uploading backup file", "error", err)
		}
		partInfo, _ := os.Stat(filepath.Join(tmpPath, fileName))
		if partInfo != nil {
			totalSize += partInfo.Size()
		}
	}

	if totalSize > 0 {
		backupSize = totalSize
	}

	// Delete old backup
	if config.prune {
		err := s3Storage.Prune(config.backupRetention)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Error deleting old backup from %s storage: %s ", config.storage, err))
		}
	}

	duration := goutils.FormatDuration(time.Since(startTime), 2)
	if len(filesToUpload) > 1 {
		logger.Info("Backup files uploaded to S3 storage", "parts", len(filesToUpload), "destination", config.remotePath)
		logger.Info("Backup completed", "parts", len(filesToUpload), "total_size", goutils.ConvertBytes(uint64(backupSize)), "duration", duration)
	} else {
		logger.Info("Backup file uploaded to  S3 storage", "file", finalFileName, "destination", storagePath)
		logger.Info("Backup completed", "file", finalFileName, "size", goutils.ConvertBytes(uint64(backupSize)), "duration", duration)
	}

	// Send notification
	notificationFile := finalFileName
	if len(filesToUpload) > 1 {
		notificationFile = fmt.Sprintf("%s (%d parts)", finalFileName, len(filesToUpload))
	}
	utils.NotifySuccess(&utils.NotificationData{
		File:           notificationFile,
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
func s3Restore(db *dbConfig, conf *RestoreConfig) {
	logger.Info("Restore database from s3")
	awsConfig := initAWSConfig()
	if conf.remotePath == "" {
		conf.remotePath = awsConfig.remotePath
	}
	s3Storage, err := s3.NewStorage(s3.Config{
		Endpoint:       awsConfig.endpoint,
		Bucket:         awsConfig.bucket,
		AccessKey:      awsConfig.accessKey,
		SecretKey:      awsConfig.secretKey,
		Region:         awsConfig.region,
		DisableSsl:     awsConfig.disableSsl,
		ForcePathStyle: awsConfig.forcePathStyle,
		RemotePath:     conf.remotePath,
		LocalPath:      tmpPath,
	})
	if err != nil {
		logger.Fatal("Error creating s3 storage", "error", err)
	}

	// If multipart is enabled, download all part files
	if conf.multipart {
		logger.Info("Downloading multipart backup files from S3...")
		partNum := 1
		for {
			partFileName := fmt.Sprintf("%s.part%03d", conf.file, partNum)
			err := s3Storage.CopyFrom(partFileName)
			if err != nil {
				break
			}
			partNum++
		}
		if partNum > 1 {
			logger.Info("Downloaded multipart backup files from S3", "parts", partNum-1)
		} else {
			logger.Fatal("No multipart files found", "base_file", conf.file)
		}
	} else {
		err = s3Storage.CopyFrom(conf.file)
		if err != nil {
			logger.Fatal("Error download file from S3 storage", "error", err)
		}
	}
	RestoreDatabase(db, conf)
}
