package utils

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"net/http"
	"os"
	"path/filepath"
)

// CreateSession creates a new AWS session
func CreateSession() (*session.Session, error) {

	//key := aws.String("testobject")
	endPoint := os.Getenv("S3_ENDPOINT")
	//bucket := os.Getenv("BUCKET_NAME")
	region := os.Getenv("REGION")
	accessKey := os.Getenv("ACCESS_KEY")
	secretKey := os.Getenv("SECRET_KEY")

	// Configure to use MinIO Server
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endPoint),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(true),
	}
	return session.NewSession(s3Config)

}

// UploadFileToS3 uploads a file to S3 with a given prefix
func UploadFileToS3(filePath, key, bucket, prefix string) error {
	sess, err := CreateSession()
	if err != nil {
		return err
	}

	svc := s3.New(sess)

	file, err := os.Open(filepath.Join(filePath, key))
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	objectKey := fmt.Sprintf("%s/%s", prefix, key)

	buffer := make([]byte, fileInfo.Size())
	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(bucket),
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
func DownloadFile(destinationPath, key, bucket, prefix string) error {

	sess, err := CreateSession()
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(destinationPath, key))
	if err != nil {
		fmt.Println("Failed to create file", err)
		return err
	}
	defer file.Close()

	objectKey := fmt.Sprintf("%s/%s", prefix, key)

	downloader := s3manager.NewDownloader(sess)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(objectKey),
		})
	if err != nil {
		fmt.Println("Failed to download file", err)
		return err
	}
	fmt.Println("Bytes size", numBytes)
	Info("Backup downloaded to ", file.Name())
	return nil
}
