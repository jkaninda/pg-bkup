package main

/*****
*   PostgreSQL Backup & Restore
* @author    Jonas Kaninda
* @license   MIT License <https://opensource.org/licenses/MIT>
* @link      https://github.com/jkaninda/pg-bkup
**/
import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jkaninda/pg-bkup/utils"
	flag "github.com/spf13/pflag"
)

var appVersion string = os.Getenv("VERSION")

const s3MountPath string = "/s3mnt"

var (
	operation          string = "bakup"
	storage            string = "local"
	file               string = ""
	s3Path             string = "/pg-bkup"
	dbName             string = ""
	dbHost             string = ""
	dbPort             string = "5432"
	dbPassword         string = ""
	dbUserName         string = ""
	executionMode      string = "default"
	storagePath        string = "/backup"
	accessKey          string = ""
	secretKey          string = ""
	bucketName         string = ""
	s3Endpoint         string = ""
	s3fsPasswdFile     string = "/etc/passwd-s3fs"
	disableCompression bool   = false
	startBackup        bool   = true
	outputContent      string = ""
	potimeout          int    = 30
	period             string = "0 1 * * *"
)

func init() {
	var (
		operationFlag          = flag.StringP("operation", "o", "backup", "Set operation")
		storageFlag            = flag.StringP("storage", "s", "local", "Set storage. local or s3")
		fileFlag               = flag.StringP("file", "f", "", "Set file name")
		pathFlag               = flag.StringP("path", "P", "/mysql-bkup", "Set s3 path, without file name")
		dbnameFlag             = flag.StringP("dbname", "d", "", "Set database name")
		modeFlag               = flag.StringP("mode", "m", "default", "Set execution mode. default or scheduled")
		periodFlag             = flag.StringP("period", "", "0 1 * * *", "Set schedule period time")
		timeoutFlag            = flag.IntP("timeout", "t", 30, "Set timeout")
		disableCompressionFlag = flag.BoolP("disable-compression", "", false, "Disable backup compression")
		portFlag               = flag.IntP("port", "p", 5432, "Set database port")
		helpFlag               = flag.BoolP("help", "h", false, "Print this help message")
		versionFlag            = flag.BoolP("version", "v", false, "shows version information")
	)
	flag.Parse()

	operation = *operationFlag
	storage = *storageFlag
	file = *fileFlag
	s3Path = *pathFlag
	dbName = *dbnameFlag
	executionMode = *modeFlag
	dbPort = fmt.Sprint(*portFlag)
	potimeout = *timeoutFlag
	period = *periodFlag
	disableCompression = *disableCompressionFlag

	flag.Usage = func() {
		fmt.Print("Usage: bkup -o backup -s s3 -d databasename --path /my_path ...\n")
		fmt.Print("       bkup -o backup -d databasename --disable-compression ...\n")
		fmt.Print("       Restore: bkup -o restore -d databasename -f db_20231217_051339.sql.gz ...\n\n")
		flag.PrintDefaults()
	}

	if *helpFlag {
		startBackup = false
		flag.Usage()
		os.Exit(0)
	}
	if *versionFlag {
		startBackup = false
		version()
		os.Exit(0)
	}
	if *dbnameFlag != "" {
		os.Setenv("DB_NAME", dbName)
	}
	if *pathFlag != "" {
		s3Path = *pathFlag
		os.Setenv("S3_PATH", fmt.Sprint(*pathFlag))

	}
	if *fileFlag != "" {
		file = *fileFlag
		os.Setenv("FILE_NAME", fmt.Sprint(*fileFlag))

	}
	if *portFlag != 3306 {
		os.Setenv("DB_PORT", fmt.Sprint(*portFlag))
	}
	if *periodFlag != "" {
		os.Setenv("SCHEDULE_PERIOD", fmt.Sprint(*periodFlag))
	}
	if *storageFlag != "" {
		os.Setenv("STORAGE", fmt.Sprint(*storageFlag))
	}
	dbHost = os.Getenv("DB_HOST")
	dbPassword = os.Getenv("DB_PASSWORD")
	dbUserName = os.Getenv("DB_USERNAME")
	dbName = os.Getenv("DB_NAME")
	dbPort = os.Getenv("DB_PORT")
	period = os.Getenv("SCHEDULE_PERIOD")
	storage = os.Getenv("STORAGE")

	accessKey = os.Getenv("ACCESS_KEY")
	secretKey = os.Getenv("SECRET_KEY")
	bucketName = os.Getenv("BUCKETNAME")
	s3Endpoint = os.Getenv("S3_ENDPOINT")

}

