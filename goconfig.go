package goconfig

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/leonidasdeim/goconfig/internal/defaults"
)

type Config[T any] struct {
	mu      sync.Mutex
	data    T
	time    string
	subs    map[string](chan bool)
	handler ConfigHandler
}

type ConfigHandler interface {
	Load(data any) error
	Save(data any) error
}

// Initialize library. Returns goconfig instance.
// Receives config handler.
// To use default builtin JSON file handler:
// c, err := goconfig.Init[ConfigStruct](handler.New())
func Init[T any](handler ConfigHandler) (*Config[T], error) {
	c := &Config[T]{
		handler: handler,
	}

	c.subs = make(map[string]chan bool)

	err := c.handler.Load(&c.data)
	if err != nil {
		return nil, fmt.Errorf("failed at load from config: %v", err)
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

	err = c.handler.Save(c.data)
	if err != nil {
		return nil, fmt.Errorf("failed at write to config: %v", err)
	}

	return c, nil
}

// Update configuration data. After update subscribers will be notified.
func (c *Config[T]) Update(newConfig T) error {
	c.data = newConfig

	err := c.handler.Save(c.data)
	if err != nil {
		return err
	}

	c.updateTimestamp()

	for _, channel := range c.subs {
		// Do not notify subscriber through channel if it was already notified
		if len(channel) != 0 {
			continue
		}
		channel <- true
	}

	return nil
}

// Get subscriber read only channel by key.
func (c *Config[T]) GetSubscriber(key string) <-chan bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.subs[key]
}

// Register new subscriber.
func (c *Config[T]) AddSubscriber(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subs[key] = make(chan bool, 1)
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

func (c *Config[T]) validate() error {
	validate := validator.New()
	return validate.Struct(c.data)
}

func (c *Config[T]) updateTimestamp() {
	c.time = strconv.FormatInt(time.Now().Unix(), 10)
}
