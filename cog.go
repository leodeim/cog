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

type Subscriber[T any] func(T) error
type Callback[T any] func(T)

type C[T any] struct {
	sync.Mutex
	config      T
	timestamp   string
	handler     ConfigHandler
	subscribers map[int](Subscriber[T])
	callbacks   map[int](Callback[T])
}

type ConfigHandler interface {
	Load(any) error
	Save(any) error
}

// Initialize library. Returns cog instance.
// Receives config handler.
// To use default builtin JSON file handler:
// c, err := cog.Init[ConfigStruct](handler.New())
func Init[T any](handler ...ConfigHandler) (*C[T], error) {
	cog := C[T]{
		callbacks:   make(map[int]Callback[T]),
		subscribers: make(map[int]Subscriber[T]),
	}

	if len(handler) > 0 {
		cog.handler = handler[0]
	} else {
		cog.handler, _ = fh.New() // default DYNAMIC file handler
	}

	cog.load()

	if err := cog.defaults(); err != nil {
		return nil, err
	}

	if err := validate(cog.Config()); err != nil {
		return nil, err
	}

	if err := cog.save(); err != nil {
		return nil, err
	}

	return &cog, nil
}

// Update configuration data. After update subscribers will be notified.
func (cog *C[T]) Update(new T) error {
	cog.Lock()
	defer cog.Unlock()

	if err := validate(new); err != nil {
		return err
	}

	if err := cog.notify(new); err != nil {
		return err
	}

	cog.config = new

	if err := cog.save(); err != nil {
		return err
	}

	return nil
}

// Register new callback function. It will be called after config update in non blocking goroutine.
// This method returns callback id (int). It can be used to remove callback by calling cog.RemoveCallback(id).
func (cog *C[T]) AddCallback(f Callback[T]) int {
	cog.Lock()
	defer cog.Unlock()

	l := len(cog.callbacks) + 1
	cog.callbacks[l] = f

	return l
}

// Remove callback by id.
func (cog *C[T]) RemoveCallback(id int) error {
	cog.Lock()
	defer cog.Unlock()

	if _, ok := cog.callbacks[id]; ok {
		delete(cog.callbacks, id)
		return nil
	}

	return fmt.Errorf("callback with id=%d not found", id)
}

// Register new subscriber function. It will be called after config update and wait for every subscriber to be updated.
// If at least one subscriber returns an error, update stops and rollback is initiated for all updated subscribers.
// This method returns subscriber id (int). It can be used to remove subscriber by calling cog.RemoveSubscriber(id).
func (cog *C[T]) AddSubscriber(f Subscriber[T]) int {
	cog.Lock()
	defer cog.Unlock()

	l := len(cog.subscribers) + 1
	cog.subscribers[l] = f

	return l
}

// Remove subscriber by id.
func (cog *C[T]) RemoveSubscriber(id int) error {
	cog.Lock()
	defer cog.Unlock()

	if _, ok := cog.subscribers[id]; ok {
		delete(cog.subscribers, id)
		return nil
	}

	return fmt.Errorf("subscriber with id=%d not found", id)
}

// Get timestamp of the configuration. It reflects when configuration has been updated or loaded last time.
func (cog *C[T]) GetTimestamp() string {
	cog.Lock()
	defer cog.Unlock()

	return cog.timestamp
}

// Get configuration data.
func (cog *C[T]) Config() T {
	cog.Lock()
	defer cog.Unlock()

	return cog.config
}

func (cog *C[T]) load() {
	if err := cog.handler.Load(&cog.config); err != nil {
		cog.config = *new(T)
	}
}

func (cog *C[T]) save() error {
	cog.updateTimestamp()

	if err := cog.handler.Save(cog.config); err != nil {
		return err
	}
	return nil
}

func (cog *C[T]) notify(config T) error {
	updated := []Subscriber[T]{}

	for _, f := range cog.subscribers {
		if f == nil {
			continue
		}
		if err := f(config); err != nil {
			cog.rollback(updated)
			return fmt.Errorf("subscriber returned an error on update: %v", err)
		}
		updated = append(updated, f)
	}

	for _, f := range cog.callbacks {
		if f == nil {
			continue
		}
		go f(config)
	}

	return nil
}

func (cog *C[T]) rollback(subscribers []Subscriber[T]) {
	for _, f := range subscribers {
		if f == nil {
			continue
		}
		f(cog.config)
	}
}

func (cog *C[T]) defaults() error {
	if err := defaults.Set(&cog.config); err != nil {
		return fmt.Errorf("failed to set env/default values: %v", err)
	}
	return nil
}

func (cog *C[T]) updateTimestamp() {
	cog.timestamp = strconv.FormatInt(time.Now().Unix(), 10)
}

func validate[T any](data T) error {
	if err := validator.New().Struct(data); err != nil {
		return fmt.Errorf("failed at validate config: %v", err)
	}
	return nil
}
