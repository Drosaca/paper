package stat

import (
	"os"
)

type Path struct {
	Exists bool
	IsDir bool
}

func IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && !fileInfo.IsDir()
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && fileInfo.IsDir()
}
