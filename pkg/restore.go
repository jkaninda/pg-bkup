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
	"github.com/jkaninda/encryptor"
	"github.com/jkaninda/go-storage/pkg/local"
	"github.com/jkaninda/logger"
	"github.com/jkaninda/pg-bkup/utils"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
)

func StartRestore(cmd *cobra.Command) {
	intro()
	deleteTemp()
	dbConf = initDbConfig(cmd)
	restoreConf := initRestoreConfig(cmd)

	switch restoreConf.storage {
	case LocalStorage:
		localRestore(dbConf, restoreConf)
	case S3Storage:
		s3Restore(dbConf, restoreConf)
	case SSHStorage, SFTPStorage, RemoteStorage:
		remoteRestore(dbConf, restoreConf)
	case FTPStorage:
		ftpRestore(dbConf, restoreConf)
	case AzureStorage:
		azureRestore(dbConf, restoreConf)
	default:
		localRestore(dbConf, restoreConf)
	}
}
func localRestore(dbConf *dbConfig, restoreConf *RestoreConfig) {
	logger.Info("Restore database from local")
	basePath := filepath.Dir(restoreConf.file)
	fileName := filepath.Base(restoreConf.file)
	restoreConf.file = fileName
	if basePath == "" || basePath == "." {
		basePath = storagePath
	}
	localStorage := local.NewStorage(local.Config{
		RemotePath: basePath,
		LocalPath:  tmpPath,
	})

	// If multipart is enabled, download all part files
	if restoreConf.multipart {
		logger.Info("Downloading multipart backup files...")
		// Download all part files (filename.part001, filename.part002, etc.)
		partNum := 1
		for {
			partFileName := fmt.Sprintf("%s.part%03d", fileName, partNum)
			err := localStorage.CopyFrom(partFileName)
			if err != nil {
				// No more parts found
				break
			}
			partNum++
		}
		if partNum > 1 {
			logger.Info("Downloaded multipart backup files", "parts", partNum-1)
		} else {
			logger.Info("No multipart parts found, checking for base file...")
			err := localStorage.CopyFrom(fileName)
			if err != nil {
				logger.Fatal("No multipart files or base file found", "base_file", fileName)
			}
		}
	} else {
		err := localStorage.CopyFrom(fileName)
		if err != nil {
			logger.Fatal("Error copying backup file", "error", err)
		}
	}
	RestoreDatabase(dbConf, restoreConf)

}

// RestoreDatabase restores the database from a backup file
func RestoreDatabase(db *dbConfig, conf *RestoreConfig) {
	if conf.file == "" {
		logger.Fatal("Error, file required")
	}

	// Handle multipart restore if enabled
	if conf.multipart {
		restorationFile, err := handleMultipartRestore(conf)
		if err != nil {
			logger.Fatal("Error handling multipart restore", "error", err)
		}
		conf.file = filepath.Base(restorationFile)
	}

	filePath := filepath.Join(tmpPath, conf.file)
	rFile, err := os.ReadFile(filePath)
	if err != nil {
		logger.Fatal("Error reading backup file", "error", err)
	}

	extension := filepath.Ext(filePath)
	outputFile := RemoveLastExtension(filePath)

	if extension == ".gpg" {
		decryptBackup(conf, rFile, outputFile)
	}

	restorationFile := filepath.Join(tmpPath, conf.file)
	if !utils.FileExists(restorationFile) {
		logger.Fatal("File not found", "file", restorationFile)
	}

	if err := testDatabaseConnection(db); err != nil {
		logger.Fatal("Error connecting to the database", "error", err)
	}

	logger.Info("Restoring database...")
	restoreDatabaseFile(db, restorationFile)
}

