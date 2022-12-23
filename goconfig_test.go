package goconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type TestConfig struct {
	Name    string
	Version int `validate:"required"`
}

var testData = TestConfig{Name: "config_test", Version: 123}

const CONFIG_NAME = "test"
const TEST_DIR = "testDir/"

const testString = "{\"name\":\"config_test\",\"version\":123}"
const testStringWithoutVersion = "{\"name\":\"config_test\"}"

func setUp(file string, path string, data string, subscribers []string) (*config[TestConfig], error) {
	if path != "" {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	configName := filepath.Join(path, fmt.Sprintf(file, CONFIG_NAME))
	err := os.WriteFile(configName, []byte(data), RW_RW_R_PERMISSION)

	if err != nil {
		return nil, err
	}

	c, err := Init[TestConfig](WithName(CONFIG_NAME), WithPath(path))
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
	os.RemoveAll(TEST_DIR)
}

func Test_Init(t *testing.T) {
	t.Run("No configuration files", func(t *testing.T) {
		_, err := Init[TestConfig](WithName("not_exist"))
		if err == nil {
			t.Errorf("Error is not returned")
		}
	})

	t.Run("Check loaded config data", func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(DEFAULT_CONFIG, "", testString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		want := testData
		got := *c.GetCfg()

		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Check loaded config data from active config", func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(ACTIVE_CONFIG, "", testString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		want := testData
		got := *c.GetCfg()
		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Create active config file", func(t *testing.T) {
		t.Cleanup(cleanUp)

		_, err := setUp(ACTIVE_CONFIG, "", testString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		if !fileExists(fmt.Sprintf(ACTIVE_CONFIG, CONFIG_NAME)) {
			t.Error("Expected active config file to be created, but it does not exist")
		}
		os.Remove(fmt.Sprintf(ACTIVE_CONFIG, CONFIG_NAME))
	})

	t.Run("Check active config file content", func(t *testing.T) {
		t.Cleanup(cleanUp)

		_, err := setUp(DEFAULT_CONFIG, "", testString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

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
		t.Cleanup(cleanUp)

		c, err := setUp(DEFAULT_CONFIG, "", testString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		if c.GetTimestamp() == "" {
			t.Error("Timestamp is not set")
		}
	})

	t.Run("Check subscribers being created", func(t *testing.T) {
		t.Cleanup(cleanUp)

		subscribers := [5]string{"test1", "test2", "test3", "test4", "test5"}
		c, err := setUp(DEFAULT_CONFIG, "", testString, subscribers[:])
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		if len(c.subscribers) != len(subscribers) {
			t.Error("Expected number of subscribers is not correct")
		}
	})

	t.Run("Check subscribers not being notified", func(t *testing.T) {
		t.Cleanup(cleanUp)

		subscribers := [5]string{"test1"}
		c, err := setUp(DEFAULT_CONFIG, "", testString, subscribers[:])
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}

		if len(c.GetSubscriber("test1")) != 0 {
			t.Error("Subscribers has been notified")
		}
	})

	t.Run("Custom config path", func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(DEFAULT_CONFIG, TEST_DIR, testString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		defaultConfigPth := filepath.Join(TEST_DIR, fmt.Sprintf(DEFAULT_CONFIG, CONFIG_NAME))
		if _, err := os.Stat(defaultConfigPth); err != nil {
			t.Error("Cannot find default config in expected location")
			t.FailNow()
		}

		activeConfigPth := filepath.Join(TEST_DIR, fmt.Sprintf(ACTIVE_CONFIG, CONFIG_NAME))
		if _, err := os.Stat(activeConfigPth); err != nil {
			t.Error("Cannot find active config in expected location")
			t.FailNow()
		}

		want := testData
		got := *c.GetCfg()

		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
			t.FailNow()
		}
	})

	t.Run("Check required fields validation", func(t *testing.T) {
		t.Cleanup(cleanUp)

		_, err := setUp(DEFAULT_CONFIG, "", testStringWithoutVersion, []string{})
		if err == nil {
			t.Errorf("Error is not returned")
			t.FailNow()
		}
		if !strings.Contains(err.Error(), "failed at validate config") {
			t.Errorf("Validation error is not returned")
		}
	})
}

func Test_Update(t *testing.T) {
	newData := TestConfig{Name: "new_data", Version: 456}

	t.Run("Check if config is updated", func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(DEFAULT_CONFIG, "", testString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		c.Update(newData)

		want := newData
		got := *c.GetCfg()
		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Check if subscribers are being notified", func(t *testing.T) {
		t.Cleanup(cleanUp)

		subscribers := [5]string{"test1", "test2", "test3"}
		c, err := setUp(DEFAULT_CONFIG, "", testString, subscribers[:])
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		c.Update(newData)

		if len(c.subscribers["test1"]) != 1 || len(c.subscribers["test2"]) != 1 || len(c.subscribers["test3"]) != 1 {
			t.Error("Subscribers not being notified")
		}
	})

	t.Run("Check if channels not being overloaded", func(t *testing.T) {
		t.Cleanup(cleanUp)

		subscribers := [1]string{"test1"}
		c, err := setUp(DEFAULT_CONFIG, "", testString, subscribers[:])
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		c.Update(newData)
		c.Update(newData)
		c.Update(newData)

		if len(c.subscribers["test1"]) != 1 {
			t.Error("Subscribers not being notified")
		}
	})
}
