// Package pkg /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package pkg

import (
	"bytes"
	"fmt"
	"github.com/jkaninda/pg-bkup/utils"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func intro() {
	utils.Info("Starting PostgreSQL Backup...")
	utils.Info("Copyright (c) 2024 Jonas Kaninda ")
}

// copyToTmp copy file to temporary directory
func deleteTemp() {
	utils.Info("Deleting %s ...", tmpPath)
	err := filepath.Walk(tmpPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if the current item is a file
		if !info.IsDir() {
			// Delete the file
			err = os.Remove(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		utils.Error("Error deleting files: %v", err)
	} else {
		utils.Info("Deleting %s ... done", tmpPath)
	}
}

// TestDatabaseConnection  tests the database connection
func testDatabaseConnection(db *dbConfig) {

	utils.Info("Connecting to %s database ...", db.dbName)
	// Test database connection
	query := "SELECT version();"

	// Set the environment variable for the database password
	err := os.Setenv("PGPASSWORD", db.dbPassword)
	if err != nil {
		return
	}
	// Prepare the psql command
	cmd := exec.Command("psql",
		"-U", db.dbUserName, // database user
		"-d", db.dbName, // database name
		"-h", db.dbHost, // host
		"-p", db.dbPort, // port
		"-c", query, // SQL command to execute
	)
	// Capture the output
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Run the command and capture any errors
	err = cmd.Run()
	if err != nil {
		utils.Fatal("Error running psql command: %v\nOutput: %s\n", err, out.String())
		return
	}
	utils.Info("Successfully connected to %s database", db.dbName)

}

// checkPubKeyFile checks gpg public key
func checkPubKeyFile(pubKey string) (string, error) {
	// Define possible key file names
	keyFiles := []string{filepath.Join(gpgHome, "public_key.asc"), filepath.Join(gpgHome, "public_key.gpg"), pubKey}

	// Loop through key file names and check if they exist
	for _, keyFile := range keyFiles {
		if _, err := os.Stat(keyFile); err == nil {
			// File exists
			return keyFile, nil
		} else if os.IsNotExist(err) {
			// File does not exist, continue to the next one
			continue
		} else {
			// An unexpected error occurred
			return "", err
		}
	}

	// Return an error if neither file exists
	return "", fmt.Errorf("no public key file found")
}

// checkPrKeyFile checks private key
func checkPrKeyFile(prKey string) (string, error) {
	// Define possible key file names
	keyFiles := []string{filepath.Join(gpgHome, "private_key.asc"), filepath.Join(gpgHome, "private_key.gpg"), prKey}

	// Loop through key file names and check if they exist
	for _, keyFile := range keyFiles {
		if _, err := os.Stat(keyFile); err == nil {
			// File exists
			return keyFile, nil
		} else if os.IsNotExist(err) {
			// File does not exist, continue to the next one
			continue
		} else {
			// An unexpected error occurred
			return "", err
		}
	}

	// Return an error if neither file exists
	return "", fmt.Errorf("no public key file found")
}

// readConf reads config file and returns Config
func readConf(configFile string) (*Config, error) {
	if utils.FileExists(configFile) {
		buf, err := os.ReadFile(configFile)
		if err != nil {
			return nil, err
		}

		c := &Config{}
		err = yaml.Unmarshal(buf, c)
		if err != nil {
			return nil, fmt.Errorf("in file %q: %w", configFile, err)
		}

		return c, err
	}
	return nil, fmt.Errorf("config file %q not found", configFile)
}

// checkConfigFile checks config files and returns one config file
func checkConfigFile(filePath string) (string, error) {
	// Define possible config file names
	configFiles := []string{filepath.Join(workingDir, "config.yaml"), filepath.Join(workingDir, "config.yml"), filePath}

	// Loop through config file names and check if they exist
	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			// File exists
			return configFile, nil
		} else if os.IsNotExist(err) {
			// File does not exist, continue to the next one
			continue
		} else {
			// An unexpected error occurred
			return "", err
		}
	}

	// Return an error if neither file exists
	return "", fmt.Errorf("no config file found")
}
func RemoveLastExtension(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		return filename[:idx]
	}
	return filename
}
