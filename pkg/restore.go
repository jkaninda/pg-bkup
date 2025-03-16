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
	"github.com/jkaninda/encryptor"
	"github.com/jkaninda/go-storage/pkg/local"
	"github.com/jkaninda/pg-bkup/utils"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
)

func StartRestore(cmd *cobra.Command) {
	intro()
	dbConf = initDbConfig(cmd)
	restoreConf := initRestoreConfig(cmd)

	switch restoreConf.storage {
	case "local":
		localRestore(dbConf, restoreConf)
	case "s3", "S3":
		s3Restore(dbConf, restoreConf)
	case "ssh", "SSH", "remote":
		remoteRestore(dbConf, restoreConf)
	case "ftp", "FTP":
		ftpRestore(dbConf, restoreConf)
	case "azure":
		azureRestore(dbConf, restoreConf)
	default:
		localRestore(dbConf, restoreConf)
	}
}
func localRestore(dbConf *dbConfig, restoreConf *RestoreConfig) {
	utils.Info("Restore database from local")
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
	err := localStorage.CopyFrom(fileName)
	if err != nil {
		utils.Fatal("Error copying backup file: %s", err)
	}
	RestoreDatabase(dbConf, restoreConf)

}

// RestoreDatabase restores the database from a backup file
func RestoreDatabase(db *dbConfig, conf *RestoreConfig) {
	if conf.file == "" {
		utils.Fatal("Error, file required")
	}

	filePath := filepath.Join(tmpPath, conf.file)
	rFile, err := os.ReadFile(filePath)
	if err != nil {
		utils.Fatal("Error reading backup file: %v", err)
	}

	extension := filepath.Ext(filePath)
	outputFile := RemoveLastExtension(filePath)

	if extension == ".gpg" {
		decryptBackup(conf, rFile, outputFile)
	}

	restorationFile := filepath.Join(tmpPath, conf.file)
	if !utils.FileExists(restorationFile) {
		utils.Fatal("File not found: %s", restorationFile)
	}

	if err := testDatabaseConnection(db); err != nil {
		utils.Fatal("Error connecting to the database: %v", err)
	}

	utils.Info("Restoring database...")
	restoreDatabaseFile(db, restorationFile)
}

func decryptBackup(conf *RestoreConfig, rFile []byte, outputFile string) {
	if conf.usingKey {
		utils.Info("Decrypting backup using private key...")
		prKey, err := os.ReadFile(conf.privateKey)
		if err != nil {
			utils.Fatal("Error reading private key: %v", err)
		}
		if err := encryptor.DecryptWithPrivateKey(rFile, outputFile, prKey, conf.passphrase); err != nil {
			utils.Fatal("Error decrypting backup: %v", err)
		}
	} else {
		if conf.passphrase == "" {
			utils.Fatal("Passphrase or private key required for GPG file.")
		}
		utils.Info("Decrypting backup using passphrase...")
		if err := encryptor.Decrypt(rFile, outputFile, conf.passphrase); err != nil {
			utils.Fatal("Error decrypting file: %v", err)
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
		utils.Fatal("Unknown file extension: %s", extension)
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.Fatal("Error restoring database: %v\nOutput: %s", err, string(output))
	}

	utils.Info("Database has been restored successfully.")
	deleteTemp()
}
