package filehandler

import (
	"os"
)

const filePermissions = 0664

type _utils struct{}

var Utils = _utils{}

func (_utils) fileExists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}

	return false
}

func (_utils) fetWorkDir() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return wd
}

func (_utils) writeFile(name string, data []byte) error {
	return os.WriteFile(name, data, filePermissions)
}
