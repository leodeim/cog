package cog

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/leonidasdeim/cog/pkg/defaults"
	fh "github.com/leonidasdeim/cog/pkg/filehandler"
)

type Callback[T any] func(T)
type Bound[T any] func(T) error

type Config[T any] struct {
	mutex       sync.Mutex
	data        T
	timestamp   string
	subscribers map[string](chan bool)
	callbacks   []Callback[T]
	bounds      []Bound[T]
	handler     ConfigHandler
}

type ConfigHandler interface {
	Load(data any) error
	Save(data any) error
}

// Initialize library. Returns cog instance.
// Receives config handler.
// To use default builtin JSON file handler:
// c, err := cog.Init[ConfigStruct](handler.New())
func Init[T any](handler ...ConfigHandler) (*Config[T], error) {
	c := Config[T]{
		subscribers: make(map[string]chan bool),
	}

	if len(handler) > 0 {
		c.handler = handler[0]
	} else {
		c.handler, _ = fh.New() // default DYNAMIC file handler
	}

	c.load()

	if err := c.defaults(); err != nil {
		return nil, err
	}

	if err := validate(c.GetCfg()); err != nil {
		return nil, err
	}

	if err := c.save(); err != nil {
		return nil, err
	}

	return &c, nil
}

// Update configuration data. After update subscribers will be notified.
func (c *Config[T]) Update(new T) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := validate(new); err != nil {
		return err
	}

	if err := c.bound(new); err != nil {
		return err
	}

	c.data = new

	if err := c.save(); err != nil {
		return err
	}

	c.notify()

	return nil
}

// Get subscriber read only channel by key.
// Returns an error if subscriber key does not exist.
func (c *Config[T]) GetSubscriber(key string) (<-chan bool, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if ch, ok := c.subscribers[key]; ok {
		return ch, nil
	}
	return nil, fmt.Errorf("subscriber is not registered: %s", key)
}

// Register new subscriber.
func (c *Config[T]) AddSubscriber(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.subscribers[key] = make(chan bool, 1)
}

// Register new callback function. It will be called after config update.
func (c *Config[T]) AddCallback(f Callback[T]) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.callbacks = append(c.callbacks, f)
}

// Register new bound callback function. It will be called after config update.
// If bound callback returns error, config update is aborted and old config is restored.
// Bound callbacks will be blocking functions.
func (c *Config[T]) AddBound(f Bound[T]) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.bounds = append(c.bounds, f)
}

// Remove subscriber by key.
func (c *Config[T]) RemoveSubscriber(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.subscribers, key)
}

// Get timestamp of the configuration. It reflects when configuration has been updated or loaded last time.
func (c *Config[T]) GetTimestamp() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.timestamp
}

// Get configuration data.
func (c *Config[T]) GetCfg() T {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.data
}

func (c *Config[T]) load() {
	if err := c.handler.Load(&c.data); err != nil {
		c.data = *new(T)
	}
}

func (c *Config[T]) save() error {
	c.updateTimestamp()

	if err := c.handler.Save(c.data); err != nil {
		return err
	}
	return nil
}

func (c *Config[T]) notify() {
	for _, ch := range c.subscribers {
		select {
		case ch <- true:
		default:
		}
	}

	for _, cb := range c.callbacks {
		go cb(c.data)
	}
}

func (c *Config[T]) bound(new T) error {
	used := []Bound[T]{}

	for _, f := range c.bounds {
		if err := f(new); err != nil {
			c.rollback(used)
			return fmt.Errorf("bound callback returned an error: %v", err)
		}
		used = append(used, f)
	}
	return nil
}

func (c *Config[T]) rollback(bounds []Bound[T]) {
	for _, f := range bounds {
		f(c.data)
	}
}

func (c *Config[T]) defaults() error {
	if err := defaults.Set(&c.data); err != nil {
		return fmt.Errorf("failed to set env/default values: %v", err)
	}
	return nil
}

func (c *Config[T]) updateTimestamp() {
	c.timestamp = strconv.FormatInt(time.Now().Unix(), 10)
}

func validate[T any](data T) error {
	if err := validator.New().Struct(data); err != nil {
		return fmt.Errorf("failed at validate config: %v", err)
	}
	return nil
}
