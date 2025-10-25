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

type StorageType string
type Database struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Path     string `yaml:"path"`
}
type Config struct {
	CronExpression   string     `yaml:"cronExpression"`
	BackupRescueMode bool       `yaml:"backupRescueMode"`
	Databases        []Database `yaml:"databases"`
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
	storage            StorageType
	cronExpression     string
	all                bool
	allInOne           bool
	customName         string
	allowCustomName    bool
	schemaOnly         bool
	dataOnly           bool
	tables             []string
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
