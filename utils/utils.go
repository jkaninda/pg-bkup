// Package utils /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package utils

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"io/fs"
	"os"
	"strconv"
)

// FileExists checks if the file does exist
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func WriteToFile(filePath, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}
func DeleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	return nil
}
func CopyFile(src, dst string) error {
	// Open the source file for reading
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer sourceFile.Close()

	// Create the destination file
	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destinationFile.Close()

	// Copy the content from source to destination
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	// Flush the buffer to ensure all data is written
	err = destinationFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file: %v", err)
	}

	return nil
}
func ChangePermission(filePath string, mod int) {
	if err := os.Chmod(filePath, fs.FileMode(mod)); err != nil {
		Fatal("Error changing permissions of %s: %v\n", filePath, err)
	}

}
func IsDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == nil {
		return false, nil
	}
	return true, nil
}

func GetEnv(cmd *cobra.Command, flagName, envName string) string {
	value, _ := cmd.Flags().GetString(flagName)
	if value != "" {
		err := os.Setenv(envName, value)
		if err != nil {
			return value
		}
	}
	return os.Getenv(envName)
}
func FlagGetString(cmd *cobra.Command, flagName string) string {
	value, _ := cmd.Flags().GetString(flagName)
	if value != "" {
		return value

	}
	return ""
}
func FlagGetBool(cmd *cobra.Command, flagName string) bool {
	value, _ := cmd.Flags().GetBool(flagName)
	return value
}

func SetEnv(key, value string) {

	err := os.Setenv(key, value)
	if err != nil {
		return
	}
}
func GetEnvVariable(envName, oldEnvName string) string {
	value := os.Getenv(envName)
	if value == "" {
		value = os.Getenv(oldEnvName)
		if value != "" {
			err := os.Setenv(envName, value)
			if err != nil {
				return value
			}
			Warn("%s is deprecated, please use %s instead! ", oldEnvName, envName)
		}
	}
	return value
}

// CheckEnvVars checks if all the specified environment variables are set
func CheckEnvVars(vars []string) error {
	missingVars := []string{}

	for _, v := range vars {
		if os.Getenv(v) == "" {
			missingVars = append(missingVars, v)
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing environment variables: %v", missingVars)
	}

	return nil
}

// MakeDir create directory
func MakeDir(dirPath string) error {
	err := os.Mkdir(dirPath, 0700)
	if err != nil {
		return err
	}
	return nil
}

// MakeDirAll create directory
func MakeDirAll(dirPath string) error {
	err := os.MkdirAll(dirPath, 0700)
	if err != nil {
		return err
	}
	return nil
}
func GetIntEnv(envName string) int {
	val := os.Getenv(envName)
	if val == "" {
		return 0
	}
	ret, err := strconv.Atoi(val)
	if err != nil {
		Error("Error: %v", err)
	}
	return ret
}
