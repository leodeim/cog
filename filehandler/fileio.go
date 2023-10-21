package filehandler

import (
	"fmt"
	"path/filepath"
)

type FileType string

const (
	JSON    FileType = "json"
	YAML    FileType = "yaml"
	TOML    FileType = "toml"
	DYNAMIC FileType = "dynamic"
)

var availableImpl = []FileType{
	JSON,
	YAML,
	TOML,
}

type fileIO interface {
	write(data any, file string) error
	read(data any, file string) error
	extension() string
}

func buildFileIO(o *Optional) fileIO {
	switch resolveType(o) {
	case JSON:
		return &jsonFile{}
	case YAML:
		return &yamlFile{}
	case TOML:
		return &tomlFile{}
	default:
		return nil
	}
}

func resolveType(o *Optional) FileType {
	if o.Type != DYNAMIC {
		return o.Type
	}

	for _, t := range availableImpl {
		if Utils.fileExists(filepath.Join(o.Path, fmt.Sprintf(defaultConfig, o.Name, t))) {
			return t
		}
	}
	return JSON
}
