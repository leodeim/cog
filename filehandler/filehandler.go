package filehandler

import (
	"fmt"
	"path/filepath"
)

const (
	defaultConfig = "%s.default.%s"
	activeConfig  = "%s.%s"
)

// FileHandler implements ConfigHandler using local files.
// It supports JSON, YAML, and TOML formats.
type FileHandler struct {
	file   string
	fileIO FileIO
}

// Optional holds configuration for creating a FileHandler.
type Optional struct {
	Name string
	Path string
	Type FileType
}

// Option is a functional option for configuring a FileHandler.
type Option func(f *Optional)

// WithName sets a custom config file name. Default is "app".
func WithName(n string) Option {
	return func(o *Optional) {
		o.Name = n
	}
}

// WithPath sets a custom config file directory. Default is the working directory.
func WithPath(p string) Option {
	return func(o *Optional) {
		o.Path = p
	}
}

// WithType sets the config file format.
// Supported: filehandler.DYNAMIC (default), filehandler.JSON, filehandler.YAML, filehandler.TOML.
func WithType(t FileType) Option {
	return func(o *Optional) {
		o.Type = t
	}
}

// New creates a new FileHandler with the given options.
func New(opts ...Option) (*FileHandler, error) {
	wd, err := getWorkDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	o := &Optional{
		Name: "app",
		Path: wd,
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

// Load reads the config file into data.
func (h *FileHandler) Load(data any) error {
	return h.fileIO.Read(data, h.file)
}

// Save writes data to the config file.
func (h *FileHandler) Save(data any) error {
	return h.fileIO.Write(data, h.file)
}

func (h *FileHandler) initActiveFile(defaultFile string, activeFile string) error {
	if fileExists(activeFile) {
		return nil
	}

	if !fileExists(defaultFile) {
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
