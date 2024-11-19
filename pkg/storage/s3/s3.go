package s3

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	pkg "github.com/jkaninda/pg-bkup/pkg/storage"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type s3Storage struct {
	*pkg.Backend
	client *session.Session
	bucket string
}

// Config holds the AWS S3 config
type Config struct {
	Endpoint       string
	Bucket         string
	AccessKey      string
	SecretKey      string
	Region         string
	DisableSsl     bool
	ForcePathStyle bool
	LocalPath      string
	RemotePath     string
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

// NewStorage creates new Storage
func NewStorage(conf Config) (pkg.Storage, error) {
	sess, err := createSession(conf)
	if err != nil {
		return nil, err
	}
	return &s3Storage{
		client: sess,
		bucket: conf.Bucket,
		Backend: &pkg.Backend{
			RemotePath: conf.RemotePath,
			LocalPath:  conf.LocalPath,
		},
	}, nil
}

// Copy copies file to S3 storage
func (s s3Storage) Copy(fileName string) error {
	svc := s3.New(s.client)
	file, err := os.Open(filepath.Join(s.LocalPath, fileName))
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	objectKey := filepath.Join(s.RemotePath, fileName)
	buffer := make([]byte, fileInfo.Size())
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}
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

// CopyFrom copies a file from S3 to local storage
func (s s3Storage) CopyFrom(fileName string) error {
	file, err := os.Create(filepath.Join(s.LocalPath, fileName))
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v\n", err)
			return
		}
	}(file)

	objectKey := filepath.Join(s.RemotePath, fileName)

	downloader := s3manager.NewDownloader(s.client)
	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(objectKey),
		})
	if err != nil {
		return err
	}
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
				// Object is older than retention days, delete it
				_, err := svc.DeleteObject(&s3.DeleteObjectInput{
					Bucket: aws.String(s.bucket),
					Key:    object.Key,
				})
				if err != nil {
					fmt.Printf("failed to delete object %s: %v", *object.Key, err)
				} else {
					fmt.Printf("Deleted object %s", *object.Key)
				}
			}
		}
		return !lastPage
	})
	if err != nil {
		return fmt.Errorf("failed to list objects: %v", err)
	}

	return nil

}

// Name returns the storage name
func (s s3Storage) Name() string {
	return "s3"
}
