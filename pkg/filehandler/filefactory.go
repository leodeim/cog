package filehandler

type FileType string

type fileIO interface {
	Write(data any, file string) error
	Read(data any, file string) error
	GetExtension() string
}

func fileIOFactory(t FileType) fileIO {
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
