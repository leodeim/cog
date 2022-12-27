package goconfig

import (
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/leonidasdeim/goconfig/internal/defaults"
	"github.com/leonidasdeim/goconfig/internal/files"
)

type Config[T any] struct {
	mu          sync.Mutex
	data        T
	timestamp   string
	subscribers map[string](chan bool)
	file        string
}

const (
	defaultConfig = "%s.default.json"
	activeConfig  = "%s.json"
)

type Optional struct {
	Name string
	Path string
}

type Option func(f *Optional)

// Add custom filename. By default it is set to "app".
func WithName(name string) Option {
	return func(o *Optional) {
		o.Name = name
	}
}

// Add custom config file path. By default library searches work directory.
func WithPath(path string) Option {
	return func(o *Optional) {
		o.Path = path
	}
}

// Initialize goconfig library. Returns goconfig instance and error if something goes wrong.
// Receives optional parameters to set custom filename and path.
func Init[T any](opts ...Option) (*Config[T], error) {
	workDir := files.GetWorkDir()
	c := &Config[T]{}

	optional := &Optional{
		Name: "app",   // Default configuration name for application
		Path: workDir, // Default configuration path
	}

	for _, opt := range opts {
		opt(optional)
	}

	c.subscribers = make(map[string]chan bool)
	activeFile := filepath.Join(optional.Path, fmt.Sprintf(activeConfig, optional.Name))
	defaultFile := filepath.Join(optional.Path, fmt.Sprintf(defaultConfig, optional.Name))

	if files.Exists(activeFile) {
		c.file = activeFile
	} else if files.Exists(defaultFile) {
		c.file = defaultFile
	} else {
		return nil, fmt.Errorf("no configuration files found")
	}

	err := files.Load(&c.mu, &c.data, c.file)
	if err != nil {
		return nil, fmt.Errorf("failed at load from file: %v", err)
	}

	err = c.validate()
	if err != nil {
		return nil, fmt.Errorf("failed at validate config: %v", err)
	}

	err = defaults.Set(&c.data)
	if err != nil {
		return nil, fmt.Errorf("failed to set default values in config: %v", err)
	}

	c.updateTimestamp()

	if !files.Exists(activeFile) {
		c.file = activeFile
		err = files.Persist(&c.mu, c.data, c.file)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Config[T]) updateTimestamp() {
	c.timestamp = strconv.FormatInt(time.Now().Unix(), 10)
}

// Update configuration data. After update subscribers will be notified.
func (c *Config[T]) Update(newConfig T) error {
	c.data = newConfig

	err := files.Persist(&c.mu, c.data, c.file)
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

func (c *Config[T]) validate() error {
	validate := validator.New()
	return validate.Struct(c.data)
}

// Get subscriber read only channel by key.
func (c *Config[T]) GetSubscriber(key string) <-chan bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.subscribers[key]
}

// Register new subscriber.
func (c *Config[T]) AddSubscriber(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscribers[key] = make(chan bool, 1)
}

// Remove subscriber by key.
func (c *Config[T]) RemoveSubscriber(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscribers, key)
}

// Get timestamp of the configuration. It reflects when configuration has been updated or loaded last time.
func (c *Config[T]) GetTimestamp() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.timestamp
}

// Get configuration data.
func (c *Config[T]) GetCfg() T {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data
}
