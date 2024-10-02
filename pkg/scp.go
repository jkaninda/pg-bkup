// Package pkg /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package pkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/jkaninda/pg-bkup/utils"
	"golang.org/x/crypto/ssh"
	"os"
	"path/filepath"
)

// createSSHClientConfig sets up the SSH client configuration based on the provided SSHConfig
func createSSHClientConfig(sshConfig *SSHConfig) (ssh.ClientConfig, error) {
	if sshConfig.identifyFile != "" && utils.FileExists(sshConfig.identifyFile) {
		return auth.PrivateKey(sshConfig.user, sshConfig.identifyFile, ssh.InsecureIgnoreHostKey())
	} else {
		if sshConfig.password == "" {
			return ssh.ClientConfig{}, errors.New("SSH_PASSWORD environment variable is required if SSH_IDENTIFY_FILE is empty")
		}
		utils.Warn("Accessing the remote server using password, which is not recommended.")
		return auth.PasswordKey(sshConfig.user, sshConfig.password, ssh.InsecureIgnoreHostKey())
	}
}

// CopyToRemote copies a file to a remote server via SCP
func CopyToRemote(fileName, remotePath string) error {
	// Load environment variables
	sshConfig, err := loadSSHConfig()
	if err != nil {
		return fmt.Errorf("failed to load SSH configuration: %w", err)
	}

	// Initialize SSH client config
	clientConfig, err := createSSHClientConfig(sshConfig)
	if err != nil {
		return fmt.Errorf("failed to create SSH client config: %w", err)
	}

	// Create a new SCP client
	client := scp.NewClient(fmt.Sprintf("%s:%s", sshConfig.hostName, sshConfig.port), &clientConfig)

	// Connect to the remote server
	err = client.Connect()
	if err != nil {
		return errors.New("Couldn't establish a connection to the remote server\n")
	}

	// Open the local file
	filePath := filepath.Join(tmpPath, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer client.Close()
	// Copy file to the remote server
	err = client.CopyFromFile(context.Background(), *file, filepath.Join(remotePath, fileName), "0655")
	if err != nil {
		return fmt.Errorf("failed to copy file to remote server: %w", err)
	}

	return nil
}

func CopyFromRemote(fileName, remotePath string) error {
	// Load environment variables
	sshConfig, err := loadSSHConfig()
	if err != nil {
		return fmt.Errorf("failed to load SSH configuration: %w", err)
	}

	// Initialize SSH client config
	clientConfig, err := createSSHClientConfig(sshConfig)
	if err != nil {
		return fmt.Errorf("failed to create SSH client config: %w", err)
	}

	// Create a new SCP client
	client := scp.NewClient(fmt.Sprintf("%s:%s", sshConfig.hostName, sshConfig.port), &clientConfig)

	// Connect to the remote server
	err = client.Connect()
	if err != nil {
		return errors.New("Couldn't establish a connection to the remote server\n")
	}
	// Close client connection after the file has been copied
	defer client.Close()
	file, err := os.OpenFile(filepath.Join(tmpPath, fileName), os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println("Couldn't open the output file")
	}
	defer file.Close()

	err = client.CopyFromRemote(context.Background(), file, filepath.Join(remotePath, fileName))

	if err != nil {
		utils.Error("Error while copying file %s ", err)
		return err
	}
	return nil

}
