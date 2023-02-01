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

var implementations = []FileType{
	JSON,
	YAML,
	TOML,
}

type FileIO interface {
	Write(data any, file string) error
	Read(data any, file string) error
	GetExtension() string
}

func BuildFileIO(o *Optional) FileIO {
	t := o.Type

	if t == DYNAMIC {
		t = resolveDynamic(o)
	}

	switch t {
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

func resolveDynamic(o *Optional) FileType {
	for _, t := range implementations {
		if files.Exists(filepath.Join(o.Path, fmt.Sprintf(defaultConfig, o.Name, t))) {
			return t
		}
	}
	return JSON
}
