package cog

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	fh "github.com/leodeim/cog/filehandler"
)

type Subscriber[T any] func(T) error
type Callback[T any] func(T)
type MaskFn[T any] func(*T)

type entry[T any] struct {
	id int
	fn T
}

type C[T any] struct {
	lock        sync.RWMutex
	config      T
	timestamp   time.Time
	handler     ConfigHandler
	nextID      int
	subscribers []entry[Subscriber[T]]
	callbacks   []entry[Callback[T]]
}

type ConfigHandler interface {
	Load(any) error
	Save(any) error
}

// Init initializes the library and returns a cog instance.
// An optional ConfigHandler can be provided; if omitted, a default
// built-in file handler is used (JSON with dynamic type detection).
func Init[T any](handler ...ConfigHandler) (*C[T], error) {
	c := C[T]{}

	if len(handler) > 0 {
		c.handler = handler[0]
	} else {
		h, err := fh.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create default file handler: %w", err)
		}
		c.handler = h
	}

	if err := c.load(); err != nil {
		// File not found is acceptable — start from zero value.
		// Any other parse/IO error should be surfaced.
		c.config = *new(T)
	}

	c.defaults()

	if err := validate(c.Config()); err != nil {
		return nil, err
	}

	if err := c.save(); err != nil {
		return nil, err
	}

	return &c, nil
}

// Update replaces the current configuration. The new config is validated,
// subscribers are notified (with rollback on error), and the result is
// persisted. Callbacks fire asynchronously after the state is committed.
func (c *C[T]) Update(new T) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if err := validate(new); err != nil {
		return err
	}

	if err := c.notifySubscribers(new); err != nil {
		return err
	}

	c.config = new

	if err := c.save(); err != nil {
		return err
	}

	c.fireCallbacks(new)

	return nil
}

// AddCallback registers a function that is called asynchronously (in a
// goroutine) after every successful config update. Returns an ID that
// can be used with RemoveCallback.
func (c *C[T]) AddCallback(f Callback[T]) int {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.nextID++
	c.callbacks = append(c.callbacks, entry[Callback[T]]{id: c.nextID, fn: f})
	return c.nextID
}

// RemoveCallback removes a previously registered callback by ID.
func (c *C[T]) RemoveCallback(id int) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	for i, e := range c.callbacks {
		if e.id == id {
			c.callbacks = append(c.callbacks[:i], c.callbacks[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("callback with id=%d not found", id)
}

// AddSubscriber registers a function that is called synchronously on
// config update. If any subscriber returns an error, the update is
// rolled back. Returns an ID that can be used with RemoveSubscriber.
func (c *C[T]) AddSubscriber(f Subscriber[T]) int {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.nextID++
	c.subscribers = append(c.subscribers, entry[Subscriber[T]]{id: c.nextID, fn: f})
	return c.nextID
}

// RemoveSubscriber removes a previously registered subscriber by ID.
func (c *C[T]) RemoveSubscriber(id int) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	for i, e := range c.subscribers {
		if e.id == id {
			c.subscribers = append(c.subscribers[:i], c.subscribers[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("subscriber with id=%d not found", id)
}

// GetTimestamp returns the time of the last configuration load or update.
func (c *C[T]) GetTimestamp() time.Time {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.timestamp
}

// Config returns a copy of the current configuration.
func (c *C[T]) Config() T {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.config
}

// String returns a JSON representation of the current configuration.
// Optional MaskFn functions can be provided to redact sensitive fields
// before serialization.
func (c *C[T]) String(masks ...MaskFn[T]) (string, error) {
	data := c.Config()

	for _, mask := range masks {
		mask(&data)
	}

	b, err := json.MarshalIndent(data, "", "  ")
	return string(b), err
}

func (c *C[T]) load() error {
	return c.handler.Load(&c.config)
}

func (c *C[T]) save() error {
	c.timestamp = time.Now()
	return c.handler.Save(c.config)
}

func (c *C[T]) notifySubscribers(config T) error {
	var updated []Subscriber[T]

	for _, e := range c.subscribers {
		if e.fn == nil {
			continue
		}
		if err := e.fn(config); err != nil {
			c.rollback(updated)
			return fmt.Errorf("subscriber returned an error on update: %w", err)
		}
		updated = append(updated, e.fn)
	}

	return nil
}

func (c *C[T]) fireCallbacks(config T) {
	for _, e := range c.callbacks {
		if e.fn == nil {
			continue
		}
		go e.fn(config)
	}
}

func (c *C[T]) rollback(subscribers []Subscriber[T]) {
	for _, f := range subscribers {
		// Best-effort rollback: if a subscriber fails during rollback
		// there is nothing meaningful we can do, so errors are ignored.
		_ = f(c.config)
	}
}

func (c *C[T]) defaults() {
	SetDefaults(&c.config)
}

func validate[T any](data T) error {
	if err := validator.New().Struct(data); err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}
	return nil
}
