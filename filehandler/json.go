package filehandler

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const (
	marshalIndent = "\t"
)

type Json struct {
	m sync.Mutex
}

func (j *Json) Write(data any, file string) error {
	j.m.Lock()
	defer j.m.Unlock()

	b, err := json.MarshalIndent(data, "", marshalIndent)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	if err := writeFile(file, b); err != nil {
		return fmt.Errorf("failed to write json file: %w", err)
	}

	return nil
}

func (j *Json) Read(data any, file string) error {
	j.m.Lock()
	defer j.m.Unlock()

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open json file: %w", err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(data); err != nil {
		return fmt.Errorf("failed to read json file: %w", err)
	}

	return nil
}

func (j *Json) GetExtension() string {
	return "json"
}
