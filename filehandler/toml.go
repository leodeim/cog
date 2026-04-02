package filehandler

import (
	"fmt"
	"os"
	"sync"

	"github.com/pelletier/go-toml/v2"
)

type Toml struct {
	m sync.Mutex
}

func (t *Toml) Write(data any, file string) error {
	t.m.Lock()
	defer t.m.Unlock()

	b, err := toml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal toml: %w", err)
	}

	if err := writeFile(file, b); err != nil {
		return fmt.Errorf("failed to write toml file: %w", err)
	}

	return nil
}

func (t *Toml) Read(data any, file string) error {
	t.m.Lock()
	defer t.m.Unlock()

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open toml file: %w", err)
	}
	defer f.Close()

	if err := toml.NewDecoder(f).Decode(data); err != nil {
		return fmt.Errorf("failed to read toml file: %w", err)
	}

	return nil
}

func (t *Toml) GetExtension() string {
	return "toml"
}