func decryptBackup(conf *RestoreConfig, rFile []byte, outputFile string) {
	if conf.usingKey {
		logger.Info("Decrypting backup using private key...")
		prKey, err := os.ReadFile(conf.privateKey)
		if err != nil {
			logger.Fatal("Error reading private key", "error", err)
		}
		if err := encryptor.DecryptWithPrivateKey(rFile, outputFile, prKey, conf.passphrase); err != nil {
			logger.Fatal("Error decrypting backup", "error", err)
		}
	} else {
		if conf.passphrase == "" {
			logger.Fatal("Passphrase or private key required for GPG file.")
		}
		logger.Info("Decrypting backup using passphrase...")
		if err := encryptor.Decrypt(rFile, outputFile, conf.passphrase); err != nil {
			logger.Fatal("Error decrypting file", "error", err)
		}
		conf.file = RemoveLastExtension(conf.file)
	}
}

func restoreDatabaseFile(db *dbConfig, restorationFile string) {
	extension := filepath.Ext(restorationFile)
	var cmdStr string

	switch extension {
	case ".gz":
		cmdStr = "zcat " + restorationFile + " | psql -h " + db.dbHost + " -p " + db.dbPort + " -U " + db.dbUserName + " -v -d " + db.dbName
	case ".sql":
		cmdStr = "cat " + restorationFile + " | psql -h " + db.dbHost + " -p " + db.dbPort + " -U " + db.dbUserName + " -v -d " + db.dbName
	default:
		logger.Fatal("Unknown file extension", "extension", extension)
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error restoring database: %v\nOutput: %s", err, string(output)))
	}

	logger.Info("Database has been restored successfully.")
	deleteTemp()
}

// findPartFiles finds all part files for a given base filename
// Only matches files with pattern: filename.partNNN (where NNN is 3 digits)
// Also verifies files are non-empty (S3 library may create empty files on failed downloads)
func findPartFiles(tmpPath, baseFileName string) ([]string, error) {
	var partFiles []string
	entries, err := os.ReadDir(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read temp directory: %w", err)
	}

	partPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(baseFileName) + `\.part\d{3}$`)
	for _, entry := range entries {
		if !entry.IsDir() && partPattern.MatchString(entry.Name()) {
			filePath := filepath.Join(tmpPath, entry.Name())
			info, err := os.Stat(filePath)
			if err != nil {
				continue
			}
			if info.Size() == 0 {
				os.Remove(filePath)
				continue
			}
			partFiles = append(partFiles, entry.Name())
		}
	}

	sort.Strings(partFiles)
	return partFiles, nil
}

// mergePartFiles merges multiple part files into a single file
// Returns the path to the merged file
func mergePartFiles(tmpPath, baseFileName string, partFiles []string) (string, error) {
	mergedFilePath := filepath.Join(tmpPath, baseFileName)
	mergedFile, err := os.Create(mergedFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create merged file: %w", err)
	}
	defer mergedFile.Close()

	for _, partFile := range partFiles {
		partPath := filepath.Join(tmpPath, partFile)
		partData, err := os.Open(partPath)
		if err != nil {
			return "", fmt.Errorf("failed to open part file %s: %w", partFile, err)
		}

		_, err = io.Copy(mergedFile, partData)
		partData.Close()
		if err != nil {
			return "", fmt.Errorf("failed to copy part file %s: %w", partFile, err)
		}
	}

	logger.Info("Multipart files merged", "parts", len(partFiles), "output", baseFileName)
	return mergedFilePath, nil
}

// handleMultipartRestore handles merging multipart files before restore
// Returns the path to the file to restore
func handleMultipartRestore(conf *RestoreConfig) (string, error) {
	baseFileName := conf.file

	// Check if this is a multipart restore (user specified multipart flag)
	// or auto-detect by looking for part files
	partFiles, err := findPartFiles(tmpPath, baseFileName)
	if err != nil {
		return "", err
	}

	if len(partFiles) == 0 {
		// No part files found, return the original file path
		return filepath.Join(tmpPath, baseFileName), nil
	}

	logger.Info("Found multipart backup files", "parts", len(partFiles))

	// Merge all parts
	mergedPath, err := mergePartFiles(tmpPath, baseFileName, partFiles)
	if err != nil {
		return "", err
	}

	// Update the conf.file to point to merged file
	conf.file = baseFileName

	return mergedPath, nil
}
