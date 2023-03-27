package filehandler

import (
	"fmt"
	"path/filepath"

	"github.com/leonidasdeim/goconfig/internal/files"
	"github.com/leonidasdeim/goconfig/pkg/filehandler/impl"
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

type FileIO interface {
	Write(data any, file string) error
	Read(data any, file string) error
	GetExtension() string
}

var _ FileIO = (*impl.Json)(nil)
var _ FileIO = (*impl.Yaml)(nil)
var _ FileIO = (*impl.Toml)(nil)

func BuildFileIO(o *Optional) FileIO {
	switch resolveType(o) {
	case JSON:
		return &impl.Json{}
	case YAML:
		return &impl.Yaml{}
	case TOML:
		return &impl.Toml{}
	default:
		return nil
	}
}

func resolveType(o *Optional) FileType {
	if o.Type != DYNAMIC {
		return o.Type
	}

	for _, t := range availableImpl {
		if files.Exists(filepath.Join(o.Path, fmt.Sprintf(defaultConfig, o.Name, t))) {
			return t
		}
	}
	return JSON
}
