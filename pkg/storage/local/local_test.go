package local

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const content = "Lorem ipsum dolor sit amet. Eum eius voluptas sit vitae vitae aut sequi molestias hic accusamus consequatur"
const inputFile = "file.txt"
const localPath = "./tests/local"
const RemotePath = "./tests/remote"

func TestCopy(t *testing.T) {

	err := os.MkdirAll(localPath, 0777)
	if err != nil {
		t.Error(err)
	}
	err = os.MkdirAll(RemotePath, 0777)
	if err != nil {
		t.Error(err)
	}

	_, err = createFile(filepath.Join(localPath, inputFile), content)
	if err != nil {
		t.Error(err)
	}

	l := NewStorage(Config{
		LocalPath:  "./tests/local",
		RemotePath: "./tests/remote",
	})
	err = l.Copy(inputFile)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("File copied to %s\n", filepath.Join(RemotePath, inputFile))
}
func createFile(fileName, content string) ([]byte, error) {
	// Create a file named hello.txt
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
			return
		}
	}(file)

	// Write the message to the file
	_, err = file.WriteString(content)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return nil, err
	}

	fmt.Printf("Successfully wrote to %s\n", fileName)
	fileBytes, err := os.ReadFile(fileName)
	return fileBytes, err
}
