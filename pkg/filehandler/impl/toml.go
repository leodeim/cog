package impl

import (
	"fmt"
	"os"
	"sync"

	"github.com/leonidasdeim/goconfig/internal/files"
	"github.com/pelletier/go-toml/v2"
)

type Toml struct {
	m sync.Mutex
}

func (t *Toml) Write(data any, file string) error {
	t.m.Lock()
	defer t.m.Unlock()

	toml, err := toml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed at marshal toml: %v", err)
	}

	err = files.Write(file, toml)
	if err != nil {
		return fmt.Errorf("failed at write to toml file: %v", err)
	}

	return nil
}

func (t *Toml) Read(data any, file string) error {
	t.m.Lock()
	defer t.m.Unlock()

	configFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed at open toml file: %v", err)
	}

	tomlParser := toml.NewDecoder(configFile)
	if err = tomlParser.Decode(data); err != nil {
		return fmt.Errorf("failed at reading from toml file: %v", err)
	}

	return nil
}

func (t *Toml) GetExtension() string {
	return "toml"
}
