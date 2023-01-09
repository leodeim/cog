package files

import (
	"os"
)

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
