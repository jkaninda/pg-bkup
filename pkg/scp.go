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

func CopyToRemote(fileName, remotePath string) error {
	sshUser := os.Getenv("SSH_USER")
	sshPassword := os.Getenv("SSH_PASSWORD")
	sshHostName := os.Getenv("SSH_HOST_NAME")
	sshPort := os.Getenv("SSH_PORT")
	sshIdentifyFile := os.Getenv("SSH_IDENTIFY_FILE")

	err := utils.CheckEnvVars(sshVars)
	if err != nil {
		utils.Fatal(fmt.Sprintf("Error checking environment variables\n: %s", err))
	}

	clientConfig, _ := auth.PasswordKey(sshUser, sshPassword, ssh.InsecureIgnoreHostKey())
	if sshIdentifyFile != "" && utils.FileExists(sshIdentifyFile) {
		clientConfig, _ = auth.PrivateKey(sshUser, sshIdentifyFile, ssh.InsecureIgnoreHostKey())

	} else {
		if sshPassword == "" {
			return errors.New("SSH_PASSWORD environment variable is required if SSH_IDENTIFY_FILE is empty")
		}
		utils.Warn("Accessing the remote server using password, password is not recommended")
		clientConfig, _ = auth.PasswordKey(sshUser, sshPassword, ssh.InsecureIgnoreHostKey())

	}
	// Create a new SCP client
	client := scp.NewClient(fmt.Sprintf("%s:%s", sshHostName, sshPort), &clientConfig)

	// Connect to the remote server
	err = client.Connect()
	if err != nil {
		return errors.New("couldn't establish a connection to the remote server")
	}

	// Open a file
	file, _ := os.Open(filepath.Join(tmpPath, fileName))

	// Close client connection after the file has been copied
	defer client.Close()
	// Close the file after it has been copied
	defer file.Close()
	// the context can be adjusted to provide time-outs or inherit from other contexts if this is embedded in a larger application.
	err = client.CopyFromFile(context.Background(), *file, filepath.Join(remotePath, fileName), "0655")
	if err != nil {
		fmt.Println("Error while copying file ")
		return err
	}
	return nil
}

func CopyFromRemote(fileName, remotePath string) error {
	sshUser := os.Getenv("SSH_USER")
	sshPassword := os.Getenv("SSH_PASSWORD")
	sshHostName := os.Getenv("SSH_HOST_NAME")
	sshPort := os.Getenv("SSH_PORT")
	sshIdentifyFile := os.Getenv("SSH_IDENTIFY_FILE")

	err := utils.CheckEnvVars(sshVars)
	if err != nil {
		utils.Fatal("Error checking environment variables\n: %s", err)
	}

	clientConfig, _ := auth.PasswordKey(sshUser, sshPassword, ssh.InsecureIgnoreHostKey())
	if sshIdentifyFile != "" && utils.FileExists(sshIdentifyFile) {
		clientConfig, _ = auth.PrivateKey(sshUser, sshIdentifyFile, ssh.InsecureIgnoreHostKey())

	} else {
		if sshPassword == "" {
			return errors.New("SSH_PASSWORD environment variable is required if SSH_IDENTIFY_FILE is empty\n")
		}
		utils.Warn("Accessing the remote server using password, password is not recommended\n")
		clientConfig, _ = auth.PasswordKey(sshUser, sshPassword, ssh.InsecureIgnoreHostKey())

	}
	// Create a new SCP client
	client := scp.NewClient(fmt.Sprintf("%s:%s", sshHostName, sshPort), &clientConfig)

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

	// the context can be adjusted to provide time-outs or inherit from other contexts if this is embedded in a larger application.
	err = client.CopyFromRemote(context.Background(), file, filepath.Join(remotePath, fileName))

	if err != nil {
		utils.Error("Error while copying file %s ", err)
		return err
	}
	return nil

}
