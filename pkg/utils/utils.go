package utils

import (
	"os"
)

const filePermissions = 0664

func Exists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}

	return false
}

func GetWorkDir() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return wd
}

func Write(name string, data []byte) error {
	return os.WriteFile(name, data, filePermissions)
}
