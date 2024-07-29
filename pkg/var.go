package pkg

const s3MountPath string = "/s3mnt"
const s3fsPasswdFile string = "/etc/passwd-s3fs"
const cronLogFile = "/var/log/pg-bkup.log"
const tmpPath = "/tmp/backup"
const backupCronFile = "/usr/local/bin/backup_cron.sh"
const algorithm = "aes256"
const gpgExtension = "gpg"

var (
	storage            = "local"
	file               = ""
	s3Path             = "/pg-bkup"
	dbPassword         = ""
	dbUserName         = ""
	dbName             = ""
	dbHost             = ""
	dbPort             = "5432"
	executionMode      = "default"
	storagePath        = "/backup"
	disableCompression = false
	encryption         = false
)
