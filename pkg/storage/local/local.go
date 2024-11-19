package local

import (
	pkg "github.com/jkaninda/pg-bkup/pkg/storage"
	"io"
	"os"
	"path/filepath"
	"time"
)

type localStorage struct {
	*pkg.Backend
}
type Config struct {
	LocalPath  string
	RemotePath string
}

// NewStorage creates new Storage
func NewStorage(conf Config) pkg.Storage {
	return &localStorage{
		Backend: &pkg.Backend{
			LocalPath:  conf.LocalPath,
			RemotePath: conf.RemotePath,
		},
	}
}

// Copy copies file to the local destination path
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

// CopyFrom copies file from a Path to local path
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

// Name returns the storage name
func (l localStorage) Name() string {
	return "local"
}

// copyFile copies file
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(in *os.File) {
		err := in.Close()
		if err != nil {
			return
		}
	}(in)

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		err := out.Close()
		if err != nil {
			return err
		}
		return err
	}
	return out.Close()
}
