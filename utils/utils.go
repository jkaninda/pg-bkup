/*
MIT License

Copyright (c) 2023 Jonas Kaninda

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package utils

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"io"
	"io/fs"
	"os"
	"strconv"
	"time"
)

var Version = ""

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
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return

		}
	}(file)

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
	defer func(sourceFile *os.File) {
		err := sourceFile.Close()
		if err != nil {
			return
		}
	}(sourceFile)

	// Create the destination file
	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer func(destinationFile *os.File) {
		err := destinationFile.Close()
		if err != nil {
			return

		}
	}(destinationFile)

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
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			return
		}
	}(f)

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

func EnvWithDefault(envName string, defaultValue string) string {
	value := os.Getenv(envName)
	if value == "" {
		return defaultValue
	}
	return value
}

// IsValidCronExpression verify cronExpression and returns boolean
func IsValidCronExpression(cronExpr string) bool {
	// Parse the cron expression
	_, err := cron.ParseStandard(cronExpr)
	return err == nil
}

// CronNextTime returns cronExpression next time
func CronNextTime(cronExpr string) time.Time {
	// Parse the cron expression
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		Error("Error parsing cron expression: %s", err)
		return time.Time{}
	}
	// Get the current time
	now := time.Now()
	// Get the next scheduled time
	next := schedule.Next(now)
	return next
}

// ConvertBytes converts bytes to a human-readable string with the appropriate unit (bytes, MiB, or GiB).
func ConvertBytes(bytes uint64) string {
	const (
		MiB = 1024 * 1024
		GiB = MiB * 1024
	)
	switch {
	case bytes >= GiB:
		return fmt.Sprintf("%.2f GiB", float64(bytes)/float64(GiB))
	case bytes >= MiB:
		return fmt.Sprintf("%.2f MiB", float64(bytes)/float64(MiB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
