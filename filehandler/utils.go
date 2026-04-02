package filehandler

import "os"

const filePermissions = 0664

func fileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func getWorkDir() (string, error) {
	return os.Getwd()
}

func writeFile(name string, data []byte) error {
	return os.WriteFile(name, data, filePermissions)
}
