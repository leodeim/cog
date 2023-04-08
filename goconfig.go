package goconfig

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/leonidasdeim/goconfig/internal/defaults"
	fh "github.com/leonidasdeim/goconfig/pkg/filehandler"
)

type UpdateCallback[T any] func(T) error

type Config[T any] struct {
	mu        sync.Mutex
	data      T
	time      string
	subs      map[string](chan bool)
	callbacks []UpdateCallback[T]
	handler   ConfigHandler
}

type ConfigHandler interface {
	Load(data any) error
	Save(data any) error
}

// Initialize library. Returns goconfig instance.
// Receives config handler.
// To use default builtin JSON file handler:
// c, err := goconfig.Init[ConfigStruct](handler.New())
func Init[T any](handler ...ConfigHandler) (*Config[T], error) {
	c := Config[T]{}

	if len(handler) > 0 {
		c.handler = handler[0]
	} else {
		c.handler, _ = fh.New() // default DYNAMIC file handler
	}
	c.subs = make(map[string]chan bool)

	c.load()

	err := c.defaults()
	if err != nil {
		return nil, err
	}

	err = c.validate()
	if err != nil {
		return nil, err
	}

	err = c.save()
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// Update configuration data. After update subscribers will be notified.
func (c *Config[T]) Update(new T) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	old := c.data
	c.data = new

	err := c.validate()
	if err != nil {
		c.data = old
		return err
	}

	err = c.save()
	if err != nil {
		return err
	}

	for _, channel := range c.subs {
		// Do not notify subscriber through channel if it was already notified
		if len(channel) != 0 {
			continue
		}
		channel <- true
	}

	for _, cb := range c.callbacks {
		go cb(c.data)
	}

	return nil
}

// Get subscriber read only channel by key.
// Returns an error if subscriber key does not exist.
func (c *Config[T]) GetSubscriber(key string) (<-chan bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ch, ok := c.subs[key]; ok {
		return ch, nil
	}
	return nil, fmt.Errorf("subscriber is not registered: %s", key)
}

// Register new subscriber.
func (c *Config[T]) AddSubscriber(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subs[key] = make(chan bool, 1)
}

// Register new callback function. It will be called after config update.
func (c *Config[T]) AddCallback(f UpdateCallback[T]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.callbacks = append(c.callbacks, f)
}

// Remove subscriber by key.
func (c *Config[T]) RemoveSubscriber(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subs, key)
}

// Get timestamp of the configuration. It reflects when configuration has been updated or loaded last time.
func (c *Config[T]) GetTimestamp() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.time
}

// Get configuration data.
func (c *Config[T]) GetCfg() T {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data
}

func (c *Config[T]) load() {
	err := c.handler.Load(&c.data)
	if err != nil {
		c.data = *new(T)
	}
}

func (c *Config[T]) save() error {
	c.updateTimestamp()
	err := c.handler.Save(c.data)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config[T]) validate() error {
	validate := validator.New()
	err := validate.Struct(c.data)
	if err != nil {
		return fmt.Errorf("failed at validate config: %v", err)
	}
	return nil
}

func (c *Config[T]) defaults() error {
	err := defaults.Set(&c.data)
	if err != nil {
		return fmt.Errorf("failed to set env/default values: %v", err)
	}
	return nil
}

func (c *Config[T]) updateTimestamp() {
	c.time = strconv.FormatInt(time.Now().Unix(), 10)
}
