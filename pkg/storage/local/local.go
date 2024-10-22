package local

import (
	"github.com/jkaninda/pg-bkup/pkg/storage"
	"github.com/jkaninda/pg-bkup/utils"
	"io"
	"os"
	"path/filepath"
	"time"
)

type localStorage struct {
	*storage.Backend
}
type Config struct {
	LocalPath string
	RemotePath string
}

func NewStorage(conf Config) storage.Storage {
	return &localStorage{
		Backend: &storage.Backend{
			LocalPath: conf.LocalPath,
			RemotePath: conf.RemotePath,
		},
	}
}
func (l localStorage) Copy(file string) error {
	if _, err := os.Stat(filepath.Join(l.LocalPath, file)); os.IsNotExist(err) {
		return err
	}
	err := copyFile(filepath.Join(l.LocalPath, file), filepath.Join(l.RemotePath, file))
	if err != nil {
		return err
	}
	return nil
}

func (l localStorage) CopyFrom(file string) error {
	if _, err := os.Stat(filepath.Join(l.RemotePath, file)); os.IsNotExist(err) {
		return err
	}
	err := copyFile(filepath.Join(l.RemotePath, file), filepath.Join(l.LocalPath, file))
	if err != nil {
		return err
	}
	return nil
}

// Prune deletes old backup created more than specified days
func (l localStorage) Prune(retentionDays int) error {
	currentTime := time.Now()
	// Delete file
	deleteFile := func(filePath string) error {
		err := os.Remove(filePath)
		if err != nil {
			utils.Fatal("Error:", err)
		} else {
			utils.Info("File %s deleted successfully", filePath)
		}
		return err
	}
	// Walk through the directory and delete files modified more than specified days ago
	err := filepath.Walk(l.RemotePath, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if it's a regular file and if it was modified more than specified days ago
		if fileInfo.Mode().IsRegular() {
			timeDiff := currentTime.Sub(fileInfo.ModTime())
			if timeDiff.Hours() > 24*float64(retentionDays) {
				err := deleteFile(filePath)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (l localStorage) Name() string {
	return "local"
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
