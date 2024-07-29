package pkg

const s3MountPath string = "/s3mnt"
const s3fsPasswdFile string = "/etc/passwd-s3fs"
const cronLogFile = "/var/log/pg-bkup.log"
const tmpPath = "/tmp/pg-bkup"
const backupCronFile = "/usr/local/bin/backup_cron.sh"

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
)
