package files

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const (
	marshalIndent   = "	"
	emptySpace      = ""
	permissionRwRwR = 0664
)

func Persist[T any](m *sync.Mutex, data T, file string) error {
	m.Lock()
	defer m.Unlock()

	json, err := json.MarshalIndent(data, emptySpace, marshalIndent)
	if err != nil {
		return fmt.Errorf("failed at marshal json: %v", err)
	}

	err = os.WriteFile(file, json, permissionRwRwR)
	if err != nil {
		return fmt.Errorf("failed at write to file: %v", err)
	}

	return nil
}

func Load[T any](m *sync.Mutex, data *T, file string) error {
	m.Lock()
	defer m.Unlock()

	configFile, err := os.Open(file)
	if err != nil {
		return err
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&data); err != nil {
		return err
	}

	return nil
}

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
