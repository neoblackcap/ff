package utils

import (
	"fmt"
	"os"
)

func CreateFolder(prefix string) string {
	_, err := os.Stat(prefix)
	if os.IsNotExist(err) {
		_ = os.Mkdir(prefix, 0744)
		return prefix
	}
	return createFolderWithIndex(prefix, 1)
}

func createFolderWithIndex(prefix string, index int) string {
	folder := fmt.Sprintf("%s_%d", prefix, index)
	_, err := os.Stat(folder)
	if os.IsNotExist(err) {
		_ = os.Mkdir(folder, 0744)
		return folder
	}
	return createFolderWithIndex(prefix, index+1)

}
