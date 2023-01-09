package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const JSON FileType = "json"

type Json struct {
	m sync.Mutex
}

func (j *Json) Write(data any, file string) error {
	j.m.Lock()
	defer j.m.Unlock()

	json, err := json.MarshalIndent(data, emptySpace, marshalIndent)
	if err != nil {
		return fmt.Errorf("failed at marshal json: %v", err)
	}

	err = os.WriteFile(file, json, filePermissions)
	if err != nil {
		return fmt.Errorf("failed at write to json file: %v", err)
	}

	return nil
}

func (j *Json) Read(data any, file string) error {
	j.m.Lock()
	defer j.m.Unlock()

	configFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed at open json file: %v", err)
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(data); err != nil {
		return fmt.Errorf("failed at reading from json file: %v", err)
	}

	return nil
}

func (j *Json) GetExtension() string {
	return string(JSON)
}
