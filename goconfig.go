package goconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

var (
	defaultConfig = "app_config.default.json"
	activeConfig  = "app_config.active.json"
)

type config[T any] struct {
	data        T
	timestamp   string
	subscribers []chan bool
	activeFile  string
}

func Init[T any](numberOfSubs int) (*config[T], error) {
	c := &config[T]{}

	// use active config if exists
	activeFileExists := fileExists(activeConfig)
	defaultFileExists := fileExists(defaultConfig)

	if activeFileExists {
		c.activeFile = activeConfig
	} else if defaultFileExists {
		c.activeFile = defaultConfig
	} else {
		return nil, fmt.Errorf("no configuration files found")
	}

	// load config file
	err := c.load()
	if err != nil {
		return nil, fmt.Errorf("failed at load from file: %v", err)
	}
	c.updateTimestamp()

	// create subscribers
	for i := 0; i < numberOfSubs; i++ {
		c.subscribers = append(c.subscribers, make(chan bool, 1))
	}

	// create active config file if needed
	if !activeFileExists {
		c.activeFile = activeConfig
		err = c.persist()
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *config[T]) Get() *T {
	return &c.data
}

func (c *config[T]) Update(newConfig T) error {
	c.data = newConfig

	err := c.persist()
	if err != nil {
		return err
	}
	c.updateTimestamp()

	// notify subscribers
	for i := 0; i < len(c.subscribers); i++ {
		if len(c.subscribers[i]) != 0 {
			continue
		}
		c.subscribers[i] <- true
	}
	return nil
}

func (c *config[T]) GetTimestamp() string {
	return c.timestamp
}

func (c *config[T]) GetSubscriber(i int) *chan bool {
	if i < len(c.subscribers) {
		return &c.subscribers[i]
	}
	return nil
}

func (c *config[T]) updateTimestamp() {
	c.timestamp = strconv.FormatInt(time.Now().Unix(), 10)
}

func (c *config[T]) load() error {
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

func (c *config[T]) persist() error {
	file, err := json.MarshalIndent(c.data, "", "	")
	if err != nil {
		return fmt.Errorf("failed at marshal json: %v", err)
	}

	err = ioutil.WriteFile(c.activeFile, file, 0644)
	if err != nil {
		return fmt.Errorf("failed at write to file: %v", err)
	}

	return nil
}

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}

	return false
}
