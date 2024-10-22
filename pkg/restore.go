// Package pkg /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package pkg

import (
	"github.com/jkaninda/encryptor"
	"github.com/jkaninda/pg-bkup/pkg/storage/ftp"
	"github.com/jkaninda/pg-bkup/pkg/storage/local"
	"github.com/jkaninda/pg-bkup/pkg/storage/s3"
	"github.com/jkaninda/pg-bkup/pkg/storage/ssh"
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
		restoreFromS3(dbConf, restoreConf)
	case "ssh", "SSH", "remote":
		restoreFromRemote(dbConf, restoreConf)
	case "ftp", "FTP":
		restoreFromFTP(dbConf, restoreConf)
	default:
		localRestore(dbConf, restoreConf)
	}
}
func localRestore(dbConf *dbConfig, restoreConf *RestoreConfig) {
	utils.Info("Restore database from local")
	localStorage := local.NewStorage(local.Config{
		RemotePath: storagePath,
		LocalPath:  tmpPath,
	})
	err := localStorage.CopyFrom(restoreConf.file)
	if err != nil {
		utils.Fatal("Error copying backup file: %s", err)
	}
	RestoreDatabase(dbConf, restoreConf)

}
func restoreFromS3(db *dbConfig, conf *RestoreConfig) {
	utils.Info("Restore database from s3")
	//err := DownloadFile(tmpPath, conf.file, conf.bucket, conf.s3Path)
	//if err != nil {
	//	utils.Fatal("Error download file from s3 %s %v ", conf.file, err)
	//}
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
		RemotePath:     awsConfig.remotePath,
		LocalPath:      tmpPath,
	})
	if err != nil {
		utils.Fatal("Error creating s3 storage: %s", err)
	}
	err = s3Storage.CopyFrom(conf.file)
	if err != nil {
		utils.Fatal("Error download file from S3 storage: %s", err)
	}
	RestoreDatabase(db, conf)
}
func restoreFromRemote(db *dbConfig, conf *RestoreConfig) {
	utils.Info("Restore database from remote server")
	sshConfig, err := loadSSHConfig()
	if err != nil {
		utils.Fatal("Error loading ssh config: %s", err)
	}

	sshStorage, err := ssh.NewStorage(ssh.Config{
		Host:       sshConfig.hostName,
		Port:       sshConfig.port,
		User:       sshConfig.user,
		Password:   sshConfig.password,
		RemotePath: conf.remotePath,
		LocalPath:  tmpPath,
	})
	if err != nil {
		utils.Fatal("Error creating SSH storage: %s", err)
	}
	err = sshStorage.CopyFrom(conf.file)
	if err != nil {
		utils.Fatal("Error copying backup file: %w", err)
	}
	RestoreDatabase(db, conf)
}
func restoreFromFTP(db *dbConfig, conf *RestoreConfig) {
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

// RestoreDatabase restore database
func RestoreDatabase(db *dbConfig, conf *RestoreConfig) {
	if conf.file == "" {
		utils.Fatal("Error, file required")
	}
	extension := filepath.Ext(filepath.Join(tmpPath, conf.file))
	rFile, err := os.ReadFile(filepath.Join(tmpPath, conf.file))
	outputFile := RemoveLastExtension(filepath.Join(tmpPath, conf.file))
	if err != nil {
		utils.Fatal("Error reading backup file: %s ", err)
	}

	if extension == ".gpg" {

		if conf.usingKey {
			utils.Info("Decrypting backup using private key...")
			utils.Warn("Backup decryption using a private key is not fully supported")
			prKey, err := os.ReadFile(conf.privateKey)
			if err != nil {
				utils.Fatal("Error reading public key: %s ", err)
			}
			err = encryptor.DecryptWithPrivateKey(rFile, outputFile, prKey, conf.passphrase)
			if err != nil {
				utils.Fatal("error during decrypting backup %v", err)
			}
			utils.Info("Decrypting backup using private key...done")

		} else {
			if conf.passphrase == "" {
				utils.Error("Error, passphrase or private key required")
				utils.Fatal("Your file seems to be a GPG file.\nYou need to provide GPG keys. GPG_PASSPHRASE or GPG_PRIVATE_KEY environment variable is required.")
			} else {
				utils.Info("Decrypting backup using passphrase...")
				//decryptWithGPG file
				err := encryptor.Decrypt(rFile, outputFile, conf.passphrase)
				if err != nil {
					utils.Fatal("Error decrypting file %s %v", file, err)
				}
				utils.Info("Decrypting backup using passphrase...done")
				//Update file name
				conf.file = RemoveLastExtension(file)
			}
		}

	}

	if utils.FileExists(filepath.Join(tmpPath, conf.file)) {

		err := os.Setenv("PGPASSWORD", db.dbPassword)
		if err != nil {
			return
		}
		testDatabaseConnection(db)
		utils.Info("Restoring database...")

		extension := filepath.Ext(conf.file)
		// Restore from compressed file / .sql.gz
		if extension == ".gz" {
			str := "zcat " + filepath.Join(tmpPath, conf.file) + " | psql -h " + db.dbHost + " -p " + db.dbPort + " -U " + db.dbUserName + " -v -d " + db.dbName
			_, err := exec.Command("sh", "-c", str).Output()
			if err != nil {
				utils.Fatal("Error, in restoring the database %v", err)
			}
			utils.Info("Restoring database... done")
			utils.Info("Database has been restored")
			//Delete temp
			deleteTemp()

		} else if extension == ".sql" {
			//Restore from sql file
			str := "cat " + filepath.Join(tmpPath, conf.file) + " | psql -h " + db.dbHost + " -p " + db.dbPort + " -U " + db.dbUserName + " -v -d " + db.dbName
			_, err := exec.Command("sh", "-c", str).Output()
			if err != nil {
				utils.Fatal("Error in restoring the database %v", err)
			}
			utils.Info("Restoring database... done")
			utils.Info("Database has been restored")
			//Delete temp
			deleteTemp()
		} else {
			utils.Fatal("Unknown file extension: %s", extension)
		}

	} else {
		utils.Fatal("File not found in %s", filepath.Join(tmpPath, conf.file))
	}
}
