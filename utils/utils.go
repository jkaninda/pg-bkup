package utils

/*****
*   PostgreSQL Backup & Restore
* @author    Jonas Kaninda
* @license   MIT License <https://opensource.org/licenses/MIT>
* @link      https://github.com/jkaninda/pg-bkup
**/
import (
	"fmt"
	"io/fs"
	"os"
)

func Info(v ...any) {
	fmt.Println("[INFO] ", fmt.Sprint(v...))
}
func Infof(msg string, v ...any) {
	fmt.Printf("[INFO] "+msg, v...)
}
func Warning(message string) {
	fmt.Println("[WARNING]", message)
}
func Warningf(msg string, v ...any) {
	fmt.Printf("[WARNING] "+msg, v...)
}
func Fatal(v ...any) {
	fmt.Println("[ERROR] ", fmt.Sprint(v...))
	os.Exit(1)
}
func Fatalf(msg string, v ...any) {
	fmt.Printf("[ERROR] "+msg, v...)
	os.Exit(1)
}

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
func ChangePermission(filePath string, mod int) {
	if err := os.Chmod(filePath, fs.FileMode(mod)); err != nil {
		Fatalf("Error changing permissions of %s: %v\n", filePath, err)
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
