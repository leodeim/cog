package filehandler

import (
	"fmt"
	"path/filepath"
)

const (
	defaultConfig = "%s.default.%s"
	activeConfig  = "%s.%s"
)

type FileHandler struct {
	file   string
	fileIO FileIO
}

type Optional struct {
	Name string
	Path string
	Type FileType
}

type Option func(f *Optional)

// Add custom filename. By default it is set to "app".
func WithName(n string) Option {
	return func(o *Optional) {
		o.Name = n
	}
}

// Add custom config file path. By default library uses work directory.
func WithPath(p string) Option {
	return func(o *Optional) {
		o.Path = p
	}
}

// Specify handler type.
// - filehandler.DYNAMIC (default)
// - filehandler.JSON
// - filehandler.YAML
// - filehandler.TOML
func WithType(t FileType) Option {
	return func(o *Optional) {
		o.Type = t
	}
}

func New(opts ...Option) (*FileHandler, error) {

	// Set defaults
	o := &Optional{
		Name: "app",
		Path: Utils.GetWorkDir(),
		Type: DYNAMIC,
	}

	for _, opt := range opts {
		opt(o)
	}

	h := FileHandler{}
	h.fileIO = BuildFileIO(o)
	if h.fileIO == nil {
		return nil, fmt.Errorf("bad file type, or dynamic type has not been resolved: %s", string(o.Type))
	}

	e := h.fileIO.GetExtension()
	h.file = filepath.Join(o.Path, fmt.Sprintf(activeConfig, o.Name, e))
	defaultFile := filepath.Join(o.Path, fmt.Sprintf(defaultConfig, o.Name, e))

	if err := h.initActiveFile(defaultFile, h.file); err != nil {
		return nil, err
	}

	return &h, nil
}

func (h *FileHandler) Load(data any) error {
	return h.fileIO.Read(data, h.file)
}

func (h *FileHandler) Save(data any) error {
	return h.fileIO.Write(data, h.file)
}

func (h *FileHandler) initActiveFile(defaultFile string, activeFile string) error {
	if Utils.FileExists(activeFile) {
		return nil
	}

	if !Utils.FileExists(defaultFile) {
		return nil
	}

	var t interface{}

	if err := h.fileIO.Read(&t, defaultFile); err != nil {
		return err
	}

	if err := h.fileIO.Write(t, activeFile); err != nil {
		return err
	}

	return nil
}
