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
	"github.com/jkaninda/pg-bkup/utils"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

type Database struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Path     string `yaml:"path"`
}
type Config struct {
	Databases      []Database `yaml:"databases"`
	CronExpression string     `yaml:"cronExpression"`
}

type dbConfig struct {
	dbHost     string
	dbPort     string
	dbName     string
	dbUserName string
	dbPassword string
}
type targetDbConfig struct {
	targetDbHost     string
	targetDbPort     string
	targetDbUserName string
	targetDbPassword string
	targetDbName     string
}
type TgConfig struct {
	Token  string
	ChatId string
}
type BackupConfig struct {
	backupFileName     string
	backupRetention    int
	disableCompression bool
	prune              bool
	remotePath         string
	encryption         bool
	usingKey           bool
	passphrase         string
	publicKey          string
	storage            string
	cronExpression     string
}
type FTPConfig struct {
	host       string
	user       string
	password   string
	port       int
	remotePath string
}
type AzureConfig struct {
	accountName   string
	accountKey    string
	containerName string
}

// SSHConfig holds the SSH connection details
type SSHConfig struct {
	user         string
	password     string
	hostName     string
	port         int
	identifyFile string
}
type AWSConfig struct {
	endpoint       string
	bucket         string
	accessKey      string
	secretKey      string
	region         string
	remotePath     string
	disableSsl     bool
	forcePathStyle bool
}

func initDbConfig(cmd *cobra.Command) *dbConfig {
	jdbcUri := os.Getenv("DB_URL")
	if len(jdbcUri) != 0 {
		config, err := convertJDBCToDbConfig(jdbcUri)
		if err != nil {
			utils.Fatal("Error: %v", err.Error())
		}
		return config
	}
	// Set env
	utils.GetEnv(cmd, "dbname", "DB_NAME")
	dConf := dbConfig{}
	dConf.dbHost = os.Getenv("DB_HOST")
	dConf.dbPort = os.Getenv("DB_PORT")
	dConf.dbName = os.Getenv("DB_NAME")
	dConf.dbUserName = os.Getenv("DB_USERNAME")
	dConf.dbPassword = os.Getenv("DB_PASSWORD")

	err := utils.CheckEnvVars(dbHVars)
	if err != nil {
		utils.Error("Please make sure all required environment variables for database are set")
		utils.Fatal("Error checking environment variables: %s", err)
	}
	return &dConf
}

func getDatabase(database Database) *dbConfig {
	return &dbConfig{
		dbHost:     database.Host,
		dbPort:     database.Port,
		dbName:     database.Name,
		dbUserName: database.User,
		dbPassword: database.Password,
	}
}

// loadSSHConfig loads the SSH configuration from environment variables
func loadSSHConfig() (*SSHConfig, error) {
	utils.GetEnvVariable("SSH_HOST", "SSH_HOST_NAME")
	sshVars := []string{"SSH_USER", "SSH_HOST", "SSH_PORT", "REMOTE_PATH"}
	err := utils.CheckEnvVars(sshVars)
	if err != nil {
		return nil, fmt.Errorf("error missing environment variables: %w", err)
	}

	return &SSHConfig{
		user:         os.Getenv("SSH_USER"),
		password:     os.Getenv("SSH_PASSWORD"),
		hostName:     os.Getenv("SSH_HOST"),
		port:         utils.GetIntEnv("SSH_PORT"),
		identifyFile: os.Getenv("SSH_IDENTIFY_FILE"),
	}, nil
}
func loadFtpConfig() *FTPConfig {
	// Initialize data configs
	fConfig := FTPConfig{}
	fConfig.host = utils.GetEnvVariable("FTP_HOST", "FTP_HOST_NAME")
	fConfig.user = os.Getenv("FTP_USER")
	fConfig.password = os.Getenv("FTP_PASSWORD")
	fConfig.port = utils.GetIntEnv("FTP_PORT")
	fConfig.remotePath = os.Getenv("REMOTE_PATH")
	err := utils.CheckEnvVars(ftpVars)
	if err != nil {
		utils.Error("Please make sure all required environment variables for FTP are set")
		utils.Fatal("Error missing environment variables: %s", err)
	}
	return &fConfig
}
func loadAzureConfig() *AzureConfig {
	// Initialize data configs
	aConfig := AzureConfig{}
	aConfig.containerName = os.Getenv("AZURE_STORAGE_CONTAINER_NAME")
	aConfig.accountName = os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	aConfig.accountKey = os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")

	err := utils.CheckEnvVars(azureVars)
	if err != nil {
		utils.Error("Please make sure all required environment variables for Azure Blob storage are set")
		utils.Fatal("Error missing environment variables: %s", err)
	}
	return &aConfig
}
func initAWSConfig() *AWSConfig {
	// Initialize AWS configs
	aConfig := AWSConfig{}
	aConfig.endpoint = utils.GetEnvVariable("AWS_S3_ENDPOINT", "S3_ENDPOINT")
	aConfig.accessKey = utils.GetEnvVariable("AWS_ACCESS_KEY", "ACCESS_KEY")
	aConfig.secretKey = utils.GetEnvVariable("AWS_SECRET_KEY", "SECRET_KEY")
	aConfig.bucket = utils.GetEnvVariable("AWS_S3_BUCKET_NAME", "BUCKET_NAME")
	aConfig.remotePath = utils.GetEnvVariable("AWS_S3_PATH", "S3_PATH")

	aConfig.region = os.Getenv("AWS_REGION")
	disableSsl, err := strconv.ParseBool(os.Getenv("AWS_DISABLE_SSL"))
	if err != nil {
		disableSsl = false
	}
	forcePathStyle, err := strconv.ParseBool(os.Getenv("AWS_FORCE_PATH_STYLE"))
	if err != nil {
		forcePathStyle = false
	}
	aConfig.disableSsl = disableSsl
	aConfig.forcePathStyle = forcePathStyle
	err = utils.CheckEnvVars(awsVars)
	if err != nil {
		utils.Error("Please make sure all required environment variables for AWS S3 are set")
		utils.Fatal("Error checking environment variables: %s", err)
	}
	return &aConfig
}
func initBackupConfig(cmd *cobra.Command) *BackupConfig {
	utils.SetEnv("STORAGE_PATH", storagePath)
	utils.GetEnv(cmd, "cron-expression", "BACKUP_CRON_EXPRESSION")
	utils.GetEnv(cmd, "path", "REMOTE_PATH")
	utils.GetEnv(cmd, "config", "BACKUP_CONFIG_FILE")
	// Get flag value and set env
	remotePath := utils.GetEnvVariable("REMOTE_PATH", "SSH_REMOTE_PATH")
	storage = utils.GetEnv(cmd, "storage", "STORAGE")
	prune := false
	backupRetention := utils.GetIntEnv("BACKUP_RETENTION_DAYS")
	if backupRetention > 0 {
		prune = true
	}
	disableCompression, _ = cmd.Flags().GetBool("disable-compression")
	_, _ = cmd.Flags().GetString("mode")
	passphrase := os.Getenv("GPG_PASSPHRASE")
	_ = utils.GetEnv(cmd, "path", "AWS_S3_PATH")
	cronExpression := os.Getenv("BACKUP_CRON_EXPRESSION")

	publicKeyFile, err := checkPubKeyFile(os.Getenv("GPG_PUBLIC_KEY"))
	if err == nil {
		encryption = true
		usingKey = true
	} else if passphrase != "" {
		encryption = true
		usingKey = false
	}
	// Initialize backup configs
	config := BackupConfig{}
	config.backupRetention = backupRetention
	config.disableCompression = disableCompression
	config.prune = prune
	config.storage = storage
	config.encryption = encryption
	config.remotePath = remotePath
	config.passphrase = passphrase
	config.publicKey = publicKeyFile
	config.usingKey = usingKey
	config.cronExpression = cronExpression
	return &config
}

