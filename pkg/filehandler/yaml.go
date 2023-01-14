package filehandler

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

const YAML FileType = "yaml"

type Yaml struct {
	m sync.Mutex
}

func (j *Yaml) Write(data any, file string) error {
	j.m.Lock()
	defer j.m.Unlock()

	yaml, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed at marshal yaml: %v", err)
	}

	err = os.WriteFile(file, yaml, filePermissions)
	if err != nil {
		return fmt.Errorf("failed at write to yaml file: %v", err)
	}

	return nil
}

func (j *Yaml) Read(data any, file string) error {
	j.m.Lock()
	defer j.m.Unlock()

	configFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed at open yaml file: %v", err)
	}

	yamlParser := yaml.NewDecoder(configFile)
	if err = yamlParser.Decode(data); err != nil {
		return fmt.Errorf("failed at reading from yaml file: %v", err)
	}

	return nil
}

func (j *Yaml) GetExtension() string {
	return string(YAML)
}
