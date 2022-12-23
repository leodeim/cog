package goconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

type TestConfig struct {
	Name    string
	Version int
}

var testData = TestConfig{"config_test", 123}

const CONFIG_NAME = "test"

const testString = "{\"name\":\"config_test\",\"version\":123}"

func setUp(file string, data string, subscribers []string) (*config[TestConfig], error) {
	err := os.WriteFile(fmt.Sprintf(file, CONFIG_NAME), []byte(data), RW_RW_R_PERMISSION)

	if err != nil {
		return nil, err
	}

	c, err := Init[TestConfig](WithName(CONFIG_NAME))
	if err != nil {
		return nil, err
	}

	for _, subscriber := range subscribers {
		c.AddSubscriber(subscriber)
	}

	return c, nil
}

func cleanUp() {
	os.Remove(fmt.Sprintf(DEFAULT_CONFIG, CONFIG_NAME))
	os.Remove(fmt.Sprintf(ACTIVE_CONFIG, CONFIG_NAME))
}

func Test_Init(t *testing.T) {
	t.Run("No configuration files", func(t *testing.T) {
		_, err := Init[TestConfig](WithName("not_exist"))
		if err == nil {
			t.Errorf("Error is not returned unexpectedly")
		}
	})

	t.Run("Check loaded config data", func(t *testing.T) {
		c, err := setUp(ACTIVE_CONFIG, testString, []string{})

		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		want := testData
		got := *c.GetCfg()

		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Check loaded config data from active config", func(t *testing.T) {
		c, err := setUp(ACTIVE_CONFIG, testString, []string{})
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		want := testData
		got := *c.GetCfg()
		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Create active config file", func(t *testing.T) {
		_, err := setUp(ACTIVE_CONFIG, testString, []string{})
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		if !fileExists(fmt.Sprintf(ACTIVE_CONFIG, CONFIG_NAME)) {
			t.Error("Expected active config file to be created, but it does not exist")
		}
		os.Remove(fmt.Sprintf(ACTIVE_CONFIG, CONFIG_NAME))
	})

	t.Run("Check active config file content", func(t *testing.T) {
		_, err := setUp(DEFAULT_CONFIG, testString, []string{})
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		fileContent := TestConfig{}
		configFile, err := os.Open(fmt.Sprintf(ACTIVE_CONFIG, CONFIG_NAME))
		if err != nil {
			t.Error("Opening activeConfig file", err.Error())
		}

		jsonParser := json.NewDecoder(configFile)
		if err = jsonParser.Decode(&fileContent); err != nil {
			t.Error("Parsing activeConfig file", err.Error())
		}

		want := testData
		got := fileContent

		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Check timestamp is created", func(t *testing.T) {
		c, err := setUp(DEFAULT_CONFIG, testString, []string{})
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		if c.GetTimestamp() == "" {
			t.Error("Timestamp is not set")
		}
	})

	t.Run("Check subscribers being created", func(t *testing.T) {
		subscribers := [5]string{"test1", "test2", "test3", "test4", "test5"}

		c, err := setUp(DEFAULT_CONFIG, testString, subscribers[:])

		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}

		defer cleanUp()

		if len(c.subscribers) != len(subscribers) {
			t.Error("Expected number of subscribers is not correct")
		}
	})

	t.Run("Check subscribers not being notified", func(t *testing.T) {
		subscribers := [5]string{"test1"}
		c, err := setUp(DEFAULT_CONFIG, testString, subscribers[:])
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		if len(c.GetSubscriber("test1")) != 0 {
			t.Error("Subscribers has been notified")
		}
	})
}

func Test_Update(t *testing.T) {
	newData := TestConfig{"new_data", 456}

	t.Run("Check if config is updated", func(t *testing.T) {
		c, err := setUp(DEFAULT_CONFIG, testString, []string{})
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		c.Update(newData)

		want := newData
		got := *c.GetCfg()
		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Check if subscribers are being notified", func(t *testing.T) {
		subscribers := [5]string{"test1", "test2", "test3"}

		c, err := setUp(DEFAULT_CONFIG, testString, subscribers[:])

		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}

		defer cleanUp()

		c.Update(newData)

		if len(c.subscribers["test1"]) != 1 || len(c.subscribers["test2"]) != 1 || len(c.subscribers["test3"]) != 1 {
			t.Error("Subscribers not being notified")
		}
	})

	t.Run("Check if channels not being overloaded", func(t *testing.T) {
		subscribers := [1]string{"test1"}
		c, err := setUp(DEFAULT_CONFIG, testString, subscribers[:])

		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}

		defer cleanUp()

		c.Update(newData)
		c.Update(newData)
		c.Update(newData)

		if len(c.subscribers["test1"]) != 1 {
			t.Error("Subscribers not being notified")
		}
	})
}
