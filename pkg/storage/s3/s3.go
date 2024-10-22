package s3

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jkaninda/pg-bkup/pkg/storage"
	"github.com/jkaninda/pg-bkup/utils"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type s3Storage struct {
	*storage.Backend
	client *session.Session
	bucket string
}
type Config struct {
	Endpoint       string
	Bucket         string
	AccessKey      string
	SecretKey      string
	Region         string
	DisableSsl     bool
	ForcePathStyle bool
	LocalPath string
	RemotePath string
}

// CreateSession creates a new AWS session
func createSession(conf Config) (*session.Session, error) {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(conf.AccessKey, conf.SecretKey, ""),
		Endpoint:         aws.String(conf.Endpoint),
		Region:           aws.String(conf.Region),
		DisableSSL:       aws.Bool(conf.DisableSsl),
		S3ForcePathStyle: aws.Bool(conf.ForcePathStyle),
	}

	return session.NewSession(s3Config)
}

func NewStorage(conf Config) (storage.Storage, error) {
	sess, err := createSession(conf)
	if err != nil {
		return nil, err
	}
	return &s3Storage{
		client: sess,
		bucket: conf.Bucket,
		Backend: &storage.Backend{
			RemotePath: conf.RemotePath,
			LocalPath: conf.LocalPath,
		},
	}, nil
}
func (s s3Storage) Copy(fileName string) error {
	svc := s3.New(s.client)
	file, err := os.Open(filepath.Join(s.LocalPath, fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	objectKey := filepath.Join(s.RemotePath, fileName)
	buffer := make([]byte, fileInfo.Size())
	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(objectKey),
		Body:          fileBytes,
		ContentLength: aws.Int64(fileInfo.Size()),
		ContentType:   aws.String(fileType),
	})
	if err != nil {
		return err
	}

	return nil
}

// CopyFrom copies a file from the remote server to local storage
func (s s3Storage) CopyFrom(fileName string) error {
	utils.Info("Copy data from S3 storage...")
	file, err := os.Create(filepath.Join(s.LocalPath, fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	objectKey := filepath.Join(s.RemotePath, fileName)

	downloader := s3manager.NewDownloader(s.client)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(objectKey),
		})
	if err != nil {
		utils.Error("Failed to download file %s", fileName)
		return err
	}
	utils.Info("Backup downloaded:  %s , bytes size: %d ", file.Name(), uint64(numBytes))

	return nil
}

// Prune deletes old backup created more than specified days
func (s s3Storage) Prune(retentionDays int) error {
	svc := s3.New(s.client)

	// Get the current time
	now := time.Now()
	backupRetentionDays := now.AddDate(0, 0, -retentionDays)

	// List objects in the bucket
	listObjectsInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(s.RemotePath),
	}
	err := svc.ListObjectsV2Pages(listObjectsInput, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, object := range page.Contents {
			if object.LastModified.Before(backupRetentionDays) {
				utils.Info("Deleting old backup: %s", *object.Key)
				// Object is older than retention days, delete it
				_, err := svc.DeleteObject(&s3.DeleteObjectInput{
					Bucket: aws.String(s.bucket),
					Key:    object.Key,
				})
				if err != nil {
					utils.Info("Failed to delete object %s: %v", *object.Key, err)
				} else {
					utils.Info("Deleted object %s", *object.Key)
				}
			}
		}
		return !lastPage
	})
	if err != nil {
		utils.Error("Failed to list objects: %v", err)
	}

	utils.Info("Deleting old backups...done")
	return nil

}

// Name returns the storage name
func (s s3Storage) Name() string {
	return "s3"
}
