package pkg

import (
	"fmt"
	"github.com/jlaffaye/ftp"
	"io"
	"os"
	"path/filepath"
	"time"
)

// initFtpClient initializes and authenticates an FTP client
func initFtpClient() (*ftp.ServerConn, error) {
	ftpConfig := initFtpConfig()
	ftpClient, err := ftp.Dial(fmt.Sprintf("%s:%s", ftpConfig.host, ftpConfig.port), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to FTP: %w", err)
	}

	err = ftpClient.Login(ftpConfig.user, ftpConfig.password)
	if err != nil {
		return nil, fmt.Errorf("failed to log in to FTP: %w", err)
	}

	return ftpClient, nil
}

// CopyToFTP uploads a file to the remote FTP server
func CopyToFTP(fileName, remotePath string) (err error) {
	ftpConfig := initFtpConfig()
	ftpClient, err := initFtpClient()
	if err != nil {
		return err
	}
	defer ftpClient.Quit()

	filePath := filepath.Join(tmpPath, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", fileName, err)
	}
	defer file.Close()

	remoteFilePath := filepath.Join(ftpConfig.remotePath, fileName)
	err = ftpClient.Stor(remoteFilePath, file)
	if err != nil {
		return fmt.Errorf("failed to upload file %s: %w", fileName, err)
	}

	return nil
}

// CopyFromFTP downloads a file from the remote FTP server
func CopyFromFTP(fileName, remotePath string) (err error) {
	ftpClient, err := initFtpClient()
	if err != nil {
		return err
	}
	defer ftpClient.Quit()

	remoteFilePath := filepath.Join(remotePath, fileName)
	r, err := ftpClient.Retr(remoteFilePath)
	if err != nil {
		return fmt.Errorf("failed to retrieve file %s: %w", fileName, err)
	}
	defer r.Close()

	localFilePath := filepath.Join(tmpPath, fileName)
	outFile, err := os.Create(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", fileName, err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, r)
	if err != nil {
		return fmt.Errorf("failed to copy data to local file %s: %w", fileName, err)
	}

	return nil
}
