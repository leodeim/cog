package goconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
)

type Config[T any] struct {
	mu          sync.Mutex
	data        T
	timestamp   string
	subscribers map[string](chan bool)
	activeFile  string
}

const (
	defaultConfig   = "%s.default.json"
	activeConfig    = "%s.json"
	marshalIndent   = "	"
	emptySpace      = ""
	permissionRwRwR = 0664
	defaultTag      = "default"
)

type Optional struct {
	Name string
	Path string
}

type Option func(f *Optional)

func WithName(name string) Option {
	return func(o *Optional) {
		o.Name = name
	}
}

func WithPath(path string) Option {
	return func(o *Optional) {
		o.Path = path
	}
}

func Init[T any](opts ...Option) (*Config[T], error) {
	workDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	c := &Config[T]{}

	optional := &Optional{
		Name: "app",   // Default configuration name for application
		Path: workDir, // Default configuration path
	}

	for _, opt := range opts {
		opt(optional)
	}

	c.subscribers = make(map[string]chan bool)
	activeConfigFilename := filepath.Join(optional.Path, fmt.Sprintf(activeConfig, optional.Name))
	defaultConfigFilename := filepath.Join(optional.Path, fmt.Sprintf(defaultConfig, optional.Name))
	activeFileExists := fileExists(activeConfigFilename)
	defaultFileExists := fileExists(defaultConfigFilename)

	if activeFileExists {
		c.activeFile = activeConfigFilename
	} else if defaultFileExists {
		c.activeFile = defaultConfigFilename
	} else {
		return nil, fmt.Errorf("no configuration files found")
	}

	err = c.load()
	if err != nil {
		return nil, fmt.Errorf("failed at load from file: %v", err)
	}

	err = c.validate()
	if err != nil {
		return nil, fmt.Errorf("failed at validate config: %v", err)
	}

	err = c.setDefault()
	if err != nil {
		return nil, fmt.Errorf("failed to set default values in config: %v", err)
	}

	c.updateTimestamp()

	if !activeFileExists {
		c.activeFile = activeConfigFilename
		err = c.persist()
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Config[T]) setDefault() error {
	v := reflect.ValueOf(&c.data).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		if defaultVal := t.Field(i).Tag.Get(defaultTag); defaultVal != "" {
			if err := c.setField(v.Field(i), defaultVal); err != nil {
				return err
			}

		}
	}
	return nil
}

func (c *Config[T]) setField(field reflect.Value, defaultVal string) error {

	if !field.CanSet() {
		return fmt.Errorf("can't set value")
	}

	if !IsEmpty(field) {
		// field already set.
		return nil
	}

	switch field.Kind() {

	case reflect.Int:
		if val, err := strconv.Atoi(defaultVal); err == nil {
			field.Set(reflect.ValueOf(int(val)).Convert(field.Type()))
		}
	case reflect.String:
		field.Set(reflect.ValueOf(defaultVal).Convert(field.Type()))
	case reflect.Bool:
		if val, err := strconv.ParseBool(defaultVal); err == nil {
			field.Set(reflect.ValueOf(bool(val)).Convert(field.Type()))
		}
	}

	return nil
}

func (c *Config[T]) updateTimestamp() {
	c.timestamp = strconv.FormatInt(time.Now().Unix(), 10)
}

func (c *Config[T]) Update(newConfig T) error {
	c.data = newConfig

	err := c.persist()

	if err != nil {
		return err
	}

	c.updateTimestamp()

	for _, channel := range c.subscribers {
		// Do not notify subscriber through channel if it was already notified
		if len(channel) != 0 {
			continue
		}

		channel <- true
	}

	return nil
}

func (c *Config[T]) persist() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := json.MarshalIndent(c.data, emptySpace, marshalIndent)

	if err != nil {
		return fmt.Errorf("failed at marshal json: %v", err)
	}

	err = os.WriteFile(c.activeFile, file, permissionRwRwR)

	if err != nil {
		return fmt.Errorf("failed at write to file: %v", err)
	}

	return nil
}

func (c *Config[T]) load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	configFile, err := os.Open(c.activeFile)

	if err != nil {
		return err
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&c.data); err != nil {
		return err
	}

	return nil
}

func (c *Config[T]) validate() error {
	validate := validator.New()
	return validate.Struct(c.data)
}

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}

	return false
}

func (c *Config[T]) GetSubscriber(key string) <-chan bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.subscribers[key]
}

func (c *Config[T]) AddSubscriber(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscribers[key] = make(chan bool, 1)
}

func (c *Config[T]) RemoveSubscriber(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscribers, key)
}

func (c *Config[T]) GetTimestamp() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.timestamp
}

func (c *Config[T]) GetCfg() *T {
	c.mu.Lock()
	defer c.mu.Unlock()
	return &c.data
}

// Helpers
func IsEmpty(v reflect.Value) bool {
	return !v.IsValid() || reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
