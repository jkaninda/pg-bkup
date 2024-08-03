package pkg

import (
	"github.com/jkaninda/pg-bkup/utils"
	"os"
	"path/filepath"
	"time"
)

func copyToTmp(sourcePath string, backupFileName string) {
	//Copy backup from storage to /tmp
	err := utils.CopyFile(filepath.Join(sourcePath, backupFileName), filepath.Join(tmpPath, backupFileName))
	if err != nil {
		utils.Fatal("Error copying file %s %v", backupFileName, err)

	}
}
func moveToBackup(backupFileName string, destinationPath string) {
	//Copy backup from tmp folder to storage destination
	err := utils.CopyFile(filepath.Join(tmpPath, backupFileName), filepath.Join(destinationPath, backupFileName))
	if err != nil {
		utils.Fatal("Error copying file %s %v", backupFileName, err)

	}
	//Delete backup file from tmp folder
	err = utils.DeleteFile(filepath.Join(tmpPath, backupFileName))
	if err != nil {
		utils.Error("Error deleting file: %s", err)

	}
	utils.Done("Database has been backed up and copied to  %s", filepath.Join(destinationPath, backupFileName))
}
func deleteOldBackup(retentionDays int) {
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
			utils.Done("File %s deleted successfully", filePath)
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
			if timeDiff.Hours() > 24*float64(retentionDays) {
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
