package file

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func FindDockerFile(dfPath string) (string, error) {
	dfPath, err := GetFullPath(dfPath)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(dfPath)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return "", errors.New("not a file")
	}

	return dfPath, nil
}

func GetFullPath(p string) (string, error) {
	info, err := os.Stat(filepath.Clean(p))
	if err != nil {
		return "", err
	}

	path, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(path, info.Name()), nil
}

func ArePathsEqual(p1, p2 string) bool {
	return strings.EqualFold(filepath.Clean(p1), filepath.Clean(p2))
}
