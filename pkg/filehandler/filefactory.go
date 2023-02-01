package filehandler

import (
	"fmt"
	"path/filepath"

	"github.com/leonidasdeim/goconfig/internal/files"
)

type FileType string

type FileIO interface {
	Write(data any, file string) error
	Read(data any, file string) error
	GetExtension() string
}

func FileIOFactory(o *Optional) FileIO {
	t := o.Type

	if t == DYNAMIC {
		t = resolveDynamic(o)
	}

	switch t {
	case JSON:
		return &Json{}
	case YAML:
		return &Yaml{}
	case TOML:
		return &Toml{}
	default:
		return nil
	}
}

var types = []FileType{
	JSON,
	YAML,
	TOML,
}

const DYNAMIC FileType = "dynamic"

func resolveDynamic(o *Optional) FileType {
	for _, t := range types {
		if files.Exists(filepath.Join(o.Path, fmt.Sprintf(defaultConfig, o.Name, t))) {
			return t
		}
	}
	return JSON
}
