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
	"time"
)

const tmpPath = "/tmp/backup"
const gpgHome = "/config/gnupg"
const gpgExtension = "gpg"
const timeFormat = "2006-01-02 at 15:04:05"

var (
	storage = "local"
	file    = ""

	storagePath              = "/backup"
	workingDir               = "/config"
	disableCompression       = false
	encryption               = false
	usingKey                 = false
	backupSize         int64 = 0
	startTime                = time.Now()
	backupRescueMode         = false
)

// dbHVars Required environment variables for database
var dbHVars = []string{
	"DB_HOST",
	"DB_PASSWORD",
	"DB_USERNAME",
}
var tdbRVars = []string{
	"TARGET_DB_HOST",
	"TARGET_DB_NAME",
	"TARGET_DB_USERNAME",
	"TARGET_DB_PASSWORD",
}

var dbConf *dbConfig
var targetDbConf *targetDbConfig

var ftpVars = []string{
	"FTP_HOST_NAME",
	"FTP_USER",
	"FTP_PASSWORD",
	"FTP_PORT",
}
var azureVars = []string{
	"AZURE_STORAGE_CONTAINER_NAME",
	"AZURE_STORAGE_ACCOUNT_NAME",
	"AZURE_STORAGE_ACCOUNT_KEY",
}

// AwsVars Required environment variables for AWS S3 storage
var awsVars = []string{
	"AWS_S3_ENDPOINT",
	"AWS_S3_BUCKET_NAME",
	"AWS_ACCESS_KEY",
	"AWS_SECRET_KEY",
	"AWS_REGION",
}
