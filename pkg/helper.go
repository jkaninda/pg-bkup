/*
 *  MIT License
 *
 * Copyright (c) 2024 Jonas Kaninda
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 */

package pkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jkaninda/logger"
	"github.com/jkaninda/pg-bkup/utils"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func intro() {
	fmt.Println("Starting PG-BKUP...")
	fmt.Printf("Version: %s\n", utils.Version)
	fmt.Println("Copyright (c) 2024 Jonas Kaninda")
}

// copyToTmp copy file to temporary directory
func deleteTemp() {
	logger.Info(fmt.Sprintf("Deleting %s ...", tmpPath))
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
		logger.Error("Error deleting files", "error", err)
	} else {
		logger.Info(fmt.Sprintf("Deleting %s ... done", tmpPath))
	}
}

// TestDatabaseConnection  tests the database connection
func testDatabaseConnection(db *dbConfig) error {

	logger.Info(fmt.Sprintf("Connecting to %s database ...", db.dbName))
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", db.dbUserName, db.dbPassword, db.dbHost, db.dbPort, db.dbName)
	// Create the PostgresSQL client config file
	if err := createPGConfigFile(*db); err != nil {
		return errors.New(err.Error())
	}
	// Set database name for notification error
	utils.DatabaseName = db.dbName
	if db.dbName == "" {
		connString = fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable", db.dbUserName, db.dbPassword, db.dbHost, db.dbPort)
	}

	// Attempt to connect to the PostgreSQL server
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer func(conn *pgx.Conn, ctx context.Context) {
		err = conn.Close(ctx)
		if err != nil {
			logger.Error("Error closing connexion", "error", err)

		}
	}(conn, context.Background())

	// Optionally, execute a simple query to verify the connection
	var version string
	err = conn.QueryRow(context.Background(), "SELECT version()").Scan(&version)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	logger.Info(fmt.Sprintf("Successfully connected to %s database", db.dbName))
	return nil

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
	// Remove the quotes
	filePath = strings.Trim(filePath, `"`)
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
func convertJDBCToDbConfig(jdbcURI string) (*dbConfig, error) {
	// Remove the "jdbc:" prefix
	jdbcURI = strings.TrimPrefix(jdbcURI, "jdbc:")
	// Parse the URI
	u, err := url.Parse(jdbcURI)
	if err != nil {
		return &dbConfig{}, fmt.Errorf("failed to parse JDBC URI: %v", err)
	}
	// Extract components
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "5432" // Default PostgreSQL port
	}
	database := strings.TrimPrefix(u.Path, "/")
	params, _ := url.ParseQuery(u.RawQuery)
	username := params.Get("user")
	password := params.Get("password")
	// Validate essential fields
	if host == "" || database == "" || username == "" {
		return &dbConfig{}, fmt.Errorf("incomplete JDBC URI: missing host, database, or username")
	}

	return &dbConfig{
		dbHost:     host,
		dbPort:     port,
		dbName:     database,
		dbUserName: username,
		dbPassword: password,
	}, nil
}

// Create mysql client config file
func createPGConfigFile(db dbConfig) error {

	// Set the environment variable for the database password
	err := os.Setenv("PGPASSWORD", db.dbPassword)
	if err != nil {
		return fmt.Errorf("failed to set PGPASSWORD environment variable: %v", err)
	}
	return nil
}
