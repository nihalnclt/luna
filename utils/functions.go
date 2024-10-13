package utils

import (
	"os"
	"path/filepath"
	"regexp"
)

func IsURLFriendly(s string) bool {
	// Use a regular expression to allow only URL-friendly characters
	regexp := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	return regexp.MatchString(s)
}

func GetBaseFolderName() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Base(cwd), nil
}
