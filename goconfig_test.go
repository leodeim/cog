package goconfig

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

type TestConfig struct {
	Name    string
	Version int
}

var testData = TestConfig{"goconfig_test", 123}

const testString = "{\"name\":\"goconfig_test\",\"version\":123}"

func setUp(file string, data string, subs int) (*config[TestConfig], error) {
	err := ioutil.WriteFile(file, []byte(data), 0644)
	if err != nil {
		return nil, err
	}

	c, err := Init[TestConfig](subs)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func cleanUp() {
	os.Remove(defaultConfig)
	os.Remove(activeConfig)
}

func Test_Init(t *testing.T) {
	t.Run("No configuration files", func(t *testing.T) {
		_, err := Init[TestConfig](0)
		if err == nil {
			t.Errorf("Error is not returned unexpectedly")
		}
	})

	t.Run("Check loaded config data", func(t *testing.T) {
		c, err := setUp(defaultConfig, testString, 0)
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		want := testData
		got := *c.Get()
		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Check loaded config data from active config", func(t *testing.T) {
		c, err := setUp(activeConfig, testString, 0)
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		want := testData
		got := *c.Get()
		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Create active config file", func(t *testing.T) {
		_, err := setUp(defaultConfig, testString, 0)
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		if !fileExists(activeConfig) {
			t.Error("Expected active config file to be created, but it does not exist")
		}
		os.Remove(activeConfig)
	})

	t.Run("Check active config file content", func(t *testing.T) {
		_, err := setUp(defaultConfig, testString, 0)
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		fileContent := TestConfig{}
		configFile, err := os.Open(activeConfig)
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
		c, err := setUp(defaultConfig, testString, 0)
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
		const NUM_OF_SUBS = 5
		c, err := setUp(defaultConfig, testString, NUM_OF_SUBS)
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		if len(c.subscribers) != NUM_OF_SUBS {
			t.Error("Expected number of subscribers is not correct")
		}
	})

	t.Run("Check subscribers not being notified", func(t *testing.T) {
		const NUM_OF_SUBS = 1
		c, err := setUp(defaultConfig, testString, NUM_OF_SUBS)
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		if len(*c.GetSubscriber(0)) != 0 {
			t.Error("Subscribers has been notified")
		}
	})
}

func Test_Update(t *testing.T) {
	newData := TestConfig{"new_data", 456}

	t.Run("Check if config is updated", func(t *testing.T) {
		c, err := setUp(defaultConfig, testString, 0)
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		c.Update(newData)

		want := newData
		got := *c.Get()
		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Check if subscribers are being notified", func(t *testing.T) {
		c, err := setUp(defaultConfig, testString, 3)
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		c.Update(newData)

		if len(c.subscribers[0]) != 1 || len(c.subscribers[1]) != 1 || len(c.subscribers[2]) != 1 {
			t.Error("Subscribers not being notified")
		}
	})

	t.Run("Check if channels not being overloaded", func(t *testing.T) {
		c, err := setUp(defaultConfig, testString, 1)
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}
		defer cleanUp()

		c.Update(newData)
		c.Update(newData)
		c.Update(newData)

		if len(c.subscribers[0]) != 1 {
			t.Error("Subscribers not being notified")
		}
	})
}
