package filehandler

import (
	"fmt"
	"os"
	"sync"

	"github.com/pelletier/go-toml/v2"
)

type tomlFile struct {
	m sync.Mutex
}

func (t *tomlFile) write(data any, file string) error {
	t.m.Lock()
	defer t.m.Unlock()

	toml, err := toml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed at marshal toml: %v", err)
	}

	err = Utils.writeFile(file, toml)
	if err != nil {
		return fmt.Errorf("failed at write to toml file: %v", err)
	}

	return nil
}

func (t *tomlFile) read(data any, file string) error {
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

func (t *tomlFile) extension() string {
	return "toml"
}
