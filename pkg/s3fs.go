// Package pkg /*
/*
Copyright © 2024 Jonas Kaninda  <jonaskaninda.gmail.com>
*/
package pkg

import (
	"fmt"
	"github.com/jkaninda/pg-bkup/utils"
	"os"
	"os/exec"
)

const s3MountPath string = "/s3mnt"
const s3fsPasswdFile string = "/etc/passwd-s3fs"

var (
	accessKey  = ""
	secretKey  = ""
	bucketName = ""
	s3Endpoint = ""
)

func init() {
	accessKey = os.Getenv("ACCESS_KEY")
	secretKey = os.Getenv("SECRET_KEY")
	bucketName = os.Getenv("BUCKETNAME")
	s3Endpoint = os.Getenv("S3_ENDPOINT")
}

// MountS3Storage Mount s3 storage using s3fs
func MountS3Storage(s3Path string) {
	if accessKey == "" || secretKey == "" || bucketName == "" {
		utils.Fatal("Please make sure all environment variables are set")
	} else {
		storagePath := fmt.Sprintf("%s%s", s3MountPath, s3Path)
		err := os.Setenv("STORAGE_PATH", storagePath)
		if err != nil {
			return
		}

		//Write file
		err = utils.WriteToFile(s3fsPasswdFile, fmt.Sprintf("%s:%s", accessKey, secretKey))
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
				utils.Fatal("Error mounting Object storage:", err)
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
