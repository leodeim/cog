package filehandler

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type Yaml struct {
	m sync.Mutex
}

func (y *Yaml) Write(data any, file string) error {
	y.m.Lock()
	defer y.m.Unlock()

	b, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}

	if err := writeFile(file, b); err != nil {
		return fmt.Errorf("failed to write yaml file: %w", err)
	}

	return nil
}

func (y *Yaml) Read(data any, file string) error {
	y.m.Lock()
	defer y.m.Unlock()

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open yaml file: %w", err)
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(data); err != nil {
		return fmt.Errorf("failed to read yaml file: %w", err)
	}

	return nil
}

func (y *Yaml) GetExtension() string {
	return "yaml"
}
