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
	"github.com/jkaninda/logger"
	"github.com/jkaninda/pg-bkup/utils"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
)

func initDbConfig(cmd *cobra.Command) *dbConfig {
	jdbcUri := os.Getenv("DB_URL")
	if len(jdbcUri) != 0 {
		config, err := convertJDBCToDbConfig(jdbcUri)
		if err != nil {
			logger.Fatal("Error converting JDBC to DB config", "error", err.Error())
		}
		return config
	}
	// Set env
	utils.GetEnv(cmd, "dbname", "DB_NAME")
	dConf := dbConfig{}
	dConf.dbHost = os.Getenv("DB_HOST")
	dConf.dbPort = utils.EnvWithDefault("DB_PORT", defaultDbPort)
	dConf.dbName = os.Getenv("DB_NAME")
	dConf.dbUserName = os.Getenv("DB_USERNAME")
	dConf.dbPassword = os.Getenv("DB_PASSWORD")

	err := utils.CheckEnvVars(dbHVars)
	if err != nil {
		logger.Error("Please make sure all required environment variables for database are set")
		logger.Fatal("Error checking environment variables", "error", err)
	}
	return &dConf
}

func getDatabase(database Database) *dbConfig {
	// Set default values from environment variables if not provided
	database.User = getEnvOrDefault(database.User, "DB_USERNAME", database.Name, "")
	database.Password = getEnvOrDefault(database.Password, "DB_PASSWORD", database.Name, "")
	database.Host = getEnvOrDefault(database.Host, "DB_HOST", database.Name, "")
	database.Port = getEnvOrDefault(database.Port, "DB_PORT", database.Name, defaultDbPort)
	return &dbConfig{
		dbHost:     database.Host,
		dbPort:     database.Port,
		dbName:     database.Name,
		dbUserName: database.User,
		dbPassword: database.Password,
	}
}

// Helper function to get environment variable or use a default value
func getEnvOrDefault(currentValue, envKey, suffix, defaultValue string) string {
	// Return the current value if it's already set
	if currentValue != "" {
		return currentValue
	}

	// Check for suffixed or prefixed environment variables if a suffix is provided
	if suffix != "" {
		suffixUpper := strings.ToUpper(suffix)
		envSuffix := os.Getenv(fmt.Sprintf("%s_%s", envKey, suffixUpper))
		if envSuffix != "" {
			return envSuffix
		}

		envPrefix := os.Getenv(fmt.Sprintf("%s_%s", suffixUpper, envKey))
		if envPrefix != "" {
			return envPrefix
		}
	}

	// Fall back to the default value using a helper function
	return utils.EnvWithDefault(envKey, defaultValue)
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
		logger.Error("Please make sure all required environment variables for FTP are set")
		logger.Fatal("Error missing environment variables", "error", err)
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
		logger.Error("Please make sure all required environment variables for Azure Blob storage are set")
		logger.Fatal("Error missing environment variables", "error", err)
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
		logger.Error("Please make sure all required environment variables for AWS S3 are set")
		logger.Fatal("Error checking environment variables", "error", err)
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
	customName, _ := cmd.Flags().GetString("custom-name")
	all, _ := cmd.Flags().GetBool("all-databases")
	allInOne, _ := cmd.Flags().GetBool("all-in-one")
	if allInOne {
		all = true
	}
	_, _ = cmd.Flags().GetString("mode")
	passphrase := os.Getenv("GPG_PASSPHRASE")
	_ = utils.GetEnv(cmd, "path", "AWS_S3_PATH")
	cronExpression := os.Getenv("BACKUP_CRON_EXPRESSION")

	// TODO: Update cron expression

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
	config.storage = StorageType(storage)
	config.encryption = encryption
	config.remotePath = remotePath
	config.passphrase = passphrase
	config.publicKey = publicKeyFile
	config.usingKey = usingKey
	config.cronExpression = cronExpression
	config.all = all
	config.allInOne = allInOne
	config.customName = customName
	return &config
}

type RestoreConfig struct {
	s3Path     string
	remotePath string
	storage    StorageType
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
	rConfig.storage = StorageType(strings.ToLower(storage))
	rConfig.bucket = bucket
	rConfig.file = file
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
			logger.Fatal("Error", "error", err.Error())
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
	tdbConfig.targetDbPort = utils.EnvWithDefault("TARGET_DB_PORT", defaultDbPort)
	tdbConfig.targetDbName = os.Getenv("TARGET_DB_NAME")
	tdbConfig.targetDbUserName = os.Getenv("TARGET_DB_USERNAME")
	tdbConfig.targetDbPassword = os.Getenv("TARGET_DB_PASSWORD")

	err := utils.CheckEnvVars(tdbRVars)
	if err != nil {
		logger.Error("Please make sure all required environment variables for the target database are set")
		logger.Fatal("Error checking target database environment variables", "error", err)
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