type RestoreConfig struct {
	s3Path     string
	remotePath string
	storage    string
	file       string
	bucket     string
	usingKey   bool
	passphrase string
	privateKey string
}

func initRestoreConfig(cmd *cobra.Command) *RestoreConfig {
	utils.SetEnv("STORAGE_PATH", storagePath)
	utils.GetEnv(cmd, "path", "REMOTE_PATH")

	// Get flag value and set env
	s3Path := utils.GetEnv(cmd, "path", "AWS_S3_PATH")
	remotePath := utils.GetEnvVariable("REMOTE_PATH", "SSH_REMOTE_PATH")
	storage = utils.GetEnv(cmd, "storage", "STORAGE")
	file = utils.GetEnv(cmd, "file", "FILE_NAME")
	bucket := utils.GetEnvVariable("AWS_S3_BUCKET_NAME", "BUCKET_NAME")
	passphrase := os.Getenv("GPG_PASSPHRASE")
	privateKeyFile, err := checkPrKeyFile(os.Getenv("GPG_PRIVATE_KEY"))
	if err == nil {
		usingKey = true
	} else if passphrase != "" {
		usingKey = false
	}

	// Initialize restore configs
	rConfig := RestoreConfig{}
	rConfig.s3Path = s3Path
	rConfig.remotePath = remotePath
	rConfig.storage = storage
	rConfig.bucket = bucket
	rConfig.file = file
	rConfig.storage = storage
	rConfig.passphrase = passphrase
	rConfig.usingKey = usingKey
	rConfig.privateKey = privateKeyFile
	return &rConfig
}
func initTargetDbConfig() *targetDbConfig {
	jdbcUri := os.Getenv("TARGET_DB_URL")
	if len(jdbcUri) != 0 {
		config, err := convertJDBCToDbConfig(jdbcUri)
		if err != nil {
			utils.Fatal("Error: %v", err.Error())
		}
		return &targetDbConfig{
			targetDbHost:     config.dbHost,
			targetDbPort:     config.dbPort,
			targetDbName:     config.dbName,
			targetDbPassword: config.dbPassword,
			targetDbUserName: config.dbUserName,
		}
	}
	tdbConfig := targetDbConfig{}
	tdbConfig.targetDbHost = os.Getenv("TARGET_DB_HOST")
	tdbConfig.targetDbPort = utils.EnvWithDefault("TARGET_DB_PORT", "5432")
	tdbConfig.targetDbName = os.Getenv("TARGET_DB_NAME")
	tdbConfig.targetDbUserName = os.Getenv("TARGET_DB_USERNAME")
	tdbConfig.targetDbPassword = os.Getenv("TARGET_DB_PASSWORD")

	err := utils.CheckEnvVars(tdbRVars)
	if err != nil {
		utils.Error("Please make sure all required environment variables for the target database are set")
		utils.Fatal("Error checking target database environment variables: %s", err)
	}
	return &tdbConfig
}
func loadConfigFile() (string, error) {
	backupConfigFile, err := checkConfigFile(os.Getenv("BACKUP_CONFIG_FILE"))
	if err == nil {
		return backupConfigFile, nil
	}
	return "", fmt.Errorf("backup config file not found")
}