func version() {
	fmt.Printf("Version: %s \n", appVersion)
	fmt.Print()
}
func main() {
	os.Setenv("STORAGE_PATH", storagePath)

	if startBackup {
		start()
	}

}
func start() {

	if executionMode == "default" {
		if operation != "backup" {
			if storage != "s3" {
				utils.Info("Restore from local")
				restore()
			} else {
				utils.Info("Restore from s3")
				s3Restore()
			}
		} else {
			if storage != "s3" {
				utils.Info("Backup to local storage")
				backup()
			} else {
				utils.Info("Backup to s3 storage")
				s3Backup()
			}
		}
	} else if executionMode == "scheduled" {
		scheduledMode()
	} else {
		utils.Fatal("Error, unknown execution mode!")
	}
}
func backup() {
	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_NAME") == "" || os.Getenv("DB_USERNAME") == "" || os.Getenv("DB_PASSWORD") == "" {
		utils.Fatal("Please make sure all required environment variables for database are set")
	} else {
		testDatabaseConnection()
		// Backup database
		utils.Info("Backing up database...")
		bkFileName := fmt.Sprintf("%s_%s.sql.gz", dbName, time.Now().Format("20060102_150405"))

		os.Setenv("PGPASSWORD", dbPassword)
		if disableCompression {
			bkFileName = fmt.Sprintf("%s_%s.sql", dbName, time.Now().Format("20060102_150405"))
			cmd := exec.Command("pg_dump",
				"-h", dbHost,
				"-p", dbPort,
				"-U", dbUserName,
				"-d", dbName,
			)
			output, err := cmd.Output()
			if err != nil {
				log.Fatal(err)
			}

			file, err := os.Create(fmt.Sprintf("%s/%s", storagePath, bkFileName))
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			_, err = file.Write(output)
			if err != nil {
				log.Fatal(err)
			}
			utils.Info("Database has been backed up")

		} else {
			cmd := exec.Command("pg_dump",
				"-h", dbHost,
				"-p", dbPort,
				"-U", dbUserName,
				"-d", dbName,
			)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Fatal(err)
			}
			gzipCmd := exec.Command("gzip")
			gzipCmd.Stdin = stdout
			gzipCmd.Stdout, err = os.Create(fmt.Sprintf("%s/%s", storagePath, bkFileName))
			gzipCmd.Start()
			if err != nil {
				log.Fatal(err)
			}
			if err := cmd.Run(); err != nil {
				log.Fatal(err)
			}
			if err := gzipCmd.Wait(); err != nil {
				log.Fatal(err)
			}
			utils.Info("Database has been backed up")

		}

		historyFile, err := os.OpenFile(fmt.Sprintf("%s/history.txt", storagePath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer historyFile.Close()
		if _, err := historyFile.WriteString(bkFileName + "\n"); err != nil {
			log.Fatal(err)
		}
	}
}
func restore() {
	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_NAME") == "" || os.Getenv("DB_USERNAME") == "" || os.Getenv("DB_PASSWORD") == "" || file == "" {
		utils.Fatal("Please make sure all required environment variables are set")
	} else {

		if utils.FileExists(fmt.Sprintf("%s/%s", storagePath, file)) {
			testDatabaseConnection()
			os.Setenv("PGPASSWORD", dbPassword)

			extension := filepath.Ext(fmt.Sprintf("%s/%s", storagePath, file))
			// GZ compressed file
			if extension == ".gz" {
				str := "zcat " + fmt.Sprintf("%s/%s", storagePath, file) + " | psql -h " + os.Getenv("DB_HOST") + " -p " + os.Getenv("DB_PORT") + " -U " + os.Getenv("DB_USERNAME") + " -v -d " + os.Getenv("DB_NAME")
				output, err := exec.Command("bash", "-c", str).Output()
				if err != nil {
					utils.Fatal("Error, in restoring the database")
				}
				outputContent = string(output)
				utils.Info("Database has been restored")

			} else if extension == ".sql" {
				//SQL file
				str := "cat " + fmt.Sprintf("%s/%s", storagePath, file) + " | psql -h " + os.Getenv("DB_HOST") + " -p " + os.Getenv("DB_PORT") + " -U " + os.Getenv("DB_USERNAME") + " -v -d " + os.Getenv("DB_NAME")
				output, err := exec.Command("bash", "-c", str).Output()
				if err != nil {
					utils.Fatalf("Error in restoring the database", err)
				}
				outputContent = string(output)
				utils.Info("Database has been restored")
			} else {
				utils.Fatal("Unknown file extension ", extension)
			}

		} else {
			utils.Fatal("File not found in ", fmt.Sprintf("%s/%s", storagePath, file))
		}

	}
}
func s3Backup() {
	// Implement S3 backup logic
	s3Mount()
	backup()
}

// Run in scheduled mode
func scheduledMode() {
	// Verify operation
	if operation == "backup" {

		fmt.Println()
		fmt.Println("**********************************")
		fmt.Println("     Starting PostgreSQL Bkup...   ")
		fmt.Println("***********************************")
		utils.Info("Running in Scheduled mode")
		utils.Info("Log file in /var/log/pg-bkup.log")
		utils.Info("Execution period ", os.Getenv("SCHEDULE_PERIOD"))
		testDatabaseConnection()
		utils.Info("Creating backup job...")
		createCrontabScript()
		supervisordCmd := exec.Command("supervisord", "-c", "/etc/supervisor/supervisord.conf")
		if err := supervisordCmd.Run(); err != nil {
			utils.Fatalf("Error starting supervisord: %v\n", err)
		}
	} else {
		utils.Fatal("Scheduled mode supports only backup operation")
	}
}

