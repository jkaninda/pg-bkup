// Package pkg /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package pkg

import (
	"errors"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/jkaninda/pg-bkup/utils"
	"os"
	"os/exec"
	"strings"
)

// Decrypt decrypts backup file using a passphrase
func Decrypt(inputFile string, passphrase string) error {
	utils.Info("Decrypting backup using passphrase...")

	//Create gpg home dir
	err := utils.MakeDirAll(gpgHome)
	if err != nil {
		return err
	}
	utils.SetEnv("GNUPGHOME", gpgHome)
	cmd := exec.Command("gpg", "--batch", "--passphrase", passphrase, "--output", RemoveLastExtension(inputFile), "--decrypt", inputFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}
	utils.Info("Decrypting backup using passphrase...done")
	utils.Info("Backup file decrypted successful!")
	return nil
}

// Encrypt encrypts backup using a passphrase
func Encrypt(inputFile string, passphrase string) error {
	utils.Info("Encrypting backup using passphrase...")

	//Create gpg home dir
	err := utils.MakeDirAll(gpgHome)
	if err != nil {
		return err
	}
	utils.SetEnv("GNUPGHOME", gpgHome)
	cmd := exec.Command("gpg", "--batch", "--passphrase", passphrase, "--symmetric", "--cipher-algo", algorithm, inputFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}
	utils.Info("Encrypting backup using passphrase...done")
	utils.Info("Backup file encrypted successful!")
	return nil
}

// encrypt encrypts backup using a public key
func encrypt(inputFile string, publicKey string) error {
	utils.Info("Encrypting backup using public key...")
	// Read the public key
	pubKeyBytes, err := os.ReadFile(publicKey)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading public key: %s", err))
	}
	// Create a new keyring with the public key
	publicKeyObj, err := crypto.NewKeyFromArmored(string(pubKeyBytes))
	if err != nil {
		return errors.New(fmt.Sprintf("Error parsing public key: %s", err))
	}

	keyRing, err := crypto.NewKeyRing(publicKeyObj)
	if err != nil {

		return errors.New(fmt.Sprintf("Error creating key ring: %v", err))
	}

	// Read the file to encrypt
	fileContent, err := os.ReadFile(inputFile)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading file: %v", err))
	}

	// Encrypt the file
	message := crypto.NewPlainMessage(fileContent)
	encMessage, err := keyRing.Encrypt(message, nil)
	if err != nil {
		return errors.New(fmt.Sprintf("Error encrypting file: %v", err))
	}

	// Save the encrypted file
	err = os.WriteFile(fmt.Sprintf("%s.%s", inputFile, gpgExtension), encMessage.GetBinary(), 0644)
	if err != nil {
		return errors.New(fmt.Sprintf("Error saving encrypted file: %v", err))
	}
	utils.Info("Encrypting backup using public key...done")
	utils.Info("Backup file encrypted successful!")
	return nil

}

// decrypt decrypts backup file using a private key and passphrase.
// privateKey GPG private key
// passphrase GPG passphrase
func decrypt(inputFile, privateKey, passphrase string) error {
	utils.Info("Encrypting backup using private key...")

	// Read the private key
	priKeyBytes, err := os.ReadFile(privateKey)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading private key: %s", err))
	}

	// Read the password for the private key (if it’s password-protected)
	password := []byte(passphrase)

	// Create a key object from the armored private key
	privateKeyObj, err := crypto.NewKeyFromArmored(string(priKeyBytes))
	if err != nil {
		return errors.New(fmt.Sprintf("Error parsing private key: %s", err))
	}

	// Unlock the private key with the password
	if passphrase != "" {
		// Unlock the private key with the password
		_, err = privateKeyObj.Unlock(password)
		if err != nil {
			return errors.New(fmt.Sprintf("Error unlocking private key: %s", err))
		}

	}

	// Create a new keyring with the private key
	keyRing, err := crypto.NewKeyRing(privateKeyObj)
	if err != nil {
		return errors.New(fmt.Sprintf("Error creating key ring: %v", err))
	}

	// Read the encrypted file
	encFileContent, err := os.ReadFile(inputFile)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading encrypted file: %s", err))
	}

	// Decrypt the file
	encryptedMessage := crypto.NewPGPMessage(encFileContent)
	message, err := keyRing.Decrypt(encryptedMessage, nil, 0)
	if err != nil {
		return errors.New(fmt.Sprintf("Error decrypting file: %s", err))
	}

	// Save the decrypted file
	err = os.WriteFile(RemoveLastExtension(inputFile), message.GetBinary(), 0644)
	if err != nil {
		return errors.New(fmt.Sprintf("Error saving decrypted file: %s", err))
	}
	utils.Info("Encrypting backup using public key...done")
	fmt.Println("File successfully decrypted!")
	return nil
}
func RemoveLastExtension(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		return filename[:idx]
	}
	return filename
}