// Mount s3 using s3fs
func s3Mount() {
	if accessKey == "" || secretKey == "" || bucketName == "" {
		utils.Fatal("Please make sure all environment variables are set")
	} else {
		storagePath = fmt.Sprintf("%s%s", s3MountPath, s3Path)
		os.Setenv("STORAGE_PATH", storagePath)

		//Write file
		err := utils.WriteToFile(s3fsPasswdFile, fmt.Sprintf("%s:%s", accessKey, secretKey))
		if err != nil {
			utils.Fatal("Error creating file")
		}
		//Change file permission
		utils.ChangePermission(s3fsPasswdFile, 0600)
		utils.Info("Mounting Object storage in", s3MountPath)
		if isEmpty, _ := utils.IsDirEmpty(s3MountPath); isEmpty {
			cmd := exec.Command("s3fs", bucketName, s3MountPath,
				"-o", "passwd_file="+s3fsPasswdFile,
				"-o", "use_cache=/tmp/s3cache",
				"-o", "allow_other",
				"-o", "url="+s3Endpoint,
				"-o", "use_path_request_style",
			)

			if err := cmd.Run(); err != nil {
				utils.Fatalf("Error mounting Object storage:", err)
			}

			if err := os.MkdirAll(storagePath, os.ModePerm); err != nil {
				utils.Fatalf("Error creating directory %v %v", storagePath, err)
			}

		} else {
			utils.Info("Object storage already mounted in " + s3MountPath)
			if err := os.MkdirAll(storagePath, os.ModePerm); err != nil {
				utils.Fatal("Error creating directory "+storagePath, err)
			}

		}

	}
}
func s3Restore() {
	// Implement S3 restore logic\
	s3Mount()
	restore()
}

func createCrontabScript() {
	task := "/usr/local/bin/backup_cron.sh"
	touchCmd := exec.Command("touch", task)
	if err := touchCmd.Run(); err != nil {
		utils.Fatalf("Error creating file %s: %v\n", task, err)
	}
	var disableC string = ""
	if disableCompression {
		disableC = "--disable-compression"
	}

	var scriptContent string

	if storage == "s3" {
		scriptContent = fmt.Sprintf(`#!/usr/bin/env bash
set -e
bkup --operation backup --dbname %s --port %s --storage s3 --path %s %v
`, os.Getenv("DB_NAME"), os.Getenv("DB_PORT"), os.Getenv("S3_PATH"), disableC)
	} else {
		scriptContent = fmt.Sprintf(`#!/usr/bin/env bash
set -e
bkup --operation backup --dbname %s --port %s %v
`, os.Getenv("DB_NAME"), os.Getenv("DB_PORT"), disableC)
	}

	if err := utils.WriteToFile(task, scriptContent); err != nil {
		utils.Fatalf("Error writing to %s: %v\n", task, err)
	}

	chmodCmd := exec.Command("chmod", "+x", "/usr/local/bin/backup_cron.sh")
	if err := chmodCmd.Run(); err != nil {
		utils.Fatalf("Error changing permissions of %s: %v\n", task, err)
	}

	lnCmd := exec.Command("ln", "-s", "/usr/local/bin/backup_cron.sh", "/usr/local/bin/backup_cron")
	if err := lnCmd.Run(); err != nil {
		utils.Fatalf("Error creating symbolic link: %v\n", err)

	}

	cronJob := "/etc/cron.d/backup_cron"
	touchCronCmd := exec.Command("touch", cronJob)
	if err := touchCronCmd.Run(); err != nil {
		utils.Fatalf("Error creating file %s: %v\n", cronJob, err)
	}

	cronContent := fmt.Sprintf(`%s root exec /bin/bash -c ". /run/supervisord.env; /usr/local/bin/backup_cron.sh >> /var/log/mysql-bkup.log"
`, os.Getenv("SCHEDULE_PERIOD"))

	if err := utils.WriteToFile(cronJob, cronContent); err != nil {
		utils.Fatalf("Error writing to %s: %v\n", cronJob, err)
	}
	utils.ChangePermission("/etc/cron.d/backup_cron", 0644)

	crontabCmd := exec.Command("crontab", "/etc/cron.d/backup_cron")
	if err := crontabCmd.Run(); err != nil {
		utils.Fatal("Error updating crontab: ", err)
	}
	utils.Info("Starting backup in scheduled mode")
}

// testDatabaseConnection tests the database connection
func testDatabaseConnection() {
	utils.Info("Testing database connection...")
	// Test database connection

}
